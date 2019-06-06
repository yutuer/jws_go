package client

import (
	"bytes"
	"encoding/gob"
	"time"

	"fmt"

	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
)

//Client ..
type Client struct {
	mc      *manageConnect
	modules *modules

	gid      uint32
	groups   []uint32
	shardIDs []uint32

	isClosed bool
}

//NewClient ..
func NewClient(gid uint32, shardIDs []uint32) *Client {
	c := &Client{
		gid:      gid,
		shardIDs: shardIDs,

		isClosed: false,
	}
	return c
}

//Start ..
func (c *Client) Start() error {
	modules := newModules()
	c.modules = modules
	c.modules.loadModules()

	mc := newManageConnect(c.gid, c.shardIDs)
	for _, g := range c.groups {
		mc.addService(g)
	}
	c.mc = mc
	if err := mc.start(); nil != err {
		return err
	}
	return nil
}

//Stop ..
func (c *Client) Stop() {
	c.isClosed = true
	c.mc.stop()
}

//AddGroupIDs ..
func (c *Client) AddGroupIDs(gl []uint32) {
	c.groups = append(c.groups, gl...)
}

//SetConnPoolMax ..
func (c *Client) SetConnPoolMax(pm uint32) {
	c.mc.setConnPoolMax(pm)
}

//CallSync ..
func (c *Client) CallSync(groupID uint32, moName string, meName string, source string, param module.Param) (module.Ret, int, error) {
	me := c.modules.getMethod(moName, meName)
	if nil == me {
		return nil, message.ErrCodeUnknownMethod, nil
	}

	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(param)

	req := &message.Request{
		ProtocolID: message.ProtocolSyncReq,
		Module:     moName,
		Method:     meName,
		GroupID:    groupID,
		HashSource: source,

		Data:    buf.Bytes(),
		DataLen: uint32(len(buf.Bytes())),
	}

	msgReq, err := message.EncodeRequestToMessage(req)
	if nil != err {
		return nil, message.ErrCodeEncode, err
	}

	conn, err := c.mc.getConn(groupID)
	if nil != err {
		return nil, message.ErrCodeInner, err
	}
	defer conn.Release()

	if err := conn.Send(msgReq); nil != err {
		if c.isClosed {
			return nil, message.ErrCodeClosed, err
		}
		return nil, message.ErrCodeInner, err
	}

	timeout := time.Now().Add(DefaultRecvTimeout)
	msgRsp, err := conn.RecvWithTimeout(timeout)
	if nil != err {
		return nil, message.ErrCodeInner, err
	}

	rsp, err := message.DecodeMessageToResponse(msgRsp)
	if nil != err {
		return nil, message.ErrCodeInner, err
	}
	if message.ProtocolSyncRsp != rsp.ProtocolID {
		return nil, message.ErrCodeInner, fmt.Errorf("get un-expect message protocal %d", rsp.ProtocolID)
	}

	if message.ErrCodeOK != rsp.ErrCode {
		if message.ErrCodeInner == rsp.ErrCode {
			return nil, int(rsp.ErrCode), fmt.Errorf("callback error, errcode %d, ..msg %+v", rsp.ErrCode, string(rsp.Data))
		}
		return nil, int(rsp.ErrCode), fmt.Errorf("callback error, errcode %d, ..msg %+v", rsp.ErrCode, rsp)
	}

	rspBuf := bytes.NewBuffer(rsp.Data)
	ret := me.NewRet()
	if err := gob.NewDecoder(rspBuf).Decode(ret); nil != err {
		return nil, message.ErrCodeDecode, err
	}

	return ret, message.ErrCodeOK, nil
}

//CallAsync ..
func (c *Client) CallAsync(groupID uint32, moName string, meName string, source string, param module.Param) (int, error) {
	me := c.modules.getMethod(moName, meName)
	if nil == me {
		return message.ErrCodeUnknownMethod, nil
	}

	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(param)

	req := &message.Request{
		ProtocolID: message.ProtocolAsyncReq,
		Module:     moName,
		Method:     meName,
		GroupID:    groupID,
		HashSource: source,

		Data:    buf.Bytes(),
		DataLen: uint32(len(buf.Bytes())),
	}

	msgReq, err := message.EncodeRequestToMessage(req)
	if nil != err {
		return message.ErrCodeEncode, err
	}

	conn, err := c.mc.getConn(groupID)
	if nil != err {
		return message.ErrCodeInner, err
	}
	defer conn.Release()

	if err := conn.Send(msgReq); nil != err {
		if c.isClosed {
			return message.ErrCodeClosed, err
		}
		return message.ErrCodeInner, err
	}

	timeout := time.Now().Add(DefaultRecvTimeout)
	msgRsp, err := conn.RecvWithTimeout(timeout)
	if nil != err {
		return message.ErrCodeInner, err
	}

	rsp, err := message.DecodeMessageToResponse(msgRsp)
	if nil != err {
		return message.ErrCodeDecode, err
	}
	if message.ProtocolAsyncRsp != rsp.ProtocolID {
		return message.ErrCodeInner, fmt.Errorf("get un-expect message protocal %d", rsp.ProtocolID)
	}

	if message.ErrCodeOK != rsp.ErrCode {
		if message.ErrCodeInner == rsp.ErrCode {
			return int(rsp.ErrCode), fmt.Errorf("callback error, errcode %d, ..msg %+v", rsp.ErrCode, string(rsp.Data))
		}
		return int(rsp.ErrCode), fmt.Errorf("callback error, errcode %d, ..msg %+v", rsp.ErrCode, rsp)
	}

	return int(rsp.ErrCode), nil
}

//Pull ..
func (c *Client) Pull() (*message.Request, module.Param, error) {
	req, ok := <-c.mc.pull()
	if false == ok {
		return nil, nil, fmt.Errorf("pull failed by close")
	}
	me := c.modules.getMethod(req.Module, req.Method)
	if nil == me {
		return nil, nil, fmt.Errorf("unknown mothod %s.%s", req.Module, req.Method)
	}
	param := me.NewParam()
	gob.NewDecoder(bytes.NewBuffer(req.Data)).Decode(param)
	return req, param, nil
}
