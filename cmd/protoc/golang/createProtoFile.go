package golang

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	inject_tags2 "github.com/no-mole/neptune/cmd/protoc/inject_tags"

	"github.com/no-mole/neptune/utils"
)

var metadataTempalte = `package {{.PackageName}}

import (
	"github.com/no-mole/neptune/registry"
)

func Metadata() *registry.Metadata {
	return &registry.Metadata{
		ServiceDesc: {{.ServiceDesc}},
		Namespace:   "neptune",
		Version:     "v1",
	}
}`

func InitGolangProto(targetFile string, paths []string) error {
	if len(targetFile) == 0 {
		return errors.New("No corresponding address found")
	}
	err := checkProtocGenGo()
	if err != nil {
		return err
	}
	err = checkProtocGenGoGrpc()
	if err != nil {
		return err
	}

	curDir := utils.GetWorkdir()
	return initProtoFiles(curDir, targetFile, paths)
}

func checkProtocGenGo() error {
	checkProtoc := exec.Command("sh", "-c", "protoc-gen-go --version")
	errReader, _ := checkProtoc.StderrPipe()
	outReader, _ := checkProtoc.StdoutPipe()

	err := checkProtoc.Start()
	if err != nil {
		return err
	}

	stderr := bytes.NewBuffer(nil)
	errIn := bufio.NewScanner(errReader)
	errFlag := false

	for errIn.Scan() {
		if errIn.Text() != "" {
			errFlag = true
			stderr.WriteString(errIn.Text())
			stderr.WriteString("\n")
		}
	}

	stdout := bytes.NewBuffer(nil)
	outIn := bufio.NewScanner(outReader)
	for outIn.Scan() {
		if outIn.Text() != "" {
			stdout.WriteString(outIn.Text())
			stdout.WriteString("\n")
		}
	}

	err = checkProtoc.Wait()
	if err != nil {
		return err
	}

	if errFlag {
		return errors.New(stderr.String())
	}

	if !strings.Contains(stdout.String(), "protoc-gen-go v1.28.0") {
		return errors.New("version mismatch please run: go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0")
	}
	return nil
}

func checkProtocGenGoGrpc() error {
	checkProtoc := exec.Command("sh", "-c", "protoc-gen-go-grpc --version")
	errReader, _ := checkProtoc.StderrPipe()
	outReader, _ := checkProtoc.StdoutPipe()

	err := checkProtoc.Start()
	if err != nil {
		return err
	}

	stderr := bytes.NewBuffer(nil)
	errIn := bufio.NewScanner(errReader)
	errFlag := false

	for errIn.Scan() {
		if errIn.Text() != "" {
			errFlag = true
			stderr.WriteString(errIn.Text())
			stderr.WriteString("\n")
		}
	}

	stdout := bytes.NewBuffer(nil)
	outIn := bufio.NewScanner(outReader)
	for outIn.Scan() {
		if outIn.Text() != "" {
			stdout.WriteString(outIn.Text())
			stdout.WriteString("\n")
		}
	}

	err = checkProtoc.Wait()
	if err != nil {
		return err
	}

	if errFlag {
		return errors.New(stderr.String())
	}

	if !strings.Contains(stdout.String(), "protoc-gen-go-grpc 1.2.0") {
		return errors.New("version mismatch please run: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0")
	}
	return nil

}

func initProtoFiles(curDir, targetFile string, includePaths []string) error {

	filePath := fmt.Sprintf("%s/%s", curDir, targetFile)

	paths := strings.Split(path.Base(filePath), ".")
	if len(paths) < 2 {
		return errors.New("not match file .proto")
	}
	if paths[1] != "proto" {
		return errors.New("not match file .proto")
	}

	fileName := path.Base(filePath)
	upDir := path.Dir(filePath)
	if len(targetFile) > 0 && targetFile[0] == '/' {
		targetFile = targetFile[1:]
	}
	cmdStr := fmt.Sprintf("protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative   --proto_path=.  --proto_path=../ ")

	for _, v := range includePaths {
		cmdStr += fmt.Sprintf(" --proto_path=%s ", v)
	}
	cmdStr += targetFile
	println(cmdStr)
	if path.Base(upDir)+".proto" != fileName {
		println("Warning: Package name and file name are different")
	}

	protoInit := exec.Command("sh", "-c", cmdStr)
	errReader, _ := protoInit.StderrPipe()

	err := protoInit.Start()
	if err != nil {
		return err
	}

	stderr := bytes.NewBuffer(nil)
	errFlag := false

	in := bufio.NewScanner(errReader)
	for in.Scan() {
		if in.Text() != "" {
			errFlag = true
			stderr.WriteString(in.Text())
			stderr.WriteString("\n")
		}
	}

	_ = protoInit.Wait()

	if errFlag {
		return errors.New(stderr.String())
	}

	gprcpbPath := strings.Replace(filePath, ".proto", "_grpc.pb.go", -1)
	pbFiles := strings.Replace(filePath, ".proto", ".pb.go", -1)
	if strings.Contains(fileName, "*") {
		files, err := os.ReadDir(upDir)
		if err != nil {
			return err
		}
		for _, file := range files {
			if strings.Contains(file.Name(), "_grpc.pb.go") {
				gprcpbPath = fmt.Sprintf("%s/%s", path.Dir(gprcpbPath), file.Name())
			} else if strings.Contains(file.Name(), "_grpc.pb.go") {
				pbFiles = fmt.Sprintf("%s/%s", path.Dir(gprcpbPath), file.Name())
			}
		}
	}

	packageName, serviceDesc, err := ProtoServiceDesc(gprcpbPath)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"PackageName": packageName,
		"ServiceDesc": serviceDesc,
	}

	//生成proto的文件夹
	//protoDir :=path.Join(upDir,"proto")
	buf := bytes.NewBufferString("")
	tpl, err := template.New("metadata").Parse(metadataTempalte)
	if err != nil {
		return err
	}
	err = tpl.Execute(buf, data)
	if err != nil {
		return err
	}

	err = os.WriteFile(upDir+"/metadata.go", buf.Bytes(), os.ModePerm)
	if err != nil {
		return err
	}
	err = inject_tags2.ParseAndWrite(pbFiles, nil, false)
	if err != nil {
		return err
	}
	return nil
}

func ProtoServiceDesc(filePath string) (packageName, serviceDesc string, err error) {
	fset := token.NewFileSet()
	fileAst, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	packageName = fileAst.Name.Name
	for _, decl := range fileAst.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.VAR {
			continue

		}
		if len(genDecl.Specs) == 0 {
			continue
		}
		valueSpec, ok := genDecl.Specs[0].(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range valueSpec.Names {
			if strings.HasSuffix(name.Name, "_ServiceDesc") {
				serviceDesc = name.Name
				return
			}
		}
	}
	err = errors.New("not Match ServiceDesc")
	return
}
