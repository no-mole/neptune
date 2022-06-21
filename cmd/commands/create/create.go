package create

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/no-mole/neptune/cmd/commands"
	"github.com/no-mole/neptune/utils"
)

func init() {
	commands.Registry("new", CreateCommand)
}

var CreateCommand = &commands.Command{
	Run:       New,
	UsageLine: "new [$modName]",
	Helper: `
	Creates Neptune application.

	├── bootstrap
	│ ├── database.go		// init database,support drivers:mysql、clickhouse、postgreSql
	│ ├── grpc_server.go	// init grpc server,for discovery third party service
	│ ├── logger.go         // init logger
	│ ├── redis.go			//init redis connection pool
	│ ├── router.go         //init http router
	│ └── service.go        //registry grpc service
	├── config
	│ ├── app.yaml		    //app conf,appName、namespace、version、grpcPort、httpPort....
	│ ├── dev
	│ │ ├── config.yaml     //config center conf
	│ │ └── registry.yaml   //registry center conf
	│ ├── grey
	│ ├── prod
	│ └── test
	├── controller
	│ └── bar
	│     └── bar.go        //implementation gin handle (is optional)
	├── model
	│ ├── bar
	│ │ ├── log.go
	│ │ └── model.go       
	│ └── model.go          //model common variable
	├── service
	│ └── bar
	│     └── service.go    //implementation grpc interface
	├── .env.example        //move to .env,mark envMode & envDebug，default mode=prod,debug=false
	├── Dockerfile          //multi-stage construction
	├── Makefile            //make image 
	├── docker_build.sh     //or  sh ./docker_build.sh tag=v0.1
	├── main.go 
	└── go.mod
`,
}

//go:embed template template/.env.example template/.gitignore template/.gitlab-ci.yml
var tpls embed.FS

// Warp cmd.run().
//
// Because some commands print the errors to stdout. Other commands might print to stderr but return an error code of 0。
// See https://stackoverflow.com/questions/18159704/how-to-debug-exit-status-1-error-when-running-exec-command-in-golang

func CmdRun(cmd *exec.Cmd) (output string, err error) {
	fmt.Printf("execute %s", cmd.String())

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

func New(_ *commands.Command, args []string) {
	if len(args) == 0 {
		println("command [new] must have a mod name,use [neptune help new] for more tips!")
		return
	}
	modName := strings.Trim(strings.TrimSpace(args[0]), "/")
	if len(modName) == 0 {
		println("command [new] must have a mod name,use [neptune help new] for more tips!")
		return
	}

	curDir := utils.GetWorkdir()
	baseDir := path.Join(curDir, modName) //实际上的目录
	fmt.Printf("curDir=%s\nbaseDir=%s\n", curDir, baseDir)

	err := os.MkdirAll(baseDir, os.ModePerm)
	if err != nil {
		println(err.Error())
		return
	}

	stack := []string{"template"}
	data := map[string]interface{}{
		"ModName": modName,
	}

	for len(stack) > 0 {
		dirPath := stack[0]
		stack = stack[1:] //pop one
		println("create dir:", path.Join(baseDir, strings.Trim(strings.TrimPrefix(dirPath, "template"), "/")))
		err = os.MkdirAll(path.Join(baseDir, strings.Trim(strings.TrimPrefix(dirPath, "template"), "/")), os.ModePerm)
		if err != nil {
			println(err.Error())
			return
		}

		dirInfo, err := tpls.ReadDir(dirPath)
		if err != nil {
			panic(err)
		}
		for _, f := range dirInfo {
			if f.IsDir() {
				stack = append(stack, path.Join(dirPath, f.Name())) //dir path需要保持层级关机
				continue
			}

			filePath := path.Join(dirPath, f.Name())
			fileBody, err := tpls.ReadFile(filePath) //read file
			if err != nil {
				println(err.Error())
				return
			}

			writeFileName := path.Join(baseDir, strings.Trim(strings.TrimSuffix(strings.TrimPrefix(filePath, "template"), "template"), "/")) //去掉前后的template

			buf := bytes.NewBufferString("")
			if strings.HasSuffix(filePath, ".gotemplate") { //是template 文件
				tpl, err := template.New(f.Name()).Parse(string(fileBody))
				if err != nil {
					println(err.Error())
					return
				}
				err = tpl.Execute(buf, data)
				if err != nil {
					println(err.Error())
					return
				}
			} else {
				buf.Write(fileBody)
			}
			println("create file:", writeFileName)
			err = os.WriteFile(writeFileName, buf.Bytes(), os.ModePerm)
			if err != nil {
				println(err.Error())
				return
			}
		}
	}

	modInit := exec.Command("sh", "-c", fmt.Sprintf("cd %s && go mod init %s && go mod tidy && go mod vendor", baseDir, modName))
	output, err := CmdRun(modInit)

	fmt.Println(output)

	if err != nil {
		fmt.Printf("create new app[%s] failed, the error log is above.", modName)
	} else {
		fmt.Printf("create new app [%s] success,don`t forgot edit [.env] for dev mode!", modName)
	}
}
