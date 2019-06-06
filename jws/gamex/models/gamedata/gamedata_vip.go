package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
)

type VIPConfig struct {
	VIPLevel                    uint32
	RMBpoints                   uint32
	EnergyPurchaseLimit         uint32
	SprintPurchaseLimit         int32
	SCPurchaseLimit             uint32
	StoreRefreshLimit           uint32
	PvpStoreRefreshLimit        uint32
	TpvpStoreRefreshLimit       uint32
	BossStoreRefreshLimit       uint32
	GuildStoreRefreshLimit      uint32
	ExpeditionStoreRefreshLimit uint32
	HDPStoreRefreshLimit        uint32
	TPVPTimeLimit               int32
	SimplePVPTimeLimit          int32
	PrivilegeId                 int32
	SweepValid                  bool // 废弃
	SweepTenValid               bool
	SweepResourceValid          bool
	OrdinaryNotStarSweep        bool
	EliteNotStarSweep           bool
	HellNotStarSweep            bool
	BossFightSweep              bool
	GoldLevelSweep              bool
	IronLevelSweep              bool
	DcLevelSweep                bool
	EliteStagePurchase          uint32
	HellStagePurchase           uint32
	GoldLevelAdd                float32
	IronLevelAdd                float32
	DcLevelAdd                  float32
	WorldBossSweep              bool
	GuildSignMaxCount           int
	VIPDailyGift                givesData
	StarHcUpLimitDaily          int
	VIPGachaLimit               bool
	GrowFund                    bool
	DGVipNormalTimes            uint32
	DGVipAdv                    bool
	GateEnemyBuff               bool
	BaoZiPurchaseLimit          uint32
	GuildWorshipLimit           uint32
	WushuangShopRefreshLImit    uint32
	WsChallengeLimit            int32
	CSRobBuildCarTimes          uint32            //劫营夺粮运粮次数
	CSRobCarKeep                uint32            //劫营夺粮运粮时长（分钟）
	CSRobRobTimes               uint32            //劫营夺粮掠夺次数
	CSRobHelpLimit              uint32            //劫营夺粮协助次数
	StoreRefreshLimitTable      map[uint32]uint32 //商店刷新次数
	WorldBosstimes              uint32
	SurplusGachaLimit           [3]int
}

var (
	gdVipConfig        []VIPConfig
	gdPrivilegeBuy2vip map[int32]uint32
	gdMaxVipLevel      uint32
)

func loadVIPConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.VIPSETTINGS_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()

	gdVipConfig = make([]VIPConfig, 0, len(lv_data))
	gdPrivilegeBuy2vip = map[int32]uint32{}

	for _, c := range lv_data {
		cfg := VIPConfig{
			VIPLevel:                    c.GetVIPLevel(),
			RMBpoints:                   c.GetRMBpoints(),
			EnergyPurchaseLimit:         c.GetEnergyPurchaseLimit(),
			SprintPurchaseLimit:         c.GetSprintPurchaseLimit(),
			SCPurchaseLimit:             c.GetSCPurchaseLimit(),
			StoreRefreshLimit:           c.GetStoreRefreshLimit(),
			PvpStoreRefreshLimit:        c.GetPvpStoreRefreshLimit(),
			TpvpStoreRefreshLimit:       c.GetTpvpStoreRefreshLimit(),
			BossStoreRefreshLimit:       c.GetBossStoreRefreshLimit(),
			GuildStoreRefreshLimit:      c.GetGuildStoreRefreshLimit(),
			ExpeditionStoreRefreshLimit: c.GetExpeditionStoreRefreshLimit(),
			HDPStoreRefreshLimit:        c.GetHDPStoreRefreshLimit(),
			TPVPTimeLimit:               c.GetTPVPTimeLimit(),
			PrivilegeId:                 c.GetPrivilegeID(),
			SweepValid:                  c.GetSweep() > 0,
			SweepTenValid:               c.GetTenSweep() > 0,
			SweepResourceValid:          c.GetSweepResource() > 0,
			OrdinaryNotStarSweep:        c.GetNormalSweep() > 0,
			EliteNotStarSweep:           c.GetEliteSweep() > 0,
			HellNotStarSweep:            c.GetHardSweep() > 0,
			BossFightSweep:              c.GetBossFightSweep() > 0,
			GoldLevelSweep:              c.GetGoldLevelSweep() > 0,
			IronLevelSweep:              c.GetIronLevelSweep() > 0,
			DcLevelSweep:                c.GetDestinyLevelSweep() > 0,
			EliteStagePurchase:          c.GetElitePurchase(),
			HellStagePurchase:           c.GetHardPurchase(),
			GoldLevelAdd:                c.GetGoldLevelAdd(),
			IronLevelAdd:                c.GetIronLevelAdd(),
			DcLevelAdd:                  c.GetDCLevelAdd(),
			WorldBossSweep:              c.GetBOSSFightSweep() > 0,
			GuildSignMaxCount:           int(c.GetGuildSignLimit()),
			StarHcUpLimitDaily:          int(c.GetStarupUseHCLimit()),
			VIPGachaLimit:               c.GetVIPGachaLimit() > 0,
			SimplePVPTimeLimit:          c.GetPVPTimeLimit(),
			GrowFund:                    c.GetGrowFund() > 0,
			DGVipNormalTimes:            c.GetDGVIPTrain(),
			DGVipAdv:                    c.GetDGVIPContinuityTrain() > 0,
			GateEnemyBuff:               c.GetGEBuffEncourage() > 0,
			BaoZiPurchaseLimit:          c.GetHighEnergyLimitt(),
			GuildWorshipLimit:           c.GetGuildSignLimit(),
			WushuangShopRefreshLImit:    c.GetWSPVPStoreRefreshLimit(),
			WsChallengeLimit:            c.GetWSPVPTimeLimit(),
			CSRobBuildCarTimes:          c.GetCropstimes(),
			CSRobCarKeep:                c.GetCropsDuration(),
			CSRobRobTimes:               c.GetRobtimes(),
			CSRobHelpLimit:              c.GetAssisttimes(),
			StoreRefreshLimitTable:      map[uint32]uint32{},
			WorldBosstimes:              c.GetWorldBosstimes(),
			SurplusGachaLimit:           [3]int{int(c.GetSurplusGacha3()), int(c.GetSurplusGacha4()), int(c.GetSurplusGacha5())},
		}
		cfg.VIPDailyGift.AddItem(c.GetVIPDailyGift1(), c.GetCount1())
		cfg.VIPDailyGift.AddItem(c.GetVIPDailyGift2(), c.GetCount2())
		cfg.VIPDailyGift.AddItem(c.GetVIPDailyGift3(), c.GetCount3())
		cfg.VIPDailyGift.AddItem(c.GetVIPDailyGift4(), c.GetCount4())
		gdVipConfig = append(gdVipConfig, cfg)
		gdPrivilegeBuy2vip[c.GetPrivilegeID()] = c.GetVIPLevel()
		if c.GetVIPLevel() > gdMaxVipLevel {
			gdMaxVipLevel = c.GetVIPLevel()
		}

		for _, storeCfg := range c.GetStoreRefresh_Table() {
			cfg.StoreRefreshLimitTable[storeCfg.GetStoreID()] = storeCfg.GetRefreshTime()
		}
	}

	//logs.Trace("gdVipConfig %v", gdVipConfig)

}

func GetVIPCfg(vip int) *VIPConfig {
	if vip < 0 || vip >= len(gdVipConfig) {
		return nil
	}
	return &gdVipConfig[vip]
}

func Privilege2Vip(privilegeId int32) (uint32, bool) {
	vip, ok := gdPrivilegeBuy2vip[privilegeId]
	if ok {
		return vip, true
	} else {
		return 0, false
	}
}

func GetMaxVipLevel() uint32 {
	return gdMaxVipLevel
}
