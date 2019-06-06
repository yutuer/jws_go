package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// GetPackgeInfo : 获取兑礼包信息
//
const (
	NotReceived    = 0
	NotFinish      = 1
	Received       = 2
	SpecialPackage = 3
)

func (p *Account) GetPackageInfoHandler(req *reqMsgGetPackageInfo, resp *rspMsgGetPackageInfo) uint32 {
	proData := gamedata.GetHotDatas().HotKoreaPackge.GetValidData() //判断是否激活
	p.Profile.Koreapackget.UpdatePackageLimit(p.GetProfileNowTime())

	resp.PackagePropInfo = make([][]byte, 0)
	var t PackagePropInfo
	var ok bool
	for _, item := range proData {
		if t, ok = p.genPackagePropInfo(item); ok {
			resp.PackagePropInfo = append(resp.PackagePropInfo, encode(t))
		}
	}
	return 0
}

/*
获取所有可以得到的礼包信息
*/
func (p *Account) genPackagePropInfo(data *ProtobufGen.HOTPACKAGE) (PackagePropInfo, bool) {

	info := PackagePropInfo{}
	if data == nil {
		logs.Error("the package is nil")
		return info, false
	}
	info.PackageID = int64(data.GetHotPackageID())
	info.SubPackageId = int64(data.GetHotPackageSubID())
	info.PackageType = int64(data.GetHotPackageType())
	now_t := p.Profile.GetProfileNowTime()
	info.LimitType = int64(data.GetLimitType())
	logs.Debug("Try to get the info of the package %d:%d. Return to client to show", data.GetHotPackageID(), data.GetHotPackageSubID())
	Start_time := gamedata.GetHotDatas().HotKoreaPackge.GetHotStartTime(info.PackageID, info.SubPackageId)

	Duration := int64(data.GetDuration()) * util.HourSec
	if (now_t > Start_time+Duration) && info.LimitType != gamedata.Nolimit {
		logs.Debug("The package %d:%d out of duration", data.GetHotPackageID(), data.GetHotPackageSubID())
		return info, false
		/*
			超出周期限制
		*/
	}

	lev := p.Profile.GetCorp().Level
	if lev < data.GetLevelLimitMin() || lev > data.GetLevelLimitMax() {
		logs.Debug("The package %d:%d out of level", data.GetHotPackageID(), data.GetHotPackageSubID())
		return info, false
		/*
			如果不在等级限制内
		*/
	}

	info.PackageName = data.GetHotPackage()
	info.SubPackageId = int64(data.GetHotPackageSubID())
	info.IapId = data.GetIAPID()
	info.VipLevel = int64(data.GetVIPLevel())

	info.Limitcount = int64(data.GetTimesLimit())
	tCount := p.Profile.Koreapackget.GetLimitById(info.PackageID, info.SubPackageId)
	info.Count = tCount
	logs.Debug("%d the usetimes of Package %d:%d", tCount, info.PackageID, info.SubPackageId)
	haveBuy := p.Profile.Koreapackget.GetCondHaveBuy(int64(data.GetHotPackageID()))

	if data.GetHotPackageType() == gamedata.SpecialPackage {
		if !haveBuy || info.Count >= info.Limitcount {
			logs.Debug("The SpecialPackage %d:%d can not be show")
			return info, false
		}
		/*
			如果特殊条件礼包只有 触发了条件并且能够购买 才发送
		*/
	}

	if haveBuy && (info.PackageType == gamedata.ConditonPackage || info.PackageType == gamedata.LevelPackage) {
		info.Count = 1
		/*
			条件礼包和阶梯礼包的特殊限制
		*/
	}
	info.PackageItem = make([]string, 0)
	info.PackageCount = make([]int64, 0)
	info.StartTime = Start_time
	if data.GetDuration() != 999999 {
		info.EndTime = Start_time + Duration
	} else {
		info.EndTime = 0
	}

	info.BackRatio = int64(data.GetCashbackRatio())
	info.HCValue = int64(data.GetHCValue())
	info.ShowValue = int64(data.GetShowValue())
	info.ConditionType = int64(data.GetFCType())
	info.ConditionIp1 = int64(data.GetFCValueIP1())
	info.ConditionIp2 = int64(data.GetFCValueIP2())
	info.ConditionSp1 = data.GetFCValueSP1()
	info.ConditionSp2 = data.GetFCValueSP2()
	info.QuestName = data.GetQuestName()
	info.QuestDes = data.GetQuestDes()
	info.BackImage = int64(data.GetBackImage())
	info.BuyPackage = 0
	info.CurrentPos, _ = p.Profile.Koreapackget.GetCurrentPosById(info.PackageID)
	logs.Debug("Get Package CurrentPos from profile level:%d", info.CurrentPos)
	for _, value := range data.GetHotPackageGoods_Temp() {
		tV, tC := value.GetGoodsID(), value.GetGoodsCount()
		info.PackageItem = append(info.PackageItem, tV)
		info.PackageCount = append(info.PackageCount, int64(tC))
	}

	/*
		如果是条件礼包并且已经购买,更新BuyPackage的状态
	*/
	if data.GetHotPackageType() == gamedata.ConditonPackage && haveBuy {
		info.BuyPackage = 1
		if tCount == info.Limitcount {
			info.CanBuyPackage = Received
		} else {
			tCondi := account.NewCondition(uint32(data.GetFCType()), int64(data.GetFCValueIP1()), int64(data.GetFCValueIP2()),
				data.GetFCValueSP1(), data.GetFCValueSP2())
			pro, all := account.GetConditionProgress(tCondi, p.Account, data.GetFCType(), int64(data.GetFCValueIP1()), int64(data.GetFCValueIP2()),
				data.GetFCValueSP1(), data.GetFCValueSP2())
			info.Progress = int64(pro)
			info.All = int64(all)
			if pro >= all {
				info.CanBuyPackage = NotReceived
			} else {
				info.CanBuyPackage = NotFinish
			}
		}
	}

	if data.GetHotPackageType() == gamedata.LevelPackage {
		//阶梯礼包复用为已购买次数
		info.Progress = tCount
		logs.Debug("%d the usetimes of LevelPackage %d:%d", info.Progress, info.PackageID, info.SubPackageId)
	}

	return info, true
}

