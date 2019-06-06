package hero_diff

import "vcs.taiyouxi.net/platform/planx/util/logs"

func (hd *heroDiffModule) DebugCleanRank() {
	logs.Debug("herodiff debug clean rank")
	hd.deleteRank()
}
