package helper

import (
	"bytes"
	"encoding/gob"
)

//HelloReq ..
type HelloReq struct {
	ShardIDs []uint32
}

//EncodeHelloReq ..
func EncodeHelloReq(req *HelloReq) []byte {
	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(req)
	return buf.Bytes()
}

//DecodeHelloReq ..
func DecodeHelloReq(bs []byte) *HelloReq {
	req := &HelloReq{}
	gob.NewDecoder(bytes.NewBuffer(bs)).Decode(req)
	return req
}
