package pay

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"

	"strconv"

	"fmt"
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/x/api_gateway/logiclog"
	"vcs.taiyouxi.net/platform/x/api_gateway/util"
)

const (
	key_tistatus       = "tistatus"
	key_uid            = "uid"
	key_channel        = "channel"
	key_channel_uid    = "channel_uid"
	key_good_idx       = "good_idx"
	key_order_no       = "order_no"
	key_pay_time       = "pay_time"
	key_money_amount   = "money_amount"
	key_status         = "status"
	key_extras_params  = "extras_params"
	key_product        = "product_id"
	key_ver            = "ver"
	key_mobile         = "mobile"
	key_IsTest         = "is_test"
	key_note           = "note"
	tistatus_paid      = "Paid"
	tistatus_delivered = "Delivered"
	test_tag           = "1"
	gid_test_uplimit   = 100
)

type Messages struct {
	Message Message `xml:"message"`
}

type Message struct {
	IsTest        string `xml:"is_test"`       // 是否为测试订单 1为测试 0为线上正式订单
	Channel       string `xml:"channel"`       // 渠道标示ID
	ChannelUid    string `xml:"channel_uid"`   // 渠道用户唯一标示,该值从客户端GetUserId()中可获取
	Game_order    string `xml:"game_order"`    // 游戏在调用QucikSDK发起支付时传递的游戏方订单,这里会原样传回
	Order_no      string `xml:"order_no"`      // 天象唯一订单号
	Pay_time      string `xml:"pay_time"`      // 支付时间 2015-01-01 23:00:00
	Amount        string `xml:"amount"`        // 成交金额
	Status        string `xml:"status"`        // 充值状态 0 成功 1失败(为1时 应返回FAILUD失败)
	Extras_params string `xml:"extras_params"` // 游戏客户端调用SDK发起支付时填写的透传参数.没有则为空
	Note          string `xml:"note"`          // ios沙盒情况下为"sandbox"；其他情况为空
}

