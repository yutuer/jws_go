package gamedata

import (
	"github.com/golang/protobuf/proto"
	"strconv"
	"strings"
	"time"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	OrdinaryPackage = 0
	LevelPackage    = 1
	ConditonPackage = 2
	SpecialPackage  = 3
	Nolimit         = 0
	DayLimit        = 1
	WeekLimit       = 2
	MonthLimit      = 3
	AllLimit        = 4
)

type hotKoreaPackge struct {
	koreaPackgeData *ProtobufGen.HOTPACKAGE_ARRAY
}

type PackageInfo struct {
	PkgId    int
	SubPkgId int
}

type PackageFoundData struct {
	PackageIap map[int]bool //礼包表索引  当set使用
}

func (hkg *hotKoreaPackge) loadData(buffer []byte, datas *HotDatas) error {
	hkg.koreaPackgeData = &ProtobufGen.HOTPACKAGE_ARRAY{}
	dataList := &ProtobufGen.HOTPACKAGE_ARRAY{}
	if err := proto.Unmarshal(buffer, dataList); err != nil {
		return err
	}
	datas.PackageFound.PackageIap = make(map[int]bool, 0)
	/*
		维护礼包的IAPID，在give的时候进行判断
	*/
	for _, value := range dataList.Items {
		IapId := value.GetIAPID()
		tc := strings.Split(IapId, ",")
		for i := 0; i < len(tc); i++ {
			id1, err := strconv.Atoi(tc[i])
			logs.Debug("koreaPackage loadData Atoi err %v", err)
			datas.PackageFound.PackageIap[id1] = true
		}
		serverCfg, ok := datas.Activity.serverGroup[value.GetServerGroupID()]
		if !ok || !_checkServerShardValid(serverCfg) {
			continue
		}
		hkg.koreaPackgeData.Items = append(hkg.koreaPackgeData.Items, value)
	}
	datas.HotKoreaPackge = *hkg
	return nil
}

/*
获取有效的数据
*/
func (hkg *hotKoreaPackge) GetValidData() []*ProtobufGen.HOTPACKAGE {
	ret := make([]*ProtobufGen.HOTPACKAGE, 0)

	for _, value := range hkg.koreaPackgeData.GetItems() {
		if value.GetActivityValid() == uint32(1) {
			ret = append(ret, value)
		}
	}
	return ret
}

func (hkg *hotKoreaPackge) GetHotKoreaPackgeData() *ProtobufGen.HOTPACKAGE_ARRAY {
	return hkg.koreaPackgeData
}

/*
限购类型
*/
func (hkg *hotKoreaPackge) GetHotTimeType(pkgid int64, subpkgid int64) int64 {
	for _, value := range hkg.GetValidData() {
		if int64(value.GetHotPackageID()) == pkgid && int64(value.GetHotPackageSubID()) == subpkgid {
			return int64(value.GetLimitType())
		}
	}
	return 0
}

/*
开始时间
*/
func (hkg *hotKoreaPackge) GetHotStartTime(pkgid int64, subpkgid int64) int64 {
	logs.Debug("Get KoreaPackage %d:%d starttime", pkgid, subpkgid)
	for _, data := range hkg.GetValidData() {
		if int64(data.GetHotPackageID()) == pkgid && (int64(data.GetHotPackageSubID()) == subpkgid || subpkgid == -1) {
			if data.GetTimeType() == HotTime_Absolute {
				_ts, err := time.ParseInLocation("20060102_15:04", data.GetStartTime(), util.ServerTimeLocal)
				if err != nil {
					logs.Error("Convert Hot Korea Package's time error")
				}
				return _ts.Unix()
			} else if data.GetTimeType() == HotTime_Relative {
				// 转换为绝对时间
				if len(game.Cfg.ShardId) <= 0 { // multiplayer 不用加载
					continue
				}
				// 取index=0的shardId, 行不行!
				serverStartTime := game.ServerStartTime(game.Cfg.ShardId[0])

				beginTime := util.GetCurDayTimeAtHour(serverStartTime, HotTime_BeginHour)

				sdays, err := strconv.ParseFloat(data.GetStartTime(), 32)
				if err != nil {
					logs.Error("Convert Hot Korea Package's time error")
				}

				ts := time.Unix(int64(sdays)*util.DaySec+beginTime, 0).In(util.ServerTimeLocal)
				logs.Debug("Convert Absolute time begin %d-%d-%d %d:%d:%d  %v, %d",
					ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), ts.Location(), ts.Unix())
				return ts.Unix()
			} else {
				logs.Error("Illegale hot time type: %d", data.GetTimeType())
				continue
			}
		}
	}
	return 0
}

/*
限制次数
*/
func (hkg *hotKoreaPackge) GetHotTimeLimit(pkgid int64, subpkgid int64) int64 {
	for _, value := range hkg.GetValidData() {
		if int64(value.GetHotPackageID()) == pkgid && int64(value.GetHotPackageSubID()) == subpkgid {
			return int64(value.GetTimesLimit())
		}
	}
	return 0
}

/*
返回礼包
*/
func (hkg *hotKoreaPackge) GetHotPackage(pkgid int64, subpkgid int64) *ProtobufGen.HOTPACKAGE {
	if pkgid <= 0 && subpkgid <= 0{
		return nil
	}
	for _, value := range hkg.GetValidData() {
		if int64(value.GetHotPackageID()) == pkgid && (int64(value.GetHotPackageSubID()) == subpkgid || subpkgid == -1) {
			return value
		}
	}
	return nil
}

/*
返回礼包类型
*/
func (hkg *hotKoreaPackge) GetHotPackageType(pkgid int64) int64 {
	for _, value := range hkg.GetValidData() {
		if int64(value.GetHotPackageID()) == pkgid {
			return int64(value.GetHotPackageType())
		}
	}
	logs.Debug("can not get package type")
	return 0
}
