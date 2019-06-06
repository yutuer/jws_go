package pay

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"

	"vcs.taiyouxi.net/jws/gamex/protogen"

	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/x/api_gateway/logiclog"
	"vcs.taiyouxi.net/platform/x/api_gateway/util"
)

const (
	KOGPAndroidChannel = "5005"
	KOONESTOREChannel  = "5006"
	KOIOSChannel       = "5007"
)

func EnjoyKoPayNotify(typ string, data sdkPayData) gin.HandlerFunc {
	return func(c *gin.Context) {

		cfgCallbackKey := data.Sdk[typ].Params[0]

		signOrg := c.PostForm("sign")

		param := make(map[string]string, 16)
		param["order_id"] = c.PostForm("order_id")
		param["app_id"] = c.PostForm("app_id")
		param["goods_id"] = c.PostForm("goods_id")
		param["pay_description"] = c.PostForm("pay_description")
		param["pay_time"] = c.PostForm("pay_time")
		param["role_id"] = c.PostForm("role_id")
		param["server_id"] = c.PostForm("server_id")
		param["trade_provider"] = c.PostForm("trade_provider")
		param["total_money"] = c.PostForm("total_money")
		param["uid"] = c.PostForm("uid")
		logs.Debug("EnjoyKo pay notify receive param: %v, and sign is %s", param, signOrg)
		// sign 解密
		signDec, err := util.GenEnjoySign(param, cfgCallbackKey)
		if err != nil {
			logs.Error("EnjoyKoPayNotify decode sign err %s", err.Error())
			c.String(http.StatusBadRequest, "fail")
			return
		}
		logs.Debug("origin sign %s", signOrg)
		logs.Debug("decode sign %s", signDec)
		if signOrg != signDec {
			logs.Error("EnjoyKoPayNotify sign not same")
			c.String(http.StatusBadRequest, "fail")
			return
		}

		hash_key := param["order_id"]
		pay_time := param["pay_time"]
		amount_str := param["total_money"]
		amount, err := strconv.Atoi(amount_str)
		if err != nil {
			logs.Error("[%s] PayNotify Amount err %v %v", typ, amount_str, param)
			c.String(http.StatusInternalServerError, "fail")
		}
		//amount /= 100.0
		//amount_str = strconv.FormatFloat(amount, 'f', -1, 64)
		logs.Debug("convert true amount str:%s, value:%d", amount_str, amount)
		pkgid := 0
		subpkgid := 0
		extr_info := param["pay_description"]
		ex := strings.Split(extr_info, "|")
		if len(ex) < 4 {
			logs.Error("EnjoyKoPayNotify Extras_params err %v, %v",
				amount, param)
			c.String(http.StatusInternalServerError, "fail")
			return
		}
		payTime := ex[0]
		productId := ex[1]
		ver := ex[2]
		game_order_org := ex[3]
		if len(ex) > 4 {

			pkgid, err = strconv.Atoi(ex[4])
			if err != nil {
				logs.Error("[%s] packageid convert to int err %s", typ, err.Error())
			}
			logs.Debug("EnjoyKoNotify: koreaPackageIAP")
			subpkgid, err = strconv.Atoi(ex[5])
			if err != nil {
				logs.Error("[%s] subpackageid convert to int err %s", typ, err.Error())
			}
			logs.Debug("EnjoyKoNotify: koreaPackageIAP %d:%d", pkgid, subpkgid)
		}

		cpInfo := genCPInfo(typ, game_order_org)
		if cpInfo == nil {
			logs.Error("EnjoyKoPayNotify genCPInfo err Data %v", param)
			c.String(http.StatusBadRequest, "fail")
			return
		}

		uid := cpInfo.uid
		isdbDebug := cpInfo.isDBDebug
		game_order := cpInfo.game_order
		// 判断重复
		if res, resMsg, isRepeat := checkOrderNoRepeat(typ, data, isdbDebug, hash_key); !res {
			logs.Error("EnjoyKoPayNotify checkOrderNoRepeat err %s %v", resMsg, param)
			c.String(http.StatusInternalServerError, "fail")
			return
		} else {
			if isRepeat {
				logs.Warn("EnjoyKoPayNotify order_no repeat %v", param)
				c.String(http.StatusOK, "fail")
				return
			}
		}
		//// 比较钱和订单需要的钱是否一致
		//if res, msg := checkMoneyWithCfg(typ, data, game_order, uint32(amount), uid); !res {
		//	logs.Error("[%s] PayNotify checkMoneyWithCfg err %v %s, %v", typ, game_order, msg, param)
		//	c.String(http.StatusInternalServerError, "fail")
		//	return
		//}

		var cfg *ProtobufGen.IAPBASE
		value := make(map[string]interface{}, 11)
		var channelID string
		if param["trade_provider"] == "1" || param["trade_provider"] == "110" {
			value[key_channel] = KOGPAndroidChannel // 渠道标示ID
			channelID = KOGPAndroidChannel
			cfg = gamedata.GetIAPBaseConfigAndroid(param["goods_id"])
		} else if param["trade_provider"] == "4" {
			value[key_channel] = KOONESTOREChannel // 渠道标示ID
			channelID = KOONESTOREChannel
			cfg = gamedata.GetIAPBaseConfigAndroid(param["goods_id"])
		} else if param["trade_provider"] == "2" {
			value[key_channel] = KOIOSChannel // 渠道标示ID
			channelID = KOIOSChannel
			cfg = gamedata.GetIAPBaseConfigIOS(param["goods_id"])
		}

		if cfg == nil {
			logs.Error("EnjoyKoPay Can Not Find goods_id:%s", param["goods_id"])
			c.String(http.StatusInternalServerError, "fail")
			return
		}
		game_order_org = strconv.Itoa(int(cfg.GetIndex()))
		amount_str = strconv.Itoa(int(cfg.GetPrice()))

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

		value[key_channel_uid] = param["uid"] // 渠道用户唯一标示
		value[key_order_no] = hash_key        // 唯一订单号
		value[key_pay_time] = payTime         // 支付时间
		value[key_money_amount] = amount      // 成交金额
		value[key_status] = "0"               // 充值状态 0 成功 1失败(为1时 应返回FAILUD失败)
		value[key_product] = productId
		value[key_ver] = ver
		value[key_extras_params] = extr_info
		value[key_tistatus] = tistatus_paid // 通知状态，paid支付成功；delivered支付成功并玩家已拿到
		value[key_uid] = uid                // accountid
		value[key_mobile] = data.Name
		if err := data.PaySetByHashM(isdbDebug, hash_key, value); err != nil {
			mapB, _ := json.Marshal(param)
			logs.Warn("EnjoyKoPayNotify SetByHashM db err %v %s, json:%q", isdbDebug, err.Error(), mapB)
			c.String(http.StatusInternalServerError, "fail")
			return
		}

		// send mail to dynamodb
		order := timail.IAPOrder{
			Order_no:      hash_key,
			Game_order:    game_order,
			Game_order_id: game_order_org,
			Amount:        amount_str,
			Channel:       channelID,
			PayTime:       payTime,
			PkgInfo:       timail.PackageInfo{pkgid, subpkgid},
		}
		info, _ := json.Marshal(&order)

		if err := data.SendMail(isdbDebug, uid, string(info)); err != nil {
			logs.Error("EnjoyKoPayNotify SetByHashM db err %v %s, %v", isdbDebug, err.Error(), param)
			c.String(http.StatusInternalServerError, "fail")
			return
		}
		uperr := data.PayUpdateByHash(isdbDebug, hash_key,
			map[string]interface{}{
				key_tistatus: tistatus_delivered,
			})
		if uperr != nil {
			logs.Error("EnjoyKoPayNotify GetByHashM db err %v %s, %v", isdbDebug, uperr.Error(), param)
			c.String(http.StatusInternalServerError, "fail")
			return
		}

		// log
		logiclog.LogKoPay(data.Name, uid, true, channelID, uid, hash_key, pay_time,
			amount_str, "0", "", "", game_order_org, extr_info,
			game_order, tistatus_delivered, payTime, productId, ver, fmt.Sprintf("%d:%d", pkgid, subpkgid))

		logs.Debug("EnjoyKoPayNotify order %s success", hash_key)
		c.String(http.StatusOK, "success")
	}
}
