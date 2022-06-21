package commands

import (
	"bytes"
	"text/template"
)

func Help(cmdName string) {
	if cmdName == "" {
		t := template.New("Usage")
		t, err := t.Parse(usageTemplate)
		if err != nil {
			panic(err)
		}
		buf := bytes.NewBufferString("")
		err = t.Execute(buf, AvailableCommands)
		if err != nil {
			panic(err)
		}
		println(buf.String())
		return
	}
	if cmd, ok := AvailableCommands[cmdName]; ok {
		println(cmd.UsageLine)
		println(cmd.Helper)
		return
	}
	println("Unsupported features:" + cmdName)
}

var usageTemplate = `
Neptune is a Fast and Practical tool for Managing your Application.

Usage: neptune [command] [args]

AVAILABLE COMMANDS:
{{range .}}
   neptune {{.UsageLine}}{{end}}

Use neptune help [command] for more information about a command.
`
