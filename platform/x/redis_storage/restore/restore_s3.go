package restore

import (
	"errors"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

//
// S3遍历类
//

type S3KeysHander func(keys []string) error

type s3Scanner struct {
	conn     *storehelper.StoreS3
	last_idx string

	hander   S3KeysHander
	scan_len int64
}

func NewS3Scanner(conn *storehelper.StoreS3, hander S3KeysHander, scan_len int64) *s3Scanner {
	return &s3Scanner{
		conn:     conn,
		last_idx: "",
		hander:   hander,
		scan_len: scan_len,
	}
}

func (r *s3Scanner) Start() error {
	turn := 0
	for {
		logs.Info("Start %d", turn)
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

type RestoreS3 struct {
	master RestoreMaster
	land   *storehelper.StoreS3
}

func (r *RestoreS3) SetStore(s storehelper.IStore) {
	r.land = s.(*storehelper.StoreS3)
}

/// Interface
func (r *RestoreS3) cloneNewSafeReader() (storehelper.IStore, error) {
	return r.land.Clone()
}

func (r *RestoreS3) close() {
	r.land.Close()
}

func (r *RestoreS3) restore(job RestoreJob,
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

func (r *RestoreS3) RestoreOne(key string, value []byte) error {
	r.master.AddJob(RestoreJob{key, nil, nil})
	return nil
}

func (r *RestoreS3) RestoreOneWaitRes(key string, value []byte, res restoreOneResChan) error {
	r.master.AddJob(RestoreJob{key, nil, res})
	return nil
}

func (r *RestoreS3) RestoreAll() error {
	// 创建Store
	l, err := r.land.Clone()
	// err := l.Open()
	if err != nil {
		logs.Error("Open store err by %s", err.Error())
		return err
	}

	s3Store, ok := l.(*storehelper.StoreS3)
	if !ok {
		logs.Error("store is not s3Store")
		return errors.New("store is not s3Store")
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

	scanner := NewS3Scanner(s3Store, hander, int64(r.master.GetScanLen()))
	return scanner.Start()
}
