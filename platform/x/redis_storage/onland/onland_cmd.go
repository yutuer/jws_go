package onland

import (
	"vcs.taiyouxi.net/platform/x/redis_storage/command"
)

func help(ud interface{}, args []string) (result string, err error) {
	return "help TODO", nil
}

func sync_one(ud interface{}, args []string) (result string, err error) {
	onlander := ud.(*OnlandStores)

	key := ""
	if len(args) > 0 {
		key = args[0]
	}
	onlander.NewKeyDumpJob(key)
	return
}

func sync_all(ud interface{}, args []string) (result string, err error) {
	onlander := ud.(*OnlandStores)
	res, err := onlander.OnlandAll()
	return res, err
}

func (o *OnlandStores) Register(c *command.CmdService) {
	c.Register("help", o, help)
	c.Register("sync", o, sync_one)
	c.Register("sync_all", o, sync_all)
}
