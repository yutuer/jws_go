package main

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/x/tiprotogen/log"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/tools/removeAccount/config"
	"fmt"
	"strconv"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//type PerInfo struct {
//	VirtualMoney []string
//	Mat          []string
//	Xp           []string
//	Jad          map[string]int64
//	Acid         string
//	RedisAddr    string
//	RedisNum     int
//}
//
//var info []PerInfo

//func init() {
	//info = []PerInfo{
	//	{
	//		[]string{},
	//		[]string{},
	//		[]string{
	//			"USEI_XP_5",
	//			"USEI_XP_6",
	//		},
	//		map[string]int64{
	//			"JD_PURPLE_8": 1371,
	//			"JD_RED_6":    12337,
	//			"JD_CHING_6":  12345,
	//			"JD_BLUE_6":   12345,
	//			"JD_GREEN_6":  12346,
	//			"JD_YELLOW_6": 12346,
	//		},
	//		"200:1076:1ac05084-e02a-4e07-bc5a-97e73dfc143d",
	//		"10.33.4.15:16379",
	//		2,
	//	},
	//	{
	//		[]string{},
	//		[]string{},
	//		[]string{
	//			"USEI_XP_2",
	//			"USEI_XP_3",
	//			"USEI_XP_4",
	//			"USEI_XP_5",
	//			"USEI_XP_6",
	//		},
	//		map[string]int64{
	//			"JD_RED_9":    457,
	//			"JD_PURPLE_9": 457,
	//			"JD_YELLOW_9": 457,
	//			"JD_GREEN_9":  457,
	//			"JD_BLUE_9":   457,
	//			"JD_CHING_9":  457,
	//		},
	//		"200:1101:bc0666a9-d807-42b8-ab06-08d419cefd84",
	//		"10.33.4.12:16379",
	//		3,
	//	},
	//	{
	//		[]string{},
	//		[]string{},
	//		[]string{
	//			"USEI_XP_2",
	//			"USEI_XP_3",
	//			"USEI_XP_4",
	//			"USEI_XP_5",
	//			"USEI_XP_6",
	//		},
	//		map[string]int64{
	//			"JD_RED_9":    457,
	//			"JD_PURPLE_9": 457,
	//			"JD_YELLOW_9": 457,
	//			"JD_GREEN_9":  457,
	//			"JD_BLUE_9":   457,
	//			"JD_CHING_9":  457,
	//		},
	//		"200:1076:8ca91bc4-4adb-4378-ad65-a49be22d1066",
	//		"10.33.4.15:16379",
	//		2,
	//	},
	//	{
	//		[]string{},
	//		[]string{},
	//		[]string{
	//			"USEI_XP_2",
	//			"USEI_XP_3",
	//			"USEI_XP_4",
	//			"USEI_XP_5",
	//			"USEI_XP_6",
	//		},
	//		map[string]int64{
	//			"JD_RED_9":    457,
	//			"JD_PURPLE_5": 36964,
	//			"JD_YELLOW_5": 37029,
	//			"JD_CHING_5":  37037,
	//			"JD_BLUE_4":   110943,
	//			"JD_GREEN_4":  111058,
	//		},
	//		"200:1041:e09efe9a-f394-4a79-88c7-815ffd12fd77",
	//		"10.33.4.11:16379",
	//		13,
	//	},
	//	{
	//		[]string{},
	//		[]string{},
	//		[]string{
	//			"USEI_XP_5",
	//			"USEI_XP_6",
	//		},
	//		map[string]int64{
	//			"JD_RED_5":    37037,
	//			"JD_PURPLE_5": 37037,
	//			"JD_YELLOW_5": 37037,
	//			"JD_GREEN_5":  37037,
	//			"JD_BLUE_5":   37037,
	//			"JD_CHING_5":  37037,
	//		},
	//		"200:1041:cba56c96-62cb-425a-8251-a1b6a0c1cdfb",
	//		"10.33.4.11:16379",
	//		13,
	//	},
	//	{
	//		[]string{
	//			"VI_BAOZI",
	//		},
	//		[]string{
	//			"MAT_GWStone_3",
	//			"MAT_GWC_ZGL",
	//			"MAT_GWStone_4",
	//			"MAT_GWStone_5",
	//		},
	//		[]string{
	//			"USEI_XP_4",
	//			"USEI_XP_3",
	//			"USEI_XP_5",
	//			"USEI_XP_2",
	//			"USEI_XP_6",
	//		},
	//		map[string]int64{
	//			"JD_BLUE_9":   457,
	//			"JD_CHING_9":  457,
	//			"JD_GREEN_9":  457,
	//			"JD_PURPLE_9": 457,
	//			"JD_RED_9":    457,
	//			"JD_YELLOW_9": 457,
	//		},
	//		"202:3086:3f173760-fa58-40b0-82de-2308e750e002",
	//		"10.33.4.183:6379",
	//		9,
	//	},
	//}

	//info = []PerInfo{
	//	{
	//		[]string{
	//			"VI_BAOZI",
	//		},
	//		[]string{
	//			"MAT_GWStone_3",
	//			"MAT_GWC_ZGL",
	//			"MAT_GWStone_4",
	//			"MAT_GWStone_5",
	//		},
	//		[]string{
	//			"USEI_XP_4",
	//			"USEI_XP_3",
	//			"USEI_XP_5",
	//			"USEI_XP_2",
	//			"USEI_XP_6",
	//		},
	//		map[string]int64{
	//			"JD_BLUE_9":   457,
	//			"JD_CHING_9":  457,
	//			"JD_GREEN_9":  457,
	//			"JD_PURPLE_9": 457,
	//			"JD_RED_9":    457,
	//			"JD_YELLOW_9": 457,
	//		},
	//		"1:15:3e0e19b4-3fff-46f4-808a-3c7f87dae148",
	//		"10.222.3.10:16379",
	//		11,
	//	},
	//}
