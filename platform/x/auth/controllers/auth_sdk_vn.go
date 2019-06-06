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

func (dic *DeviceIDController) RegDeviceWithVNAndLogin_v2() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := struct {
			Result           string                `json:"result"`
			Authtoken        string                `json:"authtoken,omitempty"`
			DisplayName      string                `json:"display,omitempty"`
			ShardHasRoleV250 []models.ShardHasRole `json:"shardrolev250"`
			ShardHasRole     []string              `json:"shardrole"`
			LastShard        string                `json:"lastshard"`
			Err              string                `json:"error,omitempty"`
			Username         string                `json:"username"`
			Id               int64                 `json:"id"`
			Sdkuserid        int64                 `json:"sdkuserid"`
		}{
			"no",
			"", "", []models.ShardHasRole{}, []string{}, "", "",
			"", 0, 0,
		}

		_authCode := c.PostForm("authCode")
		_channelId := c.PostForm("channelId")
		_device := c.PostForm("device")
		if _authCode == "" || _channelId == "" {
			//无法校验的请求一概放弃
			errorctl.CtrlErrorReturn(c, "[Auth.DeviceWithVN]",
				fmt.Errorf(":-)"), errorctl.ClientErrorFormatVerifyFailed,
			)
			return
		}

		b_authCode, _ := secure.DefaultEncode.Decode64FromNet(_authCode)
		b_channelId, _ := secure.DefaultEncode.Decode64FromNet(_channelId)
		device := ""
		if _device != "" {
			b_device, _ := secure.DefaultEncode.Decode64FromNet(_device)
			device = string(b_device)
		}

		s_authcode := string(b_authCode)
		channelId := string(b_channelId)

		// get token
		token, err := sdk.GetVNToken(s_authcode, channelId)
		if err != nil {
			logs.Error("vn gettoken err %s %s", s_authcode, err.Error())
			errorctl.CtrlErrorReturn(c, "[Auth.DeviceWithVN]",
				fmt.Errorf(":-)"), errorctl.ClientErrorHeroSdkToken,
			)
			return
		}
		logs.Trace(" rec vn token %s channel %s device %s", token, channelId, device)

		// get user info
		userInfo, err := sdk.GetVNUserInfo(token, channelId)
		if err != nil {
			logs.Error("vn getuserinfo err %s", err.Error())
			errorctl.CtrlErrorReturn(c, "[Auth.DeviceWithHero]",
				fmt.Errorf(":-)"), errorctl.ClientErrorHeroSdkGetUser,
			)
			return
		}
		logs.Trace(" rec vn userinfo %d %d", userInfo.Id, userInfo.Sdkuserid)

		res, authToken, display, uid := login(c, "vn.com", fmt.Sprintf("%d", userInfo.Id), channelId, device)
		if !res {
			return
		}

		logs.Info("[login] vn uid %s sdkuid %d %d channelId %s", uid, userInfo.Id, userInfo.Sdkuserid, channelId)

		r.ShardHasRoleV250, r.LastShard = _getUserShardInfo(uid)
		for _, st := range r.ShardHasRoleV250 {
			r.ShardHasRole = append(r.ShardHasRole, st.Shard)
		}

		r.Result = "ok"
		r.Authtoken = authToken
		r.DisplayName = display
		r.Username = userInfo.Username
		r.Sdkuserid = userInfo.Sdkuserid
		r.Id = userInfo.Id

		c.JSON(200, r)
	}
}
