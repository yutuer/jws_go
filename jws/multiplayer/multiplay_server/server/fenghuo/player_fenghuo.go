package fenghuo

import (
	"runtime/debug"

	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/msgprocessor"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/fenghuomsg"
	"vcs.taiyouxi.net/platform/planx/funny/link"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	PlayerStatusWaiting = 0
	PlayerStatusHPSync  = 1
	PlayerStatusDead    = 2
)

type FenghuoPlayer struct {
	msgprocessor.ServiceImp
	session *link.Session
	game    *FenghuoGame
	AcID    string

	PlayerStatus int
	IDX          int
}

func NewFenghuoPlayer(s *link.Session) *FenghuoPlayer {
	r := &FenghuoPlayer{
		session:      s,
		PlayerStatus: PlayerStatusWaiting,
		IDX:          0,
	}
	r.DecodePacket = func(buf []byte) msgprocessor.IPacket {
		return fenghuomsg.GetRootAsPacket(buf, 0)
	}
	r.Init()
	return r
}

func (p *FenghuoPlayer) GetAcID() string {
	return p.AcID
}

func (p *FenghuoPlayer) GetSession() *link.Session {
	return p.session
}

func (p *FenghuoPlayer) Start() {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("Player Panic, session Err %v", err)
			debug.PrintStack()
		}
	}()

	for {
		var msg []byte
		//logs.Warn("for %v", p.session.IsClosed())
		err := p.session.Receive(&msg)
		if err != nil {
			logs.Trace("recv err, session  By %s", err.Error())
			//if p.game != nil && p.AcID != "" {
			//	p.game.PlayerLoss(p.AcID)
			//}
			return
		}

		if msg == nil || len(msg) == 0 {
			logs.Error("recv err, session  By msg nil")
			//if p.game != nil && p.AcID != "" {
			//	p.game.PlayerLoss(p.AcID)
			//}
			return
		}

		//logs.Trace("recv : %s", msg)
		if msg != nil {
			err := p.ProcessMsg(p.session, msg)
			if err != nil {
				logs.Error("processMsg err, session  By %s", err.Error())
				//if p.game != nil && p.AcID != "" {
				//	p.game.PlayerLoss(p.AcID)
				//}
				return
			}
		}

	}
}

func (p *FenghuoPlayer) Stop() {
}
