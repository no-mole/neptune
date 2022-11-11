package php

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

func InitPhpFile(args string) error {
	if len(args) == 0 {
		return errors.New("no corresponding address found")
	}

	curDir := utils.GetWorkdir()
	filePath := fmt.Sprintf("%s/%s", curDir, args)

	return initProtoFiles(filePath)
}

func initProtoFiles(filePath string) error {
	paths := strings.Split(path.Base(filePath), ".")
	if len(paths) < 2 {
		return errors.New("not match file .proto")
	}
	if paths[1] != "proto" {
		return errors.New("not match file .proto")
	}

	fileName := path.Base(filePath)
	upDir := path.Dir(filePath)
	cmdStr := fmt.Sprintf("cd %s && protoc --php_out=. ./%s", upDir, fileName)
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

	err = protoInit.Wait()
	if err != nil {
		return err
	}

	if errFlag {
		return errors.New(stderr.String())
	}
	return nil
}
