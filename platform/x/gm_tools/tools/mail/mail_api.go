package mail

import (
	"strconv"

	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
)

func delMail(c *gm_command.Context, server, accountid string, params []string) error {
	id_to_del, err := strconv.ParseInt(params[0], 10, 64)
	if err != nil {
		return err
	}

	err = DelMail(server, accountid, id_to_del)
	if err != nil {
		return err
	} else {
		c.SetData("{\"res\":\"ok\"}")
		return nil
	}
}
