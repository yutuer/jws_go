package pay

import (
	"net/http"

	"strconv"

	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/x/api_gateway/logiclog"
	"vcs.taiyouxi.net/platform/x/api_gateway/util"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

func VivoPayNotify(typ string, data sdkPayData) gin.HandlerFunc {
	return func(c *gin.Context) {
		resS := "success"
		resF := "fail"
		cpKey := data.Sdk[typ].Params[0]

		param := make(util.Para, 16)

		respCode := c.PostForm("respCode")
		if respCode != "200" {
			logs.Error("VivoPayNotify respCode err %s", respCode)
			c.String(http.StatusBadRequest, resF)
			return
		}
		param["respCode"] = respCode
		param["respMsg"] = c.PostForm("respMsg")
		param["signMethod"] = c.PostForm("signMethod")
		param["signature"] = c.PostForm("signature")
		param["tradeType"] = c.PostForm("tradeType")
		tradeStatus := c.PostForm("tradeStatus")
		param["tradeStatus"] = c.PostForm("tradeStatus")
		param["cpId"] = c.PostForm("cpId") // 定长20位数字，由vivo分发的唯一识别码
		param["appId"] = c.PostForm("appId")
		param["uid"] = c.PostForm("uid")                     // 用户在vivo这边的唯一标识
		param["cpOrderNumber"] = c.PostForm("cpOrderNumber") // 商户自定义的订单号, 商户自定义，最长 64 位字母、数字和下划线组成
		param["orderNumber"] = c.PostForm("orderNumber")     // 交易流水号,vivo订单号
		param["orderAmount"] = c.PostForm("orderAmount")     // 交易金额,单位：分，币种：人民币，为长整型，如：101，10000
		param["extInfo"] = c.PostForm("extInfo")             // 商户透传参数
		param["payTime"] = c.PostForm("payTime")             // 交易时间	yyyyMMddHHmmss

		// check signature
		verify := util.VerifySignature(param, cpKey)
		if !verify {
			logs.Error("VivoPayNotify VerifySignature err %s", param)
			c.String(http.StatusBadRequest, resF)
			return
		}
		logs.Debug("VivoPayNotify %v", param)

		// 参数
		game_order_org := param["cpOrderNumber"]
		hash_key := param["orderNumber"]
		amount_str := param["orderAmount"]
		extr_info := param["extInfo"]

		// 解析game order
		cpInfo := genCPInfo(typ, game_order_org)
		if cpInfo == nil {
			logs.Error("VivoPayNotify genCPInfo err Data %v", param)
			c.String(http.StatusBadRequest, resF)
			return
		}
		uid := cpInfo.uid
		isdbDebug := cpInfo.isDBDebug
		game_order := cpInfo.game_order

		amount, err := strconv.Atoi(amount_str)
		if err != nil {
			logs.Error("[%s] PayNotify Amount err %v %s %v", typ, game_order, amount_str, param)
			c.String(http.StatusInternalServerError, resF)
		}
		amount = amount / 100
		// extr_info
		ex := strings.Split(extr_info, ":")
		if len(ex) < 3 {
			logs.Error("VivoPayNotify Extras_params err %s %v, %v",
				game_order, amount, param)
			c.String(http.StatusInternalServerError, resF)
			return
		}
		payTime := ex[0]
		productId := ex[1]
		ver := ex[2]
		amount_str = fmt.Sprintf("%d", amount)
		pkgid := 0
		subpkgid := 0
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

		cfg := gamedata.GetIAPBaseConfigAndroid(game_order)
		if cfg == nil {
			logs.Error("vivo Can Not Find goods_id:%s", game_order)
			c.String(http.StatusInternalServerError, "fail")
			return
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
		if tradeStatus != "0000" {
			// log
			logiclog.LogPay(data.Name, uid, false, util.VivoChannel, param["uid"], hash_key, param["payTime"],
				amount_str, tradeStatus, "", "", game_order_org, extr_info,
				game_order, tistatus_delivered, payTime, productId, ver,fmt.Sprintf("%d:%d", pkgid, subpkgid))
			logs.Debug("VivoPayNotify order %s fail", hash_key)
			c.String(http.StatusOK, resS)
			return
		}

		// 判断重复
		if res, resMsg, isRepeat := checkOrderNoRepeat(typ, data, isdbDebug, hash_key); !res {
			logs.Error("VivoPayNotify checkOrderNoRepeat err %s %v", resMsg, param)
			c.String(http.StatusInternalServerError, resF)
			return
		} else {
			if isRepeat {
				logs.Warn("VivoPayNotify order_no repeat %v", param)
				c.String(http.StatusOK, resS)
				return
			}
		}

		// 比较钱和订单需要的钱是否一致
		if res, msg := checkMoneyWithCfg(typ, data, game_order, uint32(amount), uid); !res {
			logs.Error("[%s] PayNotify checkMoneyWithCfg err %v %s, %v", typ, game_order, msg, param)
			c.String(http.StatusInternalServerError, resF)
			return
		}

		// 存入dynamo
		value := make(map[string]interface{}, 11)
		value[key_channel] = util.VivoChannel  // 渠道标示ID
		value[key_channel_uid] = param["uid"]  // 渠道用户唯一标示
		value[key_good_idx] = game_order_org   // 游戏在调用QucikSDK发起支付时传递的游戏方订单,这里会原样传回
		value[key_order_no] = hash_key         // 唯一订单号
		value[key_pay_time] = param["payTime"] // 支付时间
		value[key_money_amount] = amount_str   // 成交金额
		value[key_status] = tradeStatus        // 充值状态 0 成功 1失败(为1时 应返回FAILUD失败)
		value[key_product] = productId
		value[key_ver] = ver
		value[key_extras_params] = extr_info
		value[key_tistatus] = tistatus_paid // 通知状态，paid支付成功；delivered支付成功并玩家已拿到
		value[key_uid] = uid                // accountid
		value[key_mobile] = data.Name
		// TODO by zhangzhen 测试订单 怎么识别

		if err := data.PaySetByHashM(isdbDebug, hash_key, value); err != nil {
			mapB, _ := json.Marshal(value)
			logs.Warn("QuickPayNotify SetByHashM db err %v %s, json:%q", isdbDebug, err.Error(), mapB)
			c.String(http.StatusInternalServerError, "FAILED")
			return
		}

		// send mail to dynamodb
		order := timail.IAPOrder{
			Order_no:      hash_key,
			Game_order:    game_order,
			Game_order_id: game_order_org,
			Amount:        amount_str,
			Channel:       util.VivoChannel,
			PayTime:       payTime,
			PkgInfo:       timail.PackageInfo{pkgid, subpkgid}}
		info, _ := json.Marshal(&order)

		if err := data.SendMail(isdbDebug, uid, string(info)); err != nil {
			logs.Error("VivoPayNotify SetByHashM db err %v %s, %v", isdbDebug, err.Error(), param)
			c.String(http.StatusInternalServerError, resF)
			return
		}
		uperr := data.PayUpdateByHash(isdbDebug, hash_key,
			map[string]interface{}{
				key_tistatus: tistatus_delivered,
			})
		if uperr != nil {
			logs.Error("VivoPayNotify GetByHashM db err %v %s, %v", isdbDebug, uperr.Error(), param)
			c.String(http.StatusInternalServerError, resF)
			return
		}

		// log
		logiclog.LogPay(data.Name, uid, true, util.VivoChannel, param["uid"], hash_key, param["payTime"],
			amount_str, tradeStatus, "", "", game_order_org, extr_info,
			game_order, tistatus_delivered, payTime, productId, ver,fmt.Sprintf("%d:%d", pkgid, subpkgid))

		logs.Debug("VivoPayNotify order %s success", hash_key)
		c.String(http.StatusOK, resS)
	}
}
