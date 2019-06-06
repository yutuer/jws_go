package room

import (
	"math/rand"

	"encoding/json"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/room/info"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/errorcode"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/uuid"

	"fmt"

	"time"

	gveHelper "vcs.taiyouxi.net/jws/multiplayer/helper"
)

//FIXME by YZH FenghuoRoom cmd errorCode的模式的改造
//FIXME by YZH FenghuoRoom 双方都死亡的情况,应该直接解散房间
func (r *module) processCmd(c *cmd) *cmd {
	switch c.Type {
	case cmdTypeGet:
		return r.getAll(c.P1, c.P2)
	case cmdTypeAttachObserver:
		return r.attachObserver(c.Account, c.PlayerChan)
	case cmdTypeUpdateRoom:
		return r.updateRoom(c.Account, c.SimpleInfo, c.Rooms[0])
	case cmdTypeEnterRoom:
		return r.enterRoom(c.Account, c.SimpleInfo, c.P1, c.PS)
	case cmdTypeLeaveRoom:
		return r.leaveRoom(c.Account, c.PS, c.P1)
	case cmdTypeChangeMaster:
		return r.changeRoomMaster(c.Account, c.PS, c.P1)
	case cmdTypeReady:
		return r.ready(c.Account, c.P1, c.Avatar)
	case cmdTypeCancelReady:
		return r.cancelReady(c.Account, c.P1)
	case cmdTypeStartFight:
		return r.startFight(c.Rnd, c.Account, c.P1, c.P2, c.PB)
	case cmdTypeMakeReward:
		return r.makeReward(c.Account, c.P1, c.P2)
	case cmdTypeGetRoomInfo:
		return r.getRoomInfo(c.Account, c.P1)

	}

	return nil
}

func (r *module) GetRoomInfo(ctx context.Context, acID string, roomNum int) (info.Room, error) {
	res := r.sendCmd(ctx, &cmd{
		Type:    cmdTypeGetRoomInfo,
		Account: acID,
		P1:      roomNum,
	})

	if res == nil {
		return info.Room{}, info.RoomErr_CTX_TIMEOUT
	}

	if res.Code == 0 {
		return res.Rooms[0], nil
	} else {
		return info.Room{}, info.RoomErr_GetRoomInfo_NotFound
	}
}

func (r *module) getRoomInfo(acID string, roomNum int) *cmd {
	room, ok := r.roomNumMap[roomNum]
	res := new(cmd)
	if ok {
		res.Rooms = make([]info.Room, 1, 1)
		res.Rooms[0] = *room
	} else {
		res.Code = ROOM_WARN_GETROOMINFO_NUM_WRONG
	}
	return res
}

func (r *module) GetAll(ctx context.Context, startIdx, maxCount int) ([]info.Room, error) {
	res := r.sendCmd(ctx, &cmd{
		Type: cmdTypeGet,
		P1:   startIdx,
		P2:   maxCount,
	})

	if res == nil {
		return nil, info.RoomErr_CTX_TIMEOUT
	}

	return res.Rooms[:], nil
}

func (r *module) getAll(startIdx, maxCount int) *cmd {
	if maxCount == 0 {
		maxCount = 32
	}
	res := new(cmd)
	res.Rooms = make([]info.Room, 0, maxCount)
	if startIdx < 0 || startIdx >= len(r.roomNumArray) {
		return res
	}

	for i := startIdx; i < len(r.roomNumArray) && i < maxCount; i++ {
		num := r.roomNumArray[i]
		room, ok := r.roomNumMap[num]
		if ok {
			res.Rooms = append(res.Rooms, *room)
		}
	}
	return res
}

func (r *module) AttachObserve(ctx context.Context,
	acID string, channel chan<- servers.Request) ([]info.Room, errorcode.ErrorCode) {
	res := r.sendCmd(ctx, &cmd{
		Type:       cmdTypeAttachObserver,
		Account:    acID,
		PlayerChan: channel,
	})

	if res == nil {
		return nil, info.RoomErrCode_CTX_TIMEOUT
	}

	return res.Rooms[:], nil
}

