package gm_command

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"io/ioutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/util"
)

type commandParams struct {
	Params []string `json:"params" form:"params"`
	Key    string   `json:"key" form:"key"`
}

var (
	errNoAccount = errors.New("errNoAccount")
	errNoGrants  = errors.New("errNoGrants")
)

func CommandGet(r *gin.Engine) {
	r.POST("/api/v1/command/:server_name/:account_id/*command", func(c *gin.Context) {
		account_id := c.Param("account_id")
		server_name := c.Param("server_name")
		command := util.DeleteBackslash(c.Param("command"))

		s := commandParams{}
		err := c.Bind(&s)

		if err != nil {
			c.String(400, errRes(err))
			return
		}

		s.Params = append(s.Params, "System")

		logs.Info("command get %s %s %s - %v", server_name, account_id, command, s)

		acc := GetAccountByKey(s.Key)
		if acc == nil {
			c.String(401, errRes(errNoAccount))
			return
		}

		if !CheckGrant(s.Key, command) {
			c.String(403, errRes(errNoGrants))
			return
		}

		if command == "setAllInfoByAccountID" {
			if accountInfo, err := parseFile(c); err == nil {
				s.Params[0] = string(accountInfo)
			} else {
				c.String(500, errRes(err))
				return
			}
		}

		data, err := OnCommand(command, server_name, account_id, s.Params[:])

		if data == "" {
			data = "{}"
		}

		if err == nil {
			if command == "getAllInfoByAccountID" {
				// 下载文件
				c.Header("Content-Disposition", fmt.Sprintf("attachment;fileName=%s.json", account_id))
				c.Data(200, "multipart/form-data", []byte(data))
			} else if command == "getBatchMailDetail" {
				c.Header("Content-Disposition", fmt.Sprintf("attachment;fileName=batch_mail_%s.csv", s.Params[0]))
				c.Data(200, "multipart/form-data", []byte(data))
			} else {
				c.String(200, data)
			}

		} else {
			logs.Error("command %s err %s", command, err.Error())
			c.String(500, errRes(err))
		}
	})
}

func errRes(err error) string {
	return fmt.Sprintf("{\"err\":\"%s\"}", err.Error())
}

// key, params
func parseFile(c *gin.Context) ([]byte, error) {
	file, fileName, err := c.Request.FormFile("file")
	if err != nil {
		logs.Warn("fail to load file")
		return nil, err
	}
	logs.Debug("receive fileName, %s", fileName.Filename)
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		logs.Warn("fail to parse file")
		return nil, err
	}
	return bytes, nil
}
