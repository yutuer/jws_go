package safetable

import (
	"hash/crc32"
)

//SafeTable Hash表的
type SafeTable struct {
	buckets   []*safeBucket
	bucketLen uint32
}

//NewSafeTable 构造新的表
func NewSafeTable(bLen uint32) *SafeTable {
	st := &SafeTable{}

	l := uint32(64)
	for l < bLen {
		l *= 2
	}

	st.bucketLen = l
	st.buckets = make([]*safeBucket, st.bucketLen)

	for i := uint32(0); i < st.bucketLen; i++ {
		st.buckets[i] = newSafeBucket()
	}

	return st
}

//Get 从表中取值
func (t *SafeTable) Get(key string) interface{} {
	hk := crc32.ChecksumIEEE([]byte(key))
	b := t.buckets[hk%t.bucketLen]

	return b.get(key)
}

//Set 向表中写值
func (t *SafeTable) Set(key string, val interface{}) {
	hk := crc32.ChecksumIEEE([]byte(key))
	b := t.buckets[hk%t.bucketLen]
	b.set(key, val)
	return
}

//Del 从表中删除
func (t *SafeTable) Del(key string) {
	hk := crc32.ChecksumIEEE([]byte(key))
	b := t.buckets[hk%t.bucketLen]
	b.del(key)
}