func (r *module) attachObserver(acID string, channel chan<- servers.Request) *cmd {
	res := new(cmd)
	//TODO by YZH FenghuoRoom, 这里首次监听返回固定32个房间的模式需要改进,需要分页,分批逐步发送给玩家
	res.Rooms = make([]info.Room, 0, 32)

	num2Client := 0

	for _, room := range r.roomNumMap {
		num2Client++
		res.Rooms = append(res.Rooms, *room)
		if num2Client > 32 {
			break
		}
	}
	r.playerChanRegChan <- playerChanRegInfo{
		AccountID: acID,
		Chan:      channel,
	}
	return res
}

func (r *module) DetachObserve(ctx context.Context, acID string) {
	r.playerChanRegChan <- playerChanRegInfo{
		AccountID: acID,
		IsDel:     true,
	}
}

func (r *module) UpdateRoom(ctx context.Context,
	acID string,
	simpleInfo FenghuoProfile,
	room info.Room) (info.Room, errorcode.ErrorCode) {
	res := r.sendCmd(ctx, &cmd{
		Type:       cmdTypeUpdateRoom,
		Account:    acID,
		Rooms:      []info.Room{room},
		SimpleInfo: simpleInfo,
	})

	if res == nil {
		return info.Room{}, info.RoomErrCode_CTX_TIMEOUT
	}

	return res.Rooms[0], nil
}

func (r *module) updateRoom(acID string, profile FenghuoProfile, room info.Room) *cmd {
	res := new(cmd)
	res.Rooms = make([]info.Room, 1, 1)

	if room.Num > 0 {
		// Update
		roomCurr, ok := r.roomNumMap[room.Num]
		if ok {
			pl := roomCurr.GetPlayerByID(acID)
			pl.AvatarID = profile.AvatarID
			logs.Debug("room update for avatarID by %v", room)
			res.Rooms[0] = *roomCurr
			res.Rooms[0].ToData()
			NotifyRoomEventToAll(roomCurr, info.RoomEventUpdate)
		} else {
			logs.Error("Fenghuo updateRoom for player avatar update")
		}
	} else {
		// Create
		room.ID = uuid.NewV4().String()
		room.Num = r.allocRoomNum()
		//room.MaxPlayerCount = gamedata.FegnhuoRoomMaxPlayer
		room.RoomStat = info.RoomStatWaitting
		room.AddPlayer(&info.PlayerInRoom{
			AcID:     acID,
			Name:     profile.Name,
			AvatarID: profile.AvatarID,
			CorpLv:   profile.CorpLv,
			Gs:       profile.Gs,
		})
		r.roomNumMap[room.Num] = &room
		r.playerSyncRoomCmdChan <- cmdSyncRoom{
			NewRoom: room.ToData(),
		}
		res.Rooms[0] = room
		res.Rooms[0].ToData()
		logs.Debug("Fenghuo updateRoom create by %v", room)
	}

	return res
}

func (r *module) EnterRoom(ctx context.Context,
	acID string,
	profile FenghuoProfile,
	roomNum int, roomID string) (info.Room, errorcode.ErrorCode) {
	res := r.sendCmd(ctx, &cmd{
		Type:       cmdTypeEnterRoom,
		Account:    acID,
		P1:         roomNum,
		PS:         roomID,
		SimpleInfo: profile,
	})

	if res == nil {
		return info.Room{}, info.RoomErrCode_CTX_TIMEOUT
	}

	if res.ErrCode == nil {
		return res.Rooms[0], res.ErrCode
	}

	return info.Room{}, res.ErrCode
}

func (r *module) enterRoom(acID string, profile FenghuoProfile, roomNum int, roomID string) *cmd {
	var res cmd

	room, ok := r.roomNumMap[roomNum]
	if ok {
		if room.ID != roomID {
			res.ErrCode = info.RoomWarnCode_ENTER_ROOM_NUM_WRONG
			return &res
		}

		if !room.IsJoinable() {
			logs.Error("room enter protocol called when room is not joinable")
			res.ErrCode = info.RoomWarnCode_ENTER_ROOM_NOTJOINABLE
			return &res
		}
		p := room.GetPlayerInRoom(acID)
		if p == nil {

			//logs.Trace("enterRoom cur:%d max:%d", room.GetPlayerLen(), room.MaxPlayerCount)
			if room.GetPlayerLen() >= gamedata.FegnhuoRoomMaxPlayer {
				res.ErrCode = info.RoomWarnCode_ENTER_ROOM_FULL
				return &res
			}
			room.AddPlayer(&info.PlayerInRoom{
				AcID:     acID,
				Name:     profile.Name,
				AvatarID: profile.AvatarID,
				CorpLv:   profile.CorpLv,
				Gs:       profile.Gs,
			})
			res.Rooms = make([]info.Room, 1, 1)
			res.Rooms[0] = *room
			res.Rooms[0].ToData()
			NotifyRoomEventToAll(room, info.RoomEventUpdate)
		} else {
			res.ErrCode = info.RoomWarnCode_ENTER_ROOM_ALREADYINROOM
		}
	} else {
		res.ErrCode = info.RoomWarnCode_ENTER_ROOM_NUM_WRONG
	}

	return &res
}

