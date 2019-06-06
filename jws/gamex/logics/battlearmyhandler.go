package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// ShowBattleArmy : 传输战阵信息
//
func (p *Account) ShowBattleArmyHandler(req *reqMsgShowBattleArmy, resp *rspMsgShowBattleArmy) uint32 {
	//先判断是否开启了战阵
	if !account.CondCheck(gamedata.Mod_BattleArmy, p.Account) {
		return errCode.CommonConditionFalse
	}

	battle_armys_data := p.Profile.BattleArmys.GetBattleArmys()

	resp.BattleArmysInfo = make([][]byte, 0, helper.BATTLE_ARMY_NUM_MAX*helper.BATTLE_ARMYLOC_NUM_MAX)
	for i, v := range battle_armys_data {
		for j, v1 := range v.GetBattleArmyLocs() {
			logs.Trace("[cyt]战阵详细信息:%v", BattleArmyInfo{BattleArmyID: int64(i*helper.BATTLE_ARMYLOC_NUM_MAX + j + 1),
				BattleArmyAvatarID: int64(v1.AvatarID), BattleArmyLev: int64(v1.Lev)})
			resp.BattleArmysInfo = append(resp.BattleArmysInfo,
				encode(BattleArmyInfo{BattleArmyID: int64(i*helper.BATTLE_ARMYLOC_NUM_MAX + j + 1), BattleArmyAvatarID: int64(v1.AvatarID), BattleArmyLev: int64(v1.Lev)}))
		}
	}

	logs.Trace("[cyt]传输战阵信息完成,共%v条数据", len(resp.BattleArmysInfo))
	return 0
}

const nonePreBattleArmy = 999

// BattleArmyLevUp : 战阵升级／解锁协议
//
func (p *Account) BattleArmyLevUpHandler(req *reqMsgBattleArmyLevUp, resp *rspMsgBattleArmyLevUp) uint32 {
	//先判断是否开启了战阵
	if !account.CondCheck(gamedata.Mod_BattleArmy, p.Account) {
		return errCode.CommonConditionFalse
	}
	if req.BattleArmyID != nonePreBattleArmy && req.BattleArmyID > helper.BATTLE_ARMYLOC_NUM_MAX*helper.BATTLE_ARMY_NUM_MAX {
		//无效的战阵ID
		return errCode.CommonInvalidParam
	}
	battle_army := gamedata.GetBattleArmy(int(req.BattleArmyID))
	if battle_army == nil || req.BattleArmyID < 1 {
		//无效的战阵ID
		return errCode.CommonInvalidParam
	}
	resp.BattleArmyID = req.BattleArmyID
	//服务器中的战阵信息
	battle_armys_data := p.Profile.BattleArmys.GetBattleArmys()
	//显示的战阵等级
	battle_index := req.BattleArmyID - 1
	currLev := req.CurrBattleArmyLevel
	battleLoc := &battle_armys_data[battle_index/helper.BATTLE_ARMYLOC_NUM_MAX].
		GetBattleArmyLocs()[battle_index%helper.BATTLE_ARMYLOC_NUM_MAX]
	battle_lev := battleLoc.Lev
	if battle_lev != int(currLev) {
		//服务器存档与客户端传输的等级不一致，直接返还服务器等级数据给客户端。
		resp.CurrBattleArmyLevel = req.CurrBattleArmyLevel
		return 0
	}
	logs.Trace("[cyt]BattleArmyLevUpAndUnlock：battle_index:%v CurrBattleArmyLevel:%v", battle_index, req.CurrBattleArmyLevel)
	if battle_lev == 0 {
		//尚未解锁,需要判断前置将位是否解锁
		pre_armyID := battle_army.GetPreArmyID() - 1
		if pre_armyID < helper.BATTLE_ARMYLOC_NUM_MAX*uint32(len(battle_armys_data)) {
			pre_battleLoc := battle_armys_data[pre_armyID/helper.BATTLE_ARMYLOC_NUM_MAX].
				GetBattleArmyLocs()[pre_armyID%helper.BATTLE_ARMYLOC_NUM_MAX]
			if pre_battleLoc.Lev == 0 {
				//前置将位未解锁,不满足解锁条件
				return errCode.CommonConditionFalse
			}
		}
	}
	battle_army_lev := gamedata.GetBattleArmyLevel(battle_army.GetBattleArmyLoc(), uint32(battle_lev+1))
	if battle_army_lev == nil {
		//未查询到battle_lev+1等级，说明已经满级
		return errCode.CommonMaxLimit
	}
	if p.Profile.GetCorp().Level < battle_army_lev.GetBattleArmyLevelLimit() {
		//未达到升级条件
		return errCode.CommonConditionFalse
	}

	costData := &gamedata.CostData{}
	for _, data := range battle_army_lev.BattleArmyUp_Template {
		costData.AddItem(data.GetBattleArmyCoin(), data.GetBattleArmyCost())
	}

	if !account.CostBySync(p.Account, costData, resp, "BattleArmyLevUp") {
		//钱不够
		return errCode.CommonLessMoney
	}

	//升级／解锁成功
	battleLoc.Lev++
	resp.CurrBattleArmyLevel = int64(battleLoc.Lev)
	p.Profile.GetData().SetNeedCheckMaxGS()
	resp.OnChangeBattleArmyInfo()

	logs.Trace("[cyt]战阵升级/解锁成功,等级为：%v", p.Profile.BattleArmys.GetBattleArmys()[battle_index/helper.BATTLE_ARMYLOC_NUM_MAX].
		GetBattleArmyLocs()[battle_index%helper.BATTLE_ARMYLOC_NUM_MAX].Lev)
	return 0
}

