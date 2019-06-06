package storehelper

import (
	"sync"
	"sync/atomic"

	"github.com/syndtr/goleveldb/leveldb"
	goleveldberrors "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//TODO YZH: batch write optmization of LevelDB
type StoreLevelDB struct {
	dbPath        string
	db            *leveldb.DB
	levelDBCloner uint64

	openOnce *sync.Once
}

func NewStoreLevelDB(dbPath string) *StoreLevelDB {
	n := &StoreLevelDB{
		dbPath:   dbPath,
		openOnce: &sync.Once{},
	}
	return n
}

func (s *StoreLevelDB) Open() error {
	var serr error

	//
	s.openOnce.Do(func() {
		db, err := leveldb.OpenFile(s.dbPath, nil)
		if err != nil {
			logs.Warn("Open leveldb %s failed. %s", s.dbPath, err.Error())
			if _, ok := err.(*goleveldberrors.ErrCorrupted); ok {
				logs.Warn("Try to recover leveldb.")
				db, rerr := leveldb.RecoverFile(s.dbPath, nil)
				if rerr != nil {
					logs.Error("Recover leveldb %s failed. %s", s.dbPath, rerr.Error())
					serr = rerr
					return
				}
				s.db = db
				return
			} else {
				logs.Error("Open leveldb %s failed finally. %s", s.dbPath, err.Error())
			}
		}
		atomic.AddUint64(&s.levelDBCloner, 1)
		s.db = db
	})
	return serr
}

func (s *StoreLevelDB) Close() error {
	final := atomic.AddUint64(&s.levelDBCloner, ^uint64(0))
	var me error
	if final == 1 {
		me = s.db.Close()
	}
	return me
}

func (s *StoreLevelDB) Put(key string, val []byte, rh ReadHandler) error {
	return s.db.Put([]byte(key), val, nil)
}

func (s *StoreLevelDB) Get(key string) ([]byte, error) {
	return s.db.Get([]byte(key), nil)
}

func (s *StoreLevelDB) Del(key string) error {
	return s.db.Delete([]byte(key), nil)
}

func (s *StoreLevelDB) StoreKey(key string) string {
	return key
}

func (s *StoreLevelDB) RedisKey(key_in_store string) (string, bool) {
	return key_in_store, true
}

// goleveldb is go routine safe
// https://github.com/syndtr/goleveldb/issues/55
func (s *StoreLevelDB) Clone() (IStore, error) {
	if atomic.LoadUint64(&s.levelDBCloner) == 0 {
		//如果Store还没有Open过，则必须Open一次
		if err := s.Open(); err != nil {
			return nil, err
		}
	} else {
		atomic.AddUint64(&s.levelDBCloner, 1)
	}
	return s, nil
}

func (s *StoreLevelDB) IterateAllObjects(fn func(key, value []byte)) error {
	iter := s.db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		fn(key, value)
	}
	iter.Release()
	return iter.Error()
}

func (s *StoreLevelDB) ListObject(last_idx string, scan_len int64) ([]string, error) {
	var lastIdx []byte
	if last_idx != "" {
		lastIdx = []byte(last_idx)
	}

	iter := s.db.NewIterator(&util.Range{Start: lastIdx, Limit: nil}, nil)
	keys := make([]string, 0, scan_len)

	var i int64
	for iter.Next() {
		// Use key/value.
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		key := iter.Key()
		if key != nil {
			keys = append(keys, string(key))
		}
		i++
		if i >= scan_len {
			break
		}
		// value := iter.Value()
	}
	iter.Release()

	return keys, iter.Error()
}