func (r *module) LeaveRoom(ctx context.Context, requesterAcID, leaveAcID string, roomNum int) {
	r.sendCmdWithoutRes(ctx, &cmd{
		Type:    cmdTypeLeaveRoom,
		Account: requesterAcID,
		PS:      leaveAcID,
		P1:      roomNum,
	})

	return
}

func (r *module) leaveRoom(acID, leaveAcID string, roomNum int) *cmd {
	c := new(cmd)
	room, ok := r.roomNumMap[roomNum]
	if !ok {
		c.Code = ROOM_WARN_ENTERROOM_NUM_WRONG
		logs.Error("room leaveRoom with wrong room num %d", roomNum)
		return c
	}

	if room.IsJoinable() {
		if "" != leaveAcID {
			//主机操作其他人
			if acID != room.RoomMasterAcID {
				logs.Error("room leaveRoom, try to kick others, but you are not master.")
				c.Code = ROOM_ERR_YOU_NO_PERMISSION_TO_KICK
				return c
			} else {
				//Kick other player logic
				room.DelPlayer(leaveAcID)
				NotifyRoomEvent(leaveAcID, room, info.RoomEventKicked)
				NotifyRoomEventToAll(room, info.RoomEventUpdate)
			}
		} else {
			//自己请求自己
			if acID == room.RoomMasterAcID {
				//解散房间
				NotifyRoomEventToAll(room, info.RoomEventDismiss)
				// room del
				delete(r.roomNumMap, roomNum)
				r.deallocRoomNum(roomNum)
				r.playerSyncRoomCmdChan <- cmdSyncRoom{
					DelRoom: roomNum,
				}

				return c
			} else {
				room.DelPlayer(acID)
				if room.GetPlayerLen() <= 0 {
					// room del
					delete(r.roomNumMap, roomNum)
					r.deallocRoomNum(roomNum)
					r.playerSyncRoomCmdChan <- cmdSyncRoom{
						DelRoom: roomNum,
					}
				}
				NotifyRoomEventToAll(room, info.RoomEventLeave)
			}
		}
	} else {
		// 战斗中,无论谁走,战斗仍然需要继续。
		// 记下主机曾经离开的标记,当一轮8关打完后再解散房间
		if leaveAcID == "" {
			room.DelPlayer(acID)
			if room.GetPlayerLen() <= 0 {
				// room del
				delete(r.roomNumMap, roomNum)
				r.deallocRoomNum(roomNum)
				r.playerSyncRoomCmdChan <- cmdSyncRoom{
					DelRoom: roomNum,
				}
			}
		}

	}

	return nil
}

func (r *module) ChangeRoomMaster2Other(ctx context.Context, acID, otherID string, roomNum int) {
	r.sendCmdWithoutRes(ctx, &cmd{
		Type:    cmdTypeChangeMaster,
		Account: acID,
		PS:      otherID,
		P1:      roomNum,
	})
}

func (r *module) changeRoomMaster(acID, otherID string, roomNum int) *cmd {
	room, ok := r.roomNumMap[roomNum]
	if !ok {
		return nil
	}

	if room.RoomMasterAcID != acID {
		return nil
	}

	player := room.GetPlayerInRoom(otherID)
	if player == nil {
		return nil
	}

	room.RoomMasterAcID = otherID
	room.RoomMasterName = player.Name

	NotifyRoomEventToAll(room, info.RoomEventChangeMaster)
	return nil
}

// Ready 房主Ready则代表游戏立即开始,
// 所以房主应该必须在其他所有人都ready后才能调用该接口
func (r *module) Ready(ctx context.Context, acID string, roomNum int, avatar *helper.Avatar2ClientByJson) int {
	res := r.sendCmd(ctx, &cmd{
		Type:    cmdTypeReady,
		Account: acID,
		P1:      roomNum,
		Avatar:  avatar,
	})
	if res == nil {
		return ROOM_ERR_UNKNOWN
	}
	return res.Code
}

