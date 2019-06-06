package teamboss

import (
	csCfg "vcs.taiyouxi.net/jws/crossservice/config"
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/crossservice/module/teamboss/multiplay_util"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//ParamAttack ..
type ParamStartFight struct {
	Sid  uint32
	Acid string
	Info helper.StartFightInfo
}

//RetAttack ..
type RetStartFight struct {
	Info helper.StartFightRetInfo
}

//MethodStartFight ..
type MethodStartFight struct {
	module.BaseMethod
}

func newMethodStartFight(m module.Module) *MethodStartFight {
	return &MethodStartFight{
		module.BaseMethod{Method: MethodStartFightID, Module: m},
	}
}

//NewParam ..
func (m *MethodStartFight) NewParam() module.Param {
	return &ParamStartFight{}
}

//NewRet ..
func (m *MethodStartFight) NewRet() module.Ret {
	return &RetStartFight{}
}

//Do ..
func (m *MethodStartFight) Do(t module.Transaction, p module.Param) (errCode uint32, ret module.Ret) {
	param := p.(*ParamStartFight)
	info := param.Info
	bm := m.ModuleAt().(*TeamBoss)
	logs.Info("bm: %v, param: %v", bm, param)
	// 通知multiplay，同步
	errCode = message.ErrCodeOK
	roomInfo := bm.Room.GetRoom(info.RoomID)
	if roomInfo == nil {
		logs.Info("[TeamBoss] Room: %v not exist", info.RoomID)
		ret = RetStartFight{
			Info: helper.StartFightRetInfo{
				Code: helper.RetCodeRoomNotExist,
			},
		}
		return
	}
	if !roomInfo.IsFull() || roomInfo.LeadAcID != info.AcID {
		logs.Info("[TeamBoss] Player: %v can't start fight for room: ", info.AcID, *roomInfo)
		ret = RetStartFight{
			Info: helper.StartFightRetInfo{
				Code: helper.RetCodeOptInvalid,
			},
		}
		return
	}

	battleData := make([][]byte, len(roomInfo.Players))
	acids := make([]string, len(roomInfo.Players))
	for i, item := range roomInfo.Players {
		if len(item.BattleData) <= 0 || item.SimpleInfo.AcID == "" {
			logs.Info("[TeamBoss] Player: %v can't start fight for room: %v, because no battle data some player", info.AcID, *roomInfo)
			ret = RetStartFight{
				Info: helper.StartFightRetInfo{
					Code: helper.RetCodeStartFightFailed,
				},
			}
			return
		}
		battleData[i] = item.BattleData
		acids[i] = item.SimpleInfo.AcID
	}

	mulRet, err := multiplay_util.NotifyMultiplay(&multiplay_util.TBStartFightData{
		RoomID:    info.RoomID,
		GroupID:   t.GroupID,
		Data:      battleData,
		AcID:      acids,
		GID:       csCfg.Cfg.Gid,
		SceneID:   roomInfo.SceneID,
		Level:     roomInfo.RoomLevel,
		BossID:    roomInfo.BossID,
		TeamTypID: roomInfo.TeamTypID,
		BoxStatus: roomInfo.BoxStatus,
		CostID : roomInfo.AdvanceCostID,
	})
	if err != nil {
		logs.Error("[TeamBoss] Notify multiplay server start fight err: %v", err)
		errCode = message.ErrCodeInner
		ret = RetStartFight{
			Info: helper.StartFightRetInfo{
				Code: helper.RetCodeStartFightFailed,
			},
		}
	} else {
		logs.Info("[TeamBoss] Receive ret from multfiplay: %v", mulRet)
		url, globalRoomID, err := multiplay_util.GenTeamBossMultiplayInfo(mulRet)
		if err == nil {
			roomInfo.RoomState = helper.TBRoomFight
			for _, p := range roomInfo.Players {
				p.SimpleInfo = helper.PlayerSimpleInfo{
					AcID:         p.SimpleInfo.AcID,
					Sid:          p.SimpleInfo.Sid,
					GS:           p.SimpleInfo.GS,
					Avatar:       p.SimpleInfo.Avatar,
					Name:         p.SimpleInfo.Name,
					VIP:          p.SimpleInfo.VIP,
					BattleAvatar: -1,
					InBattle:     true,
				}
				p.BattleData = nil
			}
			roomInfo.PositionAcID = [helper.RoomPlayerMaxCount]string{}
			roomInfo.BoxStatus = 0
			roomInfo.AdvanceCostID = ""
			roomInfo.GenRoomInfo()
			acids := roomInfo.GetExtraPlayer(info.AcID)
			for k, v := range acids {
				bm.PlayerStart(url, uint32(k), globalRoomID, v)
			}
			ret = RetStartFight{
				Info: helper.StartFightRetInfo{
					ServerUrl:    url,
					GlobalRoomID: globalRoomID,
				},
			}
		} else {
			logs.Error("[TeamBoss] Parse multiplay server reply err: %v", err)
			ret = RetStartFight{
				Info: helper.StartFightRetInfo{
					Code: helper.RetCodeStartFightFailed,
				},
			}
		}
	}

	return
}
