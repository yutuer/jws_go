package account

import (
	"math/rand"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/stage_star"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/mn_selector"
)

// stageLimitRewardInfo 用于计算Limit奖励的信息，每个奖励对应一个
// Reward_count项表示还剩余几个奖励，Space_num表示本轮剩余多少次数
type stageLimitRewardInfo struct {
	MN mnSelector.MNSelectorState `json:"mn"`
}

type stageSingleInfo struct {
	Id           string                 `json:"id"`
	Reward_state []stageLimitRewardInfo `json:"s"`
	T_count      int32                  `json:"c"`   // 当天打关卡次数
	T_refresh    int32                  `json:"r"`   // 当天刷新关卡次数的次数
	Sum_count    int32                  `json:"smc"` // 总次数
	MaxStar      int32                  `json:"m"`   // 最高星级
}

type chapterInfo struct {
	ChapterId   string   `json:"ch"` // 章节id
	Star        uint32   `json:"sr"` // 总星数
	Has_awardId []uint32 `json:"ha"` // 已领奖id
}

type Chapter2Client struct {
	ChapterId   string   `codec:"ch"` // 章节id
	Star        uint32   `codec:"sr"` // 总星数
	Has_awardId []uint32 `codec:"ha"` // 已领奖id
}

func randByOffset(main, offset int32, r *rand.Rand) int32 {
	if offset == 0 {
		return main
	}

	if offset < 0 {
		offset = -offset
	}

	rv := r.Int31n(offset * 2)
	return main + rv - offset
}

// 重置存储的奖励信息，当奖励计算了一轮之后重置
func (s *stageSingleInfo) ResetReward(idx int, num, space, offset int32, r *rand.Rand) {
	if idx >= len(s.Reward_state) {
		logs.Error("stage limit reward info Reset Err!")
		return
	}
	s.Reward_state[idx].MN.Reset(
		int64(num),
		int64(randByOffset(space, offset, r)))
}

/*
	AppendRewardState
	每一个玩家的每一个副本会有Limit奖励的信息，用于计算奖励结果
	这些信息不止一个，并且策划可能会增加奖励个数，
	这时，当玩家完成副本时，需要增加奖励个数，
	另一方面，当玩家第一次打副本时，会新建Limit信息，
	这两处需要先扩展Limit信息数组，再进行ResetReward操作
	这个接口用于扩展Limit信息数组，由于这个操作不频繁，所以只是简单的增加数组大小
*/
func (s *stageSingleInfo) AppendRewardState() {
	s.Reward_state = append(s.Reward_state,
		stageLimitRewardInfo{})
}

func (s *stageSingleInfo) update() {
	s.T_count = 0
	s.T_refresh = 0
}

type StageInfo struct {
	Stages           []stageSingleInfo `json:"stages"`
	Last_update_time int64             `json:"last"`     // 上次更新时间
	LastStageId      string            `json:"lstage"`   // 上次玩的副本id
	Chapters         []chapterInfo     `json:"chapters"` // 章节信息
}

func (s *StageInfo) IsStagePass(stage_id string) bool {
	for i := len(s.Stages); i > 0; i-- {
		if stage_id == s.Stages[i-1].Id {
			return (s.Stages[i-1]).MaxStar > 0
		}
	}
	return false
}

func (s *StageInfo) GetStar(stage_id string) int32 {
	// TBD by Fanyang 改为用sort库二分查找
	for i := len(s.Stages); i > 0; i-- {
		if stage_id == s.Stages[i-1].Id {
			return (s.Stages[i-1]).MaxStar
		}
	}
	return 0
}

func (s *StageInfo) GetStarCount(stage_id string) int32 {
	return stage_star.GetStarCount(s.GetStar(stage_id))
}

func (s *StageInfo) updateAll(day_update_time int64) {
	if s.Last_update_time == day_update_time {
		return
	}
	logs.Trace("stage update")
	for i := len(s.Stages) - 1; i >= 0; i-- {
		s.Stages[i].update()
	}
	s.Last_update_time = day_update_time
}

