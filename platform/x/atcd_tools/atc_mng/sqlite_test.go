package atc_mng

import (
	"testing"
)

func TestSqlite(t *testing.T) {
	res, _ := GetAllAtcdClient("./atcd.db")
	t.Logf("%v", res)
}
