package storehelper

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type StoreWithGZip struct {
	store IStore
}

func NewStoreWithGZip(store IStore) *StoreWithGZip {
	return &StoreWithGZip{
		store: store,
	}
}

func (s *StoreWithGZip) Clone() (IStore, error) {
	i, err := s.store.Clone()
	if err != nil {
		return nil, err
	}
	return &StoreWithGZip{
		store: i,
	}, nil
}

func (s *StoreWithGZip) Open() error {
	return s.store.Open()
}

func (s *StoreWithGZip) Close() error {
	return s.store.Close()
}

func (s *StoreWithGZip) Put(key string, val []byte, rh ReadHandler) error {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	defer w.Close()
	_, err := w.Write(val)
	if err != nil {
		return err
	}
	err = w.Flush()
	if err != nil {
		return err
	}
	logs.Trace("StoreWithGZip Put %s : %v", key, b.Bytes())
	return s.store.Put(key, b.Bytes(), rh)
}

func (s *StoreWithGZip) Get(key string) ([]byte, error) {
	datas, err := s.store.Get(key)
	if err != nil {
		return []byte{}, err
	}

	logs.Trace("StoreWithGZip Get %s : %v", key, datas)

	var b bytes.Buffer
	b.Write(datas)

	r, err := gzip.NewReader(&b)
	if err != nil {
		logs.Error("StoreWithGZip NewReader err %s", err.Error())
		return []byte{}, err
	}
	defer r.Close()
	undatas, _ := ioutil.ReadAll(r)
	return undatas, nil
}

func (s *StoreWithGZip) Del(key string) error {
	return s.store.Del(key)
}

func (s *StoreWithGZip) StoreKey(key string) string {
	return s.store.StoreKey(key)
}

func (s *StoreWithGZip) RedisKey(key_in_store string) (string, bool) {
	return s.store.RedisKey(key_in_store)
}
