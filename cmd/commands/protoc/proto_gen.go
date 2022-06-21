package protoc

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"

	"github.com/no-mole/neptune/cmd/commands/protoc/php"
	"github.com/no-mole/neptune/cmd/commands/protoc/python"

	"github.com/no-mole/neptune/cmd/commands/protoc/java"

	"github.com/no-mole/neptune/cmd/commands"
	"github.com/no-mole/neptune/cmd/commands/protoc/golang"
)

func init() {
	commands.Registry("proto-gen", ProtocGenCommand)
}

var ProtocGenCommand = &commands.Command{
	Run:       NewProto,
	UsageLine: "proto-gen [$ProtoFilePath] OR proto-gen *.proto OR proto-gen -l java $ProtoFilePath",
	Helper: `

	-l [golang|python|java|php] //specify language,default is golang.

 	-path $ProtoFilePath //specify proto path,default is the last parameter.

	gen files tree:
	├── [$ProtoName]
	│ ├── $ProtoName.proto
	│ ├── $ProtoName.pb.go
	│ ├── $ProtoName_grpc.pb.go
	│ └── metadata.go
	
	for go language:
		using struct custom tag,for example
		
		message Bar {
		  // @cTags: binding:"foo_bar"
		  int64 id = 1; // @cTags: binding:"foo_bar"
		  string in = 2;
		  string out = 3;
		  string create_time = 4;
		}
`,
}

func NewProto(_ *commands.Command, args []string) {
	if len(args) == 0 {
		println("command [proto-gen] must have a proto file,use [neptune help proto-gen] for more tips!")
		return
	}
	if !checkProtoc() {
		return
	}

	language := "golang"
	path := ""
	for i, v := range args {
		if v == "-l" && i+1 < len(args) {
			language = args[i+1]
		}
		if v == "-path" && i+1 < len(args) {
			path = args[i+1]
		}
	}
	if path == "" && len(args)%2 == 1 {
		path = args[len(args)-1]
	}

	switch language {
	case "golang":
		golang.InitGolangProto(path)
	case "java":
		java.InitJavaFile(path)
	case "python":
		python.InitPythonFile(path)
	case "php":
		php.InitPhpFile(path)
	}

}

func checkProtoc() bool {
	checkProtoc := exec.Command("sh", "-c", "protoc --version")
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

	if !strings.Contains(stdout.String(), "libprotoc 3.20.1") {
		println("please use  3.20.1")
		return false
	}
	return true
}
