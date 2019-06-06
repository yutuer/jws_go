package server

import (
	"testing"

	"bytes"
	"encoding/gob"

	"vcs.taiyouxi.net/jws/crossservice/module"
)

type TParam struct {
	P1 uint32
	P2 string
}

func TestMethodEncode(t *testing.T) {
	p := module.Param(&TParam{P1: 12, P2: "ht12"})
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(p)
	if nil != err {
		t.Logf("gob encode failed, %v", err)
		return
	}
	t.Logf("gob encode, %+v   -->   %v", p, buf.Bytes())

	op := module.Param(&TParam{})
	if err := gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(op); nil != err {
		t.Logf("gob decode failed, %v", err)
		return
	}
	t.Logf("gob decode, %+v", op)
}
