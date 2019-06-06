package errorctl

import (
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/secure"
	"vcs.taiyouxi.net/platform/x/auth/config"
)

type ErrorCodeForClient int

// 客户端对接使用的错误码，使得客户端能够正确处理错误跳转逻辑
// NOTE：请注意同事更新两个文档：
// - Auth&Login/Auth_ClientUI.md
// - http://wiki.taiyouxi.net/w/3k_engineer/client/interface/
// 并请和客户端沟通具体错误码的出现情况和客户端应该如何处理
// 帮助客户端合理友善的展示错误处理错误流程
const (
	ClientErrorNormal                   ErrorCodeForClient = 0   // Nothing
	ClientErrorUnknown                                     = 100 //通用：不应该出现的情况
	ClientErrorFormatVerifyFailed                          = 101 //通用： 请求的加密格式非法
	ClientErrorUsernameRegFormatIllegal                    = 103 //用户名格式非法
	ClientErrorMaybeDBProblem                              = 102 //通用： 数据库无法返回正确值
	ClientErrorHaveToUsePassword                           = 201 //设备ID认证：玩家设备绑定了帐号密码，强制要求使用用户名密码登录
	ClientErrorUsernameNotFound                            = 202 //用户名密码认证：用户名称未找到
	ClientErrorUserPasswordIncorrect                       = 203 //用户名密码认证：用户密码不正确
	ClientErrorRegCodeProblem                              = 204 //设备ID认证： 注册码验证错误
	ClientErrorRegCodeUsedProblem                          = 205 //设备ID认证： 注册码已被使用过
	ClientErrorBanByGM                                     = 206 //用户名密码认证：用户被禁止登陆
	ClientErrorHeroSdkToken                                = 210 //SDK: 英雄sdk查token返回错误
	ClientErrorHeroSdkGetUser                              = 211 //SDK: 英雄sdk查user信息返回错误
	ClientErrorQuickSdkCheckUser                           = 220 //SDK: Quicksdk checkuser错误
	ClientErrorUsernameHasBeenUsed                         = 301 //注册：用户名称已存在
	ClientErrorDeviceAlreadyBinded                         = 302 //注册：该设备关联的存档已经绑定过用户名,无法再次绑定
	ClientErrorLoginAuthtokenNotReady                      = 401 //登录网关：无法获取网关，因为认证数据无效
	ClientErrorGetGateNotExist                             = 402 //登录网关：当前分服无可用网关
	ClientErrorGetGateMaintenance                          = 403 //登录网关：服务器维护中
	ClientErrorUpdateVerParamErr                           = 501 //客户端更新重定向，参数错误
	ClientErrorUpdateVerUrlNotFound                        = 502 //客户端更新重定向，未找到url
)

func CtrlErrorReturn(c *gin.Context, prefix string, err error, clientErrorCode ErrorCodeForClient) {
	errInfo := err.Error()
	switch config.Cfg.Runmode {
	case "prod":
		//prod环境错误信息应该加密
		errInfo = secure.DefaultEncode.Encode64ForNet([]byte(errInfo))
	default:
		//do nothing
	}

	c.JSON(200, struct {
		Result    string             `json:"result"`
		Error     string             `json:"error,omitempty"`
		ForClient ErrorCodeForClient `json:"forclient,omitempty"`
	}{
		Result:    "no",
		ForClient: clientErrorCode,
		Error:     errInfo,
	})
}

func CtrlBanReturn(c *gin.Context, prefix string, err error, clientErrorCode ErrorCodeForClient, banTime int64, banReason string) {
	logs.Error("Error In:", prefix, " error: ", err.Error())
	errInfo := err.Error()
	switch config.Cfg.Runmode {
	case "prod":
		//prod环境错误信息应该加密
		errInfo = secure.DefaultEncode.Encode64ForNet([]byte(errInfo))
	default:
		//do nothing
	}

	c.JSON(200, struct {
		Result    string             `json:"result"`
		Error     string             `json:"error,omitempty"`
		BanReason string             `json:"banreason,omitempty"`
		BanTime   int64              `json"bantime,omitempty"`
		ForClient ErrorCodeForClient `json:"forclient,omitempty"`
	}{
		Result:    "no",
		ForClient: clientErrorCode,
		Error:     errInfo,
		BanTime:   banTime,
		BanReason: banReason,
	})
}
