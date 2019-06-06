package logics

import (
	"math"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/friend"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// GiveGiftToFriend : 给友人赠送礼品
//
func (p *Account) GiveGiftToFriendHandler(req *reqMsgGiveGiftToFriend, resp *rspMsgGiveGiftToFriend) uint32 {
	f := &p.Friend
	if !f.IsInFriendList(req.TargetID) {
		logs.Error("not in friend list for acid: %s", req.TargetID)
		return errCode.CommonInner
	}
	if !f.CanGiveGift(req.TargetID) {
		logs.Error("has give gift for acid: %s", req.TargetID)
		return errCode.CommonInner
	}
	friend.GetModule(p.AccountID.ShardId).GiftCommandExecAsync(friend.GiftCmd{
		Typ:   friend.GiftCmd_Give,
		Param: []string{req.TargetID, p.AccountID.String()},
	})
	f.MarkGift2Friend(req.TargetID)
	p.updateCondition(account.COND_TYP_Give_FriendGift, 1, 0, "", "", resp)
	resp.OnChangeFriendGift()
	logs.Debug("[FriendGift] gift2friend info: %v", f.GetFriendGiftInfo())
	return 0
}

// ReceiveGiftFromFriend : 收取好友赠送的礼品
//
func (p *Account) ReceiveGiftFromFriendHandler(req *reqMsgReceiveGiftFromFriend, resp *rspMsgReceiveGiftFromFriend) uint32 {
	f := &p.Friend
	if !f.CanReceiveGift() {
		return errCode.CommonConditionFalse
	}
	ok := p.getReceiveGiftRetFromeModule(req.TargetID)
	if !ok {
		return errCode.CommonConditionFalse
	}
	costData := &gamedata.CostData{}
	cfg := gamedata.GetFriendConfig()
	costData.AddItem(gamedata.VI_BaoZi, cfg.GetGiveBaoziNum())
	if !account.GiveBySync(p.Account, costData, resp, "friend gift") {
		return errCode.RewardFail
	}
	f.AddReceiveGiftTimes(1)
	resp.SyncReceiveGiftTimes = f.GetReceiveGiftTimes()
	p.updateGiftReceiveInfo(false, &resp.SyncResp)
	return 0
}

// BatchGiveGift2Friend : 批量给友人赠送礼品
//
func (p *Account) BatchGiveGift2FriendHandler(req *reqMsgBatchGiveGift2Friend, resp *rspMsgBatchGiveGift2Friend) uint32 {
	f := &p.Friend
	c := 0
	for _, item := range f.GetFriendList() {
		if f.CanGiveGift(item.AcID) {
			c++
			friend.GetModule(p.AccountID.ShardId).GiftCommandExecAsync(friend.GiftCmd{
				Typ:   friend.GiftCmd_Give,
				Param: []string{item.AcID, p.AccountID.String()},
			})
			p.updateCondition(account.COND_TYP_Give_FriendGift, 1, 0, "", "", resp)
			f.MarkGift2Friend(item.AcID)
		}
	}
	if c == 0 {
		return errCode.ALLFRIENDHADGIVEN
	}
	resp.OnChangeFriendGift()
	return 0
}

// BatchReceiveGiftFromFriend : 批量收取好友赠送的礼品
//
func (p *Account) BatchReceiveGiftFromFriendHandler(req *reqMsgBatchReceiveGiftFromFriend, resp *rspMsgBatchReceiveGiftFromFriend) uint32 {
	f := &p.Friend
	if !f.CanReceiveGift() {
		return errCode.CommonConditionFalse
	}
	cfg := gamedata.GetFriendConfig()
	info, ok := p.getFriendInfoFromModule(false)
	if !ok {
		return errCode.CommonInner
	}
	// warn 目前只支持发放包子
	curCount := p.Profile.GetSC().GetSC(gamedata.SC_BaoZi)
	if curCount >= int64(gamedata.GetCommonCfg().GetBaoZiGetLimit()) {
		return errCode.CommonConditionFalse
	}
	leaveTimes := int(cfg.GetGetBaoziTimes()) - f.GetReceiveGiftTimes()
	minL := int(math.Min(float64(len(info)), float64(leaveTimes)))
	failC := 0
	for i := len(info) - 1; i >= len(info)-minL; i-- {
		ok := p.getReceiveGiftRetFromeModule(info[i])
		if !ok {
			failC++
		}
	}
	if failC > 0 {
		logs.Error("fatal error, faicC: %d, info: %v", failC, info)
	}
	rc := minL - failC
	costData := &gamedata.CostData{}
	costData.AddItem(gamedata.VI_BaoZi, uint32(rc)*cfg.GetGiveBaoziNum())
	if !account.GiveBySync(p.Account, costData, resp, "friend batch gift") {
		return errCode.RewardFail
	}
	f.AddReceiveGiftTimes(rc)
	p.updateGiftReceiveInfo(false, &resp.SyncResp)
	return 0
}

func (p *Account) getFriendInfoFromModule(refresh bool) ([]string, bool) {
	param := []string{p.AccountID.String()}
	if refresh {
		param = append(param, "")
		param = append(param, p.Friend.GetBlackAcID()[:]...)
	}
	ret := friend.GetModule(p.AccountID.ShardId).GiftCommandExec(friend.GiftCmd{
		Typ:   friend.GiftCmd_GetInfo,
		Param: param,
	})
	if ret == nil {
		return nil, false
	}
	if info, ok := ret.Ret.([]string); ok {
		return info, true
	} else {
		logs.Error("fatal error, deserialized failed for interface %v", ret.Ret)
		return nil, false
	}
}

func (p *Account) getReceiveGiftRetFromeModule(tgtID string) bool {
	ret := friend.GetModule(p.AccountID.ShardId).GiftCommandExec(friend.GiftCmd{
		Typ:   friend.GiftCmd_Receive,
		Param: []string{p.AccountID.String(), tgtID},
	})
	if ret == nil {
		return false
	}
	if r, ok := ret.Ret.(bool); ok {
		if !r {
			return false
		}
	} else {
		return false
	}
	return true
}

// GetReceiveGiftInfo : 获得收到的好友礼品列表信息
//
func (p *Account) GetReceiveGiftInfoHandler(req *reqMsgGetReceiveGiftInfo, resp *rspMsgGetReceiveGiftInfo) uint32 {
	p.updateGiftReceiveInfo(true, &resp.SyncResp)
	return 0
}

// GetFriendGiftAcID : 获得已赠予礼品的好友的名单
//
func (p *Account) GetFriendGiftAcIDHandler(req *reqMsgGetFriendGiftAcID, resp *rspMsgGetFriendGiftAcID) uint32 {
	resp.OnChangeFriendGift()
	return 0
}

func (p *Account) updateGiftReceiveInfo(refresh bool, resp *SyncResp) {
	f := &p.Friend
	info, ok := p.getFriendInfoFromModule(refresh)
	if ok {
		resp.SyncReceiveGiftList = make([][]byte, 0)
		friendGiftInfo := friend.GetModule(p.AccountID.ShardId).GetFriendInfo(info, false)
		for i := len(info) - 1; i >= 0; i-- {
			for _, item := range friendGiftInfo {
				if item.AcID == info[i] {
					resp.SyncReceiveGiftList = append(resp.SyncReceiveGiftList, encode(item))
					break
				}
			}
		}
	}
	resp.OnChangeReceiveGiftList(true)
	resp.SyncReceiveGiftTimes = f.GetReceiveGiftTimes()
	resp.OnChangeFriendGift()
}
