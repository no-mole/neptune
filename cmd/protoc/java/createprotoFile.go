package java

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/no-mole/neptune/utils"
)

func InitJavaFile(targetFile string, includePaths []string) error {
	if len(targetFile) == 0 {
		return errors.New("no corresponding address found")
	}

	curDir := utils.GetWorkdir()
	return initProtoFiles(curDir, targetFile, includePaths)
}

func initProtoFiles(curDir, targetFile string, includePaths []string) error {
	filePath := fmt.Sprintf("%s/%s", curDir, targetFile)

	paths := strings.Split(path.Base(filePath), ".")
	if len(paths) == 0 {
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

	cmdStr := fmt.Sprintf("protoc --java_out=. ./%s", fileName)
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
	return nil
}