func (r *module) ready(acID string, roomNum int, avatar *helper.Avatar2ClientByJson) *cmd {
	c := new(cmd)
	room, ok := r.roomNumMap[roomNum]
	if !ok {
		return nil
	}

	if !room.IsJoinable() {
		logs.Error("room ready protocol called when room is not joinable")
		return nil
	}

	allotherready := room.HaveAllOthersReady()
	if room.RoomMasterAcID == acID {
		//房主还不能Ready
		if !allotherready {
			c.Code = ROOM_WARN_MASTER_WAITING_OTHERS_READY
			return c
		}
	}

	p := room.GetPlayerInRoom(acID)
	if p == nil {
		c.Code = ROOM_ERR_UNKNOWN
		return c
	}

	p.Stat = info.PlayerStatReady
	p.AvatarInfo = avatar

	if allotherready {
		//本次运行中,所有人都准备好了
		room.RoomStat = info.RoomStatFightting
		var playersIds [gamedata.FegnhuoRoomMaxPlayer]string
		var avatars [gamedata.FegnhuoRoomMaxPlayer]*helper.Avatar2ClientByJson
		for i := 0; i < room.GetPlayerLen(); i++ {
			pl := room.GetPlayerInRoomByIdx(i)
			playersIds[i] = pl.AcID
			avatars[i] = pl.AvatarInfo
		}

		data := gveHelper.FenghuoValue{
			AcIDs:      playersIds[:],
			AvatarInfo: avatars[:],
		}

		url := fmt.Sprintf("%s%s", uutil.JwsCfg.MatchUrl, gveHelper.FenghuoPostUrlAddressV1)
		retcode, ret, err := notifyMatchServer(data, url)
		if err != nil || retcode != 200 {
			//FIXME by YZH Fenghuo Ready.
			c.Code = 173
			return c
		}
		var fc gveHelper.FenghuoCreateInfo
		if err := json.Unmarshal(ret, &fc); err != nil {
			//FIXME by YZH Fenghuo Ready.
			c.Code = 172
			return c
		}

		room.MultiplayRoomID = fc.RoomID
		room.MultiplayUrl = fc.WebsktUrl
		room.MultiplayCancelUrl = fc.CancelUrl

		NotifyRoomEventToAll(room, info.RoomEventReadyForLevel)
	} else {
		NotifyRoomEventToAll(room, info.RoomEventUpdate)
	}

	return c
}

// CancelReady 取消玩家的Ready状态, 玩家只能取消自己的。
// 房主是不需要这个操作的, 房主默认永远是Ready状态。当房主标记为Ready时
// Room进入Fight状态
func (r *module) CancelReady(ctx context.Context, acID string, roomNum int) {
	r.sendCmdWithoutRes(ctx, &cmd{
		Type:    cmdTypeCancelReady,
		Account: acID,
		P1:      roomNum,
	})

	return
}

func (r *module) cancelReady(acID string, roomNum int) *cmd {
	room, ok := r.roomNumMap[roomNum]
	if !ok {
		return nil
	}

	if !room.IsJoinable() {
		logs.Error("room cancelReady protocol called when room is not joinable")
		return nil
	}

	if room.RoomMasterAcID == acID {
		//房主是不需要这个操作的
		return nil
	}

	p := room.GetPlayerInRoom(acID)
	if p == nil {
		return nil
	}

	p.Stat = info.PlayerStatNoReady
	NotifyRoomEventToAll(room, info.RoomEventUpdate)

	return nil
}

func (r *module) StartFight(ctx context.Context,
	rnd *rand.Rand, acID string,
	roomNum, subLevelIndex int, HasExtraReward bool) (int, *gamedata.FenghuoLevelData, *gamedata.PriceDatas) {

	res := r.sendCmd(ctx, &cmd{
		Type:    cmdTypeStartFight,
		Account: acID,
		P1:      roomNum,
		P2:      subLevelIndex,
		PB:      HasExtraReward,
		Rnd:     rnd,
	})

	if res == nil {
		return ROOM_ERR_UNKNOWN, nil, nil
	}

	return res.Code, res.SubLevels, res.Reward
}

