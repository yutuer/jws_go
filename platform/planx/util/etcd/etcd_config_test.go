package etcd

import (
	"testing"
)

func TestEtcdGetCfg(t *testing.T) {
	InitClient([]string{"http://127.0.0.1:2379/"})
	res, err := Get("/a3k/1/10/GameConfig/server_start_time")
	if err != nil {
		t.Errorf("get err %s", err.Error())
		return
	} else {
		t.Logf("res start_time %s", res)
	}

	res2, err2 := GetByServer("1", "10", "/GameConfig/listen")
	if err2 != nil {
		t.Errorf("get err GetByServer / %s", err2.Error())
		return
	} else {
		t.Logf("res start_time %s", res2)
	}

	res3, err3 := GetByServer("1", "10", "GameConfig/run_mode")
	if err3 != nil {
		t.Errorf("get err3 GetByServer  %s", err3.Error())
		return
	} else {
		t.Logf("res start_time %s", res3)
	}

	res4, err4 := GetAllSubKeys("/a3k/1/")
	if err4 != nil {
		t.Errorf("get err4 GetByServer  %s", err4.Error())
		return
	} else {
		t.Logf("res start_time %s", res4)
	}
}