/*
特殊条件礼包触发
*/
func (p *Account) GetSpecialPackageInfoHandler(req *reqMsgGetSpecialPackageInfo, resp *rspMsgGetSpecialPackageInfo) uint32 {
	logs.Debug("Get Special Package Info %d", req.SpackageNum)
	value := gamedata.GetHotDatas().HotKoreaPackge.GetHotPackage(req.SpackageNum, -1) //判断是否激活
	if value == nil {
		logs.Error("The package %d is an invaild package", req.SpackageNum)
		return 0
	}
	p.Profile.Koreapackget.UpdatePackageLimit(p.GetProfileNowTime())
	Start_time := gamedata.GetHotDatas().HotKoreaPackge.GetHotStartTime(req.SpackageNum, -1)
	now_t := p.Profile.GetProfileNowTime()
	resp.ContinueShow = 1
	Duration := int64(value.GetDuration()) * util.HourSec
	logs.Debug("SpecialPackagePropInfo StartTime:%d Duration:%d", Start_time, Duration)
	if (now_t > Start_time+Duration) && value.GetLimitType() != gamedata.Nolimit {
		/*
			如果超出了持续时间
		*/
		logs.Debug("This package out of duration start:%d end:%d", Start_time, Start_time+Duration)
		resp.ContinueShow = 0
		return 0
	}

	num := p.Profile.Koreapackget.GetLimitById(int64(value.GetHotPackageID()), int64(value.GetHotPackageSubID()))
	if value.GetLimitType() == gamedata.AllLimit && num >= int64(value.GetTimesLimit()) {
		logs.Debug("The Special Package %d:%d have been buy", value.GetHotPackageID(), value.GetHotPackageSubID())
		resp.ContinueShow = 0
		return 0
		//如果购买次数超过了限制 并且 是全周期限购的
	}
	//否则将特殊礼包加入conhavebuy表示需要进行显示
	if !p.Profile.Koreapackget.GetCondHaveBuy(req.SpackageNum) {
		resp.ContinueShow += 100
	}
	p.Profile.Koreapackget.InsertLimitTimeOne(int64(value.GetHotPackageID()), int64(value.GetHotPackageSubID()), now_t)
	logs.Debug("SpecialPackagePropInfo insert to profile continue to show:%d", resp.ContinueShow)
	return 0
}

