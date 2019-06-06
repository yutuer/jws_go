package global_count

import (
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	_ = iota
	GlobalCount_Cmd_GetInfo
	GlobalCount_Cmd_DelAndGet
	GlobalCount_Cmd_AddAndGet
)

type GlobalCountCmd struct {
	CmdTyp   int
	CountTyp string
	Gid      uint
	Sid      uint
	Key      GlobalCountKey
	resChan  chan GlobalCountRet
}

type GlobalCountKey struct {
	IId uint32 `json:"i"`
	SId string `json:"s"`
}

type GlobalCountRet struct {
	Counti2c map[uint32]uint32
	Counts2c map[string]uint32
	Success  bool
}

type worker struct {
	waitter   util.WaitGroupWrapper
	cmd_chan  chan GlobalCountCmd
	save_chan chan saveInfo
}

func (w *worker) start(m *GlobalCountModule) {
	w.cmd_chan = make(chan GlobalCountCmd, 2048)
	w.save_chan = make(chan saveInfo, 2048)

	w.waitter.Wrap(func() {
		for cmd := range w.cmd_chan {
			func() {
				//by YZH 这个让parent never dead, 应该如此吗？
				defer logs.PanicCatcherWithInfo("global_count Worker Panic")
				w.processCommand(m, &cmd)
			}()
		}
		close(w.save_chan)
	})

	w.waitter.Wrap(func() {
		for fr := range w.save_chan {
			if err := m.dbSave(&fr.kvs); err != nil {
				logs.Error("GlobalCountModule save err: %s", err.Error())
			}
		}
	})
}

func (w *worker) stop() {
	close(w.cmd_chan)
	w.waitter.Wait()
}

func (w *worker) processCommand(m *GlobalCountModule, cmd *GlobalCountCmd) {
	ret := GlobalCountRet{}
	gct := cmd.CountTyp

	err, kvs := m.dbLoad(gct)
	if err != nil {
		logs.Error("global_count processCommand dbLoad %v err %v", gct, err)
		cmd.resChan <- ret
		return
	}
	switch cmd.CmdTyp {
	case GlobalCount_Cmd_GetInfo:
		w.getInfo(cmd, kvs)
	case GlobalCount_Cmd_DelAndGet:
		w.delAndGet(cmd, kvs)
	case GlobalCount_Cmd_AddAndGet:
		w.addAndGet(cmd, kvs)
	}
}

func (w *worker) getInfo(cmd *GlobalCountCmd, kvs *countkvs) {
	ret := &GlobalCountRet{
		Success: true,
	}
	kvs.tranKvs2Ret(ret)
	cmd.resChan <- *ret
}

func (w *worker) delAndGet(cmd *GlobalCountCmd, kvs *countkvs) {
	ret := &GlobalCountRet{}
	for i, _ := range kvs.Counts {
		kv := &kvs.Counts[i]
		if kv.Key == cmd.Key {
			if kv.Count > 0 {
				kv.Count--
				kvs.tranKvs2Ret(ret)
				ret.Success = true
				cmd.resChan <- *ret
				w.save_chan <- saveInfo{
					kvs: *kvs,
				}
				return
			} else {
				break
			}
		}
	}
	cmd.resChan <- *ret
}

func (w *worker) addAndGet(cmd *GlobalCountCmd, kvs *countkvs) {
	ret := &GlobalCountRet{}
	for i, _ := range kvs.Counts {
		kv := &kvs.Counts[i]
		if kv.Key == cmd.Key {
			kv.Count++
			kvs.tranKvs2Ret(ret)
			ret.Success = true
			cmd.resChan <- *ret
			w.save_chan <- saveInfo{
				kvs: *kvs,
			}
			return
		}
	}
	cmd.resChan <- *ret
}

type saveInfo struct {
	kvs countkvs
}
