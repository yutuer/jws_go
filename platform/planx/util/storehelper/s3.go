package storehelper

import "vcs.taiyouxi.net/platform/planx/util/cloud_db/s3db"

type BucketCannedACL string

const (
	Private           BucketCannedACL = "private"
	PublicRead        BucketCannedACL = "public-read"
	PublicReadWrite   BucketCannedACL = "public-read-write"
	AuthenticatedRead BucketCannedACL = "authenticated-read"
)

type StoreS3 struct {
	s *s3db.S3
}

func NewStoreS3(region, bucket, accessKey string, secretKey string, format, seq string) IStore {
	s := &StoreS3{}
	s.s = s3db.NewStoreS3(region, bucket, accessKey, secretKey, format, seq)
	return s
}

func (s *StoreS3) Open() error {
	return s.s.Open()
}

func (s *StoreS3) Clone() (IStore, error) {
	cs := &StoreS3{}
	_cs, err := s.s.Clone()
	if err != nil {
		return nil, err
	}
	cs.s = _cs
	return cs, nil
}

func (s *StoreS3) Close() error {
	s.s.Close()
	return nil
}

func (s *StoreS3) Put(key string, val []byte, rh ReadHandler) error {
	return s.s.Put(key, val)
}
func (s *StoreS3) Get(key string) ([]byte, error) {
	return s.s.Get(key)
}
func (s *StoreS3) Del(key string) error {
	return s.s.Del(key)
}
func (s *StoreS3) StoreKey(key string) string {
	return s.s.StoreKey(key)
}
func (s *StoreS3) RedisKey(key_in_store string) (string, bool) {
	return s.s.RedisKey(key_in_store)
}

func (s *StoreS3) ListObject(last_idx string, scan_len int64) ([]string, error) {
	return s.s.ListObject(last_idx, scan_len)
}