func (r *module) startFight(rnd *rand.Rand, acID string, roomNum, subLevelIndex int, HasExtraReward bool) *cmd {
	resCmd := new(cmd)
	room, ok := r.roomNumMap[roomNum]
	if !ok {
		return nil
	}

	//second return is bool firstChoice in startFight, seem no use case right now.
	leveldata, _ := room.PlayerSelectSubLevel(acID, subLevelIndex)
	reward := room.PlayerGenerateReward(rnd, acID, subLevelIndex, HasExtraReward)

	resCmd.Reward = reward
	resCmd.SubLevels = &leveldata

	if room.CouldStartFight(subLevelIndex) {
		logs.Trace("room module start fight! RoomEventStartFight")
		NotifyRoomEventToAll(room, info.RoomEventStartFight)
	} else {
		NotifyRoomEventToAll(room, info.RoomEventUpdate)
	}
	return resCmd
}

// MakeReward returns
// (code, reward, useExtraReward,
// RewardPower, BattleHard, MasterAcID,)
func (r *module) MakeReward(
	ctx context.Context,
	acID string,
	roomNum, subLevelIndex int) (
	int, *gamedata.PriceDatas, bool,
	int, int, string) {

	res := r.sendCmd(ctx, &cmd{
		Type:    cmdTypeMakeReward,
		Account: acID,
		P1:      roomNum,
		P2:      subLevelIndex,
	})

	if res == nil {
		return ROOM_ERR_UNKNOWN, nil, false, 0, 0, ""
	}

	return res.Code, res.Reward, res.PB, res.P1, res.P2, res.PS
}

func (r *module) makeReward(acID string, roomNum, subLevelIndex int) *cmd {
	resCmd := new(cmd)
	room, ok := r.roomNumMap[roomNum]
	if !ok {
		return nil
	}
	resCmd.P1 = room.GetRewardPower()
	resCmd.P2 = room.Degree
	resCmd.PS = room.RoomMasterAcID

	err, useExtra, reward := room.PlayerMakeMyReward(acID, subLevelIndex)
	switch err {
	case info.RoomErr_Fight_DupMakeReward,
		info.RoomErr_Fight_NotSelectSubLevel,
		info.RoomErr_Fight_PlayerNotFound:
	case nil:
		resCmd.Reward = reward
		resCmd.PB = useExtra

	default:
		return resCmd
	}
	NotifyRoomEventToAll(room, info.RoomEventMakeReward)

	doneFinal := func() {
		//通知multiplay服务器Room可以关闭
		data := gveHelper.FenghuoValue{
			Shutdown: true,
			RoomID:   room.MultiplayRoomID,
		}

		retcode, ret, err := notifyMatchServer(data, room.MultiplayCancelUrl)
		logs.Info("Fenghuo room.PlayerFinalRewardDone try to notify multiplay server destroy room. %d, %s, %v",
			retcode, string(ret), err)

		//完成当前轮后,清理房间状态,房间继续
		// 或者原来房间房主已经离开,则房间解散
		dismiss := room.ResetRoomForNewStart()
		if dismiss {
			//解散房间
			NotifyRoomEventToAll(room, info.FinalDismiss)
			// room del
			delete(r.roomNumMap, roomNum)
			r.deallocRoomNum(roomNum)
			r.playerSyncRoomCmdChan <- cmdSyncRoom{
				DelRoom: roomNum,
			}
		} else {
			NotifyRoomEventToAll(room, info.RoomEventReJoin)
		}
	}
	//如果5second后,结果对方又发来了MkReward怎么办, room.RoomStat会阻止再次触发下面代码。
	if room.RoomStat == info.RoomStatFightting {
		room.MakeFinalRewardOnce()
		//防止如下情况: 有可能在最后一个玩家收到并发出MakeReward之前出现掉线,导致所有玩家都等待
		//因此这个goroutine会在5秒后,强制进行最终结算。并通知所有人

		go func() {
			<-time.After(5 * time.Second)
			room.DoFinalReward(doneFinal)
		}()

		FinalDone := room.AllPlayerFinalRewardDone()
		logs.Debug("Fenghuo room.PlayerFinalRewardDone(acID) %v", FinalDone)
		if FinalDone {
			room.DoFinalReward(doneFinal)
		}
	}
	return resCmd
}
