package main

import (
	"encoding/json"
	"testing"

	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func TestToNumInStr(t *testing.T) {
	/*
		sk := toNumInStr(3199999999999999999)
		t.Logf("s  : %s\n", sk)
		t.Logf("s  : %d\n", fromStrToNum(sk))
		s0 := toNumInStr(94999999999999999)  // ZWY ZWU GOL EFV
		t.Logf("s  : %s\n", s0)
		t.Logf("s  : %d\n", fromStrToNum(s0))
		s1 := toNumInStr(9999999999999999)   // CSV UNI ZIG KAP
		t.Logf("s : %s\n", s1)
		t.Logf("s  : %d\n", fromStrToNum(s1))
		s2 := toNumInStr( 3680000000000000)   // BAB UGE DMW ENO
		t.Logf("s : %s\n", s2)
		t.Logf("s  : %d\n", fromStrToNum(s2))
	*/
}

/*
func TestGiftCode(t *testing.T) {
	for j := 0; j < 100000; j++ {
		fmt.Printf("%6d -> %6d\n", int64(j), revNum(int64(j), 100000))
	}
	for i := 0; i < 100000; i++ {
		num := mkGiftCodeNum(1, 1, int64(i))
		str := toNumInStr(num)
		num_r := fromStrToNum(str)
		b, t, c, ok := parseGiftCodeNum(num_r)
		fmt.Printf("%d;%d;%d;%d;%s;%d;%d;%d;%d;%s;%v\n",
			1, 1, i, num, str, b, t, c, num_r, strings.ToLower(str), ok)
	}

}

func TestGiftCodeSimple(t *testing.T) {
	for i := 0; i < 100000; i++ {
		num := mkGiftCodeNum(1, 1, int64(i))
		str := toNumInStr(num)
		//num_r := fromStrToNum(str)
		//b,t,c, ok := parseGiftCodeNum(num_r)
		fmt.Printf("%d;%d;%d;%s;%s\n",
			1, 1, i, str, strings.ToLower(str))
	}

}
*/
func TestDynamoDB(t *testing.T) {
	//t.Logf("Start")

	db := &dynamodb.DynamoDB{}

	err := db.Connect(
		CommonConfig.AWS_Region,
		CommonConfig.AWS_AccessKey,
		CommonConfig.AWS_SecretKey,
		"")

	if err != nil {
		logs.Error("Connect Err %s", err.Error())
	}

	db.InitTable()

	var bid int64 = 3
	var tid int64 = 1
	var max int64 = 25000
	var c int64
	codes := Gen(bid, tid, max*2, false)
	for ; c < max; c++ {
		code := codes[c]
		values := make(map[string]dynamodb.Any, 8)

		values["State"] = int64(0)
		values["Begin"] = int64(1)
		values["End"] = int64(1449759661)
		b, err := json.Marshal(RedeemCodeValues{
			BatchID: "2",
			GroupID: "1",
			ItemIDs: []string{"VI_HC", "VI_SC", "NL_ALL_1_3", ""},
			Counts:  []uint32{288, 8888, 1},
			Title:   "封测礼包",
		})
		if err != nil {
			logs.Error("%d,Err,SetByHashM,%s,%d,%d,\"%v\"", c, code, bid, tid, values)
			continue
		}
		values["Value"] = string(b)
		logs.Trace("%d,Data,SetByHashM,%s,%d,%d,\"%v\"", c, code, bid, tid, values)
		err = db.SetByHashM(CommonConfig.Dynamo_DB, code, values)
		if err != nil {
			logs.Error("%d,Err,SetByHashM,%s,%d,%d,\"%v\"", c, code, bid, tid, values)
		}
	}

}

// 9   4     99999    9   9999  9       99   7     9
// 1   4     00000    0   0000  0       00   7     0
// 随机 固定  序号      随机 批号  随机     组号  固定  随机
