package linkext

import (
	"sync/atomic"

	"vcs.taiyouxi.net/platform/planx/funny/link"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// const (
//     Session_channel_command_join = iota
//     Session_channel_command_exit
//     Session_channel_command_broadcast
//     Session_channel_command_broadcast_others
//     Session_channel_command_close
// )

// type session_channel_command struct {
//     typ       int
//     session   *link.Session
//     accountid string
//     msg       []byte
// }

type Channel struct {
	ch *link.StringChannel
	// commands  chan session_channel_command
	closeFlag int32
}

func NewChannel() *Channel {
	channel := &Channel{
		ch: link.NewStringChannel(),
		// commands: make(chan session_channel_command),
	}
	// go channel.Start()
	return channel
}

// func (channel *Channel) Start() {
//     for {
//         command, ok := <-channel.commands
//         if !ok {
//             return
//         }
//         switch command.typ {
//         case Session_channel_command_join:
//             {
//                 //logs.Trace("join %d", command.session.ID())
//                 channel.join(command.session, command.accountid)
//             }
//         case Session_channel_command_exit:
//             {
//                 //logs.Trace("exit %d", command.session.ID())
//                 channel.delSession(command.accountid)
//             }
//         case Session_channel_command_broadcast:
//             {
//                 //logs.Trace("broadcast %v", command.msg)
//                 channel.broadcast(command.msg)
//             }
//         case Session_channel_command_broadcast_others:
//             {
//                 //logs.Trace("broadcast %v", command.msg)
//                 channel.broadcastOthers(command.msg, command.session)
//             }
//         case Session_channel_command_close:
//             {
//                 //logs.Trace("close")
//                 channel.close()
//             }
//         }
//     }
// }

// 玩家离开过程之前发送的消息容易引起IO阻塞,根据文档请使用codec_async转换Send函数为异步函数
func (channel *Channel) broadcast(msg []byte) error {
	channel.ch.Fetch(func(s *link.Session) {
		if err := s.Send(msg); err != nil {
			logs.Error("AsyncSend Err By %s", err.Error())
		}
	})
	return nil
}

// 玩家离开过程之前发送的消息容易引起IO阻塞,根据文档请使用codec_async转换Send函数为异步函数
func (channel *Channel) broadcastOthers(msg []byte, me *link.Session) error {
	channel.ch.Fetch(func(s *link.Session) {
		if s.Id() == me.Id() {
			return
		}
		if err := s.Send(msg); err != nil {
			logs.Error("AsyncSend Err session[%v] By %s", s.Id(), err.Error())
		}
	})

	return nil
}

func (channel *Channel) broadcastSomeOthers(msg []byte, ids map[uint64]struct{}) error {
	channel.ch.Fetch(func(s *link.Session) {
		if _, ok := ids[s.Id()]; !ok {
			return
		}
		if err := s.Send(msg); err != nil {
			logs.Error("AsyncSend Err session[%v] By %s", s.Id(), err.Error())
		}
	})

	return nil
}

func (channel *Channel) join(session *link.Session, accountId string) {
	channel.ch.Put(accountId, session)
}

func (channel *Channel) delSession(accountid string) bool {
	return channel.ch.Remove(accountid)
}

func (channel *Channel) close() {
	channel.ch.Close()
	// add by zhangzhen
	if atomic.CompareAndSwapInt32(&channel.closeFlag, 0, 1) {
		// close(channel.commands)
	}
}

func (channel *Channel) IsClosed() bool { return atomic.LoadInt32(&channel.closeFlag) != 0 }

func (channel *Channel) Broadcast(msg []byte) {
	channel.broadcast(msg)
	// channel.commands <- session_channel_command{
	//     typ: Session_channel_command_broadcast,
	//     msg: msg,
	// }
}

func (channel *Channel) BroadcastOthers(msg []byte, session *link.Session) {
	channel.broadcastOthers(msg, session)
	// channel.commands <- session_channel_command{
	//     typ:     Session_channel_command_broadcast_others,
	//     session: session,
	//     msg:     msg,
	// }
}

func (channel *Channel) BroadcastSomeOthers(msg []byte, ids map[uint64]struct{}) {
	channel.broadcastSomeOthers(msg, ids)
}

func (channel *Channel) Join(session *link.Session, accountId string) {
	channel.join(session, accountId)
	// channel.commands <- session_channel_command{
	//     typ:       Session_channel_command_join,
	//     session:   session,
	//     accountid: accountId,
	// }
}

func (channel *Channel) Exit(accountId string) {
	channel.delSession(accountId)
	// channel.commands <- session_channel_command{
	//     typ:     Session_channel_command_exit,
	//     session: session,
	//	   accountid:accountId,
	// }
}

func (channel *Channel) Close() {
	channel.close()
	// channel.commands <- session_channel_command{
	//     typ: Session_channel_command_close,
	// }
}
