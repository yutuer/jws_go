package logics

import (
	"fmt"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// WheelShowInfo : 幸运转盘展示
//
func (p *Account) WheelShowInfoHandler(req *reqMsgWheelShowInfo, resp *rspMsgWheelShowInfo) uint32 {

	now_t := p.Profile.GetProfileNowTime()
	ga := gamedata.GetHotDatas().Activity
	wheelinfo := p.Profile.GetWheelGachaInfo()
	var actInfo *gamedata.HotActivityInfo
	_actInfo := gamedata.GetHotDatas().Activity.GetActivityInfo(gamedata.ActLuckyWheel, p.Profile.ChannelQuickId)
	//检查活动的时间
	for _, v := range _actInfo {
		if now_t > v.StartTime && now_t < v.EndTime {
			actInfo = v
		}
	}
	if actInfo == nil {
		logs.Debug("Dont have Wheel Activities")
		return errCode.ActivityTimeOut
	}
	// 跟新活动Id,如果不是同一个活动，清除数据
	if actInfo.ActivityId != wheelinfo.ActId {
		logs.Debug("LuckyWheel:not same activity,old:%d --- new:%d,clear info",wheelinfo.ActId,actInfo.ActivityId)
		wheelinfo.InitInfo()
		wheelinfo.UpdateActId(actInfo.ActivityId)
	}
	// 检查解锁条件
	if !account.CondCheck(gamedata.Mod_LuckyWheel, p.Account) {
		logs.Debug("WheelActivity was locked")
		return errCode.ActivityNotValid
	}
	showInfo := ga.GetWheelGachaShow(actInfo.ActivityId)
	setInfo := ga.GetWheelSeting(actInfo.ActivityId)
	costInfo := ga.GetWheelCost(actInfo.ActivityId)

	//获取道具展示信息
	for _, info := range showInfo {
		resp.ItemID = append(resp.ItemID, info.GetItemID())
		resp.ItemCount = append(resp.ItemCount, int64(info.GetItemCount()))
		resp.Special = append(resp.Special, int64(info.GetSpecial()))
	}

	//转盘的基本信息
	resp.HcBase = int64(setInfo.GetHCBase())
	resp.Index = wheelinfo.CurNum
	resp.GetCoin = int64(setInfo.GetGetCoinItem())
	resp.GetCoinMax = int64(setInfo.GetGetCoinItemMax())
	resp.CurCoin = wheelinfo.GetCurCoin()
	resp.GachaCoin = setInfo.GetGachaCoin()

	//避免表错导致不足,获取当前需要的积分
	resp.CoinCost = int64(costInfo[len(costInfo)-1].GetCost())
	for _, value := range costInfo {
		if int64(value.GetIndex()) == wheelinfo.CurNum+1 {
			resp.CoinCost = int64(value.GetCost())
		}
	}

	//获取物品的品级，如果找不到则品级为0
	for _, value := range resp.ItemID {
		if itemcfg, ok := gamedata.GetProtoItem(value); ok {
			resp.RareLevel = append(resp.RareLevel, int64(itemcfg.GetRareLevel()))
		} else {
			resp.RareLevel = append(resp.RareLevel, int64(0))
		}
	}
	tVis := make(map[string]bool)
	for _, value := range wheelinfo.GenItem() {
		tVis[value] = true
	}
	for _, value := range resp.ItemID {
		if tVis[value] {
			resp.DontShow = append(resp.DontShow, 1)
		} else {
			resp.DontShow = append(resp.DontShow, 0)
		}
	}
	logs.Debug("LuckyWheel:HcBase %d; CurNum %d; GetCoin %d; GetCoinMax %d; CurCoin %d; GachaCoin %s; Cost %d",
		resp.HcBase, resp.Index, resp.GetCoin, resp.GetCoinMax, resp.CurCoin, resp.GachaCoin, resp.CoinCost)
	logs.Debug("WheelItemShow:")
	for i := 0; i < len(resp.ItemID); i++ {
		logs.Debug("ItemId:%s ItemCount:%d special:%d", resp.ItemID[i], resp.ItemCount[i], resp.Special[i])
	}
	logs.Debug("DontShow")
	logs.Debug("%v", resp.DontShow)
	//for _,value := range resp.DontShow{
	//	logs.Debug("")
	//}

	return 0
}

func (p *Account) UseWheelOneHandler(req *reqMsgUseWheelOne, resp *rspMsgUseWheelOne) uint32 {
	//由于rpcWarn是0  rpcError是200+
	const (
		_                               = iota + 200
		CODE_NotEough_NotReachLimit_Err //积分不足并没有达到上限
		CODE_NotEough_ReachLimit_Err    //积分不足且达到上限
		CODE_Key_Cost_Err               //扣除积分失败
		CODE_Item_Give_Err              //发放物品失败
		CODE_Item_OutTimes_Err          //抽取已经超过了10次
	)

	now_t := p.Profile.GetProfileNowTime()
	ga := gamedata.GetHotDatas().Activity
	wheelinfo := p.Profile.GetWheelGachaInfo()
	var actInfo *gamedata.HotActivityInfo
	_actInfo := gamedata.GetHotDatas().Activity.GetActivityInfo(gamedata.ActLuckyWheel, p.Profile.ChannelQuickId)
	//检查活动的时间
	for _, v := range _actInfo {
		if now_t > v.StartTime && now_t < v.EndTime {
			actInfo = v
		}
	}
	if actInfo == nil {
		return errCode.ActivityTimeOut
		logs.Debug("Dont have Wheel Activities")
	}
	// 跟新活动Id,如果不是同一个活动，清除数据
	if actInfo.ActivityId != wheelinfo.ActId {
		logs.Debug("LuckyWheel:not same activity,old:%d --- new:%d,clear info",wheelinfo.ActId,actInfo.ActivityId)
		wheelinfo.InitInfo()
		wheelinfo.UpdateActId(actInfo.ActivityId)
	}
	// 检查解锁条件
	if !account.CondCheck(gamedata.Mod_LuckyWheel, p.Account) {
		logs.Debug("WheelActivity was locked")
		return errCode.ActivityNotValid
	}

	//判断是否已经抽取了十次
	if wheelinfo.CurNum >= 10 {
		logs.Debug("LuckyWheel:used out ten times")
		return CODE_Item_OutTimes_Err
	}

	setInfo := ga.GetWheelSeting(actInfo.ActivityId)
	costInfo := ga.GetWheelCost(actInfo.ActivityId)
	costCoinNeed := costInfo[len(costInfo)-1].GetCost()

	//获取当前抽取所需要的积分
	for _, value := range costInfo {
		if int64(value.GetIndex()) == wheelinfo.CurNum+1 {
			costCoinNeed = value.GetCost()
		}
	}

	//判断是否有足够的积分
	if ok := wheelinfo.Has(int64(costCoinNeed)); !ok {
		if wheelinfo.CoinFull(setInfo.GetGetCoinItemMax()) {
			return CODE_NotEough_ReachLimit_Err
		} else {
			return CODE_NotEough_NotReachLimit_Err
		}
	}

	data := &gamedata.CostData{}
	data.AddItem(setInfo.GetGachaCoin(), uint32(costCoinNeed))
	if !account.CostBySync(p.Account, data, resp, "LuckyWheel Coin") {
		logs.Error("LuckyWheel: Dont have enough Coins")
		return CODE_Key_Cost_Err
	}

	gachaNum := wheelinfo.CurNum
	haveItem := wheelinfo.GenItem()
	noticeLimit := setInfo.GetMsgQuality()
	specialId, IsSpecial := ga.IsWhiteGachaSpecial(actInfo.ActivityId, uint32(gachaNum))
	var cfg *ProtobufGen.WHEELGACHA
	//特殊掉落组 和 普通掉落组
	if IsSpecial {
		cfg = ga.GetWheelGachaNormal(specialId).RandomConfigWheel(haveItem)
	} else {
		cfg = ga.GetWheelGachaNormal(setInfo.GetItemGroupID()).RandomConfigWheel(haveItem)
	}
	itemID := cfg.GetItemID()
	itemCount := int64(cfg.GetItemCount())

	data.AddItem(itemID, uint32(itemCount))
	if !account.GiveBySync(p.Account, data, resp, "LuckyWheelOnespecilId") {
		logs.Error("LuckyWheelOne GiveBySync Err")
		return CODE_Item_Give_Err
	}
	if itemcfg, ok := gamedata.GetProtoItem(itemID); ok {
		if uint32(itemcfg.GetRareLevel()) >= noticeLimit {
			sysnotice.NewSysRollNotice(p.AccountID.ServerString(), int32(setInfo.GetMsgIDS())).
				AddParam(sysnotice.ParamType_RollName, p.Profile.Name).
				AddParam(sysnotice.ParamType_ItemId, itemID).
				AddParam(sysnotice.ParamType_Value, fmt.Sprintf("%d", itemCount)).Send()
		}
	}
	resp.ItemID = itemID
	resp.ItemCount = itemCount
	resp.Unusual = int64(cfg.GetUnusual())
	resp.NextCost = int64(costCoinNeed)
	if IsSpecial {
		logs.Debug("LuckyWheel: Itemid:%s ItemCount:%d Unusual:%d NextCost:%d GachaId:%d SpecialId:%d",
			itemID, itemCount, resp.Unusual, resp.NextCost, gachaNum, specialId)
	}else{
		logs.Debug("LuckyWheel: Itemid:%s ItemCount:%d Unusual:%d NextCost:%d GachaId:%d NormalId:%d",
			itemID, itemCount, resp.Unusual, resp.NextCost, gachaNum, setInfo.GetItemGroupID())
	}
	wheelinfo.UpdateUseWheel()
	return 0
}
