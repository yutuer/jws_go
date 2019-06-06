package base

import (
	"container/list"

	"time"

	"strings"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type PlayerMsgInfoFromNet struct {
	From       string `json:"from"`
	SerialNo   string `json:"serial_no"`
	MsgID      string `json:"msg_id"`
	SnderID    string `json:"snder_id"`
	RecverID   string `json:"recver_id"`
	MsgType    uint   `json:"msg_type"`
	ChatType   uint   `json:"chat_type"`
	MsgContent string `json:"msg_content"`
	CTime      string `json:"c_time"`
}

type PlayerMsgInfo struct {
	MsgID      string
	SenderAcID string
	ChatType   uint
	MsgContent string
	CreateTime string
}

func (pmi *PlayerMsgInfos) AddItem(newMsg PlayerMsgInfo) {
	if pmi == nil {
		*pmi = make([]PlayerMsgInfo, 0, Cfg.MaxRetainMsgCount)
	}
	*pmi = append(*pmi, newMsg)
	if len(*pmi) > Cfg.MaxRetainMsgCount {
		*pmi = (*pmi)[1:]
	}
}

type PlayerMsgInfos []PlayerMsgInfo
type handleFunc func(newMsg PlayerMsgInfo) *MuteInfo
type timerHandleFunc func(infos *MsgInfo) []*MuteInfo

type MsgInfo struct {
	MsgInfoMap  map[string]*list.Element
	MsgInfoList *list.List
}

var (
	MsgInfos        *MsgInfo
	MsgChan         chan PlayerMsgInfoFromNet
	TimerChan       <-chan time.Time
	MsgHandler      []handleFunc
	MsgTimerHandler []timerHandleFunc
)

func init() {
	MsgInfos = &MsgInfo{
		MsgInfoMap:  make(map[string]*list.Element, 0),
		MsgInfoList: list.New(),
	}
	MsgChan = make(chan PlayerMsgInfoFromNet, 1024)
	TimerChan = time.After(time.Second)
	MsgHandler = make([]handleFunc, 0)
	MsgTimerHandler = make([]timerHandleFunc, 0)
}

func ResetTimer() {
	TimerChan = time.After(time.Second * 60)
}

func AddMsgHandler(handler handleFunc) {
	MsgHandler = append(MsgHandler, handler)

}

func AddTimerMsgHandler(handler timerHandleFunc) {
	MsgTimerHandler = append(MsgTimerHandler, handler)
}

func GenMsgInfo(netMsg *PlayerMsgInfoFromNet) PlayerMsgInfo {
	// 获取实际的聊天内容
	msgContent := strings.Split(netMsg.MsgContent, ";")
	return PlayerMsgInfo{
		MsgID:      netMsg.MsgID,
		SenderAcID: netMsg.SnderID,
		ChatType:   netMsg.ChatType,
		MsgContent: msgContent[len(msgContent)-1],
		CreateTime: netMsg.CTime,
	}
}

func (mi *MsgInfo) AddMsg(newMsg PlayerMsgInfo) {
	var msgValue PlayerMsgInfos
	e, ok := mi.MsgInfoMap[newMsg.SenderAcID]
	if ok {
		msgValue = ParseMsgValue(e.Value)
		mi.MsgInfoList.Remove(e)
	}
	msgValue.AddItem(newMsg)
	newE := mi.MsgInfoList.PushBack(msgValue)
	mi.MsgInfoMap[newMsg.SenderAcID] = newE

	if mi.MsgInfoList.Len() > Cfg.MaxListLen {
		e := mi.MsgInfoList.Front()
		infos := ParseMsgValue(e.Value)
		if len(infos) > 0 {
			mi.Remove(infos[0].SenderAcID)
		}
	}
}

func (mi *MsgInfo) Remove(acID string) {
	e, ok := mi.MsgInfoMap[acID]
	if ok {
		delete(mi.MsgInfoMap, acID)
	}
	mi.MsgInfoList.Remove(e)
}

func ParseMsgValue(value interface{}) PlayerMsgInfos {
	msgValue, ok := value.(PlayerMsgInfos)
	if !ok {
		logs.Error("fatal convert type error")
		return nil
	}
	return msgValue
}
