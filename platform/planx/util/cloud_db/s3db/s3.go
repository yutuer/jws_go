package s3db

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cenk/backoff"
	"vcs.taiyouxi.net/platform/planx/util/awshelper"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type BucketCannedACL string

const (
	Private           BucketCannedACL = "private"
	PublicRead        BucketCannedACL = "public-read"
	PublicReadWrite   BucketCannedACL = "public-read-write"
	AuthenticatedRead BucketCannedACL = "authenticated-read"
)

type S3 struct {
	s         *s3.S3
	bucket    string
	region    string
	accessKey string
	secretKey string

	count_all int
	count_err int

	time_log time.Time

	format string
	seq    string
}

func NewStoreS3(region, bucket, accessKey string, secretKey string, format, seq string) *S3 {
	if region == "" {
		region = os.Getenv("AWS_REGION")
	}

	myS3 := &S3{
		bucket:    bucket,
		region:    region,
		accessKey: accessKey,
		secretKey: secretKey,
		time_log:  time.Now(),
		format:    format,
		seq:       seq,
	}

	return myS3
}

func (s *S3) Open() error {
	mySession := awshelper.CreateAWSSession(s.region, s.accessKey, s.secretKey, 3)
	s.s = s3.New(mySession)
	return nil
}

func (s *S3) Close() error {
	return nil
}

func (s *S3) Clone() (*S3, error) {
	n := NewStoreS3(
		s.region,
		s.bucket,
		s.accessKey,
		s.secretKey,
		s.format, s.seq)
	if err := n.Open(); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *S3) LogInfo() {
	n := time.Now()
	in := n.Sub(s.time_log)

	logs.Info("put info %d / %d by %d",
		s.count_err, s.count_all, in.Seconds())
	s.count_all = 0
	s.count_err = 0
	s.time_log = n
}

func (s *S3) Put(key string, val []byte) error {
	if s.count_all >= 100 {
		s.LogInfo()
	}

	_, err := s.s.PutObject(&s3.PutObjectInput{
		ACL:         aws.String(string(Private)),
		Bucket:      aws.String(s.bucket),
		Body:        bytes.NewReader(val),
		ContentType: aws.String("application/octet-stream"),
		Key:         aws.String(key),
	})

	s.count_all++
	if err == nil {
		return nil
	}
	s.count_err++

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 100 * time.Millisecond
	b.MaxElapsedTime = 1 * time.Minute
	// 以下取默认值
	// DefaultMaxInterval         = 60 * time.Second
	// DefaultMaxElapsedTime      = 15 * time.Minute
	ticker := backoff.NewTicker(b)
	defer ticker.Stop()

	for _ = range ticker.C {
		_, err := s.s.PutObject(&s3.PutObjectInput{
			ACL:         aws.String(string(Private)),
			Bucket:      aws.String(s.bucket),
			Body:        bytes.NewReader(val),
			ContentType: aws.String("text/plain"),
			Key:         aws.String(key),
		})

		if err == nil {
			break
		}
		//logs.Warn("re put %s %v", key, k)
	}

	return err
}

func (s *S3) PutWithMeta(key string, val []byte, metaKey string, metaVal string) error {
	_imp := func() error {
		_, err := s.s.PutObject(&s3.PutObjectInput{
			ACL:         aws.String(string(Private)),
			Bucket:      aws.String(s.bucket),
			Body:        bytes.NewReader(val),
			Key:         aws.String(key),
			ContentType: aws.String("text/plain"),
			Metadata: map[string]*string{
				metaKey: aws.String(metaVal),
			},
		})
		return err
	}

	err := _imp()
	s.count_all++
	if err == nil {
		return nil
	}

	s.count_err++

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 100 * time.Millisecond
	b.MaxElapsedTime = 1 * time.Minute
	// 以下取默认值
	// DefaultMaxInterval         = 60 * time.Second
	// DefaultMaxElapsedTime      = 15 * time.Minute
	ticker := backoff.NewTicker(b)
	defer ticker.Stop()
	for _ = range ticker.C {
		err = _imp()

		if err == nil {
			break
		}
		//logs.Warn("re put %s %v", key, k)
	}

	return err
}

func (s *S3) Get(key string) ([]byte, error) {
	return s.GetWithBucket(s.bucket, key)
}

func (s *S3) GetWithBucket(bucket, key string) ([]byte, error) {
	out, err := s.s.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()

	b, _ := ioutil.ReadAll(out.Body)

	return b, err
}

func (s *S3) GetMeta(key, metaKey string) (*string, error) {
	out, err := s.s.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	res, ok := out.Metadata[metaKey]
	if !ok {
		return nil, fmt.Errorf("meta %v not exist", metaKey)
	}
	return res, nil
}

func (s *S3) Del(key string) error {
	_, err := s.s.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	return err
}

func (s *S3) ListObject(last_idx string, scan_len int64) ([]string, error) {
	return s.ListObjectWithBucket(s.bucket, last_idx, scan_len)
}

func (s *S3) ListObjectWithBucket(bucket, last_idx string, scan_len int64) ([]string, error) {
	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucket), // Required
		//Delimiter:    aws.String("Delimiter"),
		//EncodingType: aws.String("EncodingType"),
		Marker:  aws.String(last_idx),
		MaxKeys: aws.Int64(scan_len),
		//Prefix:  aws.String(last_idx),
	}

	resp, err := s.s.ListObjects(params)
	if err != nil {
		// A non-service error occurred.
		return []string{}, err
	}

	keys := make([]string, 0, scan_len)

	for _, v := range resp.Contents {
		keys = append(keys, *(v.Key))
	}

	return keys, err
}

func (s *S3) StoreKey(key string) string {
	now_time := time.Now()
	if s.format == "" {
		return key
	}
	return now_time.Format(s.format) + s.seq + key
}

func (s *S3) RedisKey(key_in_store string) (string, bool) {
	// Redis存储路径
	if s.seq == "" {
		return key_in_store, true
	}

	d := strings.Split(key_in_store, s.seq)
	if len(d) < 2 {
		logs.Error("key_in_store err : %s in %s", key_in_store, s.seq)
		return "", false
	}
	return d[len(d)-1], true
}
