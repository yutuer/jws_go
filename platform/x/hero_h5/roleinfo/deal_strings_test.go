package roleinfo

import (
	"testing"
)

type Test struct {
	my_string string
	start     int
	len       int
	ret       string
}

var myTest = []Test{
	Test{"abc", 6, 10, "c"},
	Test{"mams", 6, 11, "s"},
	Test{"abcdefghijklmn", -6, -4, ""},
}

func myTests(t *testing.T) {
	for _, v := range myTest {
		ret := Substr(v.my_string, v.start, v.len)
		print("%d add %d, want %d, but get %d", v.my_string, v.start, v.len, v.ret, ret)
		/*if ret != v.ret {
			t.Errorf("%d add %d, want %d, but get %d", v.my_string, v.start, v.len, v.ret,ret)
		}*/
	}
}
