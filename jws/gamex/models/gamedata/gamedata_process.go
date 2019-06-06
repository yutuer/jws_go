package gamedata

import "vcs.taiyouxi.net/platform/planx/util/logs"

// 最终处理
func processDataBeforeAll() {
	for id, avatarID := range gdPlayerHeroID2IDx {
		for _, d := range gdItems {
			if d.GetType() == "WholeChar" && d.GetAttrType() == id {
				gdWholeCharId[avatarID] = d.GetID()
			}
		}
	}
	logs.Warn("WholeChar %v", gdWholeCharId)
}
