package sys_roll_notice

import (
	"errors"
	"strconv"
	"time"

	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	"vcs.taiyouxi.net/platform/x/gm_tools/config"
)

func RegCommands() {
	gm_command.AddGmCommandHandle("getSysRollNotice", getSysRollNotice)
	gm_command.AddGmCommandHandle("sendSysRollNotice", sendSysRollNotice)
	gm_command.AddGmCommandHandle("updateSysRollNotice", updateSysRollNotice)
	gm_command.AddGmCommandHandle("delSysRollNotice", delSysRollNotice)
	gm_command.AddGmCommandHandle("activitySysRollNotice", activitySysRollNotice)
}

func getSysRollNotice(c *gm_command.Context, server, accountid string, params []string) error {
	res := jobManager.GetJobByServer(server)
	c.SetData(string(res))
	return nil
}

func sendSysRollNotice(c *gm_command.Context, server, accountid string, params []string) error {
	tb_Time, err := time.ParseInLocation("2006/01/02 15:04:05", params[0], util.ServerTimeLocal)
	if err != nil {
		logs.Error("sendSysRollNotice err : %v", err)
		return err
	}
	te_Time, err := time.ParseInLocation("2006/01/02 15:04:05", params[1], util.ServerTimeLocal)
	if err != nil {
		logs.Error("tsendSysRollNotice err : %v", err)
		return err
	}
	tb := tb_Time.Unix()
	te := te_Time.Unix()

	in, err := strconv.ParseInt(params[2], 10, 64)
	if err != nil {
		return err
	}
	is_sended, err := strconv.ParseInt(params[3], 10, 64)
	if err != nil {
		return err
	}

	jobManager.AddJob(tb, te, in, int(is_sended), gm_command.CommandRecord{
		AccountId:  accountid,
		ServerName: server,
		Command:    "activitySysRollNotice",
		Params:     params,
		GM:         "admin",
	})
	return nil
}

func delSysRollNotice(c *gm_command.Context, server, accountid string, params []string) error {
	id, err := strconv.ParseInt(params[0], 10, 64)
	if err != nil {
		return err
	}
	jobManager.DelJob(id)
	return nil
}

func updateSysRollNotice(c *gm_command.Context, server, accountid string, params []string) error {
	tb_Time, err := time.ParseInLocation("2006/01/02 15:04:05", params[0], util.ServerTimeLocal)
	if err != nil {
		logs.Error("updateSysRollNotice err : %v", err)
		return err
	}
	te_Time, err := time.ParseInLocation("2006/01/02 15:04:05", params[1], util.ServerTimeLocal)
	if err != nil {
		logs.Error("updateSysRollNotice err : %v", err)
		return err
	}

	tb := tb_Time.Unix()
	te := te_Time.Unix()

	in, err := strconv.ParseInt(params[2], 10, 64)
	if err != nil {
		return err
	}
	is_sended, err := strconv.ParseInt(params[3], 10, 64)
	if err != nil {
		return err
	}
	id, err := strconv.ParseInt(params[6], 10, 64)
	if err != nil {
		return err
	}
	jobManager.UpdateJob(id, tb, te, in, int(is_sended), gm_command.CommandRecord{
		AccountId:  accountid,
		ServerName: server,
		Command:    "activitySysRollNotice",
		Params:     params,
		GM:         "admin",
	})
	return nil
}

//
// 触发一个跑马灯
//
func activitySysRollNotice(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("activitySysRollNotice %s %v", server, params)

	notice := NewSysRollNoticeToChat(params[4], server, params[5:])
	data, err := notice.ToData()
	if err != nil {
		logs.Error("activitySysRollNotice ToData Err by %s", err.Error())
		return err
	}

	logs.Trace("data : %s", string(data))

	cfg := config.Cfg.GetServerCfgFromName(server)
	if cfg.SysRollNoticeUrl == "" {
		return errors.New("NoCfg")
	}
	logs.Info("post %s", cfg.SysRollNoticeUrl)
	_, err = util.HttpPost(cfg.SysRollNoticeUrl, "application/json; charset=utf-8", data)
	if err != nil {
		logs.Error("activitySysRollNotice HttpPost Err by %s", err.Error())
		return err
	}
	return nil
}
