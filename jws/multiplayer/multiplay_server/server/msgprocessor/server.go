package msgprocessor

import (
	"errors"

	flatbuffers "github.com/google/flatbuffers/go"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	MsgTypNull = iota
	MsgTypRpc
	MsgTypNotify
	MsgTypPush
	MsgTypForward
)

type ISession interface {
	SendMsg(msg []byte) error
}

type IPacket interface {
	Typ() int32
	Number() int64
	Code() int32
	DataType() byte
	Data(obj *flatbuffers.Table) bool
}

type RpcHandleFuncTyp func(req IPacket) []byte

func (f RpcHandleFuncTyp) Serve(r IPacket) []byte {
	return f(r)
}

type RpcHandler interface {
	Serve(IPacket) []byte
}

type NotifyHandleFuncTyp func(req IPacket)

func (f NotifyHandleFuncTyp) Serve(r IPacket) {
	f(r)
}
func (f NotifyHandleFuncTyp) Forward(msg []byte, typ byte) {

}

type NotifyHandler interface {
	Serve(IPacket)
}

type ServiceImp struct {
	rpc          []RpcHandler
	notify       []NotifyHandler
	forward      func(msg []byte)
	DecodePacket func(buf []byte) IPacket
}

func (s *ServiceImp) OnRpc(typ int, h RpcHandleFuncTyp) {
	for typ >= len(s.rpc) {
		s.rpc = append(s.rpc, RpcHandleFuncTyp(func(req IPacket) []byte {
			return []byte{}
		}))
	}
	s.rpc[typ] = h
}

func (s *ServiceImp) OnNotify(typ int, h NotifyHandleFuncTyp) {
	for typ >= len(s.notify) {
		s.notify = append(s.notify, NotifyHandleFuncTyp(func(req IPacket) {
			return
		}))
	}
	s.notify[typ] = h
}

func (s *ServiceImp) OnForward(f func(msg []byte)) {
	s.forward = f
}

func (s *ServiceImp) Forward(msg []byte) {
	s.forward(msg)
}

var (
	ErrMsgTypeErr = errors.New("ErrMsgTypeErr")
	ErrMsgIDErr   = errors.New("ErrMsgIDErr")
)

type SendI interface {
	Send(msg interface{}) (err error)
}

func (s *ServiceImp) ProcessMsg(session SendI, msg []byte) error {
	var (
		msgType int
		msgID   int
		rsp     []byte
	)
	//logs.Trace("processMsg %v", msg)
	if msg == nil || len(msg) == 0 {
		return nil
	}
	packet := s.DecodePacket(msg)
	msgType = int(packet.Typ())
	msgID = int(packet.DataType())
	//logs.Trace("processMsg %v %d", msgType, msgID)

	switch msgType {
	case MsgTypRpc:
		if msgID < 0 || msgID >= len(s.rpc) {
			logs.Error("MsgTypRpc ErrMsgIDErr %d %v", msgID, s.notify)
			return ErrMsgIDErr
		}
		var checker flatbuffers.Table
		if packet.Data(&checker) {
			unionType := packet.DataType()
			rsp = s.rpc[unionType].Serve(packet)
		}
		if rsp == nil {
			return ErrMsgIDErr
		}
		return session.Send(rsp)
	case MsgTypNotify:
		if msgID < 0 || msgID >= len(s.notify) {
			logs.Error("MsgTypNotify ErrMsgIDErr %d %v", msgID, s.notify)
			return ErrMsgIDErr
		}
		s.notify[msgID].Serve(packet)
		return nil
	case MsgTypForward:
		s.Forward(msg)
		return nil
	}

	return ErrMsgTypeErr
}
