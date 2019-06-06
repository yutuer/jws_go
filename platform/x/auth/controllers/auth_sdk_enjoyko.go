package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/secure"
	"vcs.taiyouxi.net/platform/x/auth/errorctl"
	"vcs.taiyouxi.net/platform/x/auth/models"
	"vcs.taiyouxi.net/platform/x/auth/models/sdk"
)

func (dic *DeviceIDController) RegDeviceWithEnjoyKoAndLogin_v2() gin.HandlerFunc {
	return func(c *gin.Context) {
		//用于客户端的校验
		r := struct {
			Result           string                `json:"result"`
			Authtoken        string                `json:"authtoken,omitempty"`
			DisplayName      string                `json:"display,omitempty"`
			ShardHasRoleV250 []models.ShardHasRole `json:"shardrolev250"`
			ShardHasRole     []string              `json:"shardrole"`
			LastShard        string                `json:"lastshard"`
			Err              string                `json:"error,omitempty"`
		}{
			"no",
			"", "", []models.ShardHasRole{}, []string{}, "", "",
		}

		_uid := c.PostForm("uid")
		_token := c.PostForm("token")
		_channelId := c.PostForm("channelId")
		_typ := c.PostForm("typ")
		_device := c.PostForm("device")
		if _uid == "" || _token == "" || _channelId == "" {
			//无法校验的请求一概放弃
			errorctl.CtrlErrorReturn(c, "[Auth.DeviceWithQuick]",
				fmt.Errorf(":-)"), errorctl.ClientErrorFormatVerifyFailed,
			)
			return
		}

		b_uid, _ := secure.DefaultEncode.Decode64FromNet(_uid)
		b_token, _ := secure.DefaultEncode.Decode64FromNet(_token)
		b_channelId, _ := secure.DefaultEncode.Decode64FromNet(_channelId)

		sdkUid := string(b_uid)
		token := string(b_token)
		channelId := string(b_channelId)
		typ := sdk.Typ_Android
		if _typ != "" {
			b_typ, _ := secure.DefaultEncode.Decode64FromNet(_typ)
			typ = string(b_typ)
		}
		device := ""
		if _device != "" {
			b_device, _ := secure.DefaultEncode.Decode64FromNet(_device)
			device = string(b_device)
		}

		logs.Trace(" rec enjoyko info %v %s %s %s %s %s",
			sdkUid, token, channelId, typ, device)
		//与Enjoy的校验
		err := sdk.CheckEnjoy(token, sdkUid)
		if err != nil {
			logs.Error("enjoyko checkuser err %s", err.Error())
			errorctl.CtrlErrorReturn(c, "[Auth.DeviceWithEnjoy]",
				fmt.Errorf(":-)"), errorctl.ClientErrorQuickSdkCheckUser, /*enjoy sdk error code?*/
			)
			return
		}
		channelId = ko_spec_channel("enjoyko.com", sdkUid, channelId)
		res, authToken, display, uid := login(c, "enjoyko.com", sdkUid, channelId, device)
		if !res {
			return
		}

		logs.Info("[login] enjoyko uid %s sdkUid: %s channelId %s device %s",
			uid, sdkUid, channelId, device)

		r.ShardHasRoleV250, r.LastShard = _getUserShardInfo(uid)
		for _, st := range r.ShardHasRoleV250 {
			r.ShardHasRole = append(r.ShardHasRole, st.Shard)
		}

		r.Result = "ok"
		r.Authtoken = authToken
		r.DisplayName = display

		c.JSON(200, r)
	}
}
