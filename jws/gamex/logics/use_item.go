package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type RequestUseItem struct {
	Req
	ItemIdx  []uint32 `codec:"idx"`
	Count    []int    `codec:"n"`
	AvatarId int      `codec:"aid"`
}

type ResponseUseItem struct {
	SyncRespWithRewards
}

func (p *Account) UseItem(r servers.Request) *servers.Response {
	acid := p.AccountID.String()
	req := &RequestUseItem{}
	resp := &ResponseUseItem{}

	initReqRsp(
		"PlayerAttr/UseItemResp",
		r.RawBytes,
		req, resp, p)

	const (
		_                         = iota
		Err_Item_Cfg              // 物品配置不存在
		Err_Item_Type             // 物品类型错误
		Err_Item_Not_Enough       // 物品数量不足
		Err_Item_Part_Not_Support // 物品part属性错误
		Err_Give                  // give失败
		Err_Param_AVATAR          // AVATAR参数错误
	)

	// 作弊检查
	for _, n := range req.Count {
		if n < 0 || n > uutil.CHEAT_INT_MAX {
			return rpcErrorWithMsg(resp, 99, "UseItem Count cheat")
		}
	}

	if req.AvatarId > helper.AVATAR_NUM_CURR-1 {
		logs.SentryLogicCritical(acid, "UseItem item AVATAR err %d", req.AvatarId)
		return rpcError(resp, Err_Param_AVATAR)
	}

	cost := account.CostGroup{}
	data := &gamedata.CostData{}
	data.AddAvatar(req.AvatarId)
	for i, item := range req.ItemIdx {
		count := req.Count[i]
		idx := gamedata.ItemIdx_t(item)
		item_data := gamedata.GetItemDataByIdx(idx)
		if item_data == nil {
			logs.Warn("UseItem item cfg not found acid %s idx= %d", acid, idx)
			continue
		}
		// 检查物品类型
		if item_data.GetType() != "UseItem" {
			logs.Warn("UseItem item type err acid %s type= %s", acid, item_data.GetType())
			continue
		}
		// 使用物品
		if !cost.AddItemByBagId(p.Account, item, uint32(count)) {
			logs.Warn("UseItem item not enough acid %s idx= %d count= %d", acid, idx, count)
			continue
		}
		// 加 TBD 加其他类型时需要重构
		if item_data.GetPart() == helper.UseItem_Exp {
			data.AddItem(gamedata.VI_XP, item_data.GetAttrValue()*uint32(count))
		} else {
			logs.Warn("UseItem item part not support acid %s %s", acid, item_data.GetPart())
			continue
		}
	}
	if !cost.CostBySync(p.Account, resp, "UseItem") {
		logs.SentryLogicCritical(acid, "UseItem CostBySync item not enough  idx= %%v count= %v", req.ItemIdx, req.Count)
		return rpcError(resp, Err_Item_Not_Enough)
	}

	if !account.GiveBySync(p.Account, data, resp, "UseItem") {
		return rpcError(resp, Err_Give)
	}

	p.Profile.GetHero().SetNeedSync()
	// TBD BY FanYang 现在升级经验丹会解锁技能
	resp.OnChangeSkillAllChange()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}
