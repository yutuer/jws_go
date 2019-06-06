package redis

import (
	"testing"
	"time"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

func TestRedisCopy(t *testing.T) {
	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		t.Errorf("Dial redis failed, %v", err)
	}
	defer c.Close()

	key_name := "test_ZUNIONSTORE"
	num := 500000
	for i := 0; i < num; i++ {
		//mem := fmt.Sprintf("mem%d", i)
		_, err := c.Do("ZADD", key_name, i, uuid.NewV4().String())
		if nil != err {
			t.Errorf("redis do failed, %v", err)
			break
		}
	}

	t.Logf("prepared data, count [%d]", num)

	t1 := time.Now()
	res, err := c.Do("ZRANGE", key_name, 0, -1, "WITHSCORES")
	if nil != err {
		t.Errorf("redis do range, %v", err)
	}
	t.Logf("range cost %v", time.Now().Sub(t1).String())

	list, err := redis.Strings(res, err)
	if nil != err {
		t.Errorf("redis do parse, %v", err)
	}
	t.Logf("range result [%d]", len(list))

	t2 := time.Now()
	_, err = c.Do("ZUNIONSTORE", key_name+"_2", 1, key_name)
	if nil != err {
		t.Errorf("redis do copy, %v", err)
	}
	t.Logf("copy cost %v", time.Now().Sub(t2).String())
	check, err := c.Do("ZCOUNT", key_name+"_2", "-inf", "inf")
	cn, err := redis.Int(check, err)
	if nil != err {
		t.Errorf("redis do ZCOUNT, %v", err)
	}
	if cn != num {
		t.Errorf("redis check ZCOUNT failed, %d != %d", cn, num)
	}

	//t2_5 := time.Now()
	//res, err = c.Do("ZSCAN", key_name, 0, "COUNT", num)
	//if nil != err {
	//	t.Errorf("redis do zscan, %v", err)
	//}
	//t.Logf("zscan cost %v", time.Now().Sub(t2_5).String())
	//
	//layer1, err := redis.Values(res, err)
	//if nil != err {
	//	t.Errorf("redis do parse, %v", err)
	//}
	//list, err = redis.Strings(layer1[1], err)
	//if nil != err {
	//	t.Errorf("redis do parse, %v", err)
	//}
	//t.Logf("scan result [%d]", len(list))
	//t.Logf("show scan 10 {%v}", list[:10])

	t3 := time.Now()
	packlen := 100
	packnum := len(list) / packlen
	str := make([][]interface{}, packnum)
	for i := 0; i < packnum; i++ {
		cmd := make([]interface{}, packlen)
		for j := 0; j < packlen; j += 2 {
			cmd[j] = list[i*packlen+j+1]
			cmd[j+1] = list[i*packlen+j]
		}
		str[i] = cmd
	}
	t.Logf("prepare add cost %v", time.Now().Sub(t3).String())
	//t.Logf("show add command {%v}", str[0])

	t4 := time.Now()
	for i := 0; i < packnum; i++ {
		//for i := 0; i < len(list); i += 2 {
		//_, err = c.Do("ZADD", key_name+"_3", list[i+1], list[i])
		cmd := append([]interface{}{key_name + "_3"}, str[i]...)
		_, err = c.Do("ZADD", cmd...)
		if nil != err {
			t.Errorf("redis do failed, %v", err)
			break
		}
	}
	t.Logf("write cost %v", time.Now().Sub(t4).String())
	check, err = c.Do("ZCOUNT", key_name+"_3", "-inf", "inf")
	cn, err = redis.Int(check, err)
	if nil != err {
		t.Errorf("redis do ZCOUNT, %v", err)
	}
	if cn != num {
		t.Errorf("redis check ZCOUNT failed, %d != %d", cn, num)
	}

	c.Do("DEL", key_name)
	c.Do("DEL", key_name+"_2")
	c.Do("DEL", key_name+"_3")
}
