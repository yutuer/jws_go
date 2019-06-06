package logics

import (

	//"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/platform/planx/servers"
	//"vcs.taiyouxi.net/platform/planx/servers/game"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RequestGetAvatarExp struct {
	Req
}

type ResponseGetAvatarExp struct {
	Resp
	Exps []uint32 `codec:"exps"`
}

func (p *Account) GetAvatarExp(r servers.Request) *servers.Response {
	req := &RequestGetAvatarExp{}
	resp := &ResponseGetAvatarExp{}

	initReqRsp(
		"PlayerAttr/GetAvatarExpResponse",
		r.RawBytes,
		req, resp, p)

	resp.Exps = p.Profile.GetAvatarExp().GetAll()

	//logs.Trace("exps %v", resp.Exps)
	//logs.Trace("resp %v", resp)

	return rpcSuccess(resp)
}

type RequestAvatarArousalAdd struct {
	Req
	Avatar_ids []int `codec:"aids"`
}

type ResponseAvatarArousalAdd struct {
	SyncResp
}

func (p *Account) AvatarArousalAdd(r servers.Request) *servers.Response {
	req := &RequestAvatarArousalAdd{}
	resp := &ResponseAvatarArousalAdd{}

	initReqRsp(
		"PlayerAttr/AvatarArousalAddResponse",
		r.RawBytes,
		req, resp, p)

	const (
		_                  = iota
		CODE_No_Data_Err   // 失败:数据错误
		CODE_Cost_Err      // 失败:没有物品
		CODE_Avatar_Lv_Err // 失败:角色等级不足
		CODE_Avatar_Lock   // 失败：角色没有解锁
	)
	/*
		for _, avatar_id := range req.Avatar_ids {

			// avatar_id 必须有效
			if avatar_id < 0 || avatar_id >= helper.AVATAR_NUM_CURR {
				return rpcError(resp, CODE_No_Data_Err)
			}

			// 角色是否解锁
			if !p.Account.IsAvatarUnblock(avatar_id) {
				return rpcErrorWithMsg(resp, CODE_Avatar_Lock, fmt.Sprintf("CODE_Avatar_Lock avatar %d", avatar_id))
			}

			//
			data := gamedata.GetAvatarArousalData(avatar_id)

			acid := p.AccountID.String()

			if data == nil {
				logs.SentryLogicCritical(acid, "AvatarArousalAdd err no data")
				return rpcError(resp, CODE_No_Data_Err)
			}

			corp_lv, _ := p.Profile.GetCorp().GetXpInfo()
			player_avatar_exp := p.Profile.GetAvatarExp()
			curr_lv := player_avatar_exp.GetArousalLv(avatar_id)

			if int(curr_lv+1) >= len(data.CostToThisLevel) {
				logs.SentryLogicCritical(acid, "curr_lv err by %d", curr_lv)
				return rpcError(resp, CODE_No_Data_Err)
			}

			avatar_lv_need := data.AvatarLvNeedByThisLevel[curr_lv+1]
			cost := data.CostToThisLevel[curr_lv+1]

			if avatar_lv_need > corp_lv {
				return rpcError(resp, CODE_Avatar_Lv_Err)
			}

			c := account.CostGroup{}
			if !c.AddCostData(p.Account, &cost) {
				return rpcError(resp, CODE_Cost_Err)
			}

			if !c.CostBySync(p.Account, resp, "AvatarArousalAdd") {
				return rpcError(resp, CODE_Cost_Err)
			}
			//logs.SentryLogicCritical(acid, "AddArousalLv:%d,%d.", avatar_id, curr_lv)
			player_avatar_exp.AddArousalLv(avatar_id)
			p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 2. 角色突破
		}

		resp.OnChangeAvatarArousal()
	*/
	// 服务器屏蔽觉醒内容
	logs.SentryLogicCritical(p.AccountID.String(), "AvatarArousal Had Ban!")
	return rpcSuccess(resp)
}
