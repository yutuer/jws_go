package store

import (
	"encoding/json"
	"errors"
	"github.com/boltdb/bolt"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var db *bolt.DB

var (
	ErrNoKey = errors.New("ErrNoKey")
)

var (
	NormalBucket = "Normal"
)

func InitStore(file string) error {
	var err error
	db, err = bolt.Open(file, 0777, nil)
	return err
}

func StopStore() error {
	logs.Warn("Store Stop")
	if db != nil {
		return db.Close()
	}
	return nil
}

func Set(bucket, key string, value []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))

		if err != nil {
			return err
		}

		return b.Put([]byte(key), value)
	})
}

func Get(bucket, key string) ([]byte, error) {
	res := []byte{}
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))

		if err != nil {
			return err
		}

		re := b.Get([]byte(key))
		if re == nil {
			return ErrNoKey
		}

		res = re
		return nil
	})
	return res, err
}

func SetIntoJson(bucket, key string, value interface{}) error {
	j, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return Set(bucket, key, j)
}

func GetFromJson(bucket, key string, value interface{}) error {
	res, err := Get(bucket, key)
	if err != nil {
		return err
	}
	return json.Unmarshal(res, value)
}