func (s *StageInfo) GetAll(day_update_time int64) []stageSingleInfo {
	s.updateAll(day_update_time)
	return s.Stages[:]
}

func (s *StageInfo) DebugCleanStage() {
	s.Stages = make([]stageSingleInfo, 0, 64)
}

func (s *StageInfo) SetLastStageId(id string) {
	s.LastStageId = id
}
func (s *StageInfo) GetLastStageId() string {
	return s.LastStageId
}

func (s *StageInfo) ResetStageCount() {
	// 将所有关卡当天限制去除，下一次取信息时Update会重置
	s.Last_update_time = -1 // 肯定不是今天
}

func (s *StageInfo) IsAllPreStagePass(pre []string) bool {
	// 是否所有前置关卡都过了
	for j := 0; j < len(pre); j++ {
		if pre[j] == "" {
			continue
		}
		is_has := false
		for i := len(s.Stages) - 1; i >= 0; i-- {
			stage := s.Stages[i]
			if stage.Id == pre[j] && stage.MaxStar > 0 {
				is_has = true
				break
			}
		}
		if !is_has {
			return false
		}
	}
	return true
}

//
// 获取奖励结算信息 对于一个副本会有多个LimitReward掉落 所以返回数组
// 返回项 Reward_count项表示还剩余几个奖励，Space_num表示本轮剩余多少次数
// 注意不能缓存指针
func (s *StageInfo) GetStageInfo(day_update_time int64, stage_id string, r *rand.Rand) *stageSingleInfo {
	s.updateAll(day_update_time)

	for i := len(s.Stages); i > 0; i-- {
		if stage_id == s.Stages[i-1].Id {
			return &(s.Stages[i-1])
		}
	}

	//没有的话就新加入
	n := stageSingleInfo{}
	n.Id = stage_id

	rewards := gamedata.GetStageRewardLimitCfg(stage_id, true) //新副本肯定没打过
	n.Reward_state = make(
		[]stageLimitRewardInfo,
		len(rewards),
		len(rewards))
	for idx, reward := range rewards {
		n.ResetReward(idx, reward.Num, reward.Space, reward.Offset, r)
	}

	s.Stages = append(s.Stages, n)
	//必然是最后一个
	return &(s.Stages[len(s.Stages)-1])

}

func (s *StageInfo) GetChapterInfo(chapterId string) *chapterInfo {
	for i, _ := range s.Chapters {
		ch := &(s.Chapters[i])
		if ch.ChapterId == chapterId {
			return ch
		}
	}

	return nil
}

func (s *StageInfo) GetChapterInfoWithInit(chapterId string) *chapterInfo {
	if !gamedata.ChapterIsExist(chapterId) {
		return nil
	}
	ch := s.GetChapterInfo(chapterId)
	if ch == nil {
		ch = &chapterInfo{ChapterId: chapterId, Has_awardId: []uint32{}}
		s.Chapters = append(s.Chapters, *ch)
		return &(s.Chapters[len(s.Chapters)-1])
	}
	return ch
}

func (s *StageInfo) AddChapterStarFromStage(stage_id string, addStar int32) bool {
	chapterId := gamedata.GetStage2Chapter(stage_id)
	if chapterId == "" {
		return false
	}
	ch := s.GetChapterInfo(chapterId)
	if ch == nil {
		ch = &chapterInfo{ChapterId: chapterId, Has_awardId: []uint32{}}
		s.Chapters = append(s.Chapters, *ch)
		chapter := &(s.Chapters[len(s.Chapters)-1])
		chapter.Star = chapter.Star + uint32(addStar)
		return true
	}
	ch.Star = ch.Star + uint32(addStar)
	return true
}

func (s *StageInfo) AddEStageTimes(day_update_time int64, stage_id string, times uint32) bool {
	s.updateAll(day_update_time)

	for i := len(s.Stages); i > 0; i-- {
		if stage_id == s.Stages[i-1].Id {
			stage := &(s.Stages[i-1])
			if stage.T_count > 0 {
				if int32(times) > stage.T_count {
					stage.T_count = 0
				} else {
					stage.T_count = stage.T_count - int32(times)
				}
				return true
			}
		}
	}
	return false
}
