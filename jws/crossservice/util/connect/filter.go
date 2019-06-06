package connect

import (
	"net"
)

//IPFilter ..
type IPFilter struct {
	filters []*net.IPNet
}

//NewIPFilter ..
func NewIPFilter() *IPFilter {
	f := &IPFilter{}
	f.filters = make([]*net.IPNet, 0)
	return f
}

func (f *IPFilter) check(str string) bool {
	if 0 == len(f.filters) {
		return true
	}
	ip := net.ParseIP(str)
	for _, f := range f.filters {
		if f.Contains(ip) {
			return true
		}
	}
	return false
}

//Add ..
func (f *IPFilter) Add(str string) {
	_, in, err := net.ParseCIDR(str)
	if nil != err {
		return
	}
	f.filters = append(f.filters, in)
}
