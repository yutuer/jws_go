package logic

import "testing"

func TestIsValidKey(t *testing.T) {
	keys := []string{"public/", "public/0/", "public/0/gonggao_3.2.0_r.json", "public/1/ public/1/gonggao_3.2.0_r.json", "public/metadata-snapshot_20151223"}
	result := []bool{false, false, true, true, false}
	for i := range keys {
		if isValidKey(keys[i]) != result[i] {
			t.Error("not equal %d", i)
			t.FailNow()
		}
	}
}

func TestParseNoticeKey(t *testing.T) {
	key := "public/0/gonggao_3.2.0_r.json"
	gid, version := parseNoticeKey(key)
	if gid != "0" || version != "3.2.0" {
		t.Errorf("gid=%s, version=%s", gid, version)
		t.FailNow()
	}
}
