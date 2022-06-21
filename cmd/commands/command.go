package commands

type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Command, args []string)

	// UsageLine is the one-line Usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Helper is the description shown in the 'go help' output.
	Helper string
}

func Registry(name string, command *Command) {
	AvailableCommands[name] = command
}

var AvailableCommands = map[string]*Command{}

func Call(name string, args []string) {
	if name == "help" {
		if len(args) == 0 {
			Help("")
		} else {
			Help(args[0])
		}
		return
	}
	if cmd, ok := AvailableCommands[name]; ok {
		cmd.Run(cmd, args)
	} else {
		println("Unsupported features:" + name)
	}
}
