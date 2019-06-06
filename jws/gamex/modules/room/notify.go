package room

import (
	"encoding/json"
	"fmt"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/codec"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/jws/gamex/modules/room/info"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	gveHelper "vcs.taiyouxi.net/jws/multiplayer/helper"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (r *module) notify() {
	if len(r.newRooms) > 0 || len(r.delRooms) > 0 {
		sync := player_msg.RoomsSyncInfo{
			RoomNew: r.newRooms,
			RoomDel: r.delRooms,
		}
		msg := &servers.Request{
			Code:     player_msg.PlayerMsgRooms,
			RawBytes: codec.Encode(sync),
		}
		r.newRooms = r.newRooms[0:0]
		r.delRooms = r.delRooms[0:0]

		//go func(msg *servers.Request) {
		logs.Trace("Start Notify All Room Listener")
		for acID, channel := range r.playerChanMap {
			SendMsg(acID, msg, channel)
		}
		logs.Trace("Stop Notify All Room Listener")
		//}(msg)
	}
}

func SendMsg(acID string, msg *servers.Request, channel chan<- servers.Request) {
	// FIXME by YZH no wait fangyang
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case channel <- *msg:
	case <-ctx.Done():
		logs.Error("[RoomSync] SendMsg chann full, cmd put timeout, this msg %v",
			msg)
		return
	}
	//logs.Trace("[RoomSync] SendMsg %s %v", accountID, msg)
}

func NotifyRoomEvent(acID string, room *info.Room, typ int) {
	player_msg.Send(acID, player_msg.PlayerMsgRoomEvent, player_msg.RoomEventNotify{
		Type: typ,
		Room: room.ToData(),
	})
}

func NotifyRoomEventToAll(room *info.Room, typ int) {
	players := room.GetPlayers()
	for _, p := range players {
		NotifyRoomEvent(p.AcID, room, typ)
	}
}

func notifyMatchServer(data gveHelper.FenghuoValue, url string) (int, []byte, error) {
	d, _ := json.Marshal(data)

	token := uutil.JwsCfg.MatchToken
	if token == "" {
		token = gveHelper.MatchDefaultToken
	}
	url += fmt.Sprintf("?token=%s", token)

	//logs.Trace("Fenghuo Room json %s", string(d))
	return util.HttpPostWCode(url, util.JsonPostTyp, d)
}
