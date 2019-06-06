package restore

import (
	"vcs.taiyouxi.net/platform/x/redis_storage/command"
)

// 注册命令

func restore_one(ud interface{}, args []string) (result string, err error) {
	restore := ud.(RestoreWorker)

	if len(args) > 0 {
		err = restore.RestoreOne(args[0], nil)
	}
	return
}

func restore_all(ud interface{}, args []string) (result string, err error) {
	restore := ud.(RestoreWorker)
	err = restore.RestoreAll()
	return "Over", nil
}

func (o *Restore) Register(c *command.CmdService) {
	c.Register("restore", o, restore_one)
	c.Register("restore_all", o, restore_all)
}
