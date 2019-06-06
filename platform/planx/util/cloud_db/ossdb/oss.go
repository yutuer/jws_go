package ossdb

import (
	"bytes"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
	"strings"
	"time"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type OSS struct {
	client *oss.Client
	cfg    OSSCfg
	bucket *oss.Bucket
}

type OSSCfg struct {
	endPoint   string
	bucketName string
	format     string
	seq        string
	accessKey  string
	secretKey  string
}

func NewStoreOSS(endPoint, accessKey string, secretKey string, bucket string, format, seq string) *OSS {
	s := &OSS{
		cfg: OSSCfg{
			bucketName: bucket,
			format:     format,
			seq:        seq,
		},
	}
	var err error
	s.client, err = oss.New(endPoint, accessKey, secretKey)
	if err != nil {
		logs.Error("<StoreOss> new OOS err %v", err)
		return nil
	}
	return s
}

func (s *OSS) Open() error {
	s.client.CreateBucket(s.cfg.bucketName)
	var err error
	s.bucket, err = s.client.Bucket(s.cfg.bucketName)
	return err
}

func (s *OSS) Close() error {
	// TODO
	return nil
}

func (s *OSS) Put(key string, val []byte) error {
	if s.bucket == nil {
		return fmt.Errorf("oss bucket is nil")
	}
	err := s.bucket.PutObject(key, bytes.NewReader(val))
	if err != nil {
		return err
	}
	// TODO 失败重试
	return nil
}

func (s *OSS) Get(key string) ([]byte, error) {
	if s.bucket == nil {
		return nil, fmt.Errorf("oss bucket is nil")
	}
	buf := new(bytes.Buffer)
	body, err := s.bucket.GetObject(key)
	if err != nil {
		return nil, err
	}
	io.Copy(buf, body)
	body.Close()
	return buf.Bytes(), nil
}

func (s *OSS) Del(key string) error {
	if s.bucket == nil {
		return nil
	}
	err := s.bucket.DeleteObject(key)
	if err != nil {
		return err
	}
	return nil
}

func (s *OSS) StoreKey(key string) string {
	now_time := time.Now()
	if s.cfg.format == "" {
		return key
	}
	return now_time.Format(s.cfg.format) + s.cfg.seq + key
}

func (s *OSS) RedisKey(key_in_store string) (string, bool) {
	// Redis存储路径
	if s.cfg.seq == "" {
		return key_in_store, true
	}

	d := strings.Split(key_in_store, s.cfg.seq)
	if len(d) < 2 {
		logs.Error("key_in_store err : %s in %s", key_in_store, s.cfg.seq)
		return "", false
	}
	return d[len(d)-1], true
}

func (s *OSS) Clone() (*OSS, error) {
	clone := NewStoreOSS(s.cfg.endPoint, s.cfg.accessKey, s.cfg.secretKey,
		s.cfg.bucketName, s.cfg.format, s.cfg.seq)
	if err := clone.Open(); err != nil {
		return nil, err
	}
	return clone, nil
}

func (s *OSS) GetWithBucket(bucket, key string) ([]byte, error) {
	var err error
	s.bucket, err = s.client.Bucket(bucket)
	if err != nil {
		logs.Error("<OSS> GetWithBucket bukcet err", err)
		return nil, err
	}
	buf := new(bytes.Buffer)
	body, err := s.bucket.GetObject(key)
	if err != nil {
		logs.Error("<OSS> GetWithBucket object err", err)
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	io.Copy(buf, body)
	body.Close()
	return buf.Bytes(), nil
}

func (s *OSS) ListObject(last_idx string, scan_len int64) ([]string, error) {
	return s.ListObjectWithBucket(s.cfg.bucketName, last_idx, scan_len)
}

func (s *OSS) ListObjectWithBucket(bucketName, last_idx string, scan_len int64) ([]string, error) {
	bucket, err := s.client.Bucket(bucketName)

	// 此处参数的使用是参照S3使用的
	resp, err := bucket.ListObjects(oss.MaxKeys(int(scan_len)), oss.Prefix(last_idx),
		oss.Marker(last_idx))
	if err != nil {
		// A non-service error occurred.
		return []string{}, err
	}

	keys := make([]string, 0, scan_len)

	for _, v := range resp.Objects {
		keys = append(keys, v.Key)
	}

	return keys, err
}

func (s *OSS) GetMeta(key, metaKey string) (*string, error) {
	props, err := s.bucket.GetObjectDetailedMeta(key)

	if err != nil {
		return nil, err
	}

	res := props.Get(metaKey)

	return &res, nil
}

func (s *OSS) PutWithMeta(key string, val []byte, metaKey string, metaVal string) error {
	err := s.bucket.PutObject(key, bytes.NewReader(val))
	if err != nil {
		return err
	}

	// 场景：设置Bucket Meta，可以设置一个或多个属性。
	// 注意：Meta不区分大小写
	options := []oss.Option{
		oss.Meta(metaKey, metaVal)}
	return s.bucket.SetObjectMeta(key, options...)
}
