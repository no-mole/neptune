package protoc

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/no-mole/neptune/cmd/protoc/golang"
	"github.com/no-mole/neptune/cmd/protoc/java"
	"github.com/no-mole/neptune/cmd/protoc/php"
	"github.com/no-mole/neptune/cmd/protoc/python"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func Run(_ *cobra.Command, args []string) error {

	if len(args) == 0 {
		return errors.New("command [proto-gen] must have a proto file")
	}
	targetFile := args[0]

	err := checkProtoc()
	if err != nil {
		return err
	}
	switch language {
	case "golang":
		err = golang.InitGolangProto(targetFile, protoFilePaths)
	case "java":
		err = java.InitJavaFile(targetFile, protoFilePaths)
	case "python":
		err = python.InitPythonFile(targetFile, protoFilePaths)
	case "php":
		err = php.InitPhpFile(targetFile, protoFilePaths)
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
