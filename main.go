package main

import (
	"os"

	"github.com/no-mole/neptune/cmd/create"
	"github.com/no-mole/neptune/cmd/protoc"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "neptune",
	Short:   "Neptune is a Fast and Practical tool for Managing your Application.",
	Version: "v0.2.2",
}

func main() {
	rootCmd.AddCommand(create.Command)
	rootCmd.AddCommand(protoc.Command)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
