package iputil

import (
	"bytes"
	"net"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//test to determine if a given ip is between two others (inclusive)
func IpBetweenString(from, to, test string) bool {
	ip1 := net.ParseIP(from)
	ip2 := net.ParseIP(to)
	testip := net.ParseIP(test)
	return IpBetween(ip1, ip2, testip)
}

func IpBetween(from net.IP, to net.IP, test net.IP) bool {
	if from == nil || to == nil || test == nil {
		logs.Warn("An ip input is nil") // or return an error!?
		return false
	}

	from16 := from.To16()
	to16 := to.To16()
	test16 := test.To16()
	if from16 == nil || to16 == nil || test16 == nil {
		logs.Warn("An ip did not convert to a 16 byte") // or return an error!?
		return false
	}

	if bytes.Compare(test16, from16) >= 0 && bytes.Compare(test16, to16) <= 0 {
		return true
	}
	return false
}
