package server

import (
	"vcs.taiyouxi.net/jws/crossservice/message"
	"vcs.taiyouxi.net/jws/crossservice/module"
)

type transaction struct {
	req    *message.Request
	sync   bool
	rsp    chan *message.Response
	method module.Method
}

func makeSyncTransaction(req *message.Request) *transaction {
	t := &transaction{
		req:  req,
		sync: true,

		rsp: make(chan *message.Response, 5),
	}
	return t
}

func makeAsyncTransaction(req *message.Request) *transaction {
	t := &transaction{
		req:  req,
		sync: false,
	}
	return t
}