/*
领取条件礼包
*/
func (p *Account) ReceiveConditionPackageHandler(req *reqMsgReceiveConditionPackage, resp *rspMsgReceiveConditionPackage) uint32 {
	logs.Debug("ReceiveConditionPackage")
	pkgid := req.PackageId
	subpkgid := req.SubPackageId
	value := gamedata.GetHotDatas().HotKoreaPackge.GetHotPackage(pkgid, subpkgid)
	if value == nil {
		logs.Error("The package %d is an invaild package", pkgid)
		return 0
	}
	p.Profile.Koreapackget.UpdatePackageLimit(p.GetProfileNowTime())
	Start_time := gamedata.GetHotDatas().HotKoreaPackge.GetHotStartTime(pkgid, subpkgid)
	giveData := &gamedata.CostData{}
	reason := "Buy Korea Package"

	now_t := p.Profile.GetProfileNowTime()
	Duration := int64(value.GetDuration()) * util.HourSec
	logs.Debug("SpecialPackagePropInfo StartTime:%d Duration:%d", Start_time, Duration)
	if (now_t > Start_time+Duration) && value.GetLimitType() != gamedata.Nolimit {
		logs.Debug("This package out of duration start:%d end:%d", Start_time, Start_time+Duration)
		return 0
	}
	if int64(value.GetTimesLimit()) <= p.Profile.Koreapackget.GetLimitById(pkgid, subpkgid) {
		logs.Debug("Receive condition package %d:%d out of limittimes", pkgid, subpkgid)
		return errCode.CommonMaxLimit
	}
	giveData.IAPPkgInfo.PkgId = 0
	giveData.IAPPkgInfo.SubPkgId = 0
	for _, tvalue := range value.GetHotPackageGoods_Temp() {
		tV, tC := tvalue.GetGoodsID(), tvalue.GetGoodsCount()
		giveData.AddItem(tV, uint32(tC))
		logs.Debug("give data add item %d:%d", tV, tC)
	}
	if value.GetHCValue() != 0 {
		giveData.AddItem(helper.VI_Hc_Give, value.GetHCValue())
	}

	p.Profile.Koreapackget.UpdateLimitTimeOne(int64(pkgid), int64(subpkgid), p.GetProfileNowTime())
	resp.OnChangeIAPGift()
	if !account.GiveBySync(p.Account, giveData, resp, reason) {
		logs.Error("Give By Sync Fail")
	}
	return 0
}

/*
每次关闭列表刷新界面，将特殊礼包剔除。 如果关闭等于购买，则修改相应的购买次数
*/
func (p *Account) CloseSendInfoHandler(req *reqMsgCloseSendInfo, rsp *rspMsgCloseSendInfo) uint32 {
	logs.Debug("Close and Update info")
	for i := 0; i < len(req.PackageId); i++ {
		value := gamedata.GetHotDatas().HotKoreaPackge.GetHotPackage(req.PackageId[i], req.SubPackageId[i])
		if value == nil {
			logs.Error("The package %d:%d is an invaild package", req.PackageId[i], req.SubPackageId[i])
			continue
		}
		logs.Debug("Close Package %d:%d  close=buy:%d", req.PackageId[i], req.SubPackageId[i], value.GetIsLost_ByClosePanel())
		if value.GetIsLost_ByClosePanel() == 1 {
			p.Profile.Koreapackget.FullUseTimes(req.PackageId[i], req.SubPackageId[i], int64(value.GetTimesLimit()))
		}
		p.Profile.Koreapackget.RemovePackage(req.PackageId[i])
	}
	return 0
}
