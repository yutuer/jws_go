package redeemCodeModule

import (
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func genredeemCodeModules(sid uint) *redeemCodeModules {
	return &redeemCodeModules{
		sid: sid,
	}
}

type command struct {
	typ     int
	acID    string
	code    string
	resChan chan RedeemCodeExchange
}

const (
	cmdTypSetCodeUsed = iota
	cmdTypGetCodeData
)

type redeemCodeModules struct {
	sid          uint
	waitter      sync.WaitGroup
	command_chan chan command
}

func (r *redeemCodeModules) AfterStart(g *gin.Engine) {
}

func (r *redeemCodeModules) BeforeStop() {
}

func (r *redeemCodeModules) Start() {
	r.waitter.Add(1)
	defer r.waitter.Done()

	err := initDB()
	if err != nil {
		logs.Error("redeemCodeModules DB Open Err by %s", err.Error())
		return
	}

	r.command_chan = make(chan command, 256)

	r.waitter.Add(1)
	go func() {
		defer r.waitter.Done()
		for {
			command, ok := <-r.command_chan
			logs.Trace("redeemCodeModules command %v", command)

			if !ok {
				logs.Warn("redeemCodeModules command_chan close")
				return
			}

			switch command.typ {
			case cmdTypGetCodeData:
				if command.resChan != nil {
					data := getRedeemCodeData(command.code)
					if data == nil {
						//logs.Error("getRedeemCodeData nil By %s", command.code)
						command.resChan <- RedeemCodeExchange{}
					} else {
						command.resChan <- *data
					}
				}
			case cmdTypSetCodeUsed:
				setRedeemCodeUsed(command.acID, command.code)
				//if err != nil {
				//	logs.SentryLogicCritical(command.acID, "SetCodeUsed Err %s", err.Error())
				//}
			}
		}
	}()

}

func (r *redeemCodeModules) Stop() {
	close(r.command_chan)
	r.waitter.Wait()
}

//SendCodeUsed 设置一个Code已经被使用
//这里其实有个问题就是两个人同时使用Code时可能两次都会被通过,
//但是一个人只能使用同批次组号的Code一次,所以没什么问题
func SetCodeUsed(shardId uint, acID, code string) {
	execCmdASync(shardId, command{
		typ:     cmdTypSetCodeUsed,
		acID:    acID,
		code:    code,
		resChan: nil,
	})
}

//GetCodeData 获取Code是否被使用过
func GetCodeData(shardId uint, code string) RedeemCodeExchange {
	resChan := make(chan RedeemCodeExchange, 1)
	return execCmd(shardId, command{
		typ:     cmdTypGetCodeData,
		code:    code,
		resChan: resChan,
	})
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func execCmdASync(shardId uint, cmd command) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case GetModule(shardId).command_chan <- cmd:
	case <-ctx.Done():
		logs.Error("redeemcode execCmdASync put timeout")
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func execCmd(shardId uint, cmd command) RedeemCodeExchange {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	select {
	case GetModule(shardId).command_chan <- cmd:
	case <-ctx.Done():
		logs.Error("redeemcode execCmdASync put timeout")
	}

	select {
	case res := <-cmd.resChan:
		return res
	case <-ctx.Done():
		logs.Error("redeemcode execCmd <-res_chan timeout")
		return RedeemCodeExchange{}
	}
}
