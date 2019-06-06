package message

//Protocol ..
const (
	ProtocolInvalid = iota
	ProtocolHelloReq
	ProtocolHelloRsp
	ProtocolSyncReq
	ProtocolSyncRsp
	ProtocolAsyncReq
	ProtocolAsyncRsp
	ProtocolPushReq
	ProtocolPushRsp
)
