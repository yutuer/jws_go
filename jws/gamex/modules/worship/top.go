package worship

import "vcs.taiyouxi.net/platform/planx/util/logs"

const (
	WorshipAccountCount = 3
	initAccountCount    = 4096
)

type worshipAccData struct {
	AccountID   string `json:"a"`
	WorshipTime int    `json:"w"`
}

type TopAccountWorship struct {
	dbKey      string
	worshipMap map[string]*worshipAccData
	Worship    []worshipAccData `json:"worship"`
}

func (t *TopAccountWorship) init(sid uint) {
	t.dbKey = getKeyNameInRedis(sid)
}

func (t *TopAccountWorship) worship(accID string) error {
	w, ok := t.worshipMap[accID]
	if ok && w != nil {
		w.WorshipTime++
	} else {
		t.Worship = append(t.Worship, worshipAccData{
			AccountID:   accID,
			WorshipTime: 1,
		})
		t.worshipMap[accID] = &t.Worship[len(t.Worship)-1]
	}
	logs.Trace("TopAccountWorship %v", t.Worship)
	return nil
}

func (t *TopAccountWorship) clean() {
	t.worshipMap = make(map[string]*worshipAccData, initAccountCount)
	t.Worship = make([]worshipAccData, 0, initAccountCount)
	go func() {
		err := t.saveDB()
		if err != nil {
			logs.Error(
				"TopAccountWorship saveDB Err by %s",
				err.Error())
		}
	}()
}

func (t *TopAccountWorship) copyTop() []worshipAccData {
	res := make([]worshipAccData, 0, len(t.Worship))
	for _, w := range t.Worship {
		res = append(res, w)
	}
	logs.Trace("TopAccountWorship %v %v", t.Worship, res)
	return res
}

func (t *TopAccountWorship) getWorship(accID string) int {
	w, ok := t.worshipMap[accID]
	if ok && w != nil {
		return w.WorshipTime
	} else {
		return 0
	}
}
