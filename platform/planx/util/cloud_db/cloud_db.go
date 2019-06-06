package cloud_db

// 云存储 S3， OSS
type CloudDb interface {
	Open() error
	Close() error
	Put(key string, val []byte) error
	PutWithMeta(key string, val []byte, metaKey string, metaVal string) error
	Get(key string) ([]byte, error)
	GetWithBucket(bucket, key string) ([]byte, error)
	GetMeta(key, metaKey string) (*string, error)
	Del(key string) error
	ListObject(last_idx string, scan_len int64) ([]string, error)
	ListObjectWithBucket(bucket, last_idx string, scan_len int64) ([]string, error)
	StoreKey(key string) string
	RedisKey(key_in_store string) (string, bool)
}
