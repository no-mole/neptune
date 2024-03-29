package protoc

import (
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "proto-gen",
	Short: "Generates a proto file in the specified language",
	Long: `
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
		}`,
	Example: "neptune proto-gen [$ProtoFilePath] OR proto-gen *.proto OR proto-gen -l java $ProtoFilePath",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := Run(cmd, args)
		if err != nil {
			println(err.Error())
		}
		return err
	},
}

var (
	language       string
	protoFilePaths []string
	namespace      string
	version        string
)

func init() {
	Command.Flags().StringVarP(&language, "language", "l", "golang", "specify language. [golang|python|java|php]")
	Command.Flags().StringSliceVarP(&protoFilePaths, "include", "i", []string{}, "specify protos path, default is the last parameter.")
	Command.Flags().StringVarP(&namespace, "namespace", "n", "neptune", "specify proto service namespace")
	Command.Flags().StringVarP(&version, "version", "v", "v1", "specify proto service version")

}
