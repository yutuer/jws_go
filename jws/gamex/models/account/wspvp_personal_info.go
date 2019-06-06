package account

import (
	"fmt"
	"sort"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/account/gs"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/title_rank"
	"vcs.taiyouxi.net/jws/gamex/modules/ws_pvp"
	"vcs.taiyouxi.net/platform/planx/util/distinct"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	WSPVP_FORMATION_COUNT = 9
)

// 无双争霸的个人信息 TODO 有些信息是否存库
type WSPVPPersonalInfo struct {
	// 以下信息会存库
	Rank               int   `json:"rank"`                  // 当前排名 默认值0 代表未上榜
	NotClaimedReward   int   `json:"not_claimed_reward"`    // 累计未领取的奖励
	LastRankChangeTime int64 `json:"last_rank_change_time"` // 上次排名发生变化的时间
	HasClaimedBox      []int `json:"has_claimed_box"`       // 已经领取的宝箱
	HasChallengeCount  int   `json:"has_challenge_count"`   // 已经挑战的次数
	BestRank           int   `json:"best_rank"`             // 最佳排名
	HasClaimedBestRank []int `json:"has_claimed_best_rank"` // 最佳排名的领取奖励

	OpponentSimpleInfo []WSPVPOppSimpleInfo `json:"opp_simple"`        // 4个对手的排名
	MyAttackFormation  []int64              `json:"my_att_formation"`  // 进攻阵型
	LockingOppInfo     *ws_pvp.WSPVPInfo    `json:"locking_opp_info"`  // 锁定的武将信息
	LockingExpireTime  int64                `json:"locing_time"`       // 锁定过期时间
	DefenseFormation   []int64              `json:"defense_formation"` // 每个位置对应的武将IDX， -1 表示没有人

	LastRefreshTime int64 `json:"wspvp_refresh_time"`

	lastRefreshRankOnline int64
}

// 用于显示4个挑战对手的信息
type WSPVPOppSimpleInfo struct {
	Acid      string `json:"acid"`  // 角色acid
	Rank      int64  `json:"rank"`  // 排名
	ServerId  int64  `json:"sid"`   // 区服
	Name      string `json:"name"`  // 名字
	GuildName string `json:"gname"` // 军团名字
	TitleId   string `json:"title_id""`
	VipLevel  int64  `json:"vip_level"`
}

// 当前正锁定的武将信息， 这里选择存库是为了保证数据一致性
type LockingOppInfo struct {
	Acid      string            `json:"acid"`
	Formation []int64           `json:"formation"` // 返回每个位置对应的武将idx，固定长度9
	HeroStar  []int64           `json:"hero_star"` // 返回每个位置武将的升星等级
	CorpGs    []int64           `json:"corpgs"`    // 每个队伍的战力，固定长度3
	HeroInfo  []LockingHeroInfo `json:"hero_info"` // 布阵界面的所有武将信息
}

type LockingHeroInfo struct {
	Idx            int64    `json:"idx"`        // 武将数字ID
	Attr           []byte   `json:"attr"`       // 武将属性
	Gs             int64    `json:"gs"`         // 武将战力
	Skills         []int64  `json:"skills"`     // 技能等级
	Fashions       []string `json:"fashions"`   // 时装tableId
	PassiveSkillId []string `json:"p_skill_id"` // 被动技能ID
	CounterSkillId []string `json:"c_skill_id"` // 被动技能ID
	TriggerSkillId []string `json:"t_skill_id"` // 被动技能ID
}

func (p *Account) GetDefenseFormation() []int64 {
	if p.Profile.WSPVPPersonalInfo.DefenseFormation == nil {
		p.initDefenseFormation()
	}
	return p.Profile.WSPVPPersonalInfo.DefenseFormation
}

func (p *Account) initDefenseFormation() {
	if p.Profile.GetCorp().Level < 60 {
		return
	}
	p.Profile.WSPVPPersonalInfo.DefenseFormation = make([]int64, 9)
	for i := range p.Profile.WSPVPPersonalInfo.DefenseFormation {
		p.Profile.WSPVPPersonalInfo.DefenseFormation[i] = -1
	}
	_, _, _, _, _, _, sortGs := gs.GetCurrAttr(NewAccountGsCalculateAdapter(p))
breakLoop:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if 3*i+j >= len(sortGs) {
				break breakLoop
			}
			p.Profile.WSPVPPersonalInfo.DefenseFormation[3*j+i] = int64(sortGs[3*i+j].HeroId)
			logs.Debug("init defense formation %d, %d", 3*j+i, 3*i+j)
		}
	}
	logs.Debug("init defense formation %v", p.Profile.WSPVPPersonalInfo.DefenseFormation)
}

func (w *WSPVPPersonalInfo) SetFormation(newFormation []int64) {
	w.DefenseFormation = newFormation
}

func (w *WSPVPPersonalInfo) GetOpponentSimpInfo(acid string) *WSPVPOppSimpleInfo {
	for _, simple := range w.OpponentSimpleInfo {
		if simple.Acid == acid {
			return &simple
		}
	}
	return nil
}

func (w *WSPVPPersonalInfo) CleanLockInfo() {
	w.LockingExpireTime = 0
	w.LockingOppInfo = nil
}

func (w *WSPVPPersonalInfo) HasClaimedBestRankReward(id int) bool {
	for _, claimId := range w.HasClaimedBestRank {
		if id == claimId {
			return true
		}
	}
	return false
}

func (w *WSPVPPersonalInfo) AddBestRankReward(id int) {
	if w.HasClaimedBestRank == nil {
		w.HasClaimedBestRank = make([]int, 0)
	}
	w.HasClaimedBestRank = append(w.HasClaimedBestRank, id)
}

