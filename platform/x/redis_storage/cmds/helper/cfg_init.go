package helper

import (
	"strings"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
	"vcs.taiyouxi.net/platform/x/redis_storage/config"
	"vcs.taiyouxi.net/platform/x/redis_storage/onland"
	"vcs.taiyouxi.net/platform/x/redis_storage/redis_monitor"
	"vcs.taiyouxi.net/platform/x/redis_storage/restore"
)

//Add(s storehelper.IStore)
func InitBackends(addFn func(storehelper.IStore)) {

	backends := make(map[string]bool)
	for _, b := range config.OnlandCfg.Backends {
		backends[b] = true
	}

	// 初始化Onland目标， 如果目标有GZip后缀则，进行gzip压缩有再保存
	for b, _ := range backends {
		addFnWithGZip := func(o storehelper.IStore) {
			addFn(storehelper.NewStoreWithGZip(o))
		}
		if strings.HasSuffix(b, "GZip") {
			gzipB := strings.TrimSuffix(b, "GZip")
			if target := config.GetTarget(gzipB); target != nil {
				if target.HasConfigured() {
					target.Setup(addFnWithGZip)
					logs.Info("Onland Target %s is all set. ", b)
				}
			}
		} else {
			if target := config.GetTarget(b); target != nil {
				if target.HasConfigured() {
					target.Setup(addFn)
					logs.Info("Onland Target %s is all set. ", b)
				}
			}
		}
	}
}

func mkNeedStoreMap(config string) map[string]bool {
	keys := strings.Split(config, ",")
	re := make(map[string]bool, len(keys))
	for i := 0; i < len(keys); i++ {
		re[keys[i]] = true
	}
	return re
}

func InitOnLandStores() *onland.OnlandStores {
	// 用于监听的连接池
	redis_source_pool := storehelper.NewRedisPool(
		config.OnlandCfg.Redis,
		config.OnlandCfg.RedisAuth,
		config.OnlandCfg.RedisDB, 10)
	// 初始化落地类
	onlandStores := onland.OnlandStores{}
	onlandStores.Init(
		redis_source_pool,
		mkNeedStoreMap(config.CommonCfg.NeedStoreHead),
		config.CommonCfg.StoreMode,
		config.OnlandCfg.Workers_Max,
		true) // 监控时需要记录存档大小日志

	return &onlandStores

}

func InitMonitor(onlandinst *onland.OnlandStores) *redis_monitor.RedisPubMonitor {
	// 用于获取数据的连接池
	redis_monitor_pool := storehelper.NewRedisPool(
		config.OnlandCfg.Redis,
		config.OnlandCfg.RedisAuth,
		config.OnlandCfg.RedisDB,
		1000,
	)

	// 初始化监听
	redisMonitor := redis_monitor.RedisPubMonitor{}
	redisMonitor.Init(
		redis_monitor_pool,
		onlandinst,
		config.OnlandCfg.RedisDB,
		mkNeedStoreMap(config.CommonCfg.NeedStoreHead),
	)
	return &redisMonitor

}

func InitRestore() *restore.Restore {
	// 用于回档目标的连接池
	redis_restore_pool := storehelper.NewRedisPool(
		config.RestoreCfg.Redis_Restore,
		config.RestoreCfg.RedisAuth_Restore,
		config.RestoreCfg.RedisDB_Restore,
		1000,
	)

	restoreFrom := config.RestoreCfg.Restore_Data_from

	return restore.CreateRestore(
		restoreFrom,
		redis_restore_pool,
		config.RestoreCfg.Workers_Restore_Max,
		config.RestoreCfg.Restore_Scan_Len,
		mkNeedStoreMap(config.CommonCfg.NeedStoreHead),
	)

	// 初始化回档类
	// restorer := restore.Restore{}
	// restorer.Init(
	//     redis_restore_pool,
	//     config.RestoreCfg.Workers_Restore_Max,
	//     config.RestoreCfg.Restore_Scan_Len,
	//     mkNeedStoreMap(config.CommonCfg.NeedStoreHead),
	// )
	// 回档数据源配置
	// if config.RestoreCfg.Restore_Data_from == "S3" {
	//     restorer.UseS3(
	//         config.RestoreCfg.S3_Restore_Bucket,
	//         config.CommonCfg.AWS_Region,
	//         config.CommonCfg.AWS_AccessKey,
	//         config.CommonCfg.AWS_SecretKey,
	//         config.S3Cfg.Format,
	//         config.S3Cfg.FormatSep)
	// } else if config.RestoreCfg.Restore_Data_from == "DynamoDB" {
	//     restorer.UseDynamoDB(
	//         config.RestoreCfg.DynamoDB_Restore_Name,
	//         config.CommonCfg.AWS_Region,
	//         config.CommonCfg.AWS_AccessKey,
	//         config.CommonCfg.AWS_SecretKey,
	//         config.DynamoDBCfg.Format,
	//         config.DynamoDBCfg.FormatSep)
	// } else if config.RestoreCfg.Restore_Data_from == "SSDB" {
	//     restorer.UseSSDB(
	//         config.RestoreCfg.SSDB_Ip,
	//         config.SSDBCfg.Format,
	//         config.SSDBCfg.FormatSep,
	//         config.RestoreCfg.SSDB_Port)
	// } else if config.RestoreCfg.Restore_Data_from == "PostgreSQL" {
	//     restorer.UsePostgreSQL(
	//         config.PostgreSQLCfg.Host,
	//         config.PostgreSQLCfg.Port,
	//         config.PostgreSQLCfg.DB,
	//         config.PostgreSQLCfg.User,
	//         config.PostgreSQLCfg.Pass,
	//     )
	// } else {
	//     logs.Error("No Data Score Set")
	//     return nil
	// }
	// return &restorer
}
