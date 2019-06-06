package teamboss

import (
	"github.com/gin-gonic/gin"
	csCfg "vcs.taiyouxi.net/jws/crossservice/config"
	"vcs.taiyouxi.net/jws/crossservice/module"
	"vcs.taiyouxi.net/jws/crossservice/module/teamboss/multiplay_util"
	"vcs.taiyouxi.net/jws/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//..
const (
	ModuleID = "teamboss"

	MethodStartFightID       = "startfight"
	MethodCreateRoomID       = "createroom"
	MethodJoinRoomID         = "joinroom"
	MethodRoomListID         = "roomlist"
	MethodReadyFightID       = "readyfight"
	MethodLeaveRoomID        = "leaveroom"
	MethodChangeAvatarID     = "changeavatar"
	MethodCostAdvanceID      = "costadvance"
	MethodChangeRoomStatusID = "changeroomstatus"
	MethodEndFightID         = "endfight"
	MethodGetPlayerDetailID  = "getplayerdetail"

	CallbackPlayerStartID = "playerstart"
	CallbackRoomInfoID    = "roominfo"
	CallbackKickID        = "kick"
)

func init() {
	module.RegModule(&Generator{})
}

//Generator ..
type Generator struct {
}

//ModuleID ..
func (g *Generator) ModuleID() string {
	return ModuleID
}

//NewModule ..
func (g *Generator) NewModule(group uint32) module.Module {
	moduleTeamBoss := &TeamBoss{
		BaseModule: module.BaseModule{
			GroupID: group,
			Module:  ModuleID,
			Methods: map[string]module.Method{},
			Static:  false,
		},
		Room: &RoomInfo{
			Rooms: make(map[uint32][]*Room, 10),
		},
		RewardLog: &RewardLog{
			Rewards: make(map[string]*Rewards, 10),
			Group:   group,
		},
	}
	moduleTeamBoss.Methods[MethodStartFightID] = newMethodStartFight(moduleTeamBoss)
	moduleTeamBoss.Methods[MethodCreateRoomID] = newMethodCreateRoom(moduleTeamBoss)
	moduleTeamBoss.Methods[MethodJoinRoomID] = newMethodJoinRoom(moduleTeamBoss)
	moduleTeamBoss.Methods[MethodRoomListID] = newMethodRoomList(moduleTeamBoss)
	moduleTeamBoss.Methods[MethodReadyFightID] = newMethodReadyFight(moduleTeamBoss)
	moduleTeamBoss.Methods[MethodLeaveRoomID] = newMethodLeaveRoom(moduleTeamBoss)
	moduleTeamBoss.Methods[MethodChangeAvatarID] = newMethodChangeAvatar(moduleTeamBoss)
	moduleTeamBoss.Methods[MethodCostAdvanceID] = newMethodCostAdvance(moduleTeamBoss)
	moduleTeamBoss.Methods[MethodChangeRoomStatusID] = newMethodChangeRoomStatus(moduleTeamBoss)
	moduleTeamBoss.Methods[MethodEndFightID] = newMethodEndFight(moduleTeamBoss)
	moduleTeamBoss.Methods[MethodGetPlayerDetailID] = newMethodGetPlayerDetail(moduleTeamBoss)

	moduleTeamBoss.Methods[CallbackPlayerStartID] = newCallbackPlayerStart(moduleTeamBoss)
	moduleTeamBoss.Methods[CallbackRoomInfoID] = newCallbackRoomInfo(moduleTeamBoss)
	moduleTeamBoss.Methods[CallbackKickID] = newCallbackKick(moduleTeamBoss)
	logs.Info("[TeamBoss] NewModule for Group [%d]", group)
	return moduleTeamBoss
}

//WorldBoss ..
type TeamBoss struct {
	module.BaseModule
	Room      *RoomInfo
	RewardLog *RewardLog
}

//HashMask ..
func (s *TeamBoss) HashMask() uint32 {
	return 32
}

//Start ..
func (s *TeamBoss) Start() {
	multiplay_util.RegEtcd(csCfg.Cfg.EtcdRoot, s.GroupID)
	multiplay_util.RegTBHttpHanlde(s.BattleStopHandle, s.GroupID)
	if err := s.RewardLog.load(); err != nil {
		panic(err)
	}

	logs.Info("[TeamBoss] Group [%d] start", s.GroupID)
}

//AfterStart ..
func (s *TeamBoss) AfterStart() {
	logs.Info("[TeamBoss] Group [%d] after start", s.GroupID)
}

//BeforeStop ..
func (s *TeamBoss) BeforeStop() {
}

//Stop ..
func (s *TeamBoss) Stop() {
	if err := s.RewardLog.save(); err != nil {
		panic(err)
	}
	logs.Info("[TeamBoss] Group [%d] stop", s.GroupID)
}

func (s *TeamBoss) BattleStopHandle(c *gin.Context) {

	info := multiplay_util.TeamBossStopInfo{}
	err := c.Bind(&info)
	if err != nil {
		c.String(400, err.Error())
		return
	}
	logs.Info("[TeamBoss] Get stop game info from multiplay: %v", info)
	status := make(map[string]bool, 0)
	for _, item := range info.AcIDs {
		status[item] = false
	}
	s.RewardLog.setReward(info.GameID, &Rewards{
		GlobalRoomID: info.GameID,
		hasReward:    info.IsSuccess,
		HasRedBox:    len(info.AcIDs) >= helper.RoomPlayerMaxCount && info.BoxStatus == helper.BoxStatusAdvance,
		Status:       status,
		Level:        info.Level,
		CostID:       info.CostID,
	})
	c.String(200, "success")
}
