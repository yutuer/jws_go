package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	//"vcs.taiyouxi.net/platform/planx/util/logs"
)

type LvlEnmy struct {
	ID    string
	Count int
}

type LvlEnmyCfg map[string][]LvlEnmy

var (
	gdLvlEnemyConfig LvlEnmyCfg
)

func GetLevelEnemyConfig(levelid string) []LvlEnmy {
	//logs.Trace("GetLevelEnemyConfig has %v", gdLvlEnemyConfig)
	//logs.Trace("GetLevelEnemyConfig get %s", levelid)
	if e, ok := gdLvlEnemyConfig[levelid]; ok {
		//logs.Trace("GetLevelEnemyConfig got %v", e)
		return e[:]
	}
	return nil
}

//loadLevelEnemyConfigi 读取关卡中配置什么兵有多少种
func loadLevelEnemyConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	// 读取武器装备表
	buffer, err := loadBin(filepath)
	errcheck(err)

	lecArray := &ProtobufGen.LEVELENEMYCONFIG_ARRAY{}
	err = proto.Unmarshal(buffer, lecArray)
	errcheck(err)

	lecs := lecArray.GetItems()
	gdLvlEnemyConfig = make(LvlEnmyCfg)
	for _, lec := range lecs {
		var enemys []LvlEnmy
		if e, ok := gdLvlEnemyConfig[lec.GetID()]; ok {
			enemys = e
		} else {
			enemys = make([]LvlEnmy, 0, 10)
		}
		enemys = append(enemys, LvlEnmy{lec.GetEnemyID(), int(lec.GetCount())})
		gdLvlEnemyConfig[lec.GetID()] = enemys
		//logs.Trace("Level ID:%s, Enemy ID:%s, Count:%d", *lec.ID, *lec.EnemyID, *lec.Count)
	}
	//logs.Trace("loadLevelEnemyConfig: %v", gdLvlEnemyConfig)
}
