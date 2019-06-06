package pay

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"

	"fmt"

	"strconv"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/x/api_gateway/logiclog"
	"vcs.taiyouxi.net/platform/x/api_gateway/util"
)

const (
	VNAndroidChannel = "5003"
	VNIOSChannel     = "5004"
)

const (
	oriTypeAndroid = 10
	oriTypeLumi    = 84
)

const (
	// TODO by ljz conf
	exRate = 20000 / 0.9
)

type VNMessages struct {
	Message VNMessage `xml:"message"`
}

type VNMessage struct {
	LoginName  string  `xml:"login_name"`
	SelfDefine string  `xml:"self_define"` // 渠道标示ID
	OrderID    string  `xml:"order_id"`    // 订单号
	PayTime    string  `xml:"pay_time"`    // 支付时间 2015-01-01 23:00:00
	Amount     float64 `xml:"amount"`      // 成交金额
	Status     string  `xml:"status"`      // 充值状态 0 成功 1失败(为1时 应返回FAILUD失败)
	OriType    int     `xml:"ori_type"`    // 原始支付类型，如 10（谷歌支付）
	OriAmount  float64 `xml:"ori_amount"`  // 原始金额
}

func VNPayNotify(typ string, data sdkPayData) gin.HandlerFunc {
	return func(c *gin.Context) {

		cfgCallbackKey := data.Sdk[typ].Params[0]

		signOrg := c.PostForm("sign")
		nt_data := c.PostForm("nt_data")
		logs.Debug("rec sign %s nt_data %s", signOrg, nt_data)

		// sign 解密
		signDec, err := util.Decode(signOrg, cfgCallbackKey)
		if err != nil {
			logs.Error("QuickPayNotify decode sign err %s", err.Error())
			c.String(http.StatusBadRequest, "DECODE SIGN FAILED")
			return
		}
		logs.Debug("decode sign %s", signDec)

		// nt_data sign
		m := map[string]string{}
		m["nt_data"] = nt_data
		ntDataSign := util.GenSign(m)
		logs.Debug("sign nt_data %s", ntDataSign)
		if ntDataSign != signDec {
			logs.Error("VNPayNotify sign nt_data not same")
			c.String(http.StatusBadRequest, "DECODE SIGN FAILED")
			return
		}

		// decode nt_data
		xmlInfo, err := util.Decode(nt_data, cfgCallbackKey)
		if err != nil {
			logs.Error("VNPayNotify decode nt_data err %s", err.Error())
			c.String(http.StatusBadRequest, "FAILED")
			return
		}

		logs.Debug("decode xml %q", xmlInfo)

		var messages VNMessages
		err = xml.Unmarshal([]byte(xmlInfo), &messages)
		if err != nil {
			logs.Error("VNPayNotify xml unmarshal err %s", err.Error())
			c.String(http.StatusBadRequest, "FAILED")
			return
		}
		logs.Debug("messages %v", messages)
		// amount
		amountStr := fmt.Sprintf("%d", int(messages.Message.OriAmount))

		extr_info := messages.Message.SelfDefine
		ex := strings.Split(extr_info, "|")
		if len(ex) < 3 {
			logs.Error("VNPayNotify Extras_params err %v, %v",
				amountStr, messages)
			c.String(http.StatusInternalServerError, "fail")
			return
		}
		payTime := ex[0]
		productId := ex[1]
		ver := ex[2]
		game_order_org := ex[3]

		cpInfo := genCPInfo(typ, game_order_org)
		if cpInfo == nil {
			logs.Error("VNPayNotify genCPInfo err Data %v", messages)
			c.String(http.StatusBadRequest, "fail")
			return
		}

		uid := cpInfo.uid
		isdbDebug := cpInfo.isDBDebug
		game_order := cpInfo.game_order
		pkgid := 0
		subpkgid := 0
		hash_key := messages.Message.OrderID
		channelID := VNAndroidChannel
		if data.IsIOS() {
			channelID = VNIOSChannel
		} else {
			channelID = VNAndroidChannel
		}
		// 成功失败
		if messages.Message.Status != "0" {
			// log
			logiclog.LogPay(data.Name, uid, false, channelID, uid, hash_key, messages.Message.PayTime,
				amountStr, "0", "", "", game_order_org, extr_info,
				game_order, tistatus_delivered, payTime, productId, ver,fmt.Sprintf("%d:%d", pkgid, subpkgid))
			logs.Debug("VNPayNotify order %s fail", hash_key)
			c.String(http.StatusBadRequest, "Status Failed")
			return
		}
		// 判断重复
		if res, resMsg, isRepeat := checkOrderNoRepeat(typ, data, isdbDebug, hash_key); !res {
			logs.Error("VNPayNotify checkOrderNoRepeat err %s %v", resMsg, messages)
			c.String(http.StatusInternalServerError, "fail")
			return
		} else {
			if isRepeat {
				logs.Warn("VNPayNotify order_no repeat %v", messages)
				c.String(http.StatusOK, "fail")
				return
			}
		}
		// lumi支付不需要比较验证，可能任意金额
		if messages.Message.OriType != oriTypeLumi {
			//比较钱和订单需要的钱是否一致
			if res, msg := checkMoneyWithCfg(typ, data, game_order, uint32(messages.Message.OriAmount), uid); !res {
				logs.Error("[%s] VNPayNotify checkMoneyWithCfg err %v %s, %v", typ, game_order, msg, messages)
				c.String(http.StatusInternalServerError, "fail")
				return
			}
		}
		orderID := game_order
		if messages.Message.OriType == oriTypeLumi {
			idx, err := strconv.Atoi(game_order)
			if err != nil {
				logs.Error("[%s] PayNotify Game_order err %s", typ, game_order)
				c.String(http.StatusInternalServerError, "fail")
				return
			}
			cfg := gamedata.GetIAPInfo(uint32(idx))
			if cfg == nil {
				logs.Error("[%s] PayNotify Game_order config err %s", typ, game_order)
				c.String(http.StatusInternalServerError, "fail")
				return
			}
			if data.IsIOS() {
				if uint32(messages.Message.OriAmount) < cfg.IOS_Rmb_Price {
					orderID = "0"
				}
			} else {
				if uint32(messages.Message.OriAmount) < cfg.Android_Rmb_Price {
					orderID = "0"
				}
			}
		}
		logs.Debug("final order id: %v", orderID)
		value := make(map[string]interface{}, 11)
		if channelID == VNAndroidChannel {
			value[key_channel] = VNAndroidChannel // 渠道标示ID
		} else {
			value[key_channel] = VNIOSChannel // 渠道标示ID
		}
		value[key_channel_uid] = messages.Message.LoginName // 渠道用户唯一标示
		value[key_order_no] = hash_key                      // 唯一订单号
		value[key_pay_time] = payTime                       // 支付时间
		value[key_money_amount] = amountStr                 // 成交金额
		value[key_status] = messages.Message.Status         // 充值状态 0 成功 1失败(为1时 应返回FAILUD失败)
		value[key_product] = productId
		value[key_ver] = ver
		value[key_extras_params] = extr_info
		value[key_tistatus] = tistatus_paid // 通知状态，paid支付成功；delivered支付成功并玩家已拿到
		value[key_uid] = uid                // accountid
		value[key_mobile] = data.Name
		if err := data.PaySetByHashM(isdbDebug, hash_key, value); err != nil {
			logs.Warn("VNPayNotify SetByHashM db err %v %s", isdbDebug, err.Error())
			c.String(http.StatusInternalServerError, "fail")
			return
		}

		// send mail to dynamodb
		order := timail.IAPOrder{
			Order_no:      hash_key,
			Game_order:    orderID,
			Game_order_id: game_order_org,
			Amount:        amountStr,
			Channel:       channelID,
			PayTime:       payTime,
			PkgInfo:       timail.PackageInfo{pkgid, subpkgid},
			PayType:       fmt.Sprintf("%d", messages.Message.OriType),
		}
		info, _ := json.Marshal(&order)

		if err := data.SendMail(isdbDebug, uid, string(info)); err != nil {
			logs.Error("VNPayNotify SetByHashM db err %v %s, %v", isdbDebug, err.Error(), messages)
			c.String(http.StatusInternalServerError, "fail")
			return
		}
		uperr := data.PayUpdateByHash(isdbDebug, hash_key,
			map[string]interface{}{
				key_tistatus: tistatus_delivered,
			})
		if uperr != nil {
			logs.Error("VNPayNotify GetByHashM db err %v %s, %v", isdbDebug, uperr.Error(), messages)
			c.String(http.StatusInternalServerError, "fail")
			return
		}

		// log
		logiclog.LogVNPay(data.Name, uid, true, channelID, uid, hash_key, messages.Message.PayTime,
			amountStr, "0", "", "", game_order_org, extr_info,
			game_order, tistatus_delivered, payTime, productId, ver, fmt.Sprintf("%d", messages.Message.OriType))

		logs.Debug("EnjoyPayNotify order %s success", hash_key)
		c.String(http.StatusOK, "SUCCESS")
	}
}
