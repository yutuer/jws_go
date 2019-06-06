package team_pvp

import (
	"encoding/json"
	"errors"
	//"fmt"
	"math/rand"
	"strconv"
	"testing"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func TestSaveQueue(t *testing.T) {
	testAllFailure(t)
	testAllSuccess(t)
	testRandom(t)
}

func testAllFailure(t *testing.T) {
	worker := &dbWorker{}
	for i := 0; i < 10; i++ {
		cmd := &dbCmd{typ: DB_Cmd_Save, chg: make(map[int]helper.AccountSimpleInfo)}
		cmd.chg[i] = helper.AccountSimpleInfo{Name: strconv.Itoa(i)}
		saveChg(worker, 1, cmd, saveDbWithAllFailth)
	}
	if len(worker.failedCmdQueue) != 10 {
		t.FailNow()
	}

}

func testAllSuccess(t *testing.T) {
	worker := &dbWorker{}
	for i := 0; i < 10; i++ {
		cmd := &dbCmd{typ: DB_Cmd_Save, chg: make(map[int]helper.AccountSimpleInfo)}
		cmd.chg[i] = helper.AccountSimpleInfo{Name: strconv.Itoa(i)}
		saveChg(worker, 1, cmd, saveDbWithAllOk)
	}
	if len(worker.failedCmdQueue) != 0 {
		t.FailNow()
	}
}

func testRandom(t *testing.T) {
	worker := &dbWorker{}
	for i := 0; i < 10; i++ {
		cmd := &dbCmd{typ: DB_Cmd_Save, chg: make(map[int]helper.AccountSimpleInfo)}
		cmd.chg[i] = helper.AccountSimpleInfo{Name: strconv.Itoa(i)}
		saveChg(worker, 1, cmd, saveDbWithAllFailth)
	}
	//fmt.Println(worker.failedCmdQueue)
	for {
		if len(worker.failedCmdQueue) > 0 {
			cmd := &dbCmd{typ: DB_Cmd_Save, chg: make(map[int]helper.AccountSimpleInfo)}
			cmd.chg[0] = helper.AccountSimpleInfo{Name: strconv.Itoa(0)}
			saveChg(worker, 1, cmd, saveDbWithRandom)
			//fmt.Println(worker.failedCmdQueue)
		} else {
			break
		}
	}
	if len(worker.failedCmdQueue) > 0 {
		t.FailNow()
	}
}

func saveChg(w *dbWorker, sid uint, cmd *dbCmd, sFunc saveFunc) {
	cb := redis.NewCmdBuffer()

	if cmd.chg != nil && len(cmd.chg) > 0 {
		for r, p := range cmd.chg {
			jInfo, err := json.Marshal(p)
			if err != nil {
				logs.Error("TeamPvp dbWorker saveChg Marshal err %s", err.Error())
			}
			cb.Send("HSET", TableTeamPvpRank(sid), r, string(jInfo))
		}
	}

	if w.failedCmdQueue == nil {
		w.failedCmdQueue = make([]*dbCmd, 0, 1)
	}
	w.failedCmdQueue = append(w.failedCmdQueue, cmd)

	resultIndex := -1
	for i, tempCb := range w.failedCmdQueue {
		if err := sFunc(tempCb, i); err != nil {
			//fmt.Println("TeamPvp dbWorker saveChg DoCmdBufferWrapper err %s", err.Error())
			break
		} else {
			resultIndex = i
		}
	}

	w.failedCmdQueue = w.failedCmdQueue[resultIndex+1:]
}

type saveFunc func(cmd *dbCmd, random int) error

func saveDbWithAllFailth(cmd *dbCmd, random int) error {
	return errors.New("for test")
}

func saveDbWithAllOk(cmd *dbCmd, random int) error {
	return nil
}

func saveDbWithRandom(cmd *dbCmd, random int) error {
	if rand.Intn(2) > 0 || random < 2 {
		return nil
	}
	return errors.New("for test")
}
