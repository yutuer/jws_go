package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	//"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/platform/planx/servers"
	//"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RequestAvatarSkillAdd struct {
	Req
	Avatar_ids []int `codec:"aids"`
	Skill_ids  []int `codec:"s"`
	Lv_add     []int `codec:"ladd"`
}

type ResponseAvatarSkillAdd struct {
	SyncResp
}

func (p *Account) AddSkill(r servers.Request) *servers.Response {
	acid := p.AccountID.String()
	req := &RequestAvatarSkillAdd{}
	resp := &ResponseAvatarSkillAdd{}

	initReqRsp(
		"PlayerAttr/AvatarSkillAddResponse",
		r.RawBytes,
		req, resp, p)

	const (
		_                  = iota
		CODE_No_Data_Err   // 失败:数据错误
		CODE_Cost_Err      // 失败:没有物品
		CODE_Avatar_Lv_Err // 失败:角色等级不足
	)

	if len(req.Avatar_ids) != len(req.Lv_add) {
		logs.SentryLogicCritical(acid, "Avatar_ids isnt Lv_add")
		return rpcError(resp, CODE_No_Data_Err)
	}

	if len(req.Avatar_ids) != len(req.Skill_ids) {
		logs.SentryLogicCritical(acid, "Avatar_ids isnt Skill_ids")
		return rpcError(resp, CODE_No_Data_Err)
	}

	player_avatar_exp := p.Profile.GetAvatarExp()
	player_skill := p.Profile.GetAvatarSkill()

	for idx, avatar_id := range req.Avatar_ids {

		// avatar_id 必须有效
		if avatar_id < 0 || avatar_id >= helper.AVATAR_NUM_CURR {
			return rpcError(resp, CODE_No_Data_Err)
		}

		lv_add := req.Lv_add[idx]
		skill_id := req.Skill_ids[idx]

		data := gamedata.GetSkillLevelConfig(avatar_id, skill_id)

		if data == nil {
			logs.SentryLogicCritical(acid, "skillid err by %d", skill_id)
			return rpcError(resp, CODE_No_Data_Err)
		}

		avatar_lv, _ := player_avatar_exp.Get(avatar_id)
		curr_lv := int(player_skill.Get(avatar_id, skill_id))

		if curr_lv+lv_add >= len(data.CostToThisLv) {
			logs.SentryLogicCritical(acid, "curr_lv err by %d", curr_lv)
			return rpcError(resp, CODE_No_Data_Err)
		}

		avatar_lv_need := data.AvatarLvNeed[curr_lv+lv_add]
		if avatar_lv_need > avatar_lv {
			return rpcError(resp, CODE_Avatar_Lv_Err)
		}

		c := account.CostGroup{}
		for i := 0; i < lv_add; i++ {
			cost := data.CostToThisLv[curr_lv+1+i]
			if !c.AddCostData(p.Account, &cost) {
				return rpcError(resp, CODE_Cost_Err)
			}
		}

		if !c.CostBySync(p.Account, resp, "AvatarSkillAdd") {
			return rpcError(resp, CODE_Cost_Err)
		}
		p.updateCondition(account.COND_TYP_AvatarSkill,
			lv_add, 0, "", "", resp)

		logs.Trace("[%s]AvatarSkillAdd:%d,%d,%d,%d.", avatar_id, skill_id, curr_lv, lv_add)
		player_skill.AddSkill(avatar_id, skill_id, uint32(lv_add))
	}

	p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 5. 技能升级

	resp.OnChangeSC()
	resp.OnChangeHC()
	resp.OnChangeSkillAllChange()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

// 技能修炼逻辑

func (p *Account) AddSkillPractice(r servers.Request) *servers.Response {
	acid := p.AccountID.String()
	req := &struct {
		Req
		SkillIds []int `codec:"s"`
		LvAdd    []int `codec:"ladd"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/SkillAddRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_                  = iota
		CODE_No_Data_Err   // 失败:数据错误
		CODE_Cost_Err      // 失败:没有物品
		CODE_Avatar_Lv_Err // 失败:角色等级不足
	)

	if len(req.SkillIds) != len(req.LvAdd) {
		logs.SentryLogicCritical(acid, "Avatar_ids isnt Lv_add")
		return rpcError(resp, CODE_No_Data_Err)
	}

	player_skill := p.Profile.GetAvatarSkill()

	for idx, skillIdx := range req.SkillIds {

		// avatar_id 必须有效
		if skillIdx < 0 || skillIdx >= helper.CORP_SKILLPRACTICE_MAX {
			return rpcError(resp, CODE_No_Data_Err)
		}

		lv_add := req.LvAdd[idx]

		data := gamedata.GetSkillPracticeLevelInfo(skillIdx)

		if data == nil {
			logs.SentryLogicCritical(acid, "skillid err by %d", skillIdx)
			return rpcError(resp, CODE_No_Data_Err)
		}

		curr_lv := int(player_skill.GetPracticeLevel()[skillIdx])

		if curr_lv+lv_add >= len(data.CostToThisLv) {
			logs.SentryLogicCritical(acid, "curr_lv err by %d", curr_lv)
			return rpcError(resp, CODE_No_Data_Err)
		}

		c := account.CostGroup{}
		for i := 0; i < lv_add; i++ {
			cost := data.CostToThisLv[curr_lv+1+i]
			if !c.AddCostData(p.Account, &cost) {
				return rpcError(resp, CODE_Cost_Err)
			}
		}

		if !c.CostBySync(p.Account, resp, "SkillAdd") {
			return rpcError(resp, CODE_Cost_Err)
		}
		p.updateCondition(account.COND_TYP_AvatarSkill,
			lv_add, 0, "", "", resp)

		logs.Trace("[%s]SkillAdd:%d,%d,%d.", skillIdx, curr_lv, lv_add)
		for i := 0; i < lv_add; i++ {
			player_skill.AddPracticeLevel(skillIdx)
		}
		//avatarLevel, _ := p.Account.Profile.GetAvatarExp().Get(avatar_id)
		//datacollector.RoleSkillLevelUp(acid, p.Account.Profile.CorpInf.Level, avatar_id, avatarLevel, skill_id, player_skill.Get(avatar_id, skill_id))
	}

	p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 5. 技能升级

	resp.OnChangeSC()
	resp.OnChangeSkillAllChange()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}
