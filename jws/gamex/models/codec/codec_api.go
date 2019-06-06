package codec

import (
	"reflect"

	"github.com/ugorji/go/codec"
)

var mh codec.MsgpackHandle

func init() {
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	mh.RawToString = true
	mh.WriteExt = true
}

//Decode 网络包反序列化
func Decode(raw []byte, out interface{}) {
	dec := codec.NewDecoderBytes(raw, &mh)
	dec.Decode(out)
}

//Encode 网络包序列化
func Encode(value interface{}) []byte {
	var out []byte //XXX: If it comes from []byte buffer pool, it would be cool
	enc := codec.NewEncoderBytes(&out, &mh)
	enc.Encode(value)
	return out
}

//DecAsDict 网络包反序列化map
func DecAsDict(raw []byte) map[string]interface{} {
	var value map[string]interface{}
	dec := codec.NewDecoderBytes(raw, &mh)
	dec.Decode(&value)
	//fmt.Println("decAsDict:", value, "bytes:", raw)
	return value
}

//EncAsDict 网络包序列化map
func EncAsDict(value interface{}) []byte {
	var out []byte //XXX: If it comes from []byte buffer pool, it would be cool
	enc := codec.NewEncoderBytes(&out, &mh)
	enc.Encode(value)
	return out
}
