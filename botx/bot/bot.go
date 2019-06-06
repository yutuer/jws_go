package bot

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/ugorji/go/codec"

	"vcs.taiyouxi.net/platform/planx/client"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

var mh codec.MsgpackHandle
var mhw codec.MsgpackHandle

func init() {
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	mh.RawToString = false
	mh.WriteExt = true
	mhw.MapType = reflect.TypeOf(map[string]interface{}(nil))
	mhw.RawToString = false
	mhw.WriteExt = true
}

type BotCounter interface {
	Up()
	Down()
	CCUAdd()
	CCUSub()
}

type BotRpsEvent interface {
	Fire(t, event string)
}

type LogEntry struct {
	LTime   int64  `json:"logtime"`
	LogType string `json:"type"`
	Acid    string `json:"accountid"`

	Pkt    string `json:"pkt,omitempty"`
	Bin    string `json:"bin,omitempty"`
	RType  int    `json:"rtype,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Action string `json:"uri,omitempty"`

	RandSeed int64 `json"randseed,omitempty"`

	Session string `json"session,omitempty"`
}

var OnlyMsg bool

type PlayerBot struct {
	logEntry chan LogEntry
	//botDone     chan struct{}
	speedFactor float64

	account    db.Account
	loginToken string

	agent     *client.PacketConnAgent
	agentQuit chan struct{}
	agentLock sync.Mutex

	stop     bool
	stopChan chan struct{}

	rpcAddr string
	gateSer string
}

//Stop 不可逆，调用后bot无法再次使用
func (bot *PlayerBot) Stop() {
	bot.stop = true
	close(bot.stopChan)
}

func (bot *PlayerBot) quit() {
	if bot == nil {
		return
	}
	bot.GoOffline()

	fmt.Println("bot quit!!")
	bot = nil
}

func (bot *PlayerBot) Run(input io.Reader, bc BotCounter, be BotRpsEvent) {
	if bc != nil {
		bc.Up()
	}
	defer func() {
		if bc != nil {
			bc.Down()
		}
	}()

	go bot.input(input)
	bot.behavior(bc, be)
	bot.quit()
}

func (bot *PlayerBot) behavior(bc BotCounter, be BotRpsEvent) {
	//fmt.Println("bot behavior start", bot.speedFactor)
	//defer fmt.Println("bot behavior exit")
	eventFire := func(t, e string) {
		if be != nil {
			be.Fire(t, e)
		}
	}

	defer eventFire("quit", "quit")

	var online bool
	ccuup := func() {
		online = true
		if bc != nil {
			bc.CCUAdd()
		}

	}
	ccudown := func() {
		if online {
			online = false
			if bc != nil {
				bc.CCUSub()
			}
		}
	}

	defer ccudown()

	if OnlyMsg {
		defer func() {
			logs.Trace("account:%s, bot try go offline.", bot.account.String())
			err := bot.GoOffline()
			switch err {
			case ErrAlreadyOffline:
			case nil:
				ccudown()
			}
		}()

		logs.Trace("bot try go online.")
		if err := bot.TryGoOnline(); err != nil {
			logs.Error("account:%s, session try go online error %s", bot.account.String(), err.Error())
			return
		} else {
			ccuup()
		}
	}

	var lastTime int64
	var resetLastTime bool
	for le := range bot.logEntry {
		curTime := le.LTime

		if lastTime != 0 {
			timeDiff := curTime - lastTime
			if bot.speedFactor > 0 {
				// We can speedup or slowdown execution based on speedFactor
				if bot.speedFactor != 1 {
					timeDiff = int64(float64(curTime-lastTime) / bot.speedFactor)
				}
				select {
				case <-time.After(time.Duration(timeDiff)):
				case <-bot.stopChan:
					//logs.Info("behavior quit by Stop")
					return
				}
			} else {
				select {
				case <-time.After(100 * time.Millisecond):
				case <-bot.stopChan:
					//logs.Info("behavior quit by Stop")
					return
				}
			}
		}

		switch le.LogType {
		case "record":
			if !online {
				logs.Error("bot agent is nil when record coming")
				return
			}
			if client.PacketID(le.RType) != client.PacketIDPingPong {
				//XXX:机器人不replay任何PING信息
				if le.Prefix == "req" {
					if bot.speedFactor < 0.000001 {
						logs.Info("%s %s", le.Action, le.Pkt)
					}
					pkt := client.PktFromGobBase64(le.Bin)
					if pkt != nil {
						action, mvalue := client.PktToMap(pkt)
						mvalue["debugtime"] = le.LTime / int64(time.Second)
						if le.Action == "PlayerAttr/ChangeNameRequest" {
							mvalue["name"] = le.Acid
						}
						pkt = client.MapToPkt(pkt.GetId(), action, mvalue)
						eventFire("record", action)
						if ok := bot.agent.SendPacket(pkt); !ok {
							return
						}

						//logs.Trace("Action: %s Len: %d", le.Action, len(pkt.GetBytes()))
					} else {
						logs.Error("Can not be extracted to Packet. %v", le)
						return
					}
				}
			}
		case "init":
			if !online {
				eventFire("init", "failed1")
				logs.Error("bot agent is not online.")
				return
			}
			r := RequestDebugOp{
				PassthroughID: uuid.NewV4().String(),
				DebugTime:     le.LTime / int64(time.Second),
				Type:          "SetSeed",
				P1:            le.RandSeed,
			}
			if !bot.sendDebug(r) {
				eventFire("init", "failed2")
				return
			}
			//			SetCreateTime
			r = RequestDebugOp{
				PassthroughID: uuid.NewV4().String(),
				DebugTime:     le.LTime / int64(time.Second),
				Type:          "SetCreateTime",
				P1:            le.LTime / int64(time.Second),
			}
			if !bot.sendDebug(r) {
				eventFire("init", "failed3")
				return
			}
			eventFire("init", "success")
			// /Debug/DebugOpRequest
		case "session":
			switch le.Session {
			case "online":
				resetLastTime = true
				//logs.Trace("bot try go online.")
				time.Sleep(time.Second) // 上线等1s，放在过快上线下线，db没有准备好
				if !OnlyMsg {
					eventFire("online", "online")
					if err := bot.TryGoOnline(); err != nil {
						if err == ErrAlreadyOnline {
							goto GOON
						} else {
							logs.Error("session try go online error %s", err.Error())
							return
						}
					} else {
						logs.Trace("account:%s, TryGoOnline send out", bot.account.String())
						ccuup()
					}
				}
			case "offline":
				lastTime = le.LTime

				//logs.Trace("bot try go offline.")
				if !OnlyMsg {
					logs.Trace("account:%s, GoOffline will happen", bot.account.String())
					eventFire("offline", "offline")
					err := bot.GoOffline()
					switch err {
					case ErrAlreadyOffline:
						goto GOON
					case nil:
						logs.Trace("account:%s, GoOffline happened", bot.account.String())
						ccudown()
					}
				}
			}

		}
	GOON:
		if resetLastTime {
			resetLastTime = false
			lastTime = 0
		} else {
			lastTime = le.LTime
		}

	}

}

func (bot *PlayerBot) input(input io.Reader) {
	defer close(bot.logEntry)
	reader := bufio.NewReader(input)
	line := []byte{}
	for {
		b, isPrefix, err := reader.ReadLine()
		if bot.stop {
			break
		}
		if len(b) <= 0 {
			break
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			break
		}
		line = append(line, b...)

		if isPrefix {
			continue
		} else {
			var le LogEntry
			err := json.Unmarshal(line, &le)
			if err != nil {
				fmt.Fprintln(os.Stderr, "reading standard input:", err, "line:", string(b))
				continue
			}
			if ok := checkLogEntry(le); ok {
				select {
				case bot.logEntry <- le:
				case <-bot.stopChan:
					//logs.Info("input done by Stop")
					return
				}
			}
			line = []byte{}
		}
	}
	//	scanner := bufio.NewScanner(input)
	//	for scanner.Scan() {
	//		var le LogEntry
	//		b := scanner.Bytes()
	//		err := json.Unmarshal(b, &le)
	//		if err != nil {
	//			fmt.Fprintln(os.Stderr, "reading standard input:", err, "line:", string(b))
	//			continue
	//		}
	//		if ok := checkLogEntry(le); ok {
	//			bot.logEntry <- le
	//		}
	//
	//	}
	//	if err := scanner.Err(); err != nil {
	//		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	//	}

}

type RequestDebugOp struct {
	PassthroughID string `codec:"passthrough"`
	DebugTime     int64  `codec:"debugtime"`
	Type          string `codec:"typ"`
	P1            int64  `codec:"p1"`
}

func (bot *PlayerBot) sendDebug(r RequestDebugOp) bool {
	rawBytes := encode(r)
	var out []byte
	enc := codec.NewEncoderBytes(&out, &mhw)
	enc.Encode("Debug/DebugOpRequest")
	enc.Encode(rawBytes)
	pkt := client.NewPacket(out, client.PacketIDReqResp)
	logs.Trace("account:%s, init send out", bot.account.String())
	if ok := bot.agent.SendPacket(pkt); !ok {
		return false
	}
	return true
}
