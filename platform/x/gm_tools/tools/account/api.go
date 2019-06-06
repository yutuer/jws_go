package account

import (
	//"github.com/gin-gonic/gin"
	//"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	//∂"vcs.taiyouxi.net/platform/x/gm_tools/common/store"
	"errors"

	"vcs.taiyouxi.net/platform/x/gm_tools/login"
	//"vcs.taiyouxi.net/platform/x/gm_tools/util"
	"encoding/json"
	"strconv"
	"strings"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func RegCommands() {
	gm_command.AddGmCommandHandle("addUser", AddUser)
	gm_command.AddGmCommandHandle("changePass", ChangePass)
	gm_command.AddGmCommandHandle("getUser", GetUser)
	gm_command.AddGmCommandHandle("setGrant", SetGrant)
	gm_command.AddGmCommandHandle("getGrant", GetGrant)
}

//创建账号
func AddUser(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Trace("AddUser得到的参数为:%v", params)
	if len(params)!=4{
		return errors.New("错误的参数")
	}
	if _, ok := login.AccountManager.Accounts[params[0]]; !ok {
		if params[0]==""||params[2]==""{
			return errors.New("账户名或密码为空")
		}
		login.AccountManager.Add(login.NewAccount(params[0], params[1], params[2]))
		c.SetData("ok")
		return nil
	}
	return errors.New("此账号已注册")
}

//设置权限
func SetGrant(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Trace("SetGrant得到的参数为:%v", params)
	if _, ok := login.AccountManager.Accounts[params[0]]; ok {
		//如果无权限说明希望清空权限，此时给一个不可能的值，访问时会拒绝修改并清空原权限。
		if len(params[1]) == 0 {
			params[1] = "-1"
		}
		str_ids := strings.Split(params[1], "+")
		ids := make([]int, len(str_ids))
		for i := range str_ids {
			id, err := strconv.Atoi(str_ids[i])
			if err != nil {
				logs.Error("权限序号转换为数字类型时出错，传入的值为:%v", str_ids[i])
				return err
			}
			ids[i] = id
		}
		login.AccountManager.SetGrant(params[0], ids...)
		c.SetData("ok")
		return nil
	}
	return errors.New("不存在的账号")
}

//权限查询
func GetGrant(c *gm_command.Context, server, accountid string, params []string) error {
	if _, ok := login.AccountManager.Accounts[params[0]]; ok {
		err, grants := login.AccountManager.GetGrant(params[0])
		if err != nil {
			logs.Error("权限查询时出错，报错信息为:%v", err)
			return err
		}
		retJson, err := json.Marshal(grants)
		c.SetData(string(retJson))
		return nil
	}
	return errors.New("不存在的账号")
}

func ChangePass(c *gm_command.Context, server, accountid string, params []string) error {
	a := login.AccountManager.Get(params[0])
	if a == nil {
		return errors.New("NoGM")
	}
	a.Pass = params[0]
	a.Cookie = ""
	a.CookieTime = 0
	login.AccountManager.Add(*a)
	return nil
}

func GetUser(c *gm_command.Context, server, accountid string, params []string) error {
	c.SetData(string(login.AccountManager.GetAll()))
	return nil
}