func (w *WSPVPPersonalInfo) HasClaimedBoxReward(id int) bool {
	for _, count := range w.HasClaimedBox {
		if count == id {
			return true
		}
	}
	return false
}

func (w *WSPVPPersonalInfo) AddBoxReward(id int) {
	if w.HasClaimedBox == nil {
		w.HasClaimedBox = make([]int, 0)
	}
	w.HasClaimedBox = append(w.HasClaimedBox, id)
}

func (w *WSPVPPersonalInfo) TryRefresh(nowT int64) {
	if w.LastRefreshTime == 0 {
		w.LastRefreshTime = nowT
		return
	}
	if !gamedata.IsSameDayCommon(nowT, w.LastRefreshTime) {
		w.HasChallengeCount = 0
		w.HasClaimedBox = make([]int, 0)
		w.LastRefreshTime = nowT
	}
}

func (w *WSPVPPersonalInfo) IsWSPVPMarquee(gid, sid uint, uid string, name string) {
	rank := title_rank.GetModule(sid).GetWuShuangRank(uid)
	if rank == 1 {
		wsPvpGroupId := gamedata.GetWSPVPGroupCfg(uint32(sid)).GetWspvpGroupID()
		sids_ := gamedata.GetWSPVPSids(wsPvpGroupId)

		gsids := make([]string, 0, len(sids_))
		for _, sid := range sids_ {
			gsString := sysnotice.GetRealSid(fmt.Sprintf("%d:%v", gid, sid))
			if gsString != "" {
				gsids = append(gsids, gsString)
			}
		}
		disinct_gsids, err := distinct.ValuesAndDisinct(gsids)
		if err != nil {
			logs.Error("IsWSPVPMarquee Disinct Error:%v")
		}
		for _, wsGSid := range disinct_gsids {
			sysnotice.NewSysRollNotice(fmt.Sprintf("%v", wsGSid), gamedata.IDS_WuShuang).
				AddParam(sysnotice.ParamType_RollName, name).Send()
		}
	}
}

func (w *WSPVPPersonalInfo) GetMyAttackFormation() []int64 {
	if w.MyAttackFormation == nil {
		if w.DefenseFormation == nil {
			w.MyAttackFormation = make([]int64, 9)
			for i := 0; i < 9; i++ {
				w.MyAttackFormation[i] = -1
			}
		} else {
			w.MyAttackFormation = w.DefenseFormation
		}
	}
	return w.MyAttackFormation
}

func (w *WSPVPPersonalInfo) OnAfterLogin(sid uint, acid string) {
	groupId := int(gamedata.GetWSPVPGroupId(uint32(sid)))
	// 下面两个函数的顺序不能换，updateTimeReward会用到教早版本的排名
	w.updateTimeReward(groupId, acid)
	w.updateRank(groupId, acid)
}

func (w *WSPVPPersonalInfo) updateRank(groupId int, acid string) {
	w.lastRefreshRankOnline = time.Now().Unix()
	rank := ws_pvp.GetRanks(groupId, []string{acid})
	if len(rank) != 1 {
		logs.Error("update rank err")
		return
	}
	oldRank := w.Rank
	w.Rank = rank[0]
	if oldRank != w.Rank {
		w.OnRankChanged(oldRank)
	}
}

func (w *WSPVPPersonalInfo) updateTimeReward(groupId int, acid string) {
	if w.LastRankChangeTime == 0 {
		return
	}
	wspvpLogs := ws_pvp.WspvpLogArray(ws_pvp.GetWSPVPLog(groupId, acid))
	sort.Sort(wspvpLogs) //按时间从小到大排序
	reward := 0
	beginTime := w.LastRankChangeTime
	beginRank := w.Rank
	for _, log := range wspvpLogs {
		if log.Time > beginTime {
			_, tempReward := w.CalcNotClaimedReward(beginRank, beginTime, log.Time)
			reward += tempReward
			beginTime = log.Time
			beginRank = log.Rank
		}
	}
	w.NotClaimedReward += reward
	w.LastRankChangeTime = beginTime
	w.Rank = beginRank
}

func (w *WSPVPPersonalInfo) CalcNotClaimedReward(rank int, beginTime int64, endTime int64) (string, int) {
	logs.Debug("calcNotClaimedReward, time %d, %d", beginTime, endTime)
	cfg := gamedata.GetWsPvpTimeReward(rank)
	rewardCount := 0
	if cfg != nil && beginTime != 0 && beginTime < endTime {
		rewardCount += int(float32(endTime-beginTime) / 3600 * float32(cfg.GetRankLootNumber()))
	}
	logs.Debug("calccNotClaimedReward result %d", rewardCount)
	if cfg == nil {
		cfg = gamedata.GetWsPvpTimeReward(1) // 这里如果cfg==nil, 从其他已有的排名里面去获取物品ID
	}
	return cfg.GetRankLootID(), rewardCount
}

func (w *WSPVPPersonalInfo) OnRankChanged(oldRank int) {
	// 更新累计奖励
	nowTime := time.Now().Unix()
	_, rewardCount := w.CalcNotClaimedReward(oldRank, w.LastRankChangeTime, nowTime)
	w.LastRankChangeTime = nowTime
	w.NotClaimedReward += rewardCount
	// 更新最佳排名
	if w.Rank != 0 && (w.Rank < w.BestRank || w.BestRank == 0) {
		w.BestRank = w.Rank
	}
}

func (w *WSPVPPersonalInfo) TryRefreshRank(groupId int, acid string) {
	nowTime := time.Now().Unix()
	if nowTime-w.lastRefreshRankOnline > 10 && w.Rank > 0 {
		w.updateRank(groupId, acid)
	}
}