//}

func main() {
	config.LoadConfig("conf/config.toml")
	fmt.Println(config.RemoveConfig)
	gamedata.LoadGameData("")
	for idx, _ := range config.RemoveConfig.RmoveInfoConfig.PerInfo {
		tInfo := config.RemoveConfig.RmoveInfoConfig.PerInfo[idx]
		acid := tInfo.Acid
		redisAddress := tInfo.Redis
		redisNum := tInfo.RedisDb
		password := tInfo.Password
		driver.SetupRedisForSimple(redisAddress, redisNum, password, false)
		acID, _ := db.ParseAccount(acid)
		tProfile := account.NewProfile(acID)
		profile := &tProfile
		err := profile.DBLoad(false)
		if err != nil {
			log.Err("profile Dbload error %v", err)
			return
		}
		baginfo := profile.GetJadeBag()

		var i uint32
		profile.GetEquipJades().Jades = []uint32{}
		profile.GetDestGeneralJades().DestinyGeneralIds = []int{}
		profile.GetDestGeneralJades().DestinyGeneralJade = []uint32{}
		profile.GetEnergy().Value = 0
		mp := tInfo.Jade
		for i = 0; i <= baginfo.NextId; i++ {
			item, ok := baginfo.JadesMap[i]
			if !ok {
				continue
			}
			//item.JadeExp = 0
			item.CountNotInBag = 0
			for id, _ := range mp {
				if item.TableID == id && mp[id].Number > 0 && item.Count > 0 {
					if mp[id].Number >= item.Count {
						log.Debug("remove %s %d", item.TableID, item.Count)
						tValue := mp[id]
						//mp[id].Number -= item.Count
						tValue.Number -= item.Count
						mp[id] = tValue
						item.Count = 0
					} else {
						log.Debug("remove %s %d", item.TableID, mp[id])
						item.Count -= mp[id].Number
						mp[id] = config.JadeNum{0}
					}
				}
			}
			if item.Count > 0 {
				baginfo.JadesMap[i] = item
			} else {
				delete(baginfo.JadesMap, i)
			}
		}

		heroPiece := profile.GetHero()
		for i ,value := range tInfo.HeroPiece{
			id,err := strconv.Atoi(i)
			if err != nil{
				logs.Error("Convert HeroId String to Int error")
				return
			}
			if heroPiece.HeroStarPiece[id] < uint32(value.Number){
				logs.Error("hero piece is not enough")
				return
			}
			heroPiece.HeroStarPiece[id] -= uint32(value.Number)
		}

		//log.Debug(mp)
		for _, value := range tInfo.VirtualMoney {
			it := helper.SCId(value)
			profile.GetSC().Currency[it] = 0
			log.Debug("remove %s all", value)
		}

		cb := redis.NewCmdBuffer()
		if err := profile.DBSave(cb, true); err != nil {
			log.Err("profile Dbsave error %v", err)
			return
		}

		rc := driver.GetDBConn()

		if rc.IsNil() {
			log.Err("Save Error:Account DB Save, cant get redis conn")
			return
		}
		//logs.Trace("[%s]RedisSave %s", p.AccountID.String(), cb.String())

		_, err = rc.DoCmdBuffer(cb, true)
		if err != nil {
			log.Err("Save Error:DoCmdBuffer, %s", err)
			return
		}

		log.Debug("change profile info success")
		tPlayerbag := account.NewPlayerBag(acID)
		playerbag := &tPlayerbag
		err = playerbag.DBLoad(false)
		if err != nil {
			log.Err("bag Dbload error %v", err)
			return
		}

		for _, value := range tInfo.Xp {
			playerbag.RemoveWithFixedID(value)
			log.Debug("remove %s all", value)
		}

		for _, value := range tInfo.Mat {
			playerbag.RemoveWithFixedID(value)
			log.Debug("remove %s all", value)
		}
		cb1 := redis.NewCmdBuffer()
		if err := playerbag.DBSave(cb1, true); err != nil {
			log.Err("playerbag Dbsave error %v", err)
			return
		}

		_, err = rc.DoCmdBuffer(cb1, true)
		if err != nil {
			log.Err("Save Error:DoCmdBuffer, %s", err)
			return
		}

		log.Debug("change bag info success %s", tInfo.Acid)
		rc.Close()
		driver.ShutdownRedis()
	}
}
