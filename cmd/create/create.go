package create

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/no-mole/neptune/utils"
)

//go:embed template
var tpls embed.FS

// Warp cmd.run().
//
// Because some commands print the errors to stdout. Other commands might print to stderr but return an error code of 0。
// See https://stackoverflow.com/questions/18159704/how-to-debug-exit-status-1-error-when-running-exec-command-in-golang

func CmdRun(cmd *exec.Cmd) (output string, err error) {
	fmt.Printf("execute %s\n", cmd.String())

	var out bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()
	if err != nil {
		output = fmt.Sprint(err) + ":" + out.String()
		return
	}

	return out.String(), nil
}

func Run(args []string) error {
	if moduleName == defaultModuleName && len(args) > 0 {
		moduleName = args[0]
	}
	modName := strings.Trim(strings.TrimSpace(moduleName), "/")
	if len(modName) == 0 {
		return errors.New("command [new] must have a mod name")
	}

	curDir := utils.GetWorkdir()
	baseDir := path.Join(curDir, modName) //实际上的目录
	fmt.Printf("curDir=%s\nbaseDir=%s\n", curDir, baseDir)

	err := os.MkdirAll(baseDir, os.ModePerm)
	if err != nil {
		return err
	}

	stack := []string{"template"}
	data := map[string]interface{}{
		"ModName": modName,
	}
	modelTemplate := map[string][]byte{}
	for len(stack) > 0 {
		dirPath := stack[0]
		stack = stack[1:] //pop one
		println("create dir:", path.Join(baseDir, strings.Trim(strings.TrimPrefix(dirPath, "template"), "/")))
		err = os.MkdirAll(path.Join(baseDir, strings.Trim(strings.TrimPrefix(dirPath, "template"), "/")), os.ModePerm)
		if err != nil {
			return err
		}

		dirInfo, err := tpls.ReadDir(dirPath)
		if err != nil {
			return err
		}
		for _, f := range dirInfo {
			if f.IsDir() {
				stack = append(stack, path.Join(dirPath, f.Name())) //dir path需要保持层级关机
				continue
			}

			filePath := path.Join(dirPath, f.Name())
			fileBody, err := tpls.ReadFile(filePath) //read file
			if err != nil {
				return err
			}
			if strings.Contains(filePath, "model") { //model 单独处理
				modelTemplate[f.Name()] = fileBody
				continue
			}
			writeFileName := path.Join(baseDir, strings.Trim(strings.TrimSuffix(strings.TrimPrefix(filePath, "template"), "template"), "/")) //去掉前后的template
			buf := bytes.NewBufferString("")
			if strings.HasSuffix(filePath, ".gotemplate") { //是template 文件
				tpl, err := template.New(f.Name()).Parse(string(fileBody))
				if err != nil {
					return err
				}
				err = tpl.Execute(buf, data)
				if err != nil {
					return err
				}
			} else {
				buf.Write(fileBody)
			}
			println("create file:", writeFileName)
			err = os.WriteFile(writeFileName, buf.Bytes(), os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	err = initModelFile(baseDir, modelTemplate, modName)
	modInit := exec.Command("sh", "-c", fmt.Sprintf("cd %s && go mod init %s && go mod tidy && go mod vendor", baseDir, modName))
	output, err := CmdRun(modInit)

	fmt.Println(output)

	if err != nil {
		fmt.Printf("create new app[%s] failed", modName)
		return err
	}
	fmt.Printf("create new app [%s] success,don`t forgot edit [.env] for dev mode!", modName)
	return nil
}
