package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	DayType = [5]int64{0, util.DaySec, util.WeekSec, 30 * util.DaySec, 0}
)

type LevelPkg struct {
	Id    int64 `json:"Id"`
	Level int64 `json:"Level"`
}

type KoreaPackget struct {
	LastUpadateTime []int64    `json:"last_update_time"`  //上一次更新时间
	PackageUseTimes []int64    `json:"package_use_times"` //已购买次数
	PackageId       []int64    `json:"package_id"`        //礼包id
	SubPkgId        []int64    `json:"sub_package_id"`    //子礼包id
	CurrentPos      []LevelPkg `json:"current_pos"`       //当前购买层次
	CondHaveBuy     []int64    `json:"cond_have_buy"`     //已经购买的礼包
}

/*
ConHaveBuy用于判断礼包(非特殊礼包)是否被购买 和 特殊礼包是否显示
*/

func (pkg *KoreaPackget) RemovePackage(pkgid int64) {
	Type := gamedata.GetHotDatas().HotKoreaPackge.GetHotPackageType(pkgid)
	logs.Debug("Remove Package %d-type:%d", pkgid, Type)
	if Type == gamedata.SpecialPackage || Type == gamedata.ConditonPackage || Type == gamedata.LevelPackage {
		for i := 0; i < len(pkg.CondHaveBuy); i++ {
			if pkg.CondHaveBuy[i] == pkgid {
				logs.Debug("remove the package %d from the condhavebuy", pkgid)
				pkg.CondHaveBuy = append(pkg.CondHaveBuy[:i], pkg.CondHaveBuy[i+1:]...)
			}
		}
	}
}

/*
更新时间
*/
func (pkg *KoreaPackget) UpdateLastTime(start_time int64, now_time int64, timetyp int64, pos int) {
	durat := DayType[timetyp]
	if durat == 0 {
		pkg.LastUpadateTime[pos] = start_time
		return
	}
	tc := (now_time - start_time) / durat
	pkg.LastUpadateTime[pos] = start_time + durat*tc
}

/*
礼包时间到达重置的时候需要删除CondHaveBuy进行维护，维护更新时间
普通礼包：修改使用次数
条件礼包／阶段礼包：无周期性限购
特殊触发的礼包：如普通礼包一样处理
*/
func (pkg *KoreaPackget) UpdatePackageLimit(now_time int64) {
	for i := 0; i < len(pkg.PackageId); i++ {
		t := now_time - pkg.LastUpadateTime[i]
		timetype := gamedata.GetHotDatas().HotKoreaPackge.GetHotTimeType(pkg.PackageId[i], pkg.SubPkgId[i])
		if int(timetype) == gamedata.Nolimit || int(timetype) == gamedata.AllLimit {
			continue
		} else if t > DayType[timetype] {
			pkg.UpdateLastTime(pkg.LastUpadateTime[i], now_time, timetype, i)
			pkg.PackageUseTimes[i] = 0
			pkg.RemovePackage(pkg.PackageId[i])
			logs.Debug("Update package %d:%d from profile", pkg.PackageId[i], pkg.SubPkgId[i])
		}
	}
}

/*
条件礼包：
是否购买了条件礼包 or 是否显示特殊礼包
*/
func (pkg *KoreaPackget) GetCondHaveBuy(pkgid int64) bool {
	for _, value := range pkg.CondHaveBuy {
		if value == pkgid {
			logs.Debug("Can find the pkg %d from condhavebuy", pkgid)
			return true
		}
	}
	logs.Debug("Can not find the pkg %d from condhavebuy", pkgid)
	return false
}

/*
阶梯礼包：
获取当前的阶段
*/
func (pkg *KoreaPackget) GetCurrentPosById(pkgid int64) (int64, bool) {
	logs.Debug("Get profile's levelPackage %d", pkgid)
	for _, value := range pkg.CurrentPos {
		if value.Id == pkgid {
			logs.Debug("The level of Package %d is %d", pkgid, value.Level)
			return value.Level, true
		}
	}
	logs.Debug("Can not find the level of Package %d ", pkgid)
	return 0, false
}

/*
更新当前阶段
*/
func (pkg *KoreaPackget) UpdateCurrentPos(pkgid int64) {
	for i, value := range pkg.CurrentPos {
		if value.Id == pkgid {
			pkg.CurrentPos[i].Level++
			logs.Debug("The level of Package %d is %d", pkgid, pkg.CurrentPos[i].Level)
			return
		}
	}
	pkg.CurrentPos = append(pkg.CurrentPos, LevelPkg{pkgid, 1})
	logs.Debug("The level of Package %d is %d", pkgid, 1)
	//如果找不到则说明当前是第一阶段（0）的购买
}

