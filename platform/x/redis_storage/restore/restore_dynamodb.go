package restore

import (
	"errors"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

type RestoreDynamoDB struct {
	master RestoreMaster
	land   *storehelper.StoreDynamoDB
}

func (r *RestoreDynamoDB) SetStore(s storehelper.IStore) {
	r.land = s.(*storehelper.StoreDynamoDB)
}

func (r *RestoreDynamoDB) cloneNewSafeReader() (storehelper.IStore, error) {
	return r.land.Clone()
}

func (r *RestoreDynamoDB) close() {
	r.land.Close()
}

func (r *RestoreDynamoDB) restore(job RestoreJob,
	conn storehelper.IStore,
	isNeedOnLand func(string) bool) error {

	key_in_redis, ok := conn.RedisKey(job.key)
	if !ok {
		logs.Error("RedisKey Err %s", job.key)
		return errors.New("RedisKey Err")
	}

	if !isNeedOnLand(key_in_redis) {
		return nil
	}

	return r.master.RestoreToDB(key_in_redis, job.data)
}

func (r *RestoreDynamoDB) RestoreOne(key string, value []byte) error {
	l, err := r.land.Clone()
	// err := l.Open()
	if err != nil {
		logs.Error("Open store err by %s", err.Error())
		return err
	}

	dy, ok := l.(*storehelper.StoreDynamoDB)
	if !ok {
		logs.Error("store is not dy")
		return errors.New("store is not dy")
	}

	data, err := dy.Get(key)
	if err != nil {
		logs.Error("restore data %s", err)
		return err
	}

	logs.Trace("key %s", key)
	logs.Trace("data %s", string(data))
	r.master.AddJob(RestoreJob{key, data, nil})

	return nil

}

func (r *RestoreDynamoDB) RestoreOneWaitRes(key string, value []byte, res restoreOneResChan) error {
	l, err := r.land.Clone()
	// err := l.Open()
	if err != nil {
		logs.Error("Open store err by %s", err.Error())
		return err
	}

	dy, ok := l.(*storehelper.StoreDynamoDB)
	if !ok {
		logs.Error("store is not dy")
		return errors.New("store is not dy")
	}

	data, err := dy.Get(key)
	if err != nil {
		logs.Error("restore data %s", err)
		return err
	}

	logs.Trace("key %s", key)
	logs.Trace("data %s", string(data))
	r.master.AddJob(RestoreJob{key, data, res})

	return nil
}

func (r *RestoreDynamoDB) RestoreAll() error {
	// 创建Store
	l, err := r.land.Clone()
	// err := l.Open()
	if err != nil {
		logs.Error("Open store err by %s", err.Error())
		return err
	}

	dy, ok := l.(*storehelper.StoreDynamoDB)
	if !ok {
		logs.Error("store is not dy")
		return errors.New("store is not dy")
	}

	num := 0

	hander := func(idx int, key, data string) error {
		//logs.Trace("keys hander")

		if num%1000 == 0 {
			logs.Trace("key has dump %d %d",
				idx, num)
		}
		num++

		r.master.AddJob(RestoreJob{key, []byte(data), nil})

		return nil
	}

	logs.Info("Start Scan All")
	err = dy.Scan(hander,
		int64(r.master.GetScanLen()),
		int64(r.master.GetWorkerNum()))
	if err != nil {
		logs.Error("Scan err by %s", err.Error())
		return err
	}
	return nil
}
