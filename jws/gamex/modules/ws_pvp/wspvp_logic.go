package ws_pvp

import (
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type WsPvpRankPlayer struct {
	Acid      string `redis:"acid"`
	Rank      int
	Name      string `redis:"name"`
	Sid       int    `redis:"sid"`
	GuildName string `redis:"gname"`
	CorpLevel int    `redis:"clv"`
	SidStr    string
	Gs        int64 `redis:"allgs"`
}

func (w *WSPVPModule) GetTopN() []*WsPvpRankPlayer {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()
	resChan := make(chan wspvpReadResCommand, 1)
	cmd := wspvpReadCommand{
		commandType: WSPVP_PLAYER_GET_TOPN,
		resChan:     resChan,
	}
	select {
	case w.readChan <- cmd:
	case <-ctx.Done():
		logs.Error("wspvp module GetTopN channel time out")
		return nil
	}

	select {
	case cmd := <-resChan:
		return cmd.TopN
	case <-ctx.Done():
		logs.Error("wspvp module GetTopN resp time out")
		return nil
	}
}

func (w *WSPVPModule) GetBest9TopN() []*WsPvpRankPlayer {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()
	resChan := make(chan wspvpReadResCommand, 1)
	cmd := wspvpReadCommand{
		commandType: WSPVP_PLAYER_GET_BEST9_TOPN,
		resChan:     resChan,
	}
	select {
	case w.readChan <- cmd:
	case <-ctx.Done():
		logs.Error("wspvp module GetTopN best9 channel time out")
		return nil
	}

	select {
	case cmd := <-resChan:
		return cmd.TopN
	case <-ctx.Done():
		logs.Error("wspvp module GetTopN best9 resp time out")
		return nil
	}
}

func (w *WSPVPModule) UpdatePlayer(player *WSPVPInfo) {
	select {
	case w.saveChan <- player:
	default:
		logs.Warn("wspvp module updatePlayer channel time out")
	}
}

// TODO 是否需要调用
func (w *WSPVPModule) reloadTopNImmediately() {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()
	select {
	case w.reloadChan <- true:
	case <-ctx.Done():
		logs.Error("wspvp module reloadTopNImmediately channel time out")
	}
}
