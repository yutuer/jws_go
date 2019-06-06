package message

import (
	"fmt"

	"bytes"

	"encoding/gob"

	"vcs.taiyouxi.net/jws/crossservice/util/connect"
)

//Request ..
type Request struct {
	ProtocolID uint32
	Module     string
	Method     string
	GroupID    uint32
	HashSource string

	DataLen uint32
	Data    []byte
}

//Response ..
type Response struct {
	ProtocolID uint32
	ErrCode    uint32

	DataLen uint32
	Data    []byte
}

//DecodeMessageToRequest ..
func DecodeMessageToRequest(msg *connect.Message) (*Request, error) {
	req := &Request{}
	buf := bytes.NewBuffer(msg.Payload[:msg.Length])
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(req); nil != err {
		return nil, fmt.Errorf("decode message failed, %v", err)
	}
	return req, nil
}

//EncodeRequestToMessage ..
func EncodeRequestToMessage(req *Request) (*connect.Message, error) {
	msg := &connect.Message{}
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(req); nil != err {
		return nil, fmt.Errorf("encode request failed, %v", err)
	}
	msg.Payload = buf.Bytes()
	msg.Length = len(msg.Payload)
	return msg, nil
}

//DecodeMessageToResponse ..
func DecodeMessageToResponse(msg *connect.Message) (*Response, error) {
	rsp := &Response{}
	buf := bytes.NewBuffer(msg.Payload[:msg.Length])
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(rsp); nil != err {
		return nil, fmt.Errorf("decode message failed, %v", err)
	}
	return rsp, nil
}

//EncodeResponseToMessage ..
func EncodeResponseToMessage(rsp *Response) (*connect.Message, error) {
	msg := &connect.Message{}
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(rsp); nil != err {
		return nil, fmt.Errorf("encode response failed, %v", err)
	}
	msg.Payload = buf.Bytes()
	msg.Length = len(msg.Payload)
	return msg, nil
}

//MakeHelloAckResponse ..
func MakeHelloAckResponse() *Response {
	rsp := &Response{}
	rsp.ProtocolID = ProtocolHelloRsp
	rsp.ErrCode = ErrCodeOK
	rsp.DataLen = 0
	rsp.Data = []byte{}
	return rsp
}

//MakeTmpResponse ..
func MakeTmpResponse(e uint32, data []byte) *Response {
	rsp := &Response{}
	rsp.ProtocolID = ProtocolInvalid
	rsp.ErrCode = e
	rsp.DataLen = uint32(len(data))
	rsp.Data = data[:]
	return rsp
}

//MakeSyncAckResponse ..
func MakeSyncAckResponse(r *Response) *Response {
	rsp := &Response{}
	rsp.ProtocolID = ProtocolSyncRsp
	rsp.ErrCode = r.ErrCode
	rsp.DataLen = r.DataLen
	rsp.Data = r.Data[:]
	return rsp
}

//MakeAsyncAckResponse ..
func MakeAsyncAckResponse(req *Request) *Response {
	rsp := &Response{}
	rsp.ProtocolID = ProtocolAsyncRsp
	rsp.ErrCode = ErrCodeOK
	rsp.DataLen = 0
	rsp.Data = []byte{}
	return rsp
}

//MakeErrResponse ..
func MakeErrResponse(proto uint32, ec int, err error) *Response {
	rsp := &Response{}
	rsp.ProtocolID = proto
	rsp.ErrCode = uint32(ec)
	bsErr := []byte(err.Error())
	rsp.Data = bsErr
	rsp.DataLen = uint32(len(bsErr))
	return rsp
}

//MakeErrInnerResponse ..
func MakeErrInnerResponse(proto uint32, err error) *Response {
	return MakeErrResponse(proto, ErrCodeInner, err)
}
