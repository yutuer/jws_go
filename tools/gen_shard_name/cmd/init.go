package cmd

import (
	"github.com/codegangsta/cli"
	"log"
)

var (
	commands = make(map[string]*cli.Command)
)

// GetCommands
func InitCommands(cmds *[]cli.Command) {
	*cmds = make([]cli.Command, 0, len(commands))
	for _, v := range commands {
		*cmds = append(*cmds, *v)
	}
}

func register(c *cli.Command) {
	if c == nil {
		return
	}

	name := c.Name
	if _, ok := commands[name]; ok {
		log.Fatalln("tools: Register called twice for adapter " + name)
	}
	commands[name] = c
}
