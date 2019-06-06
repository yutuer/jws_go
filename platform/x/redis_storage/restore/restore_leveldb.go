package restore

import (
	"errors"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

//TODO YZH LEVELDB

type RestoreLevelDB struct {
	master RestoreMaster
	land   *storehelper.StoreLevelDB
}

func (r *RestoreLevelDB) SetStore(s storehelper.IStore) {
	r.land = s.(*storehelper.StoreLevelDB)
}

func (r *RestoreLevelDB) cloneNewSafeReader() (storehelper.IStore, error) {
	return r.land.Clone()
}

func (r *RestoreLevelDB) close() {
	r.land.Close()
}

func (r *RestoreLevelDB) restore(job RestoreJob,
	conn storehelper.IStore,
	isNeedOnLand func(string) bool) error {
	key := job.key
	if job.data != nil {
		return r.master.RestoreToDB(key, job.data)
	} else {
		data, err := conn.Get(key)
		if err != nil {
			return err
		}
		return r.master.RestoreToDB(key, data)
	}
}

func (r *RestoreLevelDB) RestoreOne(key string, value []byte) error {
	//TODO YZH  RestoreLevelDB 这里没有限制key prefix
	r.master.AddJob(RestoreJob{key, value, nil})
	return nil
}

func (r *RestoreLevelDB) RestoreOneWaitRes(key string, value []byte, res restoreOneResChan) error {
	r.master.AddJob(RestoreJob{key, value, res})
	return nil
}

func (r *RestoreLevelDB) RestoreAll() error {
	// 创建Store
	l, err := r.land.Clone()
	if err != nil {
		logs.Error("Open store err by %s", err.Error())
		return err
	}

	levelDB_store, ok := l.(*storehelper.StoreLevelDB)
	if !ok {
		logs.Error("store is not StoreLevelDB")
		return errors.New("store is not StoreLevelDB")
	}

	lerr := levelDB_store.IterateAllObjects(func(key, value []byte) {

		key_in_redis, ok := levelDB_store.RedisKey(string(key))
		if !ok {
			logs.Error("RedisKey Err %s", key)
			return
		}

		if !r.master.isNeedOnLand(key_in_redis) {
			return
		}

		//TODO YZH 甚至可以绕开job系统直接调用restore,
		// LevelDB是本地文件， Redis是单线程系统
		// 中间走了Jobs可能完全没有意义,也许可以缓冲两端读写速率？
		// 也许可以进行redis写合并后提升写效率

		// ！！注意本函数传递回来的value只在函数返回前有效！！
		// 当通过chan传递value后，value本身可能已经失效了
		// 所以这里第2个参数必须是空nil
		r.RestoreOne(key_in_redis, nil)
	})
	return lerr
}

// func (r *RestoreLevelDB) RestoreAll2() error {
//     // 创建Store
//     l, err := r.land.Clone()
//     // err := l.Open()
//     if err != nil {
//         logs.Error("Open store err by %s", err.Error())
//         return err
//     }

//     levelDB_store, ok := l.(*storehelper.StoreLevelDB)
//     if !ok {
//         logs.Error("store is not StoreLevelDB")
//         return errors.New("store is not StoreLevelDB")
//     }

//     num := 0

//     hander := func(keys []string) error {
//         //logs.Trace("keys hander")
//         for i, key := range keys {
//             if num%1000 == 0 {
//                 logs.Trace("key has dump %d %d %s",
//                     i, num, key)
//             }
//             num++
//             r.RestoreOne(key, nil)
//         }
//         return nil
//     }

//     turn := 0
//     last_idx := ""
//     scan_len := int64(r.master.GetScanLen())

//     for {
//         logs.Info("Start %d %d %s", turn, scan_len, last_idx)
//         keys, err := levelDB_store.ListObject(last_idx, scan_len)
//         if err != nil || len(keys) == 0 {
//             return err
//         }

//         if err := hander(keys); err != nil {
//             return err
//         }
//         last_idx = keys[len(keys)-1]
//         turn++
//     }
//     return nil
// }
