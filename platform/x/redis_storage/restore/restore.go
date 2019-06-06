package restore

import (
	// "encoding/base64"
	"errors"
	"sync"
	"time"

	"golang.org/x/net/context"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"

	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
	"vcs.taiyouxi.net/platform/x/redis_storage/config"
	util2 "vcs.taiyouxi.net/platform/x/redis_storage/util"
)

type restoreOneResChan chan<- bool

type RedisRestoreHandler func(con redis.Conn, key string, value []byte) error

type RestoreMaster interface {
	AddJob(job RestoreJob)
	GetWorkerNum() int
	GetScanLen() int
	RestoreToDB(key_in_redis string, data []byte) error
	isNeedOnLand(key string) bool
}

type RestoreWorker interface {
	cloneNewSafeReader() (storehelper.IStore, error)
	restore(job RestoreJob, conn storehelper.IStore, isNeedOnLand func(string) bool) error
	RestoreOne(key string, value []byte) error
	RestoreOneWaitRes(key string, value []byte, res restoreOneResChan) error
	RestoreAll() error
	close()
}

type Restore struct {
	wg      util.WaitGroupWrapper
	wg_jobs sync.WaitGroup
	jobs    chan RestoreJob

	redisRestoreFun RedisRestoreHandler

	pool             redispool.IPool
	workers_num      int
	scan_len         int
	need_store_heads map[string]bool

	worker RestoreWorker
}

type RestoreJob struct {
	key     string
	data    []byte
	resChan restoreOneResChan
}

func CreateRestore(restoreFrom string,
	pool redispool.IPool,
	workers_num,
	scan_len int,
	need_store_heads map[string]bool) *Restore {

	var storeh storehelper.IStore
	addFn := func(s storehelper.IStore) {
		s, err := s.Clone()
		if err != nil {
			logs.Error("Open store err in Restore, %v", err)
		} else {
			storeh = s
		}
	}

	if target := config.GetTarget(restoreFrom); target != nil {
		if target.HasConfigured() {
			target.Setup(addFn)
			logs.Info("Onland Target %s is all set. ", restoreFrom)
		} else {
			logs.Error("Onland Target %s failed. ", restoreFrom)
		}
	}
	if storeh == nil {
		return nil
	}
	restore_master := &Restore{}
	restore_master.Init(pool, workers_num, scan_len,
		need_store_heads, config.CommonCfg.StoreMode)

	//TODO YZH kill "switch"
	restore_master.worker = func(From string) RestoreWorker {
		switch From {
		case "S3":
			s3 := &RestoreS3{master: restore_master}
			s3.SetStore(storeh)
			return s3
		case "DynamoDB":
			dynamodb := &RestoreDynamoDB{master: restore_master}
			dynamodb.SetStore(storeh)
			return dynamodb
		case "SSDB":
			ssdb := &RestoreSSDB{master: restore_master}
			ssdb.SetStore(storeh)
			return ssdb
		case "PostgreSQL":
			progres := &RestorePostgreSQL{master: restore_master}
			progres.SetStore(storeh)
			return progres
		case "LevelDB":
			levelDb := &RestoreLevelDB{master: restore_master}
			levelDb.SetStore(storeh)
			return levelDb
		}
		return nil
	}(restoreFrom)

	return restore_master
}

func StartRestoreWorker(worker RestoreWorker,
	jobs chan RestoreJob,
	jobDone func(),
	isNeedOnLand func(string) bool) {
	defer worker.close()
	//给每个携程创建一个goroutine安全的reader
	conn, err := worker.cloneNewSafeReader()
	if err != nil {
		logs.Error("worker Clone got an error :%s", err.Error())
	}
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				logs.Warn("RestoreDynamoDB queue close")
				return
			}
			//defer r.wg_jobs.Done()
			jobDone()
			err := worker.restore(job, conn, isNeedOnLand)
			if err != nil {
				logs.Error("restore %s err %s", job.key, err.Error())
			} else {
				//logs.Info("restore success %s", job.key)
			}
			if job.resChan != nil {
				job.resChan <- true
			}

		}
	}
}

