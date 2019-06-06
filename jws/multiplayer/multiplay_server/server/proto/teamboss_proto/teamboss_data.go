package teamboss_proto

import (
	gamexHelper "vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/multiplayer/helper"
)

type TBGameDatas struct {
	PlayerDatas []*gamexHelper.Avatar2ClientByJson
}

func (g *TBGameDatas) AddPlayer(data *helper.TBStartFightData) {
	for _, item := range data.Info {
		g.PlayerDatas = append(g.PlayerDatas, item)
	}
}
