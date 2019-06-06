package pay

import (
	"net/http"

	"encoding/json"

	"fmt"
	"strings"

	"encoding/base64"

	"github.com/gin-gonic/gin"
	"strconv"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
	"vcs.taiyouxi.net/platform/x/api_gateway/logiclog"
	"vcs.taiyouxi.net/platform/x/api_gateway/util"
)

const (
	SIGN      = "sign"
	TRANSDATA = "transdata"
)

func SamsungInit(data sdkPayData) {
	util.Init(data.Sdk[typ_samsung].Params[0], data.Sdk[typ_samsung].Params[1])
	logs.Debug("SamsungInit success")
}

func SamsungPayNotify(typ string, data sdkPayData) gin.HandlerFunc {
	return func(c *gin.Context) {
		resS := "SUCCESS"
		resF := "FAILURE"

		transData := c.PostForm(TRANSDATA)
		sign := c.PostForm(SIGN)

		signB, err := base64.StdEncoding.DecodeString(sign)
		if err != nil {
			logs.Error("SamsungPayNotify Decode Sign err %v transData %s sign %s",
				err, transData, sign)
			c.String(http.StatusBadRequest, resF)
			return
		}
		if err := util.Verify([]byte(transData), signB); err != nil {
			logs.Error("SamsungPayNotify Verify Sign err %v transData %s sign %s",
				err, transData, sign)
			c.String(http.StatusBadRequest, resF)
			return
		}

		logs.Debug("SamsungPayNotify transData %s sign %s", transData, sign)

		trdata := &trans_data{}
		if err := json.Unmarshal([]byte(transData), trdata); err != nil {
			logs.Error("SamsungPayNotify transData json.Unmarshal err %v transData %s sign %s",
				err, transData, sign)
			c.String(http.StatusBadRequest, resF)
			return
		}

		logs.Debug("SamsungPayNotify transData %v", trdata)

		// 解析参数
		game_order_org := trdata.Cporderid
		hash_key := trdata.Transid
		amount := int(trdata.Money)
		extr_info := trdata.Cpprivate
		// game_order_org
		cpInfo := genCPInfo(typ, game_order_org)
		if cpInfo == nil {
			logs.Error("SamsungPayNotify genCPInfo err transData %v", trdata)
			c.String(http.StatusBadRequest, resF)
			return
		}
		uid := cpInfo.uid
		isdbDebug := cpInfo.isDBDebug
		game_order := cpInfo.game_order
		// game_order_org
		ex := strings.Split(extr_info, ":")
		if len(ex) < 3 {
			logs.Error("SamsungPayNotify Extras_params err %s %v, %v",
				game_order, amount, trdata)
			c.String(http.StatusInternalServerError, resF)
			return
		}
		payTime := ex[0]
		productId := ex[1]
		ver := ex[2]
		pkgid := 0
		subpkgid := 0

		cfg := gamedata.GetIAPBaseConfigAndroid(game_order)
		if cfg == nil {
			logs.Error("samsung Can Not Find goods_id:%s", game_order)
			c.String(http.StatusInternalServerError, "fail")
			return
		}
		if len(ex) > 3 {

			pkgid, err = strconv.Atoi(ex[3])
			if err != nil {
				logs.Error("[%s] packageid convert to int err %s", typ, err.Error())
			}
			logs.Debug("SamsungNotify")
			subpkgid, err = strconv.Atoi(ex[4])
			if err != nil {
				logs.Error("[%s] subpackageid convert to int err %s", typ, err.Error())
			}
			logs.Debug("SamsungNotify: PackageIAP %d:%d", pkgid, subpkgid)
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
		amount_str := fmt.Sprintf("%d", amount)
		tradeStatus := fmt.Sprintf("%d", trdata.Result)

		if trdata.Result != 0 { // 失败
			// log
			logiclog.LogPay(data.Name, uid, false, util.SamsungChannel, trdata.Appuserid, hash_key, trdata.Transtime,
				amount_str, tradeStatus, "", "", game_order_org, extr_info,
				game_order, tistatus_delivered, payTime, productId, ver,fmt.Sprintf("%d:%d", pkgid, subpkgid))
			logs.Debug("SamsungPayNotify order %s fail", hash_key)
			c.String(http.StatusOK, resS)
			return
		}

		// 判断重复
		if res, resMsg, isRepeat := checkOrderNoRepeat(typ, data, isdbDebug, hash_key); !res {
			logs.Error("SamsungPayNotify checkOrderNoRepeat err %s %v", resMsg, trdata)
			c.String(http.StatusInternalServerError, resF)
			return
		} else {
			if isRepeat {
				logs.Warn("SamsungPayNotify order_no repeat %v", trdata)
				c.String(http.StatusOK, resS)
				return
			}
		}

		// 比较钱和订单需要的钱是否一致
		if res, msg := checkMoneyWithCfg(typ, data, game_order, uint32(amount), uid); !res {
			logs.Error("[%s] PayNotify checkMoneyWithCfg err %v %s, %v", typ, game_order, msg, trdata)
			c.String(http.StatusInternalServerError, resF)
			return
		}

		// 存入dynamo
		value := make(map[string]interface{}, 11)
		value[key_channel] = util.SamsungChannel  // 渠道标示ID
		value[key_channel_uid] = trdata.Appuserid // 渠道用户唯一标示
		value[key_good_idx] = game_order_org      // 游戏在调用QucikSDK发起支付时传递的游戏方订单,这里会原样传回
		value[key_order_no] = hash_key            // 唯一订单号
		value[key_pay_time] = trdata.Transtime    // 支付时间
		value[key_money_amount] = amount_str      // 成交金额
		value[key_status] = tradeStatus           // 充值状态 0 成功 1失败(为1时 应返回FAILUD失败)
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
			Channel:       util.SamsungChannel,
			PayTime:       payTime,
			PkgInfo:       timail.PackageInfo{pkgid, subpkgid}}
		info, _ := json.Marshal(&order)

		if err := data.SendMail(isdbDebug, uid, string(info)); err != nil {
			logs.Error("SamsungPayNotify SetByHashM db err %v %s, %v", isdbDebug, err.Error(), trdata)
			c.String(http.StatusInternalServerError, resF)
			return
		}
		uperr := data.PayUpdateByHash(isdbDebug, hash_key,
			map[string]interface{}{
				key_tistatus: tistatus_delivered,
			})
		if uperr != nil {
			logs.Error("SamsungPayNotify GetByHashM db err %v %s, %v", isdbDebug, uperr.Error(), trdata)
			c.String(http.StatusInternalServerError, resF)
			return
		}

		// log
		logiclog.LogPay(data.Name, uid, true, util.SamsungChannel, trdata.Appuserid, hash_key, trdata.Transtime,
			amount_str, tradeStatus, "", "", game_order_org, extr_info,
			game_order, tistatus_delivered, payTime, productId, ver,fmt.Sprintf("%d:%d", pkgid, subpkgid))

		logs.Debug("SamsungPayNotify order %s success", hash_key)
		c.String(http.StatusOK, resS)
	}
}

type trans_data struct {
	Transtype int     `json:"transtype"`           // 交易类型
	Cporderid string  `json:"cporderid,omitempty"` // 商户订单号
	Transid   string  `json:"transid"`             // 交易流水号
	Appuserid string  `json:"appuserid"`           // 用户在商户应用的唯一标识
	Appid     string  `json:"appid"`               // 游戏id
	Waresid   int     `json:"waresid"`             // 商品编码
	Feetype   int     `json:"feetype"`             // 计费方式
	Money     float32 `json:"money"`               // 交易金额
	Currency  string  `json:"currency"`            // 货币类型
	Result    int     `json:"result"`              // 交易结果
	Transtime string  `json:"transtime"`           // 交易完成时间
	Cpprivate string  `json:"cpprivate,omitempty"` // 商户私有信息
	Paytype   int     `json:"paytype,omitempty"`   // 支付方式
}
