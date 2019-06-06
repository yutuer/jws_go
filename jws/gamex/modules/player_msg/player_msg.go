package player_msg

import (
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"

	"strconv"
	"strings"

	"vcs.taiyouxi.net/jws/gamex/models/codec"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func genPlayerMsgModule(sid uint) *PlayerMsgModule {
	return &PlayerMsgModule{}
}

type PlayerMsgChan struct {
	request  chan servers.Request
	response chan servers.Response
}

type PlayerMsgModule struct {
	mutex      sync.RWMutex
	playerChan map[string]*PlayerMsgChan
}

type RspInfo struct {
	Ret     int    `json:"ret"`
	Diamond string `json:"diamond"`
}

func (r *PlayerMsgModule) AfterStart(g *gin.Engine) {
	logs.Debug("After Start Player_Msg")
	g.POST("/h5shop/CostDiamondGamex", func(c *gin.Context) {
		id := c.PostForm("id")
		serverid := c.PostForm("serverid")
		num := c.PostForm("num")
		info := PlayerCostDiamondInfo{
			id,
			serverid,
			num,
		}
		logs.Debug("received info id:%s  serverid:%s  num:%s", id, serverid, num)
		rep := SendBySync(id, Playerh5MsgCostDiamond, info)
		res, err := strconv.Atoi(rep.Code)
		if err != nil {
			logs.Debug("Convert to int err")
		}
		logs.Trace("[h5shop/CostDiamond] GetMsg %s %d", info.Roleid, res)
		c.JSON(200, RspInfo{res, string(rep.RawBytes)})
	})

}

func (r *PlayerMsgModule) BeforeStop() {
}

func (r *PlayerMsgModule) Start() {
	r.playerChan = make(map[string]*PlayerMsgChan, 4096)
}

func (r *PlayerMsgModule) Stop() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.playerChan = make(map[string]*PlayerMsgChan)
}

func (r *PlayerMsgModule) IsOnline(acID string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	_, ok := r.playerChan[acID]
	return ok
}

func (r *PlayerMsgModule) OnPlayerLogin(accountID string, channel chan servers.Request) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.playerChan[accountID] = &PlayerMsgChan{
		request:  channel,
		response: make(chan servers.Response),
	}
}

func (r *PlayerMsgModule) OnPlayerLogout(accountID string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, ok := r.playerChan[accountID]
	if ok {
		delete(r.playerChan, accountID)
	}
	return ok
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *PlayerMsgModule) SendMsg(accountID string, msg servers.Request) {
	if r.playerChan == nil {
		return
	}
	r.mutex.RLock()
	msgChan, ok := r.playerChan[accountID]
	r.mutex.RUnlock()

	if ok {
		ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
		defer cancel()

		select {
		case msgChan.request <- msg:
		case <-ctx.Done():
			logs.Error("[PlayerMsgModule] %s SendMsg chann full, cmd put timeout, this msg %v",
				accountID, msg)
			return
		}
		logs.Trace("[PlayerMsgModule] SendMsg %s %v", accountID, msg)
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *PlayerMsgModule) SendMsgToPlayers(accountIDs []string, msg servers.Request) {
	var allchan []chan servers.Request
	allchan = make([]chan servers.Request, 0, len(accountIDs))
	func() {
		r.mutex.RLock()
		defer r.mutex.RUnlock()
		for _, accountID := range accountIDs {
			channel, ok := r.playerChan[accountID]
			if ok {
				allchan = append(allchan, channel.request)
			}
		}
	}()

	for _, chanRequest := range allchan {
		//go func() { // 为了保证顺序，这里不能go 出去
		//ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
		//defer cancel()

		select {
		case chanRequest <- msg:
		default:
			// 若某一个玩家的chan阻塞，则丢掉，不能影响其他人 TODO 是否有问题？兵临城下
			logs.Warn("SendMsgToPlayers chan full")
			//case <-ctx.Done():
			//	logs.Error("[PlayerMsgModule] SendMsgs chann full, cmd put timeout")
			//	return
		}
		logs.Trace("[PlayerMsgModule] SendMsg")
		//}()
	}
}

func Send(accountID, code string, msg interface{}) {
	if accountID == "" {
		return
	}
	if strings.Index(accountID, "wspvp") == 0 {
		return
	}
	a, err := db.ParseAccount(accountID)
	if err == nil {
		module := GetModule(a.ShardId)
		if module != nil {
			module.SendMsg(accountID, servers.Request{
				Code:     code,
				RawBytes: codec.Encode(msg),
			})

		}
	} else {
		logs.Error("player_msg send accountId err %s", accountID)
	}
}

func SendToPlayers(accountID []string, code string, msg interface{}) {
	if len(accountID) <= 0 || accountID[0] == "" {
		return
	}
	a, err := db.ParseAccount(accountID[0])
	if err == nil {
		GetModule(a.ShardId).SendMsgToPlayers(accountID, servers.Request{
			Code:     code,
			RawBytes: codec.Encode(msg),
		})
	} else {
		logs.Error("player_msg send accountId err %s", accountID[0])
	}
}

// 同步等待返回结果
func SendBySync(accountID, code string, msg interface{}) *servers.Response {
	if accountID == "" {
		return nil
	}
	if strings.Index(accountID, "wspvp") == 0 {
		return nil
	}
	a, err := db.ParseAccount(accountID)
	if err == nil {
		module := GetModule(a.ShardId)
		if module != nil {
			return module.SendMsgBySync(accountID, servers.Request{
				Code:     code,
				RawBytes: codec.Encode(msg),
			})

		}
	} else {
		logs.Error("player_msg send accountId err %s", accountID)
	}
	return nil
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (r *PlayerMsgModule) SendMsgBySync(accountID string, msg servers.Request) *servers.Response {
	if r.playerChan == nil {
		return nil
	}
	r.mutex.RLock()
	msgChan, ok := r.playerChan[accountID]
	r.mutex.RUnlock()

	if ok {
		ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
		defer cancel()

		select {
		case msgChan.request <- msg:
		case <-ctx.Done():
			logs.Error("[PlayerMsgModule] %s SendMsg chann full, cmd put timeout, this msg %v",
				accountID, msg)
			return nil
		}
		logs.Trace("[PlayerMsgModule] SendMsg %s %v", accountID, msg)

		rspCtx, rspCancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
		defer rspCancel()

		select {
		case rsp := <-msgChan.response:
			return &rsp
		case <-rspCtx.Done():
			logs.Error("[PlayerMsgModule] %s wait response, cmd put timeout, this msg %v",
				accountID, msg)
			return nil
		}
	}
	return nil
}

// 同步等待返回结果
func SendResponseBySync(accountID string, resp *servers.Response) {
	if accountID == "" {
		return
	}
	a, err := db.ParseAccount(accountID)
	if err == nil {
		module := GetModule(a.ShardId)
		if module != nil {
			module.mutex.RLock()
			msgChan, ok := module.playerChan[accountID]
			module.mutex.RUnlock()

			if ok {
				ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
				defer cancel()

				select {
				case msgChan.response <- *resp:
				case <-ctx.Done():
					logs.Error("[PlayerMsgModule] %s SendMsg chann full, cmd put timeout, this msg %v",
						accountID)
					return
				}
			}
		} else {
			logs.Error("player_msg send accountId err %s", accountID)
		}
	}
}
