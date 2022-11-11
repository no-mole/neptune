package main

import (
	"os"

	create2 "github.com/no-mole/neptune/cmd/create"
	protoc2 "github.com/no-mole/neptune/cmd/protoc"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "neptune",
	Short:   "Neptune is a Fast and Practical tool for Managing your Application.",
	Version: "v0.1.0",
}

func main() {
	rootCmd.AddCommand(create2.Command)
	rootCmd.AddCommand(protoc2.Command)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
