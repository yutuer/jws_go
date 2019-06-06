package account

import (
	"reflect"

	"sort"

	"github.com/ugorji/go/codec"
)

// 此文件主要为了防止dirtycheck每次都因为PlayerQuest.Has_closed是map类型，而导致codec结果不同的问题
// 用GobEncode将codec的结果稳定下来

var mh codec.MsgpackHandle

func init() {
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	mh.RawToString = true
	mh.WriteExt = true
}

type ClosedIds []uint32

func (cids ClosedIds) Len() int           { return len(cids) }
func (cids ClosedIds) Less(i, j int) bool { return cids[i] < cids[j] }
func (cids ClosedIds) Swap(i, j int)      { cids[i], cids[j] = cids[j], cids[i] }

func (p PlayerQuest) GobEncode() ([]byte, error) {
	db_info := playerQuestInDB{
		HasClosed:        make(ClosedIds, 0, len(p.Has_closed)),
		DailyHasClose:    p.Daily_has_closed,
		Account7HasClose: p.Account7_has_closed,
		HasReceived:      make([]quest, 0, p.Received_len),
	}
	closedIds := make(ClosedIds, 0, len(p.Has_closed))
	for k, _ := range p.Has_closed {
		closedIds = append(closedIds, k)
	}
	sort.Sort(closedIds)
	for _, cid := range closedIds {
		db_info.HasClosed = append(db_info.HasClosed, cid)
	}
	for i := 0; i < len(p.Received); i++ {
		if !p.Received[i].IsVailed() {
			db_info.HasReceived = append(db_info.HasReceived, p.Received[i])
		}
	}
	db_info.LastRefreshN = p.LastRefreshN

	var out []byte
	enc := codec.NewEncoderBytes(&out, &mh)
	err := enc.Encode(db_info)
	return out, err
}

func (p PlayerQuest) GobDecode(data []byte) error {
	db_info := playerQuestInDB{}
	dec := codec.NewDecoderBytes(data, &mh)
	if err := dec.Decode(&db_info); err != nil {
		return err
	}

	p.Received = db_info.HasReceived
	p.Received_len = len(db_info.HasReceived)
	p.Has_closed = make(hasClosedList, len(db_info.HasClosed))
	for _, qid := range db_info.HasClosed {
		p.Has_closed[qid] = qid
	}
	p.Daily_has_closed = db_info.DailyHasClose
	p.Account7_has_closed = db_info.Account7HasClose
	p.LastRefreshN = db_info.LastRefreshN
	return nil
}
