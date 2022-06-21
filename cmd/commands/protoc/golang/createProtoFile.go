package golang

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/no-mole/neptune/cmd/commands/protoc/inject_tags"
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

func InitGolangProto(args string) {
	if len(args) == 0 {
		println("No corresponding address found")
		return
	}

	if !checkProtocGenGo() {
		return
	}
	if !checkProtocGenGoGrpc() {
		return
	}

	curDir := utils.GetWorkdir()
	filePath := fmt.Sprintf("%s/%s", curDir, args)

	initProtoFiles(filePath)
}

func checkProtocGenGo() bool {
	checkProtoc := exec.Command("sh", "-c", "protoc-gen-go --version")
	errReader, _ := checkProtoc.StderrPipe()
	outReader, _ := checkProtoc.StdoutPipe()

	err := checkProtoc.Start()
	if err != nil {
		println(err.Error())
		return false
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

	checkProtoc.Wait()

	if errFlag {
		println(stderr.String())
		return false
	}

	if !strings.Contains(stdout.String(), "protoc-gen-go v1.28.0") {
		println("version mismatch please run: go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0")
		return false
	}
	return true
}

func checkProtocGenGoGrpc() bool {
	checkProtoc := exec.Command("sh", "-c", "protoc-gen-go-grpc --version")
	errReader, _ := checkProtoc.StderrPipe()
	outReader, _ := checkProtoc.StdoutPipe()

	err := checkProtoc.Start()
	if err != nil {
		println(err.Error())
		return false
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

	checkProtoc.Wait()

	if errFlag {
		println(stderr.String())
		return false
	}

	if !strings.Contains(stdout.String(), "protoc-gen-go-grpc 1.2.0") {
		println("version mismatch please run: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0")
		return false
	}
	return true

}

func initProtoFiles(filePath string) {
	paths := strings.Split(path.Base(filePath), ".")
	if len(paths) == 0 {
		println("not match file .proto")
		return
	}
	if paths[1] != "proto" {
		println("not match file .proto")
		return
	}

	fileName := path.Base(filePath)
	upDir := path.Dir(filePath)
	cmdStr := fmt.Sprintf("cd %s && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative   --proto_path=.  --proto_path=../   %s", upDir, fileName)

	if path.Base(upDir)+".proto" != fileName {
		println("Warning: Package name and file name are different")
	}

	protoInit := exec.Command("sh", "-c", cmdStr)
	errReader, _ := protoInit.StderrPipe()

	err := protoInit.Start()
	if err != nil {
		println(err.Error())
		return
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

	err = protoInit.Wait()
	//if err != nil {
	//	println(err.Error())
	//	return
	//}

	if errFlag {
		println(stderr.String())
		return
	}

	gprcpbPath := strings.Replace(filePath, ".proto", "_grpc.pb.go", -1)
	pbFiles := strings.Replace(filePath, ".proto", ".pb.go", -1)
	if strings.Contains(fileName, "*") {
		files, err := ioutil.ReadDir(upDir)
		if err != nil {
			println(err.Error())
			return
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
		println(err.Error())
		return
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
		println(err.Error())
		return
	}
	err = tpl.Execute(buf, data)
	if err != nil {
		println(err.Error())
		return
	}

	err = os.WriteFile(upDir+"/metadata.go", buf.Bytes(), os.ModePerm)
	if err != nil {
		println(err.Error())
		return
	}
	err = inject_tags.ParseAndWrite(pbFiles, nil, false)
	if err != nil {
		println(err.Error())
		return
	}
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
