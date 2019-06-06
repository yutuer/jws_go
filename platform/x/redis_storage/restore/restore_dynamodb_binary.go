package restore

// import (
//     "errors"
//     "sync"

//     "vcs.taiyouxi.net/platform/planx/util/logs"
//     "vcs.taiyouxi.net/platform/planx/util/redispool"
//     "vcs.taiyouxi.net/platform/planx/util/storehelper"
//     "vcs.taiyouxi.net/platform/x/redis_storage/util"
// )

// // 目前只是测试用

// type RestoreDynamoDBBinary struct {
//     redis_source_pool redispool.Pool
//     wg                sync.WaitGroup
//     wg_jobs           sync.WaitGroup

//     land        *storehelper.StoreWithGZip
//     workers_num int
//     scan_len    int
//     seq         string

//     need_store_heads map[string]bool

//     jobs chan dynamoDBRestoreJob
// }

// func (r *RestoreDynamoDBBinary) Init(
//     pool redispool.Pool,
//     workers_num,
//     scan_len int,
//     need_store_heads map[string]bool) {

//     r.redis_source_pool = pool
//     r.workers_num = workers_num
//     r.scan_len = scan_len
//     r.jobs = make(chan dynamoDBRestoreJob, scan_len*2)
//     r.need_store_heads = need_store_heads
//     r.wg.Add(1)
// }

// func (r *RestoreDynamoDBBinary) isNeedOnLand(key string) bool {
//     return util.IsNeedOnLand(r.need_store_heads, key)
// }

// func (r *RestoreDynamoDBBinary) SetStore(s *storehelper.StoreWithGZip) {
//     r.land = s
// }

// func (r *RestoreDynamoDBBinary) restore(job *dynamoDBRestoreJob, conn storehelper.IStore) error {
//     key_in_redis, ok := conn.RedisKey(job.key)
//     if !ok {
//         logs.Error("RedisKey Err %s", job.key)
//         return errors.New("RedisKey Err")
//     }

//     if !r.isNeedOnLand(key_in_redis) {
//         return nil
//     }

//     return r.restoreToRedis(key_in_redis, job.data)
// }

// func (r *RestoreDynamoDBBinary) restoreToRedis(key_in_redis string, data []byte) error {
//     conn := r.redis_source_pool.Get()
//     defer conn.Close()
//     return storehelper.FromJson(conn.RawConn(), key_in_redis, data)
// }

// func (r *RestoreDynamoDBBinary) Start() error {
//     defer r.wg.Done() // Add 在Init函数中
//     logs.Info("RestoreDynamoDBBinary Start")

//     if r.land == nil {
//         logs.Error("RestoreDynamoDBBinary:No Data Source")
//         return errors.New("RestoreDynamoDBBinary:No Data Source")
//     }

//     // 回档协程
//     for i := 0; i < r.workers_num; i++ {
//         go func() {
//             r.wg.Add(1)
//             defer r.wg.Done()

//             conn := r.land.Clone()

//             for {
//                 // 这里需要通过wg_jobs，等待所有任务完成
//                 is_stop := func() bool {
//                     job, ok := <-r.jobs
//                     if !ok {
//                         logs.Warn("RestoreDynamoDBBinary queue close")
//                         return true
//                     }

//                     defer r.wg_jobs.Done()

//                     //logs.Info("On Restore %v", job.key)
//                     err := r.restore(&job, conn)
//                     if err != nil {
//                         logs.Error("restore %s err %s", job.key, err.Error())
//                     } else {
//                         //logs.Info("restore success %s", job.key)
//                     }

//                     return false
//                 }()

//                 if is_stop {
//                     return
//                 }
//             }
//         }()
//     }

//     return nil
// }

// func (r *RestoreDynamoDBBinary) Stop() {
//     logs.Info("RestoreDynamoDBBinary Stop")
//     // 这里需要通过wg_jobs，等待所有任务完成
//     r.wg_jobs.Wait()

//     close(r.jobs)
//     // 等待各个协程退出
//     r.wg.Wait()
// }

// func (r *RestoreDynamoDBBinary) RestoreOne(key string) error {
//     l := r.land.Clone()
//     err := l.Open()
//     if err != nil {
//         logs.Error("Open store err by %s", err.Error())
//         return err
//     }

//     dy, ok := l.(*storehelper.StoreWithGZip)
//     if !ok {
//         logs.Error("store is not dy")
//         return errors.New("store is not dy")
//     }

//     data, err := dy.Get(key)
//     if err != nil {
//         logs.Error("restore data %s", err)
//         return err
//     }

//     logs.Trace("key %s", key)
//     logs.Trace("data %s", string(data))
//     r.jobs <- dynamoDBRestoreJob{
//         key,
//         data,
//     }

//     return nil

// }

// func (r *RestoreDynamoDBBinary) RestoreAll() error {
//     /*
//         // 创建Store
//         l := r.land.Clone()
//         err := l.Open()
//         if err != nil {
//             logs.Error("Open store err by %s", err.Error())
//             return err
//         }

//         dy, ok := l.(*storehelper.StoreDynamoDB)
//         if !ok {
//             logs.Error("store is not dy")
//             return errors.New("store is not dy")
//         }

//         num := 0

//         hander := func(idx int, key, data string) error {
//             //logs.Trace("keys hander")

//             if num%1000 == 0 {
//                 logs.Trace("key has dump %d %d",
//                     idx, num)
//             }
//             num++

//             r.wg_jobs.Add(1)
//             r.jobs <- dynamoDBRestoreJob{
//                 key,
//                 []byte(data),
//             }

//             return nil
//         }

//         logs.Info("Start Scan All")
//         err = dy.Scan(hander, int64(r.scan_len), int64(r.workers_num))
//         if err != nil {
//             logs.Error("Scan err by %s", err.Error())
//             return err
//         }
//         return nil
//     */
//     return nil
// }
