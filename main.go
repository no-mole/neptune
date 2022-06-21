package main

import (
	"os"

	_ "github.com/no-mole/neptune/cmd"
	"github.com/no-mole/neptune/cmd/commands"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"help"}
	}
	commands.Call(args[0], args[1:])
}
