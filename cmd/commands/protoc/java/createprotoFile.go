package java

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/no-mole/neptune/utils"
)

func InitJavaFile(args string) {
	if len(args) == 0 {
		println("No corresponding address found")
		return
	}

	curDir := utils.GetWorkdir()
	filePath := fmt.Sprintf("%s/%s", curDir, args)

	initProtoFiles(filePath)
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
	cmdStr := fmt.Sprintf("cd %s && protoc --java_out=. ./%s", upDir, fileName)
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

	protoInit.Wait()

	if errFlag {
		println(stderr.String())
		return
	}
}
