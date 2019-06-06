package gag_account

import (
	"errors"
	"fmt"
	"strings"
	//"strconv"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	gmConfig "vcs.taiyouxi.net/platform/x/gm_tools/config"
	"vcs.taiyouxi.net/platform/x/gm_tools/util"
)

func RegCommands() {
	gm_command.AddGmCommandHandle("gagAccount", gagAccount)
	gm_command.AddGmCommandHandle("getAllGagAccount", getAllGagAccount)
}

func gagAccount(c *gm_command.Context, server, accountid string, params []string) error {
	cfg := gmConfig.Cfg.GetServerCfgFromName(server)
	logs.Info("gagAccount  %s %s --> %v", accountid, server, cfg)
	acids := strings.Split(accountid, ":")
	if len(acids) < 3 {
		return errors.New("acidErr")
	}
	gid := acids[0]
	uid := acids[2]

	url := fmt.Sprintf("%s%s?time=%s&gid=%s", cfg.AuthGagApi, uid, params[0], gid)

	logs.Info(url)
	_, err := util.HttpGet(url)
	return err
}

func getAllGagAccount(c *gm_command.Context, server, accountid string, params []string) error {
	return nil
}
