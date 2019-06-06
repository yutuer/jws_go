package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RequestGetMails struct {
	Req
}

type ResponseGetMails struct {
	SyncResp
}

type mailToClient struct {
	Idx       int64    `codec:"idx"`
	IdsId     uint32   `codec:"ids"`
	Param     []string `codec:"param"`
	Tag       string   `codec:"tag"`
	IsRead    bool     `codec:"read"`
	ItemId    []string `codec:"id"`
	Count     []uint32 `codec:"c"`
	TimeBegin int64    `codec:"beg"`
	TimeEnd   int64    `codec:"end"`
}

func newMailToClient(m mailToClient) mailToClient {
	var hc_count uint32
	ItemId := make([]string, 0, len(m.ItemId))
	Count := make([]uint32, 0, len(m.Count))
	for idx, item_id := range m.ItemId {
		if item_id == gamedata.VI_Hc ||
			item_id == gamedata.VI_Hc_Buy ||
			item_id == gamedata.VI_Hc_Compensate ||
			item_id == gamedata.VI_Hc_Give {
			hc_count += m.Count[idx]
		} else {
			ItemId = append(ItemId, item_id)
			Count = append(Count, m.Count[idx])
		}
	}
	if hc_count > 0 {
		ItemId = append(ItemId, gamedata.VI_Hc)
		Count = append(Count, hc_count)
	}

	m.ItemId = ItemId
	m.Count = Count
	return m
}

// 1、第一次进主城；
// 2、支付的时候
// 3、主动打开邮箱的时候
// 会发此协议
func (p *Account) GetMails(r servers.Request) *servers.Response {
	req := &RequestGetMails{}
	resp := &ResponseGetMails{}

	initReqRsp(
		"PlayerAttr/GetMailsResponse",
		r.RawBytes,
		req, resp, p)

	player_mail := p.Profile.GetMails()
	acID := p.AccountID.String()
	err := player_mail.LoadMail(p.AccountID.ServerString(), acID, p.Profile.CreateTime)
	if err != nil {
		logs.SentryLogicCritical(acID,
			"LoadMailError by %s", err.Error())
		return rpcError(resp, 1)
	}

	player_mail.CheckErrorMail()

	sync_err := player_mail.SyncMailToDynamo()
	if sync_err != nil {
		return rpcError(resp, 3)
	}

	resp.OnChangeMail()
	resp.OnChangeUnlockAvatar() // 客户端目前会先请求这个包, 需要这条属性判断
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

type RequestReceiveMails struct {
	Req
	Idx []int64 `codec:"idxs"`
}

type ResponseReceiveMails struct {
	SyncRespWithRewards
	Money []uint32 `codec:"money"`
}

func (p *Account) ReceiveMails(r servers.Request) *servers.Response {
	const (
		_                 = iota
		CODE_IDX_Err      // 失败:请求的邮件不存在
		CODE_Sync_Err     // 失败:ID对应的数据不存在
		CODE_Reward_Err   // 失败:奖励错误
		CODE_Bag_Full_Err // 失败:包裹满
		CODE_Give_Fail
	)

	req := &RequestReceiveMails{}
	resp := &ResponseReceiveMails{}

	initReqRsp(
		"PlayerAttr/ReceiveMailsResponse",
		r.RawBytes,
		req, resp, p)

	player_mail := p.Profile.GetMails()
	acID := p.AccountID.String()
	err := player_mail.LoadMail(p.AccountID.ServerString(), acID, p.Profile.CreateTime)
	if err != nil {
		logs.SentryLogicCritical(acID,
			"LoadMailError by %s", err.Error())
		return rpcError(resp, 1)
	}

	type mailReward struct {
		reward *gamedata.CostData
		reason string
	}
	hasEquip := false
	hasJade := false
	costs := make([]mailReward, 0, len(req.Idx))
	resp.Money = make([]uint32, 0, len(req.Idx))
	var Channel string
	for i := 0; i < len(req.Idx); i++ {
		cost, money, reason, ver, channel := player_mail.GetMailReward(req.Idx[i])
		if ver != "" { //版本好检查，这是一个指定了版本号的邮件
			if ver != p.Profile.ClientInfo.MarketVer {
				return rpcWarn(resp, errCode.NotSupportedThisVersionPlsUpgrade)
			}
		}
		Channel = channel
		logs.Trace("mail %v", player_mail)
		if cost == nil {
			return rpcWarn(resp, errCode.MailReceiveMailIDErr)
		}
		// receive
		err := player_mail.ReceiveMail(req.Idx[i])
		if err != nil {
			logs.SentryLogicCritical(p.AccountID.String(),
				"ReceiveMailErr %s", err.Error())
			continue
		}
		costs = append(costs, mailReward{cost, reason})
		if cost.HasEquip {
			hasEquip = true
		}
		if cost.HasJade {
			hasJade = true
		}
		if money > 0 {
			resp.Money = append(resp.Money, money)
		}
	}
	if hasEquip {
		// 检查装备物品数量
		if p.BagProfile.GetEquipCount() >= gamedata.GetEquipCountUpLimit() {
			logs.SentryLogicCritical(p.AccountID.String(), "ReceiveMail err %d, %s", CODE_Bag_Full_Err, "Bag_Full for Equip")
			return rpcWarn(resp, CODE_Bag_Full_Err)
		}
	}
	if hasJade {
		if p.Profile.GetJadeBag().GetJadeSumCount() >= gamedata.GetJadeCountUpLimit() {
			logs.SentryLogicCritical(p.AccountID.String(), "ReceiveMail err %d, %s", CODE_Bag_Full_Err, "Bag_Full for Jade")
			return rpcWarn(resp, CODE_Bag_Full_Err)
		}
	}

	sync_err := player_mail.SyncMailToDynamo()
	if sync_err != nil {
		return rpcError(resp, CODE_Sync_Err)
	}

	for _, r := range costs {
		// IAP特殊处理
		if r.reward.IAPGoodIndex >= 0 {
			if Channel == "gm" {
				r.reason = "IAPByGM"
			} else {
				r.reason = "RmbPay"
			}
		}

		g := account.GiveGroup{}
		g.AddCostData(r.reward)
		if !g.GiveBySyncAuto(p.Account, resp, r.reason) {
			return rpcErrorWithMsg(resp, CODE_Give_Fail, "ReceiveMail give fail")
		}
	}

	resp.OnChangeMail()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

func (p *Account) ReadMails(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Idx []int64 `codec:"idxs"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	const (
		_             = iota
		CODE_Sync_Err // 失败:ID对应的数据不存在
	)

	initReqRsp(
		"PlayerAttr/ReadMailsResponse",
		r.RawBytes,
		req, resp, p)

	player_mail := p.Profile.GetMails()
	acID := p.AccountID.String()
	err := player_mail.LoadMail(p.AccountID.ServerString(), acID, p.Profile.CreateTime)
	if err != nil {
		logs.SentryLogicCritical(acID,
			"LoadMailError by %s", err.Error())
		return rpcError(resp, 1)
	}

	for i := 0; i < len(req.Idx); i++ {
		player_mail.ReadMail(req.Idx[i])
	}

	sync_err := player_mail.SyncMailToDynamo()
	if sync_err != nil {
		return rpcError(resp, CODE_Sync_Err)
	}

	resp.OnChangeMail()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}