func (r *Restore) Init(
	pool redispool.IPool,
	workers_num,
	scan_len int,
	need_store_heads map[string]bool,
	store_mode string) {

	r.pool = pool
	r.workers_num = workers_num
	r.scan_len = scan_len
	r.need_store_heads = need_store_heads

	r.jobs = make(chan RestoreJob, scan_len*2)

	switch store_mode {
	case "HGETALL":
		r.redisRestoreFun = FromJson
	case "DUMP":
		r.redisRestoreFun = restoreRedis
	}

}

func (r *Restore) isNeedOnLand(key string) bool {
	return util2.IsNeedOnLand(r.need_store_heads, key)
}

func (r *Restore) RestoreToDB(key_in_redis string, data []byte) error {
	conn := r.pool.Get()
	defer conn.Close()
	return r.redisRestoreFun(conn.RawConn(), key_in_redis, data)
}

func (r *Restore) AddJob(job RestoreJob) {
	r.wg_jobs.Add(1)
	r.jobs <- job
}

func (r *Restore) GetWorkerNum() int {
	return r.workers_num
}

func (r *Restore) GetScanLen() int {
	return r.scan_len
}

func (r *Restore) Start() error {
	r.wg.Add(1)
	defer r.wg.Done()

	isNeedOnLand := func(key string) bool {
		return r.isNeedOnLand(key)
	}

	// 回档协程
	for i := 0; i < r.workers_num; i++ {
		r.wg.Wrap(func() {
			//给每个携程创建一个goroutine安全的reader
			StartRestoreWorker(
				r.worker,
				r.jobs,
				func() { r.wg_jobs.Done() },
				isNeedOnLand,
			)
		})
	}

	return nil
}

func (r *Restore) Stop() {

	// 这里需要通过wg_jobs，等待所有任务完成
	r.wg_jobs.Wait()

	close(r.jobs)
	// 等待各个协程退出
	r.wg.Wait()

}

func (r *Restore) RestoreAll() error {
	if r.worker == nil {
		return errors.New("No Data Source Select")
	}
	return r.worker.RestoreAll()
}

func (r *Restore) RestoreOne(key string, value []byte) error {
	if r.worker == nil {
		return errors.New("No Data Source Select")
	}
	return r.worker.RestoreOne(key, value)
}

func (r *Restore) RestoreOneWaitRes(key string, value []byte) error {
	if r.worker == nil {
		return errors.New("No Data Source Select")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	resChan := make(chan bool, 1)
	err := r.worker.RestoreOneWaitRes(key, value, resChan)
	if err != nil {
		return err
	}

	select {
	case <-resChan:
		return nil
	case <-ctx.Done():
		return errors.New("TimeOut")
	}
	return nil
}

func FromJson(conn redis.Conn, key string, data []byte) error {
	value_map, err := storehelper.Json2RedisHM(data)

	if err != nil {
		return err
	}

	args := make([]interface{}, 0, len(value_map)*2+1)
	args = append(args, key)
	for k, v := range value_map {
		args = append(args, k)
		args = append(args, v)
	}

	_, err = conn.Do("HMSET", args...)
	return err
}

func restoreRedis(conn redis.Conn, key string, data []byte) error {
	//     If ttl is 0 the key is created without any expire, otherwise the specified expire time (in milliseconds) is set.
	// RESTORE will return a "Target key name is busy" error when key already exists unless you use the REPLACE modifier (Redis 3.0 or greater).
	// RESTORE checks the RDB version and data checksum. If they don't match an error is returned.

	// v, _ := base64.StdEncoding.DecodeString(string(data))
	// _, err := conn.Do("RESTORE", key, 0, v)

	_, err := conn.Do("RESTORE", key, 0, data)
	return err
}

/*


// func (r *Restore) UseDynamoDBBinaryWithGZip(db_name, region, access, secret, format, seq string) {
//     // 只允许有一个数据源
//     if r.restorer_imp != nil {
//         logs.Warn("UseDynamoDB restorer_imp has set")
//     }

//     dynamodb := &RestoreDynamoDBBinary{}
//     dynamodb.Init(r.pool, r.workers_num, r.scan_len, r.need_store_heads)

//     dynamodb_store := storehelper.NewStoreBinaryDynamoDB(
//         region,
//         db_name,
//         access,
//         secret, format, seq)
//     dynamodb.SetStore(storehelper.NewStoreWithGZip(dynamodb_store))
//     r.restorer_imp = dynamodb
// }

*/
