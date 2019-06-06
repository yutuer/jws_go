package storehelper

// rh ReadHandler 用于类似Redis_bilog这种，通过扫描一种主键，
//然后同时需要读取其他不同主键的情况。
type ReadHandler func(string) ([]byte, bool)

type IStore interface {
	Open() error
	Close() error

	//TODO YZH: batch write optmization of All IStore
	Put(key string, val []byte, rh ReadHandler) error
	Get(key string) ([]byte, error)
	Del(key string) error

	StoreKey(key string) string
	RedisKey(key_in_store string) (string, bool)

	// Clone 应该返回一个Open过的可用的接口
	//
	// 如果IStore是goroutine safe则直接返回自身指针,
	// 但同时要求Close结构保证最后一个Clone关闭后，才是实例真正Close。
	// 具体实现请参考levelDB, redis_bilog
	// 否则Clone()返回新创建的DB接口。
	Clone() (IStore, error)
}

//PostgreSQL, DynamoDB, Stdout, RedisBiLog, ElasticS 需要Put进来的内容是Json
//LevelDB, SSDB, Redis, S3等使用, 这4个驱动是对内容没有要求的

//TODO YZH ListObject 因为目前使用只在，还原过程中使用。
// Dynamodb本身支持并发扫描读
// [] s3的扫描可能也有并发扫描读的优化空间
// redis数据库并发扫描读/写的意义不大，因为是单线程的。
// []但是redis如果能处理好前缀，能够更好的适应单个库上多个shard时的单shard导出优化
const (
	S3  = "S3"
	OSS = "OSS"
)

type S3Cfg struct {
	Region           string
	BucketsVerUpdate string
	AccessKeyId      string
	SecretAccessKey  string
	Format           string
	Seq              string
}

type OSSInitCfg struct {
	EndPoint  string
	AccessId  string
	AccessKey string
	Bucket    string
}

// TODO
func NewStore(key string, options interface{}) IStore {
	switch key {
	case S3:
		cfg, ok := options.(S3Cfg)
		if !ok {
			return nil
		}
		return NewStoreS3(cfg.Region,
			cfg.BucketsVerUpdate,
			cfg.AccessKeyId,
			cfg.SecretAccessKey,
			cfg.Format,
			cfg.Seq)
	case OSS:
		cfg, ok := options.(OSSInitCfg)
		if !ok {
			return nil
		}
		return NewStoreOSS(cfg.EndPoint, cfg.AccessId, cfg.AccessKey, cfg.Bucket, "", "")
	}
	return nil
}
