package create

import (
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "new",
	Short: "new [$modName]",
	Long: `
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
	├── Dockerfile          //multi-stage construction
	├── main.go 
	└── go.mod
`,
	Example: "neptune new github.com/no-mole/neptune",
	RunE: func(cmd *cobra.Command, args []string) error {
		return Run(args)
	},
}

var moduleName string

func init() {
	Command.Flags().StringVarP(&moduleName, "module-name", "", defaultModuleName, "go module name,like github.com/no-mole/neptune")
}

const defaultModuleName = "myapp"
