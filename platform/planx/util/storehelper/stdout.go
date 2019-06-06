package storehelper

import (
	"fmt"
	"sync/atomic"

	"vcs.taiyouxi.net/platform/planx/util"
)

var stdoutChan chan string
var stdoutWorkers uint64
var stdoutWait util.WaitGroupWrapper

func init() {
	stdoutWorkers = 0
	stdoutChan = make(chan string, 64)
	stdoutWait.Wrap(
		func() {
			for s := range stdoutChan {
				fmt.Println(s)
			}
		})
}
func NewStoreStdout() *StdoutStore {
	atomic.AddUint64(&stdoutWorkers, 1)
	//logs.Info("NewStdoutStore %d", n)
	return &StdoutStore{}
}

type StdoutStore struct {
}

func (cs *StdoutStore) Clone() (IStore, error) {
	atomic.AddUint64(&stdoutWorkers, 1)
	//logs.Info("NewStdoutStore Clone %d", n)
	return &StdoutStore{}, nil
}

func (cs *StdoutStore) Put(key string, val []byte, rh ReadHandler) error {
	stdoutChan <- fmt.Sprintf("%s %s", key, string(val))
	return nil
}

func (cs *StdoutStore) StoreKey(key string) string {
	return key
}

func (cs *StdoutStore) Open() error {
	return nil
}
func (cs *StdoutStore) Get(key string) ([]byte, error) {
	return nil, fmt.Errorf("stdout Get not support")
}
func (cs *StdoutStore) Del(key string) error {
	return fmt.Errorf("stdout Del not support")
}
func (cs *StdoutStore) RedisKey(key_in_store string) (string, bool) {
	return "", false
}
func (cs *StdoutStore) Close() error {
	final := atomic.AddUint64(&stdoutWorkers, ^uint64(0))
	//logs.Info("yzh,%d", final)
	if final == 1 { //逻辑上多次推敲确认这里逻辑应该是1为最后一个
		close(stdoutChan)
		stdoutWait.Wait()
	}
	return nil
}
