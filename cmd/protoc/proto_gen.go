package protoc

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	golang2 "github.com/no-mole/neptune/cmd/protoc/golang"
	java2 "github.com/no-mole/neptune/cmd/protoc/java"
	php2 "github.com/no-mole/neptune/cmd/protoc/php"
	python2 "github.com/no-mole/neptune/cmd/protoc/python"

	"github.com/spf13/cobra"
)

func Run(_ *cobra.Command, args []string) error {
	if protoFilePath == "" {
		if len(args) == 0 {
			return errors.New("command [proto-gen] must have a proto file")
		}
		protoFilePath = args[0]
	}
	fmt.Printf("%v", args)
	println(language)

	err := checkProtoc()
	if err != nil {
		return err
	}
	switch language {
	case "golang":
		err = golang2.InitGolangProto(protoFilePath)
	case "java":
		err = java2.InitJavaFile(protoFilePath)
	case "python":
		err = python2.InitPythonFile(protoFilePath)
	case "php":
		err = php2.InitPhpFile(protoFilePath)
	}
	return err
}

func checkProtoc() error {
	checkCmd := exec.Command("sh", "-c", "protoc --version")
	errReader, _ := checkCmd.StderrPipe()
	outReader, _ := checkCmd.StdoutPipe()

	err := checkCmd.Start()
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

	err = checkCmd.Wait()
	if err != nil {
		return err
	}

	if errFlag {
		return errors.New(stderr.String())
	}

	if !strings.Contains(stdout.String(), "libprotoc 3.20.1") {
		return errors.New("please use protoc 3.20.1")
	}
	return nil
}
