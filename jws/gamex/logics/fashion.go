package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) FashionRefresh(sync helper.ISyncRsp) {
	acid := p.AccountID.String()
	now_time := p.Profile.GetProfileNowTime()
	deled := make(map[uint32]uint32, 10)
	cost := account.CostGroup{}
	for _, item := range p.Profile.GetFashionBag().GetFashionAll() {
		id := item.ID
		if ok, itemCfg := gamedata.IsFashion(item.TableID); ok {
			if !gamedata.IsFashionPerm(itemCfg, item.ExpireTimeStamp) &&
				now_time >= item.ExpireTimeStamp {
				deled[id] = id
				cost.AddFashionByBagId(p.Account, id)
				// 发邮件
				mail_sender.SendFashionTimeOut(acid, item.TableID)
			}
		}
	}
	cost.CostBySync(p.Account, sync, "RefreshFashion")
	avatarEquip := p.Profile.GetAvatarEquips()
	aeqs, l := avatarEquip.Curr()
	for avatar := 0; avatar < helper.AVATAR_NUM_CURR; avatar++ {
		for slot_in_avatar := 0; slot_in_avatar < l; slot_in_avatar++ {
			slot := l*avatar + slot_in_avatar
			if slot >= len(aeqs) {
				break
			}
			id := aeqs[slot]
			if id <= 0 {
				continue
			}
			if _, ok := deled[id]; ok {
				avatarEquip.EquipImp(avatar, slot_in_avatar,
					p.getDefaultFashionBagId(avatar, slot_in_avatar))
				sync.OnChangeAvatarEquip()
				p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 8. 时装
			} else {
				if id > 0 && !p.Profile.GetFashionBag().HasFashionByBagId(id) { // 清理还装备着，但包裹里没有的装备，应该没有这种情况才对
					avatarEquip.UnEquipImp(avatar, slot_in_avatar)
					sync.OnChangeAvatarEquip()
					logs.Warn("equip but not exist, %s %d", p.AccountID.String(), id)
				}
			}
		}
	}
}

type RequestRefreshFashion struct {
	Req
}

type ResponseRefreshFashion struct {
	SyncResp
}

func (p *Account) RefreshFashionReq(r servers.Request) *servers.Response {
	req := &RequestRefreshFashion{}
	resp := &ResponseRefreshFashion{}

	initReqRsp(
		"PlayerAttr/RefreshFashionResponse",
		r.RawBytes,
		req, resp, p)

	resp.mkInfo(p)
	return rpcSuccess(resp)
}

func (p *Account) BuyFashion(r servers.Request) *servers.Response {
	req := &struct {
		Req
		ItemId string `codec:"itemid"`
	}{}
	resp := &struct {
		SyncRespWithRewards
	}{}

	initReqRsp(
		"PlayerAttr/BuyFashionResp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Param_Err_Not_Fashion
		Err_Avatar_No_Unlock
		Err_Fashion_Already_Have
		Err_Fashion_Can_Not_Buy
		Err_Money_Not_Enough
		Err_Give_Fashion_Fail
	)
	// 是否是时装
	ok, cfg := gamedata.IsFashion(req.ItemId)
	if !ok {
		return rpcErrorWithMsg(resp, Err_Param_Err_Not_Fashion, fmt.Sprintf("Err_Param_Err_Not_Fashion %s", req.ItemId))
	}
	// 是否有价格
	if cfg.GetCoinPrice() <= 0 {
		return rpcErrorWithMsg(resp, Err_Fashion_Can_Not_Buy, fmt.Sprintf("Err_Fashion_Can_Not_Buy %s", req.ItemId))
	}
	// 角色是否解锁
	if !p.Account.IsAvatarUnblock(int(cfg.GetRoleOnly())) {
		return rpcErrorWithMsg(resp, Err_Avatar_No_Unlock, fmt.Sprintf("Err_Avatar_No_Unlock AvatarID %d", cfg.GetRoleOnly()))
	}
	// 是否已有了此时装
	if has, _ := p.Profile.GetFashionBag().HasFashionByTableId(req.ItemId, cfg,
		p.Profile.GetProfileNowTime()); has {
		return rpcErrorWithMsg(resp, Err_Fashion_Already_Have, fmt.Sprintf("Err_Fashion_Already_Have %s", req.ItemId))
	}
	// 扣钱
	cost := gamedata.CostData{}
	cost.AddItem(cfg.GetCoin(), cfg.GetCoinPrice())
	cost_group := account.CostGroup{}
	if !cost_group.AddCostData(p.Account, &cost) || !cost_group.CostBySync(p.Account, resp, "BuyFashion") {
		return rpcErrorWithMsg(resp, Err_Money_Not_Enough, fmt.Sprintf("Err_Money_Not_Enough %s", req.ItemId))
	}
	// 加时装
	give := account.GiveGroup{}
	give.AddItem(req.ItemId, 1)
	if !give.GiveBySyncAuto(p.Account, resp, "BuyFashion") {
		return rpcErrorWithMsg(resp, Err_Give_Fashion_Fail, fmt.Sprintf("Err_Money_Not_Enough %s", req.ItemId))
	}

	p.Profile.GetData().SetNeedCheckMaxGS() // MaxGS可能变化 8. 时装
	resp.mkInfo(p)
	return rpcSuccess(resp)
}

// 从背包里查找默认时装，应该都能找到
// notice by zhangzhen: 增加新时装部位时，这里需要改
func (p *Account) getDefaultFashionBagId(avatar_id, slot_in_avatar int) uint32 {
	avatar_cfg := gamedata.GetAvatarInitFashionData(avatar_id)
	if avatar_cfg != nil {
		itemId := ""
		switch slot_in_avatar {
		case gamedata.FashionPart_Weapon:
			itemId = avatar_cfg.GetInitFWeapon()
		case gamedata.FashionPart_Armor:
			itemId = avatar_cfg.GetInitFAmor()
		}
		if itemId != "" {
			_, cfg := gamedata.IsFashion(itemId)
			if has, id := p.Profile.GetFashionBag().HasFashionByTableId(itemId, cfg,
				p.Profile.GetProfileNowTime()); has {
				return id
			}
		}
	}
	logs.Error("default fashion not found, avatar %d slot %d", avatar_id, slot_in_avatar)
	return 0
}
