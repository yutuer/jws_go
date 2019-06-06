package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// GiveGiftToFriend : 给友人赠送礼品
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGiveGiftToFriend 给友人赠送礼品请求消息定义
type reqMsgGiveGiftToFriend struct {
	Req
	TargetID string `codec:"target_id"` // 友人的ACID
}

// rspMsgGiveGiftToFriend 给友人赠送礼品回复消息定义
type rspMsgGiveGiftToFriend struct {
	SyncResp
}

// GiveGiftToFriend 给友人赠送礼品:
func (p *Account) GiveGiftToFriend(r servers.Request) *servers.Response {
	req := new(reqMsgGiveGiftToFriend)
	rsp := new(rspMsgGiveGiftToFriend)

	initReqRsp(
		"Attr/GiveGiftToFriendRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GiveGiftToFriendHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// ReceiveGiftFromFriend : 收取好友赠送的礼品
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgReceiveGiftFromFriend 收取好友赠送的礼品请求消息定义
type reqMsgReceiveGiftFromFriend struct {
	Req
	TargetID string `codec:"target_id"` // 友人的ACID
}

// rspMsgReceiveGiftFromFriend 收取好友赠送的礼品回复消息定义
type rspMsgReceiveGiftFromFriend struct {
	SyncRespWithRewards
}

// ReceiveGiftFromFriend 收取好友赠送的礼品:
func (p *Account) ReceiveGiftFromFriend(r servers.Request) *servers.Response {
	req := new(reqMsgReceiveGiftFromFriend)
	rsp := new(rspMsgReceiveGiftFromFriend)

	initReqRsp(
		"Attr/ReceiveGiftFromFriendRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.ReceiveGiftFromFriendHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// BatchGiveGift2Friend : 批量给友人赠送礼品
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgBatchGiveGift2Friend 批量给友人赠送礼品请求消息定义
type reqMsgBatchGiveGift2Friend struct {
	Req
}

// rspMsgBatchGiveGift2Friend 批量给友人赠送礼品回复消息定义
type rspMsgBatchGiveGift2Friend struct {
	SyncResp
}

// BatchGiveGift2Friend 批量给友人赠送礼品:
func (p *Account) BatchGiveGift2Friend(r servers.Request) *servers.Response {
	req := new(reqMsgBatchGiveGift2Friend)
	rsp := new(rspMsgBatchGiveGift2Friend)

	initReqRsp(
		"Attr/BatchGiveGift2FriendRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.BatchGiveGift2FriendHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// BatchReceiveGiftFromFriend : 批量收取好友赠送的礼品
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgBatchReceiveGiftFromFriend 批量收取好友赠送的礼品请求消息定义
type reqMsgBatchReceiveGiftFromFriend struct {
	Req
}

// rspMsgBatchReceiveGiftFromFriend 批量收取好友赠送的礼品回复消息定义
type rspMsgBatchReceiveGiftFromFriend struct {
	SyncRespWithRewards
}

// BatchReceiveGiftFromFriend 批量收取好友赠送的礼品:
func (p *Account) BatchReceiveGiftFromFriend(r servers.Request) *servers.Response {
	req := new(reqMsgBatchReceiveGiftFromFriend)
	rsp := new(rspMsgBatchReceiveGiftFromFriend)

	initReqRsp(
		"Attr/BatchReceiveGiftFromFriendRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.BatchReceiveGiftFromFriendHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetReceiveGiftInfo : 获得收到的好友礼品列表信息
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetReceiveGiftInfo 获得收到的好友礼品列表信息请求消息定义
type reqMsgGetReceiveGiftInfo struct {
	Req
}

// rspMsgGetReceiveGiftInfo 获得收到的好友礼品列表信息回复消息定义
type rspMsgGetReceiveGiftInfo struct {
	SyncResp
}

// GetReceiveGiftInfo 获得收到的好友礼品列表信息:
func (p *Account) GetReceiveGiftInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGetReceiveGiftInfo)
	rsp := new(rspMsgGetReceiveGiftInfo)

	initReqRsp(
		"Attr/GetReceiveGiftInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetReceiveGiftInfoHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetFriendGiftAcID : 获得已赠予礼品的好友的名单
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetFriendGiftAcID 获得已赠予礼品的好友的名单请求消息定义
type reqMsgGetFriendGiftAcID struct {
	Req
}

// rspMsgGetFriendGiftAcID 获得已赠予礼品的好友的名单回复消息定义
type rspMsgGetFriendGiftAcID struct {
	SyncResp
}

// GetFriendGiftAcID 获得已赠予礼品的好友的名单:
func (p *Account) GetFriendGiftAcID(r servers.Request) *servers.Response {
	req := new(reqMsgGetFriendGiftAcID)
	rsp := new(rspMsgGetFriendGiftAcID)

	initReqRsp(
		"Attr/GetFriendGiftAcIDRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetFriendGiftAcIDHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
