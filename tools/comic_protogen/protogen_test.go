package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseProtoName(t *testing.T) {
	fmt.Println(strings.LastIndex("StartSingleBattleReq", "Req"))
	trimName := strings.Trim("StartSingleBattleReq", "Req")
	fmt.Println(trimName)
	protoName := parseProtoName("message StartSingleBattleReq {")
	if protoName != "StartSingleBattle" {
		t.Error(protoName)
		t.FailNow()
	}
}