const minValidAvatarID = -1
const minValidBattleArmyID = 1

// BattleArmyChoiceAvatarID : 选择武将
//
func (p *Account) BattleArmyChoiceAvatarIDHandler(req *reqMsgBattleArmyChoiceAvatarID, resp *rspMsgBattleArmyChoiceAvatarID) uint32 {
	//先判断是否开启了战阵
	if !account.CondCheck(gamedata.Mod_BattleArmy, p.Account) {
		return errCode.CommonConditionFalse
	}
	if req.BattleArmyID != nonePreBattleArmy && req.BattleArmyID > helper.BATTLE_ARMYLOC_NUM_MAX*helper.BATTLE_ARMY_NUM_MAX {
		//无效的战阵ID
		return errCode.CommonInvalidParam
	}
	battle_army := gamedata.GetBattleArmy(int(req.BattleArmyID))
	if battle_army == nil || req.BattleArmyID < minValidBattleArmyID {
		//无效的战阵ID
		return errCode.CommonInvalidParam
	}
	if req.CurrBattleArmyAvatarID < minValidAvatarID || req.CurrBattleArmyAvatarID >= helper.AVATAR_NUM_CURR {
		//无效的avatarID
		return errCode.CommonInvalidParam
	}

	logs.Trace("[cyt]收到的战阵ID:%v,武将ID：%v", req.BattleArmyID, req.CurrBattleArmyAvatarID)

	resp.BattleArmyID = req.BattleArmyID

	battle_index := req.BattleArmyID - 1

	targetLoc := &p.Profile.BattleArmys.GetBattleArmys()[battle_index/helper.BATTLE_ARMYLOC_NUM_MAX].
		GetBattleArmyLocs()[battle_index%helper.BATTLE_ARMYLOC_NUM_MAX]

	if targetLoc.AvatarID != int(req.CurrBattleArmyAvatarID) {
		for _, v := range p.Profile.BattleArmys.GetBattleArmys() {
			for _, v1 := range v.GetBattleArmyLocs() {
				if v1.AvatarID != minValidAvatarID && v1.AvatarID == int(req.CurrBattleArmyAvatarID) {
					//此武将已经选择在其他战阵中，无效的avatarID
					return errCode.CommonInvalidParam
				}
			}
		}
	}
	targetLoc.AvatarID = int(req.CurrBattleArmyAvatarID)
	resp.CurrBattleArmyAvatarID = req.CurrBattleArmyAvatarID

	p.Profile.GetData().SetNeedCheckMaxGS()
	resp.OnChangeBattleArmyInfo()

	logs.Trace("[cyt]向客户端传输的AvartarID：%v", resp.CurrBattleArmyAvatarID)
	logs.Trace("[cyt]更换武将成功,更换后的武将ID为:%v", p.Profile.BattleArmys.GetBattleArmys()[battle_index/helper.BATTLE_ARMYLOC_NUM_MAX].
		GetBattleArmyLocs()[battle_index%helper.BATTLE_ARMYLOC_NUM_MAX].AvatarID)
	return 0
}
