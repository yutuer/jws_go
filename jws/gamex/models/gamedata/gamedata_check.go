package gamedata

import "vcs.taiyouxi.net/jws/gamex/models/gamedata/check"

// XXX By Fanyang 添加更多的检查项

func checkGameData() {
	// 测试战队等级相关的数据
	maxCorpLv := GetCommonCfg().GetCorpLevelUpperLimit()
	check.InitByCorpData(int(maxCorpLv))

	// 基本战队信息检查
	check.ByCorpLv("GetCorpLvConfig", func(lv int) bool {
		return nil != GetCorpLvConfig(uint32(lv))
	})

	// recover
	check.RecoverCheck("Check_RecoverRetialEmpty", GetAllRecoverIds(), func(recoverId uint32) bool {
		cfg := GetRecoverCfg(recoverId)
		if cfg == nil {
			return false
		}
		if len(cfg.Retails) <= 0 {
			return false
		}
		return true
	})

	check.RecoverCheck("Check_RecoverQuest", GetAllRecoverIds(), func(recoverId uint32) bool {
		cfg := GetRecoverCfg(recoverId)
		if cfg == nil {
			return false
		}
		for _, r := range cfg.Retails {
			if r.GetRecoverType() == RecoverTyp_Quest {
				quest := GetQuestNeedCheckById(r.GetRecoverPara())
				if quest == nil {
					return false
				}
				ac_condition := quest.GetAccCon_Table()
				if len(ac_condition) <= 0 {
					return false
				}
				cond := ac_condition[0]
				if cond.GetACType() == 999 {
					return false
				}
			}
		}
		return true
	})

	// Gacha
	//check.ByCorpLv("Gacha 0", func(lv int) bool {
	//	return nil != GetGachaData(uint32(lv), 0)
	//})
	//check.ByCorpLv("Gacha 1", func(lv int) bool {
	//	return nil != GetGachaData(uint32(lv), 1)
	//})
	//check.ByCorpLv("Gacha 2", func(lv int) bool {
	//	return nil != GetGachaData(uint32(lv), 2)
	//})

}
