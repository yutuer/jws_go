package storehelper

import (
	"strings"

	"errors"

	"github.com/jackc/pgx"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type StorePostgreSQL struct {
	pool *pgx.ConnPool
	cfg  pgx.ConnPoolConfig
}

func NewStorePostgreSQL(
	host string, port int, db string,
	user, pass string) *StorePostgreSQL {

	n := &StorePostgreSQL{}
	n.cfg.Host = host
	n.cfg.Port = uint16(port)
	n.cfg.Database = db
	n.cfg.User = user
	n.cfg.Password = pass
	return n
}

func (s *StorePostgreSQL) Open() error {
	cfg := s.cfg

	pool, err := pgx.NewConnPool(cfg)
	if err != nil {
		logs.Error("NewStorePostgreSQL Err By %s", err.Error())
		return err
	} else {
		s.pool = pool
		mkAllSQLs(s.pool)
	}
	return nil
}

func (s *StorePostgreSQL) GetPool() *pgx.ConnPool {
	return s.pool
}

func (s *StorePostgreSQL) Clone() (IStore, error) {
	nn := *s
	n := &nn
	if err := n.Open(); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *StorePostgreSQL) Close() error {
	s.pool.Close()
	return nil
}

func (s *StorePostgreSQL) Put(key string, val []byte, rh ReadHandler) error {
	logs.Trace("StorePostgreSQL Put %s", key)
	keyS := strings.SplitN(key, ":", 2)
	if len(keyS) != 2 {
		return errors.New("KeyErr:" + key)
	}
	keyHead := keyS[0]
	id := keyS[1]

	return SetDataToPQ(s.pool, keyHead, id, val)
}

func (s *StorePostgreSQL) Get(key string) ([]byte, error) {
	logs.Trace("StorePostgreSQL Get %s", key)
	keyS := strings.SplitN(key, ":", 2)
	if len(keyS) != 2 {
		return []byte{}, errors.New("KeyErr:" + key)
	}
	keyHead := keyS[0]
	id := keyS[1]

	return GetDataFromPQ(s.pool, keyHead, id)
}

func (s *StorePostgreSQL) Del(key string) error {
	logs.Trace("StorePostgreSQL Del %s", key)
	keyS := strings.SplitN(key, ":", 2)
	if len(keyS) != 2 {
		return errors.New("KeyErr:" + key)
	}
	keyHead := keyS[0]
	id := keyS[1]

	return DelDataFromPQ(s.pool, keyHead, id)
}

func (s *StorePostgreSQL) StoreKey(key string) string {
	return key
}

func (s *StorePostgreSQL) RedisKey(key_in_store string) (string, bool) {
	return key_in_store, true
}
