package bag

import (
	"reflect"
	"sort"

	"github.com/ugorji/go/codec"
)

/*
实现一个serrializable ordered map, 帮助db dirtyCheck实现稳定map检查
减少不必要的计算开销
*/

type BagSerializePair struct {
	Key uint32  `codec:"k"`
	BI  BagItem `codec:"b"`
}

var mh codec.MsgpackHandle

func init() {
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	mh.RawToString = true
	mh.WriteExt = true
}

type BagPairSlice []BagSerializePair

func (p BagPairSlice) Len() int           { return len(p) }
func (p BagPairSlice) Less(i, j int) bool { return p[i].Key < p[j].Key }
func (p BagPairSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type BagSerialize map[uint32]BagItem

func (bs BagSerialize) GobEncode() ([]byte, error) {
	si := make([]BagSerializePair, 0, len(bs))
	for i, v := range bs {
		si = append(si, BagSerializePair{i, v})
	}
	sort.Sort(BagPairSlice(si))
	var out []byte
	enc := codec.NewEncoderBytes(&out, &mh)
	err := enc.Encode(si)
	return out, err
}

//TODO by YZH BagSerialize GobDecode 没有测试过
func (bs BagSerialize) GobDecode(data []byte) error {
	dec := codec.NewDecoderBytes(data, &mh)
	var bsp []BagSerializePair
	if err := dec.Decode(&bsp); err != nil {
		return err
	}

	for _, v := range bsp {
		bs[v.Key] = v.BI
	}
	return nil
}
