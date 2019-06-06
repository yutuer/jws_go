package gve

import (
	"runtime/debug"

	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/msgprocessor"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/multiplayMsg"
	"vcs.taiyouxi.net/platform/planx/funny/link"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type Player struct {
	msgprocessor.ServiceImp
	session *link.Session
	game    *GVEGame
	AcID    string
}

func NewPlayer(s *link.Session) *Player {
	r := &Player{
		session: s,
	}
	r.DecodePacket = func(buf []byte) msgprocessor.IPacket {
		return multiplayMsg.GetRootAsPacket(buf, 0)
	}
	r.Init()
	return r
}

//func (p *Player) SendMsg(msg []byte) error {
//	logs.Trace("SendMsg To %s : %v", p.AcID, msg)
//	return p.session.Send(msg)
//}

func (p *Player) GetAcID() string {
	return p.AcID
}

func (p *Player) GetSession() *link.Session {
	return p.session
}

func (p *Player) Start() {
	defer func() {
		p.session.Close()
		if err := recover(); err != nil {
			logs.Error("Player Panic, session Err %v", err)
			debug.PrintStack()
		}
		logs.Info("player close %s", p.AcID)
	}()

	logs.Trace("%s multiplay start", p.AcID)
	for {
		var msg []byte
		err := p.session.Receive(&msg)
		if err != nil {
			logs.Trace("%s multiplay recv err, session  By %s", p.AcID, err.Error())
			if p.game != nil && p.AcID != "" {
				p.game.PlayerLoss(p.AcID)
			}
			return
		}

		if msg == nil || len(msg) == 0 {
			logs.Error("%s multiplay recv err, session  By msg nil", p.AcID)
			if p.game != nil && p.AcID != "" {
				p.game.PlayerLoss(p.AcID)
			}
			return
		}

		//logs.Trace("recv : %s", msg)
		if msg != nil {
			err := p.ProcessMsg(p.session, msg)
			if err != nil {
				logs.Error("%s multiplay processMsg err, session  By %s", p.AcID, err.Error())
				if p.game != nil && p.AcID != "" {
					p.game.PlayerLoss(p.AcID)
				}
				return
			}
		}
	}
}

func (p *Player) Stop() {

}
