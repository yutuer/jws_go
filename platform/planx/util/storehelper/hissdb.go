package storehelper

import (
	"errors"
	"github.com/lessos/lessgo/data/hissdb"
	"strings"
	"time"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type StoreHISSDB struct {
	pool   *hissdb.Connector
	server string
	port   int

	format string
	seq    string
}

func NewStoreHISSDB(server string, port int, format, seq string) *StoreHISSDB {
	n := &StoreHISSDB{
		server: server,
		port:   port,
		format: format,
		seq:    seq,
	}
	return n
}

func (s *StoreHISSDB) Open() error {
	pool, err := hissdb.NewConnector(hissdb.Config{
		Host:    s.server,
		Port:    uint16(s.port),
		Timeout: 3,  // timeout in second, default to 10
		MaxConn: 32, // max connection number, default to 1
	})

	if err != nil {
		return err
	}
	s.pool = pool

	return nil
}

func (s *StoreHISSDB) Clone() (IStore, error) {
	ns := NewStoreHISSDB(
		s.server, s.port,
		s.format, s.seq)
	if err := ns.Open(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *StoreHISSDB) Close() error {
	s.pool.Close()
	return nil
}

//TODO YZH: batch write optmization of SSDB
func (s *StoreHISSDB) Put(key string, val []byte, rh ReadHandler) error {
	res := s.pool.Cmd("set", key, val)
	if res.State == "ok" {
		return nil
	} else {
		return errors.New("hissdb put err " + res.State)
	}
	return nil
}

func (s *StoreHISSDB) Get(key string) ([]byte, error) {
	res := s.pool.Cmd("get", key)
	if res.State == "ok" {
		return []byte(res.String()), nil
	} else {
		return []byte{}, errors.New("get err " + res.State)
	}
}

func (s *StoreHISSDB) Del(key string) error {
	res := s.pool.Cmd("del", key)
	if res.State == "ok" {
		return nil
	} else {
		return errors.New("hissdb del err " + res.State)
	}
	return nil
}

func (s *StoreHISSDB) ListObject(last_idx string, scan_len int64) ([]string, error) {
	//SSDB keys == scan命令
	res := s.pool.Cmd("keys", last_idx, "", scan_len)
	logs.Error("res %v", res)
	if res.State == "ok" {
		return res.List(), nil
	} else {
		return []string{}, errors.New("keys err " + res.State)
	}
}

func (s *StoreHISSDB) StoreKey(key string) string {
	now_time := time.Now()
	if s.format == "" {
		return key
	}
	return now_time.Format(s.format) + s.seq + key
}

func (s *StoreHISSDB) RedisKey(key_in_store string) (string, bool) {
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
