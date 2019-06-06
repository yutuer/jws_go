package main

import (
	"fmt"
	"reflect"

	"github.com/ugorji/go/codec"
	//"vcs.taiyouxi.net/platform/planx/client"
)

type ChatMsg struct {
	_struct bool `codec:",toarray"`
	// d7f5d5ac-b971-4790-a623-6ebb9d1cca03, 1000
	Uuid string
	Num  int
}

type ChatMsg2 struct {
	// d7f5d5ac-b971-4790-a623-6ebb9d1cca03, 1000
	Uuid string
	Num  int
}

//MUST Only can be *Struct
func (c *ChatMsg2) UnmarshalBinary(data []byte) error {
	var mh codec.MsgpackHandle
	dec := codec.NewDecoderBytes(data, &mh)
	dec.Decode(&c.Uuid)
	err := dec.Decode(&c.Num)
	return err
}

func (c *ChatMsg2) MarshalBinary() (data []byte, err error) {
	var mh codec.MsgpackHandle
	enc := codec.NewEncoderBytes(&data, &mh)
	enc.Encode(c.Uuid)
	enc.Encode(c.Num)
	err = nil //this is not the right way
	return
}

type Base struct {
	//_struct struct{} `codec:",omitempty"`
	Pass string
}

type Request struct {
	Base
	Abc string
	Bcd string
}

func main() {
	mapStrIntfTyp := reflect.TypeOf(map[string]interface{}(nil))
	mapStrIntfTypId := reflect.ValueOf(mapStrIntfTyp).Pointer()

	fmt.Println("map is %v, %v", mapStrIntfTyp, mapStrIntfTypId)

	// name := "uuid"
	name2 := "num"
	uuid := "d7f5d5ac-b971-4790-a623-6ebb9d1cca03"
	num := 1000
	msg := ChatMsg{true, uuid, num}
	msg2 := ChatMsg2{uuid, num}

	a := map[string]interface{}{
		"uuid": uuid,
		"num":  num,
		"a":    1,
	}
	fmt.Println("map len %d", len(a))

	var mh codec.MsgpackHandle
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	mh.RawToString = true

	var out []byte
	enc := codec.NewEncoderBytes(&out, &mh)
	// enc.Encode(name)
	enc.Encode(uuid)
	fmt.Println("1---%v", out)
	enc.Encode(name2)
	enc.Encode(num)
	fmt.Println("2---%v", out)

	var out2 []byte
	enc2 := codec.NewEncoderBytes(&out2, &mh)
	enc2.Encode(msg)
	fmt.Println("3.0---%v", out2)

	var out22 []byte
	enc22 := codec.NewEncoderBytes(&out22, &mh)
	err := enc22.Encode(&msg2) //& must match (c *ChatMsg2) MarshalBinary()
	fmt.Printf("3.1---%v, %v\n", out22, err)
	dec := codec.NewDecoderBytes(out22, &mh)
	var v ChatMsg2
	dec.Decode(&v)
	fmt.Printf("----%T, %v\n", v, v)

	var out3 []byte
	enc3 := codec.NewEncoderBytes(&out3, &mh)
	enc3.Encode(a)
	fmt.Println("4---%v", out3)

	var mh2 codec.MsgpackHandle
	//mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	//mh.RawToString = false
	mh2.WriteExt = true
	var out4 []byte
	abc := []byte{
		204, 131, 204, 169, 112, 114, 111, 102, 105, 108, 101, 105, 100, 204, 169, 48, 58, 48, 58, 49, 48, 48, 48, 49, 204, 177, 97, 116, 116, 114, 105, 98, 117, 116, 101, 115, 75, 101, 121, 76, 105, 115, 116, 204, 147, 204, 164, 110, 97, 109, 101, 204, 167, 115, 97, 118, 101, 118, 101, 114, 204, 163, 108, 118, 108, 204, 171, 112, 97, 115, 115, 116, 104, 114, 111, 117, 103, 104, 204, 217, 36, 49, 49, 51, 100, 55, 102, 51, 52, 45, 54, 100, 48, 97, 45, 52, 48, 55, 57, 45, 98, 51, 50, 53, 45, 98, 101, 51, 50, 101, 52, 98, 51, 101, 48, 97, 97,
	}
	enc4 := codec.NewEncoderBytes(&out4, &mh2)
	enc4.Encode(abc)
	fmt.Println("5---%v", out4)

	ooo := []byte{
		220, 0, 117, //0xdc array 16, len 117
		204, 131, 204, 169, 112, 114, 111, 102, 105, 108,
		101, 105, 100, 204, 169, 48, 58, 48, 58, 49,
		48, 48, 48, 49, 204, 177, 97, 116, 116, 114,
		105, 98, 117, 116, 101, 115, 75, 101, 121, 76,
		105, 115, 116, 204, 147, 204, 164, 110, 97, 109,
		101, 204, 167, 115, 97, 118, 101, 118, 101, 114,
		204, 163, 108, 118, 108, 204, 171, 112, 97, 115,
		115, 116, 104, 114, 111, 117, 103, 104, 204, 217,
		36, 49, 49, 51, 100, 55, 102, 51, 52, 45,
		54, 100, 48, 97, 45, 52, 48, 55, 57, 45,
		98, 51, 50, 53, 45, 98, 101, 51, 50, 101,
		52, 98, 51, 101, 48, 97, 97,
	}
	dec4 := codec.NewDecoderBytes(ooo, &mh2)
	var abcn []byte
	dec4.Decode(&abcn)
	fmt.Println("5.1---%v", abcn)

	r := Request{
		//Base: Base{"abc"},
		Abc: "1",
	}
	var out66 []byte
	enc66 := codec.NewEncoderBytes(&out66, &mh)
	err66 := enc66.Encode(r)
	fmt.Printf("6---%v, %v, %s\n", out66, err66, out66)
	dec66 := codec.NewDecoderBytes(out66, &mh)
	var r66 Request
	dec66.Decode(&r66)
	fmt.Printf("6.1----%T, %v\n", r66, r66)
}
