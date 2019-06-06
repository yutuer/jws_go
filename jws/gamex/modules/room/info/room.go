package info

import (
	"math/rand"

	"sync"

	"vcs.taiyouxi.net/jws/gamex/models/codec"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Room struct {
	ID              string `codec:"_1"`
	Num             int    `codec:"_14"`
	Type            int    `codec:"_2"`
	Password        string `codec:"_3"`
	Degree          int    `codec:"_4"`
	MultiplayRoomID string `codec:"_5"`
	MultiplayUrl    string `codec:"_6"`

	//产出倍率
	RewardPower        int    `codec:"_7"`
	MultiplayCancelUrl string `codec:"-"`
	//Buffers        []string `codec:"_8"`

	//彻底弃用, _9 可以被用来做其他值
	//ExtRewardCount int `codec:"_9"`

	//当前次数
	//Count int `codec:"_10"`
	//限制次数
	//LimitCount int `codec:"_11"`

	//MaxPlayerCount int `codec:"_12"`
	PlayerCount int `codec:"_13"`
	players     []PlayerInRoom
	Players     [][]byte `codec:"_15"`

	RoomMasterAcID string `codec:"_16"`
	RoomMasterName string `codec:"_17"`
	RoomStat       int    `codec:"_18"`
	//默认值0, 选择[1-8]
	RoomSubLevelLastSelected int `codec:"_19"`
	subLevels                []gamedata.FenghuoLevelData
	finalDismiss             bool
	finalRewardOnce          sync.Once
	finalRewardOnceUsed      bool
}

//第一个人MakeReward的时候就可以生成新的
func (r *Room) MakeFinalRewardOnce() {
	if !r.finalRewardOnceUsed {
		return
	}
	r.finalRewardOnce = sync.Once{}
}

func (r *Room) DoFinalReward(f func()) {
	r.finalRewardOnce.Do(func() {
		f()
		r.finalRewardOnceUsed = true
	})
}

func (r *Room) GetRewardPower() int {
	if r.RewardPower <= 0 {
		r.RewardPower = 1 //额外奖励次数至少是1倍
	}
	return r.RewardPower
}

func (r *Room) IsJoinable() bool {
	if r.RoomStat < RoomStatFightting {
		return true
	}
	return false
}

func (r *Room) generateSubLevels(BattleHard uint32) {
	r.subLevels = gamedata.GetFenghuoSubLevels(BattleHard)
}

func (r *Room) GetPlayerByID(acID string) *PlayerInRoom {
	for i, p := range r.players {
		if p.AcID == acID {
			return &r.players[i]
		}
	}
	return nil
}

func (r *Room) CouldStartFight(subLevelidx int) bool {
	if r.players == nil || len(r.players) == 0 {
		return false
	}
	var could bool
	could = true
	for _, p := range r.players {
		could = could && (p.SubLevelSelected[subLevelidx] == RoomSubLevelSelected)
		if !could {
			return false
		}
	}
	return could
}

// PlayerSelectSubLevel 玩家选择关卡进行游戏,
// 返回值bool代表是否玩家是首次选择
func (r *Room) PlayerSelectSubLevel(acID string, subLevelidx int) (gamedata.FenghuoLevelData, bool) {
	if r.subLevels == nil {
		r.generateSubLevels(uint32(r.Degree))
	}

	pl := r.GetPlayerByID(acID)
	if pl == nil {
		logs.Error("Fenghuo Room PlayerSelectSubLevel get Acid not in room! %s", acID)
		return gamedata.FenghuoLevelData{}, false
	}

	if acID == r.RoomMasterAcID {
		r.RoomSubLevelLastSelected = subLevelidx + 1
	}

	firstChoice, _ := pl.SelectSubLevel(subLevelidx)
	return r.subLevels[subLevelidx], firstChoice
}

func (r *Room) PlayerGenerateReward(
	rnd *rand.Rand, acID string,
	subLevelidx int, HasExtraReward bool) *gamedata.PriceDatas {

	pl := r.GetPlayerByID(acID)
	if pl == nil {
		logs.Error("Room PlayerGenerateReward not found acID %s", acID)
		return &gamedata.PriceDatas{}
	}

	giveDatas := gamedata.NewPriceDatas(8)

	rewardtimes := r.RewardPower //根据钻石数量进行奖励翻倍
	if rewardtimes <= 0 {
		rewardtimes = 1
	}
	for i := 0; i < rewardtimes; i++ {
		if give, err := gamedata.MakeFenghuoGives(rnd, r.subLevels[subLevelidx].NormalDrop); err == nil {
			giveDatas.AddOther(&give)
		} else {
			logs.Error("Room PlayerGenerateReward failed %s", err.Error())
			continue
		}
	}

	if HasExtraReward {
		extra, errEx := gamedata.MakeFenghuoGives(rnd, r.subLevels[subLevelidx].NormalDrop)
		if errEx == nil {
			pl.SubLevelConsumeExtraRewards[subLevelidx] = true
			giveDatas.AddOther(&extra)
		} else {
			logs.Error("Room PlayerGenerateReward failed %s", errEx.Error())
		}
	}

	//本轮选择如果是8局中的最后一局,则多给一个最终完成奖励
	if pl.IsFinalFightSelect() {
		//FinalRewards is needed.
		gives, err := gamedata.GetFinalReward(rnd, uint32(r.Degree))
		if err != nil {
			logs.Error("Room GetReward Final Loot failed %s", err.Error())
		} else {
			giveDatas.AddOther(&gives)
		}
	}

	pl.SetSubLevelReward(subLevelidx, &giveDatas)

	return &giveDatas
}

func (r *Room) PlayerFinalRewardDone(acID string) bool {
	pl := r.GetPlayerByID(acID)
	if pl != nil {
		return pl.HasFinalRewardDone()
	}
	return false
}

func (r *Room) AllPlayerFinalRewardDone() bool {
	alldone := true
	for _, p := range r.players {
		alldone = alldone && p.HasFinalRewardDone()
	}
	return alldone
}

// PlayerMakeMyReward return (error, useExraReward, reward)
func (r *Room) PlayerMakeMyReward(acID string, subLevelidx int) (error, bool, *gamedata.PriceDatas) {

	pl := r.GetPlayerByID(acID)
	if pl == nil {
		logs.Error("Room PlayerGenerateReward not found acID %s", acID)
		return RoomErr_Fight_PlayerNotFound, false, &gamedata.PriceDatas{}
	}

	fightStatus := RoomSubLevelDeafult
	//设置玩家战斗选关状态,防止多次领取

	fightStatus = pl.SubLevelSelected[subLevelidx]
	switch fightStatus {
	case RoomSubLevelDoneReward:
		return RoomErr_Fight_DupMakeReward, false, nil
	case RoomSubLevelSelected:
		pl.SubLevelSelected[subLevelidx] = RoomSubLevelDoneReward
		return nil, pl.SubLevelConsumeExtraRewards[subLevelidx], pl.SubLevelRewards[subLevelidx]
	default:
		return RoomErr_Fight_NotSelectSubLevel, false, nil
	}

	return RoomErr_ShouldNotBeHere, false, nil
}

func (r *Room) ResetRoomForNewStart() bool {
	if r.finalDismiss {
		return true
	}

	r.RoomStat = RoomStatWaitting
	r.RoomSubLevelLastSelected = 0
	if r.subLevels != nil {
		r.subLevels = nil
	}
	for i, _ := range r.players {
		r.players[i].ResetForNewStart()

	}
	return false
}

func (r *Room) ToData() []byte {
	r.GetRewardPower()
	r.PlayerCount = r.GetPlayerLen()
	r.Players = make([][]byte, 0, len(r.players))
	for i := 0; i < len(r.players); i++ {
		r.Players = append(r.Players, codec.Encode(r.players[i]))
	}
	return codec.Encode(*r)
}

func (r *Room) GetPlayers() []PlayerInRoom {
	return r.players[:]
}

func (r *Room) GetPlayerLen() int {
	return len(r.players)
}

func (r *Room) DelPlayer(acID string) {
	playersLen := len(r.players)
	for i := 0; i < playersLen; i++ {
		if r.players[i].AcID == acID {
			if playersLen > 1 {
				if i != playersLen-1 {
					r.players[i] = r.players[playersLen-1]
				}
				r.players = r.players[:playersLen-1]
			} else {
				r.players = r.players[0:0]
			}
			break
		}
	}

	if r.RoomStat == RoomStatFightting {
		//战斗中掉线逻辑
		if r.RoomMasterAcID == acID {
			r.RewardPower = 1 //原来是主机掉线, 取消生效的翻倍钻石消耗
			r.finalDismiss = true
			if len(r.players) >= 1 {
				r.RoomMasterAcID = r.players[0].AcID
				r.RoomMasterName = r.players[0].Name
			}
		} else {
			//这里是客机离开, 任何客机离开,不影响钻石翻倍

		}
	}
}

func (r *Room) AddPlayer(p *PlayerInRoom) {
	r.players = append(r.players, *p)
	if r.RoomMasterAcID == "" {
		r.RoomMasterAcID = p.AcID
		r.RoomMasterName = p.Name
	}
}

func (r *Room) GetPlayerInRoomByIdx(idx int) *PlayerInRoom {
	return &r.players[idx]
}

func (r *Room) GetPlayerInRoom(acID string) *PlayerInRoom {
	for i := 0; i < len(r.players); i++ {
		if r.players[i].AcID == acID {
			return &(r.players[i])
		}
	}
	return nil
}

func (r *Room) HaveAllOthersReady() bool {
	master := r.RoomMasterAcID
	for i := 0; i < len(r.players); i++ {
		if r.players[i].AcID != master {
			if r.players[i].Stat != PlayerStatReady {
				return false
			}
		}
	}
	return true
}

type PlayerInRoom struct {
	Name     string `codec:"_1"`
	AcID     string `codec:"_2"`
	AvatarID int    `codec:"_3"`
	CorpLv   uint32 `codec:"_4"`
	Gs       int    `codec:"_5"`
	Stat     int    `codec:"_6"`

	//RoomSubLevelDeafult(0) 代表了没有选择任何关卡
	SubLevelSelected            [gamedata.FenghuoStageMaxNum]int                  `codec:"_7"`
	SubLevelRewards             [gamedata.FenghuoStageMaxNum]*gamedata.PriceDatas `codec:"-"`
	SubLevelConsumeExtraRewards [gamedata.FenghuoStageMaxNum]bool                 `codec:"-"`
	AvatarInfo                  *helper.Avatar2ClientByJson                       `codec:"-"`
}

func (pl *PlayerInRoom) ResetForNewStart() {
	if pl == nil {
		return
	}
	pl.Stat = PlayerStatNoReady
	for i := 0; i < gamedata.FenghuoStageMaxNum; i++ {
		pl.SubLevelSelected[i] = RoomSubLevelDeafult
		pl.SubLevelRewards[i] = nil
		pl.SubLevelConsumeExtraRewards[i] = false
	}
}

func (pl *PlayerInRoom) HasFinalRewardDone() bool {
	if pl != nil {
		final := true
		for _, s := range pl.SubLevelSelected {
			final = final && (s == RoomSubLevelDoneReward)
		}
		return final
	}
	return false
}

func (pl *PlayerInRoom) IsFinalFightSelect() bool {
	if pl != nil {
		final := true
		for _, s := range pl.SubLevelSelected {
			final = final && (s == RoomSubLevelDeafult)
		}
		return final
	}
	return false
}

func (pl *PlayerInRoom) SetSubLevelReward(subLevelidx int, reward *gamedata.PriceDatas) {
	pl.SubLevelRewards[subLevelidx] = reward
}

// SelectSubLevel 玩家选择玩哪一个子关卡, 返回是否是(首次选择, 全面选择完毕)
func (pl *PlayerInRoom) SelectSubLevel(subLevelidx int) (firstChoice bool, finalChoice bool) {
	//计算当前玩家是否是第一次选择关卡
	fistselect := 0
	for _, s := range pl.SubLevelSelected {
		fistselect += s
	}
	if fistselect == 0 {
		firstChoice = true
	}

	pl.SubLevelSelected[subLevelidx] = RoomSubLevelSelected

	finalChoice = true
	for _, s := range pl.SubLevelSelected {
		if s == 0 {
			finalChoice = false
		}
	}
	return
}
