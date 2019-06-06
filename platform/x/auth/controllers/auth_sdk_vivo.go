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

func (dic *DeviceIDController) RegDeviceWithVivoAndLogin_v2() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		_token := c.PostForm("token")
		_channelId := c.PostForm("channelId")
		_device := c.PostForm("device")
		if _token == "" || _channelId == "" {
			//无法校验的请求一概放弃
			errorctl.CtrlErrorReturn(c, "[Auth.DeviceWithVivo]",
				fmt.Errorf(":-)"), errorctl.ClientErrorFormatVerifyFailed,
			)
			return
		}

		b_token, _ := secure.DefaultEncode.Decode64FromNet(_token)
		b_channelId, _ := secure.DefaultEncode.Decode64FromNet(_channelId)
		device := ""
		if _device != "" {
			b_device, _ := secure.DefaultEncode.Decode64FromNet(_device)
			device = string(b_device)
		}

		token := string(b_token)
		channelId := string(b_channelId)

		logs.Trace(" rec vivo info %s %s", token, channelId)

		err, id := sdk.CheckVivo(token)
		if err != nil {
			logs.Error("vivo checkuser err %s", err.Error())
			errorctl.CtrlErrorReturn(c, "[Auth.DeviceWithQuick]",
				fmt.Errorf(":-)"), errorctl.ClientErrorQuickSdkCheckUser,
			)
			return
		}

		res, authToken, display, uid := login(c, "vivo.com", id, channelId, device)
		if !res {
			return
		}

		logs.Info("[login] vivo uid %s sdkuid %s channelId %s device %s",
			uid, id, channelId, device)

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