func QuickPayNotify(typ string, data sdkPayData) gin.HandlerFunc {
	return func(c *gin.Context) {
		resS := "SUCCESS"

		cfgCallbackKey := data.Sdk[typ].Params[0]

		signOrg := c.PostForm("sign")
		nt_data := c.PostForm("nt_data")

		//		logs.Debug("rec sign %s nt_data %s", signOrg, nt_data)

		// sign 解密
		signDec, err := util.Decode(signOrg, cfgCallbackKey)
		if err != nil {
			logs.Error("QuickPayNotify decode sign err %s", err.Error())
			c.String(http.StatusBadRequest, "DECODE SIGN FAILED")
			return
		}

		//		logs.Debug("decode sign %s", signDec)

		// nt_data sign
		m := map[string]string{}
		m["nt_data"] = nt_data
		ntDataSign := util.GenSign(m)
		//		logs.Debug("sign nt_data %s", ntDataSign)
		if ntDataSign != signDec {
			logs.Error("QuickPayNotify sign nt_data not same")
			c.String(http.StatusBadRequest, "DECODE SIGN FAILED")
			return
		}

		// decode nt_data
		xmlInfo, err := util.Decode(nt_data, cfgCallbackKey)
		if err != nil {
			logs.Error("QuickPayNotify decode nt_data err %s", err.Error())
			c.String(http.StatusBadRequest, "FAILED")
			return
		}

		logs.Debug("decode xml %q", xmlInfo)

		var messages Messages
		err = xml.Unmarshal([]byte(xmlInfo), &messages)
		if err != nil {
			logs.Error("QuickPayNotify xml unmarshal err %s", err.Error())
			c.String(http.StatusBadRequest, "FAILED")
			return
		}
		messagesJson, _ := json.Marshal(messages)
		logs.Debug("messages %v", messages)

		// 解析参数
		cpInfo := genCPInfo(typ, messages.Message.Game_order)
		if cpInfo == nil {
			c.String(http.StatusBadRequest, "FAILED")
			return
		}
		uid := cpInfo.uid
		isdbDebug := cpInfo.isDBDebug
		game_order := cpInfo.game_order
		hash_key := messages.Message.Order_no
		// amount
		amount, err := strconv.ParseFloat(messages.Message.Amount, 64)
		if err != nil {
			logs.Error("[%s] PayNotify Amount err %v %s", typ, game_order, messages.Message.Amount)
			c.String(http.StatusInternalServerError, "GameOrder Param Err")
			return
		}
		// Extras
		ex := strings.Split(messages.Message.Extras_params, ":")
		if len(ex) < 3 {
			logs.Error("QuickPayNotify Extras_params err %s %v, json:%q",
				messages.Message.Game_order, messages.Message.Amount, messagesJson)
			c.String(http.StatusInternalServerError, "Extras_params Err")
			return
		}
		payTime := ex[0]
		productId := ex[1]
		ver := ex[2]
		pkgid := 0
		subpkgid := 0
		var cfg *ProtobufGen.IAPBASE
		if data.IsIOS() {
			cfg = gamedata.GetIAPBaseConfigIOS(game_order)
		} else {
			cfg = gamedata.GetIAPBaseConfigAndroid(game_order)
		}
		if cfg == nil {
			logs.Error("quick Can Not Find goods_id:%s", game_order)
			c.String(http.StatusInternalServerError, "fail")
			return
		}

		if len(ex) > 3 {

			pkgid, err = strconv.Atoi(ex[3])
			if err != nil {
				logs.Error("[%s] packageid convert to int err %s", typ, err.Error())
			}
			logs.Debug("QuickNotify")
			subpkgid, err = strconv.Atoi(ex[4])
			if err != nil {
				logs.Error("[%s] subpackageid convert to int err %s", typ, err.Error())
			}
			logs.Debug("QuickNotify: PackageIAP %d:%d", pkgid, subpkgid)
		}
		if pkgid != 0 {
			pkgValue := gamedata.GetHotDatas().HotKoreaPackge.GetHotPackage(int64(pkgid), int64(subpkgid))
			isTrue := false
			if pkgValue == nil {
				logs.Error("can not find the package %d:%d", pkgid, subpkgid)
				pkgid = 0
				subpkgid = 0
			} else {
				IapId := pkgValue.GetIAPID()
				tc := strings.Split(IapId, ",")
				for i := 0; i < len(tc); i++ {
					if tc[i] == fmt.Sprintf("%d", cfg.GetIndex()) {
						isTrue = true
						break
					}
				}
			}
			if !isTrue {
				pkgid = 0
				subpkgid = 0
			}
		}

		// 成功失败
		if messages.Message.Status != "0" {
			// log
			msg := messages.Message
			logiclog.LogPay(data.Name, uid, false, msg.Channel, msg.ChannelUid, msg.Order_no, msg.Pay_time,
				msg.Amount, msg.Status, msg.IsTest, msg.Note, msg.Game_order, msg.Extras_params,
				game_order, tistatus_delivered, payTime, productId, ver,fmt.Sprintf("%d:%d", pkgid, subpkgid))
			logs.Debug("QuickPayNotify order %s fail", hash_key)
			c.String(http.StatusBadRequest, "Status Failed")
			return
		}

		// 判断重复
		if ret, resMsg, isRepeat := checkOrderNoRepeat(typ, data, isdbDebug, hash_key); !ret {
			c.String(http.StatusInternalServerError, resMsg)
			return
		} else {
			if isRepeat {
				logs.Warn("QuickPayNotify order_no repeat %q", messagesJson)
				c.String(http.StatusOK, resS)
				return
			}
		}

		// 比较钱和订单需要的钱是否一致
		if res, msg := checkMoneyWithCfg(typ, data, game_order, uint32(amount), uid); !res {
			c.String(http.StatusInternalServerError, msg)
			return
		}

		// 存入dynamo
		value := make(map[string]interface{}, 11)
		value[key_channel] = messages.Message.Channel        // 渠道标示ID
		value[key_channel_uid] = messages.Message.ChannelUid // 渠道用户唯一标示,该值从客户端GetUserId()中可获取
		value[key_good_idx] = messages.Message.Game_order    // 游戏在调用QucikSDK发起支付时传递的游戏方订单,这里会原样传回
		value[key_order_no] = messages.Message.Order_no      // 天象唯一订单号
		value[key_pay_time] = messages.Message.Pay_time      // 支付时间
		value[key_money_amount] = messages.Message.Amount    // 成交金额
		value[key_status] = messages.Message.Status          // 充值状态 0 成功 1失败(为1时 应返回FAILUD失败)
		value[key_product] = productId
		value[key_ver] = ver
		value[key_extras_params] = messages.Message.Extras_params
		value[key_tistatus] = tistatus_paid // 通知状态，paid支付成功；delivered支付成功并玩家已拿到
		value[key_uid] = uid                // accountid
		value[key_mobile] = data.Name
		value[key_IsTest] = messages.Message.IsTest
		if messages.Message.Note != "" {
			value[key_note] = messages.Message.Note
		}

		if err := data.PaySetByHashM(isdbDebug, hash_key, value); err != nil {
			mapB, _ := json.Marshal(value)
			logs.Warn("QuickPayNotify SetByHashM db err %v %s, json:%q", isdbDebug, err.Error(), mapB)
			c.String(http.StatusInternalServerError, "FAILED")
			return
		}

		// send mail to dynamodb
		order := timail.IAPOrder{
			Order_no:      messages.Message.Order_no,
			Game_order:    game_order,
			Game_order_id: messages.Message.Game_order,
			Amount:        messages.Message.Amount,
			Channel:       messages.Message.Channel,
			PayTime:       payTime,
			PkgInfo:       timail.PackageInfo{pkgid, subpkgid}}
		info, _ := json.Marshal(&order)
		logs.Debug("protoctID: %s, cpInfo.gid: %d, data.worker.MailTemp: %v", productId, cpInfo.gid, data.worker.MailTemp)
		if productId == "203" && cpInfo.gid == 200 && data.worker.MailTemp != nil {
			if err := data.SendMailTemp(uid, string(info)); err != nil {
				logs.Warn("QuickPayNotify[TMP] SetByHashM db err %v %s, json:%q", isdbDebug, err.Error(), messagesJson)
				c.String(http.StatusInternalServerError, "FAILED")
				return
			}
		} else {
			if err := data.SendMail(isdbDebug, uid, string(info)); err != nil {
				logs.Warn("QuickPayNotify SetByHashM db err %v %s, json:%q", isdbDebug, err.Error(), messagesJson)
				c.String(http.StatusInternalServerError, "FAILED")
				return
			}
		}
		uperr := data.PayUpdateByHash(isdbDebug, hash_key,
			map[string]interface{}{
				key_tistatus: tistatus_delivered,
			})
		if uperr != nil {
			logs.Error("QuickPayNotify GetByHashM db err %v %s, json:%q", isdbDebug, uperr.Error(), messagesJson)
			c.String(http.StatusInternalServerError, "DB Failed")
			return
		}

		// log
		msg := messages.Message
		logiclog.LogPay(data.Name, uid, true, msg.Channel, msg.ChannelUid, msg.Order_no, msg.Pay_time,
			msg.Amount, msg.Status, msg.IsTest, msg.Note, msg.Game_order, msg.Extras_params,
			game_order, tistatus_delivered, payTime, productId, ver,fmt.Sprintf("%d:%d", pkgid, subpkgid))

		logs.Debug("QuickPayNotify order %s success", messages.Message.Order_no)
		c.String(http.StatusOK, resS)
	}
}
