package gamedata

import (
	"path/filepath"

	"sync"

	"fmt"
	"os"
	"strconv"

	"github.com/astaxie/beego/utils"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"
)

type HotDatas struct {
	Activity                 hotActivityData
	GvgConfig                hotGvgData
	LimitGoodConfig          hotLimitGood
	HeroFoundConfig          HeroFoundData
	RedPacketConfig          hotRedPacketData
	HotExchangeShopData      hotExchangeShopData
	HotKoreaPackge           hotKoreaPackge
	PackageFound             PackageFoundData
	HotStageLootExchangeData hotStageLootExchangeData
	HotBlackGachaData        BlackGachaData
	HotLimitHeroGachaData    HeroGachaData
	// ...
}

func init() {
	hotDataMngs = make([]hotDataMng, 0, 64)
	// Warn: 顺序很重要,不要改变
	hotDataMngs = append(hotDataMngs, hotDataMng{"severgroup.data", &hotServerGroupMng{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"sgactivity.data", &hotServerGroupActivity{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"servergroup.data", &hotActivityServerGroupData{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"hotactivitytime.data", &hotActivityMng{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"hotactivitydetail.data", &hotMarketActivity{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"channelgroup.data", &hotActivityChannelGroupData{}})
	// gvg
	hotDataMngs = append(hotDataMngs, hotDataMng{"gvghotdata.data", &hotGvgMng{}})

	// 招财猫
	hotDataMngs = append(hotDataMngs, hotDataMng{"moneygod.data", &hotMoneyCatActivity{}})
	// 限时商城
	hotDataMngs = append(hotDataMngs, hotDataMng{"limitgoods.data", &hotLimitGoodsMng{}})
	// 红包配置
	hotDataMngs = append(hotDataMngs, hotDataMng{"redpacket.data", &hotRedPacketMng{}})
	// 白盒宝箱
	hotDataMngs = append(hotDataMngs, hotDataMng{"hotgachasettings.data", &WhiteGachaSetings{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"hotnormalgacha.data", &WhiteNormalGacha{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"hotgachaspecial.data", &WhiteGachaSpecil{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"hotgachalowest.data", &WhiteGachaLowest{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"hotgachashow.data", &WhiteGachaShow{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"fall.data", &hotStageLootExchangeData{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"hotshop.data", &hotExchangeShopData{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"hotpackage.data", &hotKoreaPackge{}})
	// 黑盒宝箱
	hotDataMngs = append(hotDataMngs, hotDataMng{"boxsettings.data", &hotBlackGachaSettingsMng{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"boxshow.data", &hotBlackGachaShowMng{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"boxlowest.data", &hotBlackGachaLowestMng{}})
	// 限时神将
	hotDataMngs = append(hotDataMngs, hotDataMng{"hgrbox.data", &hotHeroGachaRaceChest{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"hgrrank.data", &hotHeroGachaRaceRank{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"hgrgachaoption.data", &hotHeroGachaRaceOption{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"hgrconfig.data", &hotHeroGachaRaceConfig{}})
	//TODO 幸运转盘
	hotDataMngs = append(hotDataMngs, hotDataMng{"wheelsettings.data", &LuckyWheelSetings{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"wheelgacha.data", &LuckyWheelGacha{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"wheelcost.data", &LuckyWheelCost{}})
	hotDataMngs = append(hotDataMngs, hotDataMng{"wheelshow.data", &LuckyWheelShow{}})
	mHotDataNotify = make(map[string]HandleHotDataNotify)
}

func GetHotDatas() *HotDatas {
	hotMutx.RLock()
	defer hotMutx.RUnlock()
	return hotDatas
}

func GetHotDataPath() string {
	workPath, _ := os.Getwd()
	workPath, _ = filepath.Abs(workPath)
	// initialize default configurations
	AppPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	appConfigPath := filepath.Join(AppPath, "confd")
	if workPath != AppPath {
		if utils.FileExists(appConfigPath) {
			os.Chdir(AppPath)
		} else {
			appConfigPath = filepath.Join(workPath, "confd")
		}
	}
	return filepath.Join(appConfigPath, GetHotDataRelPath())
}
func GetHotDataRelPath() string {
	return filepath.Join("hotdata", version.Version, game.Cfg.GetHotDataVerC())
}

type IHotDataMng interface {
	loadData(buffer []byte, datas *HotDatas) error
}

type hotDataMng struct {
	filename string
	instance IHotDataMng
}

var (
	hotDataMngs []hotDataMng
	hotDatas    *HotDatas
	hotMutx     sync.RWMutex
)

func loadHotGameDataFromInit(absPath, relPath string) {
	datas, ver, err := _loadHotGameData(absPath, relPath)
	if err != nil {
		panic(err)
	}
	setHotDatas(datas, ver)
	logs.Info("loadHotGameDataFromInit success path %s ver %v data %v", absPath, GetHotDataVerCfg(), hotDatas)
	setDataVer2Etcd()
}

func LoadHotGameDataFromUpdate(absPath, relPath string) error {
	datas, ver, err := _loadHotGameData(absPath, relPath)
	if err != nil {
		logs.Error("LoadHotGameDataFromUpdate err %s", err.Error())
		return err
	}
	setHotDatas(datas, ver)
	logs.Info("LoadHotGameDataFromUpdate success path %s ver %v data %v", absPath, GetHotDataVerCfg(), hotDatas)
	setDataVer2Etcd()
	HotDataValid = true

	//通知各模块响应热更数据变化
	muxHotDataNotify.RLock()
	for _, notify := range mHotDataNotify {
		notify(ver)
	}
	muxHotDataNotify.RUnlock()
	return nil
}

func _loadHotGameData(absPath, relPath string) (*HotDatas, *DataVerConf, error) {
	type Ver struct {
		Ver DataVerConf
	}
	var ver Ver
	cfg := config.NewConfigToml(filepath.Join(relPath, "proto_ver.toml"), &ver)
	if cfg == nil {
		logs.Error("hot gamedata load ver config nil")
		return nil, nil, fmt.Errorf("hot gamedata proto_ver.toml not exist")
	}

	datas := &HotDatas{}
	for _, di := range hotDataMngs {
		buffer, err := loadBin(filepath.Join(absPath, di.filename))
		if err != nil {
			return nil, nil, err
		}

		if err = di.instance.loadData(buffer, datas); err != nil {
			return nil, nil, err
		}
	}
	if err := datas.checkHeroGachaRace(); err != nil {
		return nil, nil, err
	}
	return datas, &ver.Ver, nil
}

func setHotDatas(datas *HotDatas, ver *DataVerConf) {
	hotMutx.Lock()
	defer hotMutx.Unlock()
	hotDatas = datas
	HotDataVerCfg = *ver
}

func GetHotDataVerCfg() *DataVerConf {
	hotMutx.RLock()
	defer hotMutx.RUnlock()
	return &HotDataVerCfg
}

func LoadHotDataVerFromEtcd() error {
	pv := loadParentHotDataVerFromEtcd()
	cv := loadChildHotDataVerFromEtcd()

	game.Cfg.HotDataVerC = pv
	if cv > 0 {
		game.Cfg.HotDataVerC = cv
	}
	logs.Info("loadHotDataVerFromEtcd verC %d", game.Cfg.HotDataVerC)
	return nil
}

func loadParentHotDataVerFromEtcd() (ver int) {
	prefix := GetHotDataEtcdRoot(game.Cfg.EtcdRoot, version.Version, fmt.Sprintf("%d", game.Cfg.Gid))
	key := fmt.Sprintf("%s/%s", prefix, etcd.KeyHotDataCurr)
	v, err := etcd.Get(key)
	if err != nil {
		logs.Debug("loadParentHotDataVerFromEtcd get key %s err %s", key, err.Error())
		return
	}
	c, err := strconv.Atoi(v)
	if err != nil {
		logs.Error("loadParentHotDataVerFromEtcd Atoi err %s %s", v, err.Error())
		return
	}
	if c <= 0 {
		logs.Debug("loadParentHotDataVerFromEtcd rec hotdataC unvalid: %d", c)
		return
	}
	return c
}

func loadChildHotDataVerFromEtcd() (ver int) {
	if len(game.Cfg.ShardId) <= 0 {
		return 0
	}
	prefix := GetHotDataEtcdRoot(game.Cfg.EtcdRoot, version.Version, fmt.Sprintf("%d", game.Cfg.Gid))
	key := fmt.Sprintf("%s/%d/%s", prefix, game.Cfg.ShardId[0], etcd.KeyHotDataCurr)
	v, err := etcd.Get(key)
	if err != nil {
		logs.Debug("loadChildHotDataVerFromEtcd get key %s err %s", key, err.Error())
		return
	}
	c, err := strconv.Atoi(v)
	if err != nil {
		logs.Error("loadChildHotDataVerFromEtcd Atoi err %s %s", v, err.Error())
		return
	}
	if c <= 0 {
		logs.Debug("loadChildHotDataVerFromEtcd rec hotdataC unvalid: %d", c)
		return
	}
	return c
}
func DebugSetActivityInfo(actId int, s, e int64) {
	hotMutx.Lock()
	defer hotMutx.Unlock()
	hotDatas.Activity.debugSetActivityTime(actId, s, e)
}

func GetHotDataEtcdRoot(etcdroot, version, gid string) string {
	return fmt.Sprintf("%s/%s/%s/%s", etcdroot, etcd.DirHotData, version, gid)
}

type HandleHotDataNotify func(ver *DataVerConf)

var (
	mHotDataNotify   map[string]HandleHotDataNotify
	muxHotDataNotify sync.RWMutex
)

func AddHotDataNotify(name string, notify HandleHotDataNotify) {
	muxHotDataNotify.Lock()
	defer muxHotDataNotify.Unlock()

	mHotDataNotify[name] = notify
}

func DelHotDataNotify(name string) {
	muxHotDataNotify.Lock()
	defer muxHotDataNotify.Unlock()
	delete(mHotDataNotify, name)
}
