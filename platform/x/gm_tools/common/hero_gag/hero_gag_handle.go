package hero_gag

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	gmConfig "vcs.taiyouxi.net/platform/x/gm_tools/config"
	"vcs.taiyouxi.net/platform/x/gm_tools/tools/profile_tool"
)

const (
	UnknowErr = "Unknow Error"
)

type GagAccountRet struct {
	IsOK    bool   `json:"is_ok"`
	Reason  string `json:"reason"`
	BanTime string `json:"ban_time"`
}

type GetAccountDataRet struct {
	GUID          string `json:"guid"`
	Nick          string `json:"account_name"`
	UserID        string `json:"userid"`
	Channel       string `json:"channel"`
	UID           string `json:"uid"`
	CorpLv        string `json:"corp"`
	VIPLv         string `json:"v"`
	CreateTime    string `json:"createtime"`
	LastLoginTime string `json:"logintime"`
	CorpName      string `json:"corp_name"`
	AllPay        string `json:"all_pay"`
	Gold          string `json:"sc"`
	Diamond       string `json:"hc"`
	GagTime       int64  `json:"gag_time"`
	GagReason     string `json:"gag_reason"`
	Name          string `json:"name"`
}

func handleGagAccount(c *gin.Context) {
	account_id := c.Param("account_id")
	server_name := c.Param("server_name")
	reason := c.Param("reason")
	banTime := c.Param("time")
	// tmp
	command := GagCommand

	logs.Info("command get %s %s %s %s ", server_name, account_id, command, reason)

	_, commandErr := gm_command.OnCommand(command, server_name, account_id, []string{banTime, "YingXiong", reason})

	// to client
	result := GagAccountRet{}
	var code int
	if commandErr != nil {
		result.IsOK = false
		result.Reason = commandErr.Error()
		code = 200

	} else {
		result.IsOK = true
		result.BanTime = banTime
		code = 200
	}
	str, err := json.Marshal(result)
	if err != nil {
		c.String(200, errRes(UnknowErr))
		return
	}

	c.String(code, string(str))

	// log elastic search
	info := CommandInfo{
		IsOK:        commandErr != nil,
		AccountID:   account_id,
		ServerName:  server_name,
		Reason:      reason,
		CommandName: command,
		Time:        getTime(),
	}
	logElasticInfo(info)
}

func handleGetAccountData(c *gin.Context) {
	result := GetAccountDataRet{}

	var ok bool = true
	command := GetAccountData
	account_id := c.Param("account_id")
	s := strings.Split(account_id, ":")
	if len(s) < 3 {
		logs.Error("account id format error")
		ok = false
		c.String(200, errRes("account id error"))
		return
	}
	server := strings.Join(s[0:2], ":")
	logs.Debug("The server is: %s", s)

	result.GUID = server

	profileData, err := profile_tool.GetProfileFromRedis(server, "profile:"+account_id)
	if err != nil {
		logs.Error("Command(GetProfileFromRedis) Error by %v", err)
		ok = false
		c.String(200, errRes("account id error"))
		return
	}
	err = json.Unmarshal(profileData, &result)
	if err != nil {
		logs.Error("Unmarshal profileData Error by %v", err)
		ok = false
		c.String(200, errRes(UnknowErr))
		return
	}

	authData, err := GetAuthData(s[2])
	if err != nil {
		logs.Error("Command(getAccountByName) Error by %v", err)
		ok = false
		c.String(200, errRes(UnknowErr))
		return
	}

	userData := struct {
		UserId    db.UserID `json:"user_id"`
		BanTime   int64     `json:"ban_time"`
		BanReason string    `json:"banreason"`
	}{}
	err = json.Unmarshal([]byte(authData), &userData)
	logs.Debug("UserData: %v", userData)
	if err != nil {
		logs.Error("Unmarshal(accountData) Error by %v", err)
		ok = false
		c.String(200, errRes(UnknowErr))
		return
	}

	result.UserID = userData.UserId.String()
	result.GagReason = userData.BanReason
	result.GagTime = userData.BanTime

	logs.Debug("AccountData: %v", result)

	str, err := json.Marshal(result)
	if err != nil {
		c.String(200, errRes(UnknowErr))
		return
	}
	c.String(200, string(str))

	// log elastic search
	info := CommandInfo{
		IsOK:        ok,
		AccountID:   account_id,
		ServerName:  "",
		Reason:      "",
		CommandName: command,
		Time:        getTime(),
	}
	logElasticInfo(info)
}

func handleGetAccountDataByNick(c *gin.Context) {
	result := GetAccountDataRet{}

	var ok bool = true
	command := GetAccountDataByNick
	server := c.Param("gid") + ":" + c.Param("sid")
	cfg := gmConfig.Cfg.GetServerCfgFromName(server)
	account_id, err := profile_tool.GetAcountIDFromRedis(cfg.RedisName, server, c.Param("nick"))
	s := strings.Split(account_id, ":")
	if len(s) < 3 {
		logs.Error("account id format error")
		ok = false
		c.String(200, errRes("account id error"))
		return
	}
	server = strings.Join(s[0:2], ":")
	logs.Debug("The server is: %s", s)

	result.GUID = server

	profileData, err := profile_tool.GetProfileFromRedis(server, "profile:"+account_id)
	if err != nil {
		logs.Error("Command(GetProfileFromRedis) Error by %v", err)
		ok = false
		c.String(200, errRes("account id error"))
		return
	}
	err = json.Unmarshal(profileData, &result)
	if err != nil {
		logs.Error("Unmarshal profileData Error by %v", err)
		ok = false
		c.String(200, errRes(UnknowErr))
		return
	}

	authData, err := GetAuthData(s[2])
	if err != nil {
		logs.Error("Command(getAccountByName) Error by %v", err)
	} else {
		userData := struct {
			UserId    db.UserID `json:"user_id"`
			BanTime   int64     `json:"ban_time"`
			BanReason string    `json:"banreason"`
		}{}
		err = json.Unmarshal([]byte(authData), &userData)
		if err != nil {
			logs.Error("Unmarshal(accountData) Error by %v", err)
		} else {
			logs.Debug("UserData: %v", userData)
			result.UserID = userData.UserId.String()
			result.GagReason = userData.BanReason
			result.GagTime = userData.BanTime
		}

	}

	logs.Debug("AccountData: %v", result)

	str, err := json.Marshal(result)
	if err != nil {
		c.String(200, errRes(UnknowErr))
		return
	}
	c.String(200, string(str))

	// log elastic search
	info := CommandInfo{
		IsOK:        ok,
		AccountID:   account_id,
		ServerName:  "",
		Reason:      "",
		CommandName: command,
		Time:        getTime(),
	}
	logElasticInfo(info)
}
