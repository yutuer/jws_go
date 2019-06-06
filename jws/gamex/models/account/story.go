package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const StoryNumMax = 512

const (
	STORY_State_No_Finish = iota
	STORY_State_Has_Finish
	STORY_State_Has_Read
)

type StoryToClient struct {
	Id       int `codec:"id"`
	State    int `codec:"stat"`
	Progress int `codec:"p"`
	All      int `codec:"a"`
}

type story struct {
	Id        int                `json:"id"`
	Condition gamedata.Condition `json:"c"`
}

func (q *story) IsVailed() bool {
	return q.Id == 0
}

func (q *story) SetVailed() {
	q.Id = 0
}

func (q *story) GetProgress(p *Account) (int, int) {
	storys := gamedata.GetStorys()

	return GetConditionProgress(&q.Condition, p,
		storys[q.Id].GetFCType(),
		int64(storys[q.Id].GetFCValueIP1()),
		int64(storys[q.Id].GetFCValueIP2()),
		storys[q.Id].GetFCValueSP1(),
		storys[q.Id].GetFCValueSP2())
}

type PlayerStory struct {
	HasRead []int   `json:"read"`
	Story   []story `json:"story"`
}

func (p *PlayerStory) ReadStory(account *Account, id int) {
	acid := account.AccountID.String()
	if id < 0 || id >= len(p.Story) || id >= StoryNumMax {
		logs.SentryLogicCritical(acid, "ReadStory Err By No Id %d", id)
		return
	}

	s := &p.Story[id]
	if s.IsVailed() {
		return
	}

	progress, all := s.GetProgress(account)
	if progress >= all {
		p.SetHasRead(id)
	}

}

func (p *PlayerStory) SetHasRead(id int) {
	for id >= len(p.HasRead) {
		p.HasRead = append(p.HasRead, 0)
	}
	p.HasRead[id] = 1
	p.Story[id].SetVailed()
}

func (p *PlayerStory) RegCondition(c *gamedata.PlayerCondition) {
	for i := 0; i < len(p.Story); i++ {
		if !p.Story[i].IsVailed() {
			c.RegCondition(&p.Story[i].Condition)
		}
	}
}

func (p *PlayerStory) RegOneCondition(qidx int, c *gamedata.PlayerCondition) {
	c.RegCondition(&p.Story[qidx].Condition)
}

func (p *PlayerStory) GetStoryState(account *Account, section int) ([]StoryToClient, bool) {
	storys := gamedata.GetStorysBySection(section)
	if storys == nil {
		return []StoryToClient{}, false
	} else {
		// STORY_State_No_Finish
		// STORY_State_Has_Finish
		// STORY_State_Has_Read
		states := make([]StoryToClient, 0, len(storys))
		for _, id := range storys {
			if id < 0 || id >= len(p.Story) {
				logs.Error("GetStoryState Err By Section %d %d",
					section, id)
				return []StoryToClient{}, false
			}
			if id >= len(p.HasRead) || p.HasRead[id] <= 0 {

				progress, all := p.Story[id].GetProgress(account)

				logs.Trace("GetStoryState %d %d/%d -> %v", id, progress, all, p.Story[id])
				if progress >= all && !p.Story[id].IsVailed() {
					states = append(states, StoryToClient{
						Id:       id,
						State:    STORY_State_Has_Finish,
						Progress: progress,
						All:      all,
					})
				} else {
					states = append(states, StoryToClient{
						Id:       id,
						State:    STORY_State_No_Finish,
						Progress: progress,
						All:      all,
					})
				}
			} else {
				states = append(states, StoryToClient{
					Id:    id,
					State: STORY_State_Has_Read,
				})
			}
		}
		return states[:], true
	}
}

func (p *PlayerStory) OnAfterLogin(c *gamedata.PlayerCondition) {
	storys := gamedata.GetStorys()
	old_len := len(p.Story)
	new_len := len(storys)

	if p.Story == nil {
		p.Story = make([]story, 0, new_len+1)
	}

	if p.HasRead == nil {
		p.HasRead = make([]int, 0, new_len+1)
	}

	// 先检查下有没有新的story
	if new_len > old_len {
		for i := old_len; i < new_len; i++ {
			data := storys[i]
			cond := NewCondition(
				data.GetFCType(),
				int64(data.GetFCValueIP1()),
				int64(data.GetFCValueIP2()),
				data.GetFCValueSP1(),
				data.GetFCValueSP2())

			p.Story = append(p.Story, story{
				Id:        i,
				Condition: *cond,
			})
		}
	}

	p.RegCondition(c)
}
