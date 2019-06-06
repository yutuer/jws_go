package cmds

import (
	"log"

	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"

	"github.com/codegangsta/cli"
)

//放在这里试图让log配置尽早生效
func init() {
	logxml := config.NewConfigPath("log.xml")
	logs.LoadLogConfig(logxml)
}

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

func Register(c *cli.Command) {
	if c == nil {
		return
	}

	name := c.Name
	if _, ok := commands[name]; ok {
		log.Fatalln("cmds: Register called twice for adapter " + name)
	}
	commands[name] = c
}
