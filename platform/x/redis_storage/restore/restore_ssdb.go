package restore

import (
	"errors"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

//
// S3遍历类
//

type SSDBKeysHander func(keys []string) error

type ssdbScanner struct {
	conn     *storehelper.StoreHISSDB
	last_idx string

	hander   SSDBKeysHander
	scan_len int64
}

func NewSSDBScanner(conn *storehelper.StoreHISSDB, hander SSDBKeysHander, scan_len int64) *ssdbScanner {
	return &ssdbScanner{
		conn:     conn,
		last_idx: "",
		hander:   hander,
		scan_len: scan_len,
	}
}

func (r *ssdbScanner) Start() error {
	turn := 0
	for {
		logs.Info("Start %d %d %s", turn, r.scan_len, r.last_idx)
		keys, err := r.conn.ListObject(r.last_idx, r.scan_len)
		if err != nil {
			return err
		}

		err = r.hander(keys)
		if err != nil {
			return err
		}

		if len(keys) == 0 {
			break
		}
		r.last_idx = keys[len(keys)-1]

		turn++
	}

	return nil
}

//

type RestoreSSDB struct {
	master RestoreMaster
	land   *storehelper.StoreHISSDB
}

func (r *RestoreSSDB) SetStore(s storehelper.IStore) {
	r.land = s.(*storehelper.StoreHISSDB)
}

func (r *RestoreSSDB) cloneNewSafeReader() (storehelper.IStore, error) {
	return r.land.Clone()
}

func (r *RestoreSSDB) close() {
	r.land.Close()
}

func (r *RestoreSSDB) restore(job RestoreJob,
	conn storehelper.IStore,
	isNeedOnLand func(string) bool) error {

	key := job.key
	res, err := conn.Get(key)
	if err != nil {
		logs.Error("Get %s Err %s", key, err.Error())
		return err
	}

	key_in_redis, ok := conn.RedisKey(key)
	if !ok {
		logs.Error("RedisKey Err %s", key)
		return errors.New("RedisKey Err")
	}

	if !isNeedOnLand(key_in_redis) {
		return nil
	}

	return r.master.RestoreToDB(key_in_redis, res)
}

func (r *RestoreSSDB) RestoreOne(key string, value []byte) error {
	r.master.AddJob(RestoreJob{key, nil, nil})
	return nil
}

func (r *RestoreSSDB) RestoreOneWaitRes(key string, value []byte, res restoreOneResChan) error {
	r.master.AddJob(RestoreJob{key, nil, res})
	return nil
}

func (r *RestoreSSDB) RestoreAll() error {
	// 创建Store
	l, err := r.land.Clone()
	// err := l.Open()
	if err != nil {
		logs.Error("Open store err by %s", err.Error())
		return err
	}

	ssdb_store, ok := l.(*storehelper.StoreHISSDB)
	if !ok {
		logs.Error("store is not StoreHISSDB")
		return errors.New("store is not StoreHISSDB")
	}

	num := 0

	hander := func(keys []string) error {
		//logs.Trace("keys hander")
		for i, key := range keys {
			if num%1000 == 0 {
				logs.Trace("key has dump %d %d %s",
					i, num, key)
			}
			num++
			r.RestoreOne(key, nil)
		}
		return nil
	}

	scanner := NewSSDBScanner(ssdb_store, hander, int64(r.master.GetScanLen()))
	return scanner.Start()
}