/*
购买一个礼包
如果profile中没有这个礼包，则新增 ；
如果是阶梯礼包根据使用次数and次数限制判断到达哪个阶段
维护CondHaveBuy
阶段礼包：维护CurrentPos[{Id , Level}]
特殊条件礼包 和 普通礼包一样处理
*/
func (pkg *KoreaPackget) UpdateLimitTimeOne(pkgid int64, subpkgid int64, now_time int64) {
	for i, _ := range pkg.PackageId {
		if pkg.PackageId[i] == pkgid && pkg.SubPkgId[i] == subpkgid {
			value := gamedata.GetHotDatas().HotKoreaPackge.GetHotPackage(pkgid, subpkgid)
			if value == nil {
				logs.Error("The package %d:%d is an invaild package", pkgid, subpkgid)
				return
			}
			logs.Debug("Update the use time of package=%d usetimes=%d", pkgid, pkg.PackageUseTimes[i])
			Type := value.GetHotPackageType()
			if Type == gamedata.SpecialPackage || Type == gamedata.ConditonPackage || Type == gamedata.LevelPackage {
				pkg.UpdateCondiPkgOne(pkgid)
			}
			pkg.PackageUseTimes[i]++
			//如果是阶梯礼包更新阶梯状态
			logs.Debug("Package:%d:%d  usetimes:%d  limittimes:%d", pkgid, subpkgid, pkg.PackageUseTimes[i], int(value.GetTimesLimit()))
			if pkg.PackageUseTimes[i] == int64(value.GetTimesLimit()) && Type == gamedata.LevelPackage {
				pkg.UpdateCurrentPos(pkgid)
			}
			return
		}
	}
	pkg.InsertId(pkgid, subpkgid, now_time, 1)
}

/*
用户中新添礼包
ustimes = 1表明买了礼包， 0表示只是想显示礼包
*/
func (pkg *KoreaPackget) InsertId(pkgid int64, subpkgid int64, now_time int64, usetimes int64) {
	logs.Debug("Insert package info to profile")
	pkg.PackageId = append(pkg.PackageId, pkgid)
	pkg.SubPkgId = append(pkg.SubPkgId, subpkgid)
	pkg.LastUpadateTime = append(pkg.LastUpadateTime, 0)
	pkg.PackageUseTimes = append(pkg.PackageUseTimes, usetimes)
	value := gamedata.GetHotDatas().HotKoreaPackge.GetHotPackage(pkgid, subpkgid)
	Len := len(pkg.LastUpadateTime)
	Type := value.GetHotPackageType()
	if Type == gamedata.SpecialPackage || Type == gamedata.ConditonPackage || Type == gamedata.LevelPackage {
		pkg.UpdateCondiPkgOne(pkgid)
	}

	if pkg.PackageUseTimes[Len-1] == int64(value.GetTimesLimit()) && value.GetHotPackageType() == gamedata.LevelPackage {
		pkg.UpdateCurrentPos(pkgid)
	}
	timetype := int64(value.GetLimitType())
	Start_time := gamedata.GetHotDatas().HotKoreaPackge.GetHotStartTime(pkgid, subpkgid)

	pkg.UpdateLastTime(Start_time, now_time, timetype, Len-1)
	logs.Debug("insert package to profile Id %d:%d lasttime:%d", pkgid, subpkgid, pkg.LastUpadateTime[Len-1])
}

/*
用户的使用次数
*/
func (pkg *KoreaPackget) GetLimitById(pkgid int64, subpkgid int64) int64 {
	for i := 0; i < len(pkg.PackageId); i++ {
		if pkg.PackageId[i] == pkgid && pkg.SubPkgId[i] == subpkgid {
			return pkg.PackageUseTimes[i]
		}
	}
	return 0
}

/*
购买条件礼包or显示特殊条件礼包
*/
func (pkg *KoreaPackget) UpdateCondiPkgOne(pkgid int64) {
	if !pkg.GetCondHaveBuy(pkgid) {
		pkg.CondHaveBuy = append(pkg.CondHaveBuy, pkgid)
	}
}

/*
特殊条件礼包的显示添加
*/
func (pkg *KoreaPackget) InsertLimitTimeOne(pkgid int64, subpkgid int64, cur_time int64) {
	/*
		如果profile中没有过这个礼包的信息
	*/

	for i := 0; i < len(pkg.PackageId); i++ {
		if pkg.PackageId[i] == pkgid && pkg.SubPkgId[i] == subpkgid {
			pkg.UpdateCondiPkgOne(pkgid)
			return
		}
	}
	pkg.InsertId(pkgid, subpkgid, cur_time, 0)

}

/*
特殊条件礼包 关闭=购买 处理
*/
func (pkg *KoreaPackget) FullUseTimes(pkgid int64, subpkgid int64, times int64) {
	for i := 0; i < len(pkg.PackageId); i++ {
		if pkg.PackageId[i] == pkgid && pkg.SubPkgId[i] == subpkgid {
			pkg.PackageUseTimes[i] = times
			logs.Debug("the Package %d:%d has been closebuy", pkgid, subpkgid)
			return
		}
	}
	logs.Error("Cant find the package %d:%d", pkgid, subpkgid)
}
