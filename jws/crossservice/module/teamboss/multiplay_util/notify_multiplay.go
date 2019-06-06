package multiplay_util

import (
	"encoding/json"

	"fmt"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/crossservice/util/http_util"
	"vcs.taiyouxi.net/jws/multiplayer/helper"
)

const tb_stop_url = "/crossservice/tb/stop"

var g *gin.Engine

type TBStartFightData helper.TBStartFightData
type TeamBossStopInfo helper.TeamBossStopInfo

func NotifyMultiplay(data *TBStartFightData) ([]byte, error) {
	matchToken := helper.TeamBossToken
	return GetNotify(matchToken).TeamBossNotify(helper.TBStartFightData(*data))
}

func RegEtcd(root string, groupID uint32) {
	http_util.RegETCD(root, genStopUrl(groupID), groupID)
}

func RegTBHttpHanlde(f func(c *gin.Context), groupID uint32) {
	http_util.POST(genStopUrl(groupID), f)
}

func GenTeamBossMultiplayInfo(data []byte) (string, string, error) {
	info := &helper.TeamBossCreateinfo{}
	err := json.Unmarshal(data, info)
	return info.WebsktUrl, info.GlobalRoomID, err
}

func genStopUrl(groupID uint32) string {
	return fmt.Sprintf("%s/%d", tb_stop_url, groupID)
}
