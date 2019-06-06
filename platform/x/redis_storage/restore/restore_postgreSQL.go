package restore

import (
	"errors"

	"vcs.taiyouxi.net/platform/planx/util/account_json"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

type RestorePostgreSQL struct {
	master RestoreMaster
	land   *storehelper.StorePostgreSQL
}

func (r *RestorePostgreSQL) SetStore(s storehelper.IStore) {
	r.land = s.(*storehelper.StorePostgreSQL)
}

func (r *RestorePostgreSQL) cloneNewSafeReader() (storehelper.IStore, error) {
	return r.land.Clone()
}

func (r *RestorePostgreSQL) close() {
	r.land.Close()
}

func (r *RestorePostgreSQL) restore(job RestoreJob,
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

	o, err := accountJson.FromPureJsonToOld(string(res))
	if err != nil {
		return err
	}

	d, err := o.Encode()
	if err != nil {
		return err
	}

	return r.master.RestoreToDB(key_in_redis, d)
}

func (r *RestorePostgreSQL) RestoreOne(key string, value []byte) error {
	//TODO YZH filter isNeeded before putting into jobs
	r.master.AddJob(RestoreJob{key, nil, nil})
	return nil

}

func (r *RestorePostgreSQL) RestoreOneWaitRes(key string, value []byte, res restoreOneResChan) error {
	r.master.AddJob(RestoreJob{key, nil, res})
	return nil
}

func (r *RestorePostgreSQL) RestoreAll() error {
	storehelper.GetAllKeys(r.land.GetPool(), func(key string) {
		r.RestoreOne(key, nil)
	})
	return nil
}
