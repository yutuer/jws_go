package chatserver

import (
	// "fmt"

	"fmt"
	"sync/atomic"
	"time"

	"unsafe"

	"vcs.taiyouxi.net/platform/planx/funny/link"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/chat/chatserver/proto/gen"
)

type Player struct {
	session  *link.Session
	room     *Room
	cityName string
	Name     string
	isReg    int32 // 状态: 是否Reg了
	roleInfo *gen.RoleInfo
	rolePos  unsafe.Pointer
}

func NewPlayer(s *link.Session) *Player {
	r := &Player{
		session: s,
		Name:    "NoName",
	}
	return r
}

func (p *Player) Start() {
	defer logs.PanicCatcherWithInfo("Player Panic, session[%v]", p.session.Id())
	go func() { // 检查建立连接但不Reg的情况
		select {
		case <-timingWheel.After(time.Second):
			if atomic.LoadInt32(&p.isReg) <= 0 {
				logs.Error("session[%d] connect but no reg, kick ... ", p.session.Id())
				p.session.Send(Gen_ErrMsg("Err:NoReg"))
				go func() {
					time.Sleep(time.Second)
					p.session.Close()
				}()
			}
		}
	}()

	defer func() {
		if p.roleInfo != nil {
			Rec_UnRegNotify(p, string(p.roleInfo.AccountId()), string(p.roleInfo.Name()))
		}
	}()

	if err := p.session.Send(Gen_PingMsg("Ping:" + fmt.Sprintf("%d", 1001))); err != nil {
		logs.Error("serv ping err: %s", err.Error())
		return
	}

	go func() { // 在业务层，服务器给客户端的ping
		i := 0
		for {
			select {
			case <-timingWheel.After(PingWait):
				if err := p.session.Send(Gen_PingMsg("Ping:" + fmt.Sprintf("%d", i))); err != nil {
					logs.Error("serv ping err: %s", err.Error())
					return
				}
			}
			i = i + 1
		}
	}()

	for {
		//TODO YZH optmize
		var msg []byte
		err := p.session.Receive(&msg)
		if err != nil {
			logs.Error("recv err, session[%d]  By %s", p.session.Id(), err.Error())
			break
		}

		// logs.Trace("recv : %v", msg)
		if msg != nil && len(msg) > 0 {
			p.ProcessMsg(msg)
		}
	}

	// for {
	//     select {
	//     case <-timingWheel.After(time.Second):
	//         logs.Info("Player is alive")
	//     }
	// }
}

func (p *Player) AccountID() string {
	if p.roleInfo != nil {
		return string(p.roleInfo.AccountId())
	}
	return ""
}

func (p *Player) Stop() {
	p.session.Send(Gen_ReassignNotify())
}

func (p *Player) ProcessMsg(msg []byte) {
	RecProto(p, msg)
}

func (p *Player) setPos(pPos *gen.RolePos) {
	atomic.StorePointer(&p.rolePos, unsafe.Pointer(pPos))
}

func (p *Player) getPos() *gen.RolePos {
	return (*gen.RolePos)(atomic.LoadPointer(&p.rolePos))
}
