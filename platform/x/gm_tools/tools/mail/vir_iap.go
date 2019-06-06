package mail

import (
	//"encoding/json"
	"errors"
	"strconv"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	//"vcs.taiyouxi.net/platform/planx/util/logs"
	//"vcs.taiyouxi.net/platform/x/gm_tools/util"
	"encoding/json"
	//"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/data"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

func VirtualIAP(c *gm_command.Context, server, accountid string, params []string) error {
	mail_data := &timail.MailReward{}

	hc_add, err := strconv.Atoi(params[0])
	if err != nil {
		return err
	}

	if hc_add < 0 {
		return errors.New("hc add < 0")
	}

	mail_data.AddReward(helper.VI_Hc_Buy, uint32(hc_add))
	mail_data.TimeBegin = time.Now().Unix()
	mail_data.TimeEnd = time.Now().Unix() + util.WeekSec*8
	err = SendMail(server, "profile:"+accountid, mail_data)
	if err != nil {
		return err
	} else {
		c.SetData("{\"res\":\"ok\"}")
		return nil
	}
}

func GetVirtualIAP(c *gm_command.Context, server, accountid string, params []string) error {
	res, err := GetAllMail(server, "profile:"+accountid)
	if err != nil {
		return err
	} else {
		c.SetData(res)
		return nil
	}
}

func DelVirtualIAP(c *gm_command.Context, server, accountid string, params []string) error {
	id_to_del, err := strconv.ParseInt(params[0], 10, 64)
	if err != nil {
		return err
	}

	err = DelMail(server, "profile:"+accountid, id_to_del)
	if err != nil {
		return err
	} else {
		c.SetData("{\"res\":\"ok\"}")
		return nil
	}
}

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
const ParamsCproductid = 0
const ParamsAmount = 1
const payAndriod = "0"

type IAPOrder struct {
	Order_no      string `json:"on"`
	Game_order    string `json:"go"`
	Game_order_id string `json:"goid"`
	Amount        string `json:"a"`
	Channel       string `json:"ch"`
	PayTime       string `json:"t"` // 客户端带过来的支付时间
}

type Mail2User struct {
	Mail timail.MailReward
	Typ  int64
	Uid  string
}

func VirtualTrueIAP(c *gm_command.Context, server, accountid string, params []string) error {
	//mail_data := &timail.MailReward{}
	logs.Debug("### I'm in params %v", params)
	hc_add, err := strconv.Atoi(params[0])
	//goodId := params[ParamsCproductid]
	if err != nil {
		return err
	}

	if hc_add < 0 {
		return errors.New("hc add < 0")
	}
	idx := accountid + time.Now().Local().String()

	logs.Debug("### key_order_no %d", idx)
	// 存入dynamo
	value := make(map[string]interface{}, 11)

	value[key_channel] = "gm"                                      // 渠道标示ID
	value[key_channel_uid] = "gm"                                  // 渠道用户唯一标示,该值从客户端GetUserId()中可获取
	value[key_good_idx] = params[0]                                // 游戏在调用QucikSDK发起支付时传递的游戏方订单,这里会原样传回
	value[key_order_no] = idx                                      // 天象唯一订单号
	value[key_pay_time] = time.Now().Format("2006-01-02 15:04:05") // 支付时间
	value[key_money_amount] = params[1]                            // 成交金额
	value[key_status] = "0"                                        // 充值状态 0 成功 1失败(为1时 应返回FAILUD失败)
	value[key_product] = params[0]
	value[key_ver] = etcd.KeyShardVersion
	value[key_extras_params] = idx + etcd.KeyActValid
	value[key_tistatus] = tistatus_paid // 通知状态，paid支付成功；delivered支付成功并玩家已拿到
	value[key_uid] = accountid          // accountid
	value[key_mobile] = "gm"
	value[key_IsTest] = "0"

	if params[2] == payAndriod {
		SetVirTrueIapPayAndroid(value)
	} else {
		SetVirTrueIapPayIOS(value)
	}

	// send mail to dynamodb
	order := IAPOrder{
		idx,
		params[0],
		accountid + ":gm" + time.Now().String() + ":101",
		params[1],
		"gm",
		string(time.Now().Unix()),
	}
	info, _ := json.Marshal(&order)
	switch params[2] {
	case "0":
		m := Mail2User{
			Typ: timail.Mail_Send_By_AndroidIAP,
			Uid: accountid,
		}

		m.Mail.TimeBegin = time.Now().Unix()
		m.Mail.TimeEnd = time.Now().Unix() + util.WeekSec*8
		m.Mail.Tag = string(info)

		err = SendMailVirtualIap(server, &m, 0)
	case "1":

		m := Mail2User{
			Typ: timail.Mail_Send_By_IOSIAP,
			Uid: accountid,
		}

		m.Mail.TimeBegin = time.Now().Unix()
		m.Mail.TimeEnd = time.Now().Unix() + util.WeekSec*8
		m.Mail.Tag = string(info)

		err = SendMailVirtualIap(server, &m, 0)
	}
	//	err = SendMail(server, "profile:"+accountid, mail_data)
	if err != nil {
		return err
	} else {
		c.SetData("{\"res\":\"ok\"}")
		return nil
	}
}
