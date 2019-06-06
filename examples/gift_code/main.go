package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"os"
	"time"

	"flag"

	"strings"

	"path"

	"github.com/tealeg/xlsx"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"vcs.taiyouxi.net/jws/gamex/modules/redeem_code"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/dynamodb"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// prod
//var (
//	Dynamo_DB     = "RedeemCode"
//	AWS_Region    = "cn-north-1"
//	AWS_AccessKey = "AKIAO4YSP5CZU5CDQS4A"
//	AWS_SecretKey = "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"
//	IsToDynamoDB  = true
//)

// qa

type CommonCfg struct {
	Dynamo_DB     string `toml:"Dynamo_DB"`
	AWS_Region    string `toml:"AWS_Region"`
	AWS_AccessKey string `toml:"AWS_AccessKey"`
	AWS_SecretKey string `toml:"AWS_SecretKey"`
	MongoURL      string `toml:"MongoURL"`
	IsToDynamoDB  bool   `toml:"IsToDynamoDB"`
}

func AddMongoDB(bid, tid, max int64,
	isLimit, isRandom bool,
	tb, te int64, title string,
	ids []string, cs []uint32) []string {
	session, err := mgo.DialWithTimeout(CommonConfig.MongoURL, 20*time.Second)
	if err != nil {
		return nil
	}
	session.SetMode(mgo.Monotonic, true)
	if err := session.DB(CommonConfig.Dynamo_DB).C(CommonConfig.Dynamo_DB).EnsureIndex(mgo.Index{
		Key:        []string{"code"},
		Unique:     true,
		Sparse:     true,
		Background: true,
		DropDups:   true,
	}); err != nil {
		logs.Critical("DBByMongoDB EnsureIndex err : %s", err.Error())
		panic(err)
	}
	collecion := session.DB(CommonConfig.Dynamo_DB).C(CommonConfig.Dynamo_DB)

	var c int64
	res := make([]string, 0, 320000)
	codes := Gen(bid, tid, max, isLimit, isRandom)
	for ; c < max; c++ {
		code := codes[c]
		b, err := json.Marshal(RedeemCodeValues{
			BatchID: fmt.Sprintf("%d", bid),
			GroupID: fmt.Sprintf("%d", tid),
			ItemIDs: ids,
			Counts:  cs,
			Title:   title,
		})
		if err != nil {
			panic(err)
			continue
		}
		info, err := collecion.Upsert(bson.D{{"code", code}},
			bson.M{"$set": redeemCodeModule.RedeemCodeExchange{
				Code:  code,
				State: int64(0),
				Begin: tb,
				End:   te,
				Value: string(b),
			}})
		if err != nil {
			logs.Error("Err %v,%d,SetByHashM,%s,%d,%d,\"%v\"", err, c, code, bid, tid, info)
			c = c - 1
			time.Sleep(3 * time.Second)
			continue
		}
		if c%100 == 0 {
			logs.Info("curr %d-%d %d/%d", bid, tid, c, max)
			//time.Sleep(1 * time.Second)
		}
		res = append(res, fmt.Sprintf("%d,%d,%d,%s,\"%v\"\n", c, bid, tid, code, info))
	}
	return res

}

func AddDynamoDB(bid, tid, max int64,
	isLimit, isRandom bool,
	tb, te int64, title string,
	ids []string, cs []uint32) []string {
	db := &dynamodb.DynamoDB{}

	if CommonConfig.IsToDynamoDB {
		err := db.Connect(
			CommonConfig.AWS_Region,
			CommonConfig.AWS_AccessKey,
			CommonConfig.AWS_SecretKey,
			"")
		logs.Info("cc %v", CC)
		if err != nil {
			logs.Error("Connect Err %s", err.Error())
			panic(err)
		}
		db.InitTable()
	}
	res := make([]string, 0, 320000)

	var c int64
	codes := Gen(bid, tid, max, isLimit, isRandom)
	for ; c < max; c++ {
		code := codes[c]
		values := make(map[string]interface{}, 8)

		values["State"] = int64(0)
		values["Begin"] = tb
		values["End"] = te
		b, err := json.Marshal(RedeemCodeValues{
			BatchID: fmt.Sprintf("%d", bid),
			GroupID: fmt.Sprintf("%d", tid),
			ItemIDs: ids,
			Counts:  cs,
			Title:   title,
		})
		if err != nil {
			logs.Error("Err %v,%d,Marshal,%s,%d,%d,\"%v\"", err, c, code, bid, tid, values)
			panic(err)
			continue
		}
		values["Value"] = string(b)
		if CommonConfig.IsToDynamoDB {
			err = db.SetByHashM(CommonConfig.Dynamo_DB, code, values)
			if err != nil {
				logs.Error("Err %v,%d,SetByHashM,%s,%d,%d,\"%v\"", err, c, code, bid, tid, values)
				c = c - 1
				time.Sleep(3 * time.Second)
				continue
			}
			if c%100 == 0 {
				logs.Info("curr %d-%d %d/%d", bid, tid, c, max)
				//time.Sleep(1 * time.Second)
			}
		}
		res = append(res, fmt.Sprintf("%d,%d,%d,%s,\"%v\"\n", c, bid, tid, code, values))

	}
	return res
}

/*
新手礼包	26500	钻石*88、金币*5888、5级紫色戒指*1	VI_HC=88;VI_SC=5888;RG_ALL_1_3=1	2015年12月24日	2016年1月10日
普通礼包	15-1	封测礼包	26500	钻石*188、精铁*2000、7级紫色项链*1	VI_HC=188;VI_FI=2000;NL_ALL_1_3=1	2015年12月24日	2016年1月10日
普通礼包	16-1	精英礼包	26500	钻石*288、金币*8888、5级金色戒指*1、升星石*10	VI_HC=288;VI_SC=8888;RG_ALL_1_4=1;MAT_StarStone=10	2015年12月24日	2016年1月10日
普通礼包	17-1	独家礼包	6000	钻石*288、金币*8888、祈星符*5、升星石*10	VI_HC=288;VI_SC=8888;MAT_StarBless=5;MAT_StarStone=10	2015年12月24日	2016年1月10日
普通礼包	18-1	豪华礼包	5000	钻石*388、金币*8888、7级金色项链*1、升星石*10	VI_HC=388;VI_SC=8888;NL_ALL_1_4=1;MAT_StarStone=10	2015年12月24日	2016年1月10日
普通礼包	19-1	Q群礼包	5000	钻石*200	VI_HC=200	2015年12月24日	2016年1月10日
通用礼包	20-1	微信礼包	1

*/
/*
func main() {
	//AddDynamoDB(1, 1, 25000, false, 1, 1449759661, "新手礼包", []string{"VI_HC", "VI_SC", "RG_ALL_1_3"}, []uint32{88, 5888, 1})
	//AddDynamoDB(2, 1, 25000, false, 1, 1449759661, "封测礼包", []string{"VI_HC", "VI_FI", "NL_ALL_1_3"}, []uint32{188, 2000, 1})
	//AddDynamoDB(8, 1, 1, true, 1, 1449763199, "UC更新礼包", []string{"VI_HC", "USEI_XP_2000", "VI_SC" }, []uint32{500, 10, 50000})
	//AddDynamoDB(9, 1, 1, true, 1, 1449763199, "老玩家更新礼包", []string{"VI_HC", "USEI_XP_2000", "VI_FI"}, []uint32{300, 10, 20000})
	//AddDynamoDB(11, 1, 20, false, false, 1, 1452441601, "封测限定", []string{"WP_GY_1_4_ACT"}, []uint32{1})
	//AddDynamoDB(12, 1, 50, false, false, 1, 1452441601, "封测限定", []string{"WP_GY_1_4"}, []uint32{1})

	// 下面这些码一次性重新生成, 因为随机数在调用过程中重置了 范杨 2015-12-25
	//AddDynamoDB(13, 1, 500, false, false, 1, 1452441601, "999钻礼包", []string{"VI_HC"}, []uint32{999})
	//AddDynamoDB(14, 1, 27000, false, false, 1, 1452441601, "新手礼包", []string{"VI_HC", "VI_SC", "RG_ALL_1_3"}, []uint32{88, 5888, 1})
	//AddDynamoDB(15, 1, 27000, false, false, 1, 1452441601, "封测礼包", []string{"VI_HC", "VI_FI", "NL_ALL_1_3"}, []uint32{188, 2000, 1})
	//AddDynamoDB(16, 1, 27000, false, false, 1, 1452441601, "精英礼包", []string{"VI_HC", "VI_SC", "RG_ALL_1_4", "MAT_StarStone"}, []uint32{288, 8888, 1, 10})
	//AddDynamoDB(17, 1, 6000, false, false, 1, 1452441601, "独家礼包", []string{"VI_HC", "VI_SC", "MAT_StarBless", "MAT_StarStone"}, []uint32{288, 8888, 1, 10})
	//AddDynamoDB(18, 1, 6000, false, false, 1, 1452441601, "豪华礼包", []string{"VI_HC", "VI_SC", "NL_ALL_1_4", "MAT_StarStone"}, []uint32{388, 8888, 1, 10})
	//AddDynamoDB(19, 1, 6000, false, false, 1, 1452441601, "Q群礼包", []string{"VI_HC"}, []uint32{200})
	//AddDynamoDB(20, 1, 1, true, false, 1, 1452441601, "微信礼包", []string{"VI_HC", "VI_SC"}, []uint32{200, 8888})
	// 上面这些码一次性重新生成, 因为随机数在调用过程中重置了 范杨 2015-12-25

	//AddDynamoDB(21, 1, 6000, false, false, 1, 1452441601, "独家礼包", []string{"VI_HC", "VI_SC", "MAT_StarBless", "MAT_StarStone"}, []uint32{288, 8888, 5, 10})

	// 下面这些码一次性重新生成, 因为随机数在调用过程中重置了 范杨 2016-4-8 --> 2016-6-5 : 1465142399

		AddDynamoDB(26, 1, 9500+50, false, false, 1, 1465142399, "新手礼包",
			// VI_HC=188;VI_SC=8888;VI_FI=2888;VI_DC=2000;VI_EN=50
			[]string{"VI_HC", "VI_SC", "VI_FI", "VI_DC", "VI_EN"},
			[]uint32{188, 8888, 2888, 2000, 50})

		AddDynamoDB(27, 1, 6500+50, false, false, 1, 1465142399, "精英礼包",
			//VI_HC=288;JD_RED_2=1;JD_BLUE_2=1;JD_GREEN_2=1;SI_SB_1=2
			[]string{"VI_HC", "JD_RED_2", "JD_BLUE_2", "JD_GREEN_2", "SI_SB_1"},
			[]uint32{288, 1, 1, 1, 2})

		AddDynamoDB(28, 1, 4000+50, false, false, 1, 1465142399, "豪华礼包",
			// VI_HC=288;VI_SC=18888;JD_PURPLE_3=1;JD_YELLOW_3=1;SI_SB_2=1
			[]string{"VI_HC", "VI_SC", "JD_PURPLE_3", "JD_YELLOW_3", "SI_SB_2"},
			[]uint32{288, 18888, 1, 1, 1})

		AddDynamoDB(29, 1, 3500+50, false, false, 1, 1465142399, "贴吧关注礼包",
			//  VI_HC=88;VI_EN=50;VI_DC=200;VI_FI=1000
			[]string{"VI_HC", "VI_EN", "VI_DC", "VI_FI"},
			[]uint32{88, 50, 200, 1000})

		AddDynamoDB(30, 1, 20+50, false, false, 1, 1465142399, "2088钻石礼包",
			// VI_HC=2088
			[]string{"VI_HC"},
			[]uint32{2088})

		AddDynamoDB(31, 1, 20+50, false, false, 1, 1465142399, "1088钻石礼包",
			// VI_HC=1088
			[]string{"VI_HC"},
			[]uint32{1088})

		AddDynamoDB(32, 1, 20+50, false, false, 1, 1465142399, "588钻石礼包",
			//  VI_HC=588
			[]string{"VI_HC"},
			[]uint32{588})

		AddDynamoDB(33, 1, 350+50, false, false, 1, 1465142399, "贴吧建议礼包",
			// VI_HC=288;JD_RED_3=1;VI_DC=2000
			[]string{"VI_HC", "JD_RED_3", "VI_DC"},
			[]uint32{288, 1, 2000})

		AddDynamoDB(34, 1, 350+50, false, false, 1, 1465142399, "贴吧竞猜礼包",
			// VI_HC=288;JD_PURPLE_3=1;VI_SC=20000
			[]string{"VI_HC", "JD_PURPLE_3", "VI_SC"},
			[]uint32{288, 1, 20000})
*/
// 上面这些码一次性重新生成, 因为随机数在调用过程中重置了 范杨 2016-4-8

// 下面这些码一次性重新生成, 因为随机数在调用过程中重置了 范杨 2016-4-11 --> 2016-6-5 : 1465142399
/*
	AddDynamoDB(35, 1, 5500, false, false, 1, 1465142399, "论坛专属礼包",
		// VI_HC=88;VI_EN=50;VI_DC=200;VI_FI=1000
		[]string{"VI_HC", "VI_EN", "VI_DC", "VI_FI"},
		[]uint32{88, 50, 200, 1000})

	AddDynamoDB(36, 1, 5500, false, false, 1, 1465142399, "论坛活跃礼包",
		//VI_HC=88;JD_RED_1=1;JD_BLUE_1=1;JD_GREEN_1=1;VI_SC=10000
		[]string{"VI_HC", "JD_RED_1", "JD_BLUE_1", "JD_GREEN_1", "VI_SC"},
		[]uint32{88, 1, 1, 1, 10000})

	AddDynamoDB(37, 1, 2600, false, false, 1, 1465142399, "Q群礼包",
		// VI_HC=200
		[]string{"VI_HC"},
		[]uint32{200})

	AddDynamoDB(38, 1, 1, true, false, 1, 1465142399, "微信礼包",
		//  VI_HC=200;VI_SC=8888
		[]string{"VI_HC", "VI_SC"},
		[]uint32{200, 8888})
*/
// 上面这些码一次性重新生成, 因为随机数在调用过程中重置了 范杨 2016-4-11

// 下面这些码一次性重新生成, 因为随机数在调用过程中重置了 范杨 2016-4-12 --> 2016-6-5 : 1465142399
/*
	AddDynamoDB(39, 1, 2100, false, false, 1, 1465142399, "官网礼包",
		//VI_HC=200;VI_SC=20000;VI_DC=2000;VI_FI=4000
		[]string{"VI_HC", "VI_SC", "VI_DC", "VI_FI"},
		[]uint32{200, 20000, 2000, 4000})
*/
// 上面这些码一次性重新生成, 因为随机数在调用过程中重置了 范杨 2016-4-12

// 下面这些码一次性重新生成, 因为随机数在调用过程中重置了 范杨 2016-4-13 --> 2016-6-5 : 1465142399
/*AddDynamoDB(40, 1, 2100, false, false, 1, 1465142399, "媒体专属A礼包",
	// VI_HC=108;VI_FI=2000;JD_RED_3=1;JD_GREEN_3=1
	[]string{"VI_HC", "VI_FI", "JD_RED_3", "JD_GREEN_3"},
	[]uint32{108, 2000, 1, 1})

AddDynamoDB(41, 1, 2100, false, false, 1, 1465142399, "媒体专属B礼包",
	// VI_HC=108;VI_SC=10000;JD_GREEN_3=1;JD_BLUE_3=1
	[]string{"VI_HC", "VI_SC", "JD_GREEN_3", "JD_BLUE_3"},
	[]uint32{108, 10000, 1, 1})

AddDynamoDB(42, 1, 2100, false, false, 1, 1465142399, "媒体专属C礼包",
	// VI_HC=108;VI_DC=200;JD_BLUE_3=1;JD_PURPLE_3=1
	[]string{"VI_HC", "VI_DC", "JD_BLUE_3", "JD_PURPLE_3"},
	[]uint32{108, 200, 1, 1})

AddDynamoDB(43, 1, 2100, false, false, 1, 1465142399, "媒体专属D礼包",
	// VI_HC=108;VI_GT=3;JD_PURPLE_3=1;JD_CHING_3=1
	[]string{"VI_HC", "VI_GT", "JD_PURPLE_3", "JD_CHING_3"},
	[]uint32{108, 3, 1, 1})

AddDynamoDB(44, 1, 5100, false, false, 1, 1465142399, "媒体新手礼包",
	//  VI_HC=128;VI_SC=8888;VI_FI=2888;VI_EN=50
	[]string{"VI_HC", "VI_SC", "VI_FI", "VI_EN"},
	[]uint32{128, 8888, 2888, 50})*/
// 上面这些码一次性重新生成, 因为随机数在调用过程中重置了 范杨 2016-4-13

// 下面这些码一次性重新生成, 因为随机数在调用过程中重置了 范杨 2016-4-27 --> 2016-6-5 : 1465142399
/*
	AddDynamoDB(45, 1, 2600, false, false, 1, 1465142399, "更新贴吧礼包",
		// VI_HC=288;VI_GT=10
		[]string{"VI_HC", "VI_GT"},
		[]uint32{288, 10})

	AddDynamoDB(46, 1, 1, true, false, 1, 1465142399, "更新微信礼包",
		//  VI_HC=200;VI_DC=1000
		[]string{"VI_HC", "VI_DC"},
		[]uint32{200, 1000})
*/
// 上面这些码一次性重新生成, 因为随机数在调用过程中重置了 范杨 2016-4-27
// 下面这些码一次性重新生成, 因为随机数在调用过程中重置了 范杨 2016-4-28 --> 2016-6-5 : 1465142399
/*
	AddDynamoDB(47, 1, 2600, false, false, 1, 1465142399, "五一Q群礼包",
		// VI_HC=2000;SI_SB_2=10;VI_DC=2000
		[]string{"VI_HC", "SI_SB_2", "VI_DC"},
		[]uint32{2000, 10, 2000})
*/
// 上面这些码一次性重新生成, 因为随机数在调用过程中重置了 范杨 2016-4-28

/*
	//普通礼包	48-1	新手礼包	15000	钻石*108、金币*8888、精铁*2888、体力*50
	// VI_HC=108;VI_SC=8888;VI_FI=2888;VI_EN=50	2016年5月16日	2016年6月30日
	AddDynamoDB(48, 1, 15000, false, false, 1, 1467302399, "新手礼包",
		// VI_HC=108;VI_SC=8888;VI_FI=2888;VI_EN=50
		[]string{"VI_HC", "VI_SC", "VI_FI", "VI_EN"},
		[]uint32{108, 8888, 2888, 50})

	//普通礼包	49-1	精英礼包	15000	钻石*188、2级红龙玉*1、2级蓝龙玉*1、2级绿龙玉*1，初级祝福石*2
	// VI_HC=188;JD_RED_2=1;JD_BLUE_2=1;JD_GREEN_2=1;SI_SB_1=2	2016年5月16日	2016年6月30日
	AddDynamoDB(49, 1, 15000, false, false, 1, 1467302399, "精英礼包",
		// VI_HC=188;JD_RED_2=1;JD_BLUE_2=1;JD_GREEN_2=1;SI_SB_1=2
		[]string{"VI_HC", "JD_RED_2", "JD_BLUE_2", "JD_GREEN_2", "SI_SB_1"},
		[]uint32{188, 1, 1, 1, 2})

	//普通礼包	50-1	独家礼包	15000	钻石*288、金币*18888、3级紫龙玉*1、3级黄龙玉*1、中级祝福石*1
	// VI_HC=288;VI_SC=18888;JD_PURPLE_3=1;JD_YELLOW_3=1;SI_SB_2=1	2016年5月16日	2016年6月30日
	AddDynamoDB(50, 1, 15000, false, false, 1, 1467302399, "独家礼包",
		//  VI_HC=288;VI_SC=18888;JD_PURPLE_3=1;JD_YELLOW_3=1;SI_SB_2=1
		[]string{"VI_HC", "VI_SC", "JD_PURPLE_3", "JD_YELLOW_3", "SI_SB_2"},
		[]uint32{288, 18888, 1, 1, 1})

	//普通礼包	51-1	端午礼包	15000	钻石*188、体力*100，2级红龙玉*1、2级蓝龙玉*1、2级绿龙玉*1
	// VI_HC=188;VI_EN=100;JD_RED_2=1;JD_BLUE_2=1;JD_GREEN_2=1	2016年5月16日	2016年6月30日
	AddDynamoDB(51, 1, 15000, false, false, 1, 1467302399, "端午礼包",
		// VI_HC=188;VI_EN=100;JD_RED_2=1;JD_BLUE_2=1;JD_GREEN_2=1
		[]string{"VI_HC", "VI_EN", "JD_RED_2", "JD_BLUE_2", "JD_GREEN_2"},
		[]uint32{188, 100, 1, 1, 1})

	//普通礼包	52-1	公会礼包	10000	钻石*288、大乔碎片*10、小乔碎片*10、中级祝福石*1
	// VI_HC=288;VI_GENERAL_DQIAO=10;VI_GENERAL_XQIAO=10;SI_SB_2=1	2016年5月16日	2016年6月30日
	AddDynamoDB(52, 1, 10000, false, false, 1, 1467302399, "公会礼包",
		// VI_HC=288;VI_GENERAL_DQIAO=10;VI_GENERAL_XQIAO=10;SI_SB_2=1
		[]string{"VI_HC", "VI_GENERAL_DQIAO", "VI_GENERAL_XQIAO", "SI_SB_2"},
		[]uint32{288, 10, 10, 1})

	//普通礼包	53-1	会长礼包	1000	钻石*388、郭嘉碎片*10、贾诩碎片*10、中级祝福石*3
	// VI_HC=388;VI_GENERAL_GJIA=10;VI_GENERAL_JXU=10;SI_SB_2=3	2016年5月16日	2016年6月30日
	AddDynamoDB(53, 1, 1000, false, false, 1, 1467302399, "会长礼包",
		// VI_HC=388;VI_GENERAL_GJIA=10;VI_GENERAL_JXU=10;SI_SB_2=3
		[]string{"VI_HC", "VI_GENERAL_GJIA", "VI_GENERAL_JXU", "SI_SB_2"},
		[]uint32{388, 10, 10, 3})

	//普通礼包	54-1	豪华礼包	10000	钻石*288、金币*18888、3级紫龙玉*1、3级黄龙玉*1、中级祝福石*1
	// VI_HC=288;VI_SC=18888;JD_PURPLE_3=1;JD_YELLOW_3=1;SI_SB_2=1	2016年5月16日	2016年6月30日
	AddDynamoDB(54, 1, 10000, false, false, 1, 1467302399, "豪华礼包",
		// VI_HC=288;VI_SC=18888;JD_PURPLE_3=1;JD_YELLOW_3=1;SI_SB_2=1
		[]string{"VI_HC", "VI_SC", "JD_PURPLE_3", "JD_YELLOW_3", "SI_SB_2"},
		[]uint32{288, 18888, 1, 1, 1})

	//普通礼包	55-1	520礼包	15000	520钻，1314天命	VI_HC=520;VI_DC=1314	2016年5月16日	2016年6月30日
	AddDynamoDB(55, 1, 15000, false, false, 1, 1467302399, "520礼包",
		// VI_HC=520;VI_DC=1314
		[]string{"VI_HC", "VI_DC"},
		[]uint32{520, 1314})

	//普通礼包	56-1	登陆活动礼包	15000	钻石*188，3级红龙玉*1，天命*500
	// VI_HC=188;JD_RED_3=1;VI_DC=500	2016年5月16日	2016年6月30日
	AddDynamoDB(56, 1, 15000, false, false, 1, 1467302399, "登陆活动礼包",
		// VI_HC=188;JD_RED_3=1;VI_DC=500
		[]string{"VI_HC", "JD_RED_3", "VI_DC"},
		[]uint32{188, 1, 500})

	//普通礼包	57-1	召集活动礼包	15000	钻石*288，4级紫龙玉*1，金币*20000
	// VI_HC=288;JD_PURPLE_4=1;VI_SC=20000	2016年5月16日	2016年6月30日
	AddDynamoDB(57, 1, 15000, false, false, 1, 1467302399, "召集活动礼包",
		// VI_HC=288;JD_PURPLE_4=1;VI_SC=20000
		[]string{"VI_HC", "JD_PURPLE_4", "VI_SC"},
		[]uint32{288, 1, 20000})

	//普通礼包	58-1	论坛至尊礼包1	2000	钻石*108，精铁*2000，2级红龙玉*1，2级绿龙玉*1
	// VI_HC=108;VI_FI=2000;JD_RED_2=1;JD_GREEN_2=1	2016年5月16日	2016年6月30日
	AddDynamoDB(58, 1, 2000, false, false, 1, 1467302399, "论坛至尊礼包1",
		// VI_HC=108;VI_FI=2000;JD_RED_2=1;JD_GREEN_2=1
		[]string{"VI_HC", "VI_FI", "JD_RED_2", "JD_GREEN_2"},
		[]uint32{108, 2000, 1, 1})

	//普通礼包	59-1	论坛至尊礼包2	1500	钻石*188，金币*10000，3级绿龙玉*1，3级蓝龙玉*1
	// VI_HC=188;VI_SC=10000;JD_GREEN_3=1;JD_BLUE_3=1	2016年5月16日	2016年6月30日
	AddDynamoDB(59, 1, 1500, false, false, 1, 1467302399, "论坛至尊礼包2",
		// VI_HC=188;VI_SC=10000;JD_GREEN_3=1;JD_BLUE_3=1
		[]string{"VI_HC", "VI_SC", "JD_GREEN_3", "JD_BLUE_3"},
		[]uint32{188, 10000, 1, 1})

	//普通礼包	60-1	论坛至尊礼包3	1000	钻石*288，天命*1000，4级紫龙玉*1
	// VI_HC=288;VI_DC=1000;JD_PURPLE_4=1	2016年5月16日	2016年6月30日
	AddDynamoDB(60, 1, 1000, false, false, 1, 1467302399, "论坛至尊礼包3",
		// VI_HC=288;VI_DC=1000;JD_PURPLE_4=1
		[]string{"VI_HC", "VI_DC", "JD_PURPLE_4"},
		[]uint32{288, 1000, 1})

	//普通礼包	61-1	冲榜礼包1	30	钻石*800，曹操碎片*10，吕布碎片*10，诸葛碎片*10
	// VI_HC=800;VI_GENERAL_CCAO=10;VI_GENERAL_LBU=10;VI_GENERAL_ZGLIANG=10	2016年5月16日	2016年6月30日
	AddDynamoDB(61, 1, 30, false, false, 1, 1467302399, "冲榜礼包1",
		// VI_HC=800;VI_GENERAL_CCAO=10;VI_GENERAL_LBU=10
		[]string{"VI_HC", "VI_GENERAL_CCAO", "VI_GENERAL_LBU"},
		[]uint32{800, 10, 10})

	//普通礼包	62-1	冲榜礼包2	30	钻石*1000，曹操碎片*20，吕布碎片*20，诸葛碎片*20
	// VI_HC=1000;VI_GENERAL_CCAO=20;VI_GENERAL_LBU=20;VI_GENERAL_ZGLIANG=20	2016年5月16日	2016年6月30日
	AddDynamoDB(62, 1, 30, false, false, 1, 1467302399, "冲榜礼包2",
		// VI_HC=1000;VI_GENERAL_CCAO=20;VI_GENERAL_LBU=20;VI_GENERAL_ZGLIANG=20
		[]string{"VI_HC", "VI_GENERAL_CCAO", "VI_GENERAL_LBU", "VI_GENERAL_ZGLIANG"},
		[]uint32{1000, 20, 20, 20})

	//普通礼包	63-1	冲榜礼包3	30	钻石*2000，曹操碎片*30，吕布碎片*30，诸葛碎片*30
	// VI_HC=2000;VI_GENERAL_CCAO=30;VI_GENERAL_LBU=30;VI_GENERAL_ZGLIANG=30	2016年5月16日	2016年6月30日
	AddDynamoDB(63, 1, 30, false, false, 1, 1467302399, "冲榜礼包3",
		// VI_HC=2000;VI_GENERAL_CCAO=30;VI_GENERAL_LBU=30;VI_GENERAL_ZGLIANG=30
		[]string{"VI_HC", "VI_GENERAL_CCAO", "VI_GENERAL_LBU", "VI_GENERAL_ZGLIANG"},
		[]uint32{2000, 30, 30, 30})

	//普通礼包	64-1	连续登录3天礼包	5000	钻石*688，天命*2000，3级黄龙玉*1，3级紫龙玉*1，3级青龙玉*1
	// VI_HC=688;VI_DC=2000;JD_YELLOW_3=1;JD_PURPLE_3=1;JD_CHING_3=1	2016年5月16日	2016年6月30日
	AddDynamoDB(64, 1, 5000, false, false, 1, 1467302399, "连续登录3天礼包",
		// VI_HC=688;VI_DC=2000;JD_YELLOW_3=1;JD_PURPLE_3=1;JD_CHING_3=1
		[]string{"VI_HC", "VI_DC", "JD_YELLOW_3", "JD_PURPLE_3", "JD_CHING_3"},
		[]uint32{688, 2000, 1, 1, 1})

	//普通礼包	65-1	贴吧关注礼包	2000	钻石*88，体力*50，天命*200，精铁*1000
	// VI_HC=88;VI_EN=50;VI_DC=200;VI_FI=1000	2016年5月16日	2016年6月30日
	AddDynamoDB(65, 1, 2000, false, false, 1, 1467302399, "贴吧关注礼包",
		// VI_HC=88;VI_EN=50;VI_DC=200;VI_FI=1000
		[]string{"VI_HC", "VI_EN", "VI_DC", "VI_FI"},
		[]uint32{88, 50, 200, 1000})

	//普通礼包	66-1	论坛专属礼包	2000	钻石*88，体力*50，天命*200，精铁*1000
	// VI_HC=88;VI_EN=50;VI_DC=200;VI_FI=1000	2016年5月16日	2016年6月30日
	AddDynamoDB(66, 1, 2000, false, false, 1, 1467302399, "论坛专属礼包",
		// VI_HC=88;VI_EN=50;VI_DC=200;VI_FI=1000
		[]string{"VI_HC", "VI_EN", "VI_DC", "VI_FI"},
		[]uint32{88, 50, 200, 1000})

	//普通礼包	67-1	论坛活跃礼包	2000	钻石*88，1级红龙玉*1，1级蓝龙玉*1，1级绿龙玉*1，金币*10000
	// VI_HC=88;JD_RED_1=1;JD_BLUE_1=1;JD_GREEN_1=1;VI_SC=10000	2016年5月16日	2016年6月30日
	AddDynamoDB(67, 1, 2000, false, false, 1, 1467302399, "论坛活跃礼包",
		// VI_HC=88;JD_RED_1=1;JD_BLUE_1=1;JD_GREEN_1=1;VI_SC=10000
		[]string{"VI_HC", "JD_RED_1", "JD_BLUE_1", "JD_GREEN_1", "VI_SC"},
		[]uint32{88, 1, 1, 1, 10000})

	//普通礼包	68-1	Q群礼包	2000	钻石*200	VI_HC=200;	2016年5月16日	2016年6月30日
	AddDynamoDB(68, 1, 2000, false, false, 1, 1467302399, "Q群礼包",
		// VI_HC=200
		[]string{"VI_HC"},
		[]uint32{200})

	//通用礼包	69-1	微信礼包	1	钻石*200，金币*8888	VI_HC=200;VI_SC=8888	2016年5月16日	2016年6月30日
	AddDynamoDB(69, 1, 1, true, false, 1, 1467302399, "微信礼包",
		// VI_HC=200;VI_SC=8888
		[]string{"VI_HC", "VI_SC"},
		[]uint32{200, 8888})

	//普通礼包	70-1	官网礼包	2000	钻石*200，金币*20000，天命*2000，精铁*4000
	// VI_HC=200;VI_SC=20000;VI_DC=2000;VI_FI=4000	2016年5月16日	2016年6月30日
	AddDynamoDB(70, 1, 2000, false, false, 1, 1467302399, "官网礼包",
		// VI_HC=200;VI_SC=20000;VI_DC=2000;VI_FI=4000
		[]string{"VI_HC", "VI_SC", "VI_DC", "VI_FI"},
		[]uint32{200, 20000, 2000, 4000})

	//普通礼包	71-1	表情包大战礼包	2000	钻石*500	VI_HC=500;	2016年5月16日	2016年6月30日
	AddDynamoDB(71, 1, 2000, false, false, 1, 1467302399, "表情包大战礼包",
		// VI_HC=500;
		[]string{"VI_HC"},
		[]uint32{500})

	//普通礼包	72-1	脑洞大开礼包	2000	钻石*500	VI_HC=500;	2016年5月16日	2016年6月30日
	AddDynamoDB(72, 1, 2000, false, false, 1, 1467302399, "脑洞大开礼包",
		// VI_HC=500;
		[]string{"VI_HC"},
		[]uint32{500})

	//普通礼包	73-1	观看直播礼包	2000	钻石*1000	VI_HC=1000;	2016年5月16日	2016年6月30日
	AddDynamoDB(73, 1, 2000, false, false, 1, 1467302399, "观看直播礼包",
		// VI_HC=1000;
		[]string{"VI_HC"},
		[]uint32{1000})

	//普通礼包	74-1	战力排行礼包	2000	钻石*1000	VI_HC=1000;	2016年5月16日	2016年6月30日
	AddDynamoDB(74, 1, 2000, false, false, 1, 1467302399, "战力排行礼包",
		// VI_HC=1000;
		[]string{"VI_HC"},
		[]uint32{1000})

	//普通礼包	75-1	限时问答礼包	2000	钻石*200	VI_HC=200;	2016年5月16日	2016年6月30日
	AddDynamoDB(75, 1, 2000, false, false, 1, 1467302399, "限时问答礼包",
		// VI_HC=200;
		[]string{"VI_HC"},
		[]uint32{200})

	//普通礼包	76-1	61QQ群礼包	2000	钻石*616，体力*100，金币*6161
	// VI_HC=616;VI_EN=100;VI_HC=6161	2016年5月16日	2016年6月30日
	AddDynamoDB(76, 1, 2000, false, false, 1, 1467302399, "61QQ群礼包",
		// VI_HC=616;VI_EN=100;VI_SC=6161
		[]string{"VI_HC", "VI_EN", "VI_SC"},
		[]uint32{616, 100, 6161})

	//通用礼包	77-1	61微信礼包	1	钻石*616，体力*100，精铁*6161
	// VI_HC=616;VI_EN=100;VI_FI=6161	2016年5月16日	2016年6月30日
	AddDynamoDB(77, 1, 1, true, false, 1, 1467302399, "61微信礼包",
		// VI_HC=616;VI_EN=100;VI_FI=6161
		[]string{"VI_HC", "VI_EN", "VI_FI"},
		[]uint32{616, 100, 6161})

	//普通礼包	78-1	61贴吧礼包	2000	钻石*616，体力*100，巡查令牌*6
	// VI_HC=616;VI_EN=100;VI_GT=6	2016年5月16日	2016年6月30日
	AddDynamoDB(78, 1, 2000, false, false, 1, 1467302399, "61贴吧礼包",
		// VI_HC=616;VI_EN=100;VI_GT=6
		[]string{"VI_HC", "VI_EN", "VI_GT"},
		[]uint32{616, 100, 6})

	///普通礼包	79-1	Q群连签礼包1	2000	钻石*500	VI_HC=500;	2016年5月16日	2016年6月30日
	AddDynamoDB(79, 1, 2000, false, false, 1, 1467302399, "Q群连签礼包1",
		// VI_HC=500;
		[]string{"VI_HC"},
		[]uint32{500})

	//普礼包	80-1	Q群连签礼包2	2000	钻石*1000	VI_HC=1000;	2016年5月16日	2016年6月30日
	AddDynamoDB(80, 1, 2000, false, false, 1, 1467302399, "Q群连签礼包2",
		// VI_HC=2000;SI_SB_2=10;VI_DC=2000
		[]string{"VI_HC"},
		[]uint32{1000})

	//普通礼包	81-1	Q群连签礼包3	2000	钻石*2000	VI_HC=2000;	2016年5月16日	2016年6月30日
	AddDynamoDB(81, 1, 2000, false, false, 1, 1467302399, "Q群连签礼包3",
		// VI_HC=2000
		[]string{"VI_HC"},
		[]uint32{2000})

	//普通礼包	82-1	预约测试礼包	5000	钻石*500	VI_HC=500;	2016年5月16日	2016年6月30日
	AddDynamoDB(82, 1, 5000, false, false, 1, 1467302399, "预约测试礼包",
		// VI_HC=500
		[]string{"VI_HC"},
		[]uint32{500})

	//普通礼包	83-1	冲级测试礼包	1000	钻石*2000	VI_HC=2000;	2016年5月16日	2016年6月30日
	AddDynamoDB(83, 1, 1000, false, false, 1, 1467302399, "冲级测试礼包",
		// VI_HC=2000
		[]string{"VI_HC"},
		[]uint32{2000})


	// 通用码 客服微信礼包 钻石*288  VI_HC=288
	// VI_HC=288
	AddDynamoDB(84, 1, 1, true, false, 1, 1467302399, "客服微信礼包",
		// VI_HC=288
		[]string{"VI_HC"},
		[]uint32{288})
	logs.Close()
}
*/

func init() {
	SetTimeLocal("Asia/Shanghai")
}

var local *time.Location

func SetTimeLocal(local_str string) {
	l, err := time.LoadLocation(local_str)
	if err != nil {
		panic(err)
	}
	local = l
}

var CommonConfig CommonCfg

func loadConfig(confStr string) {
	var common_cfg struct{ CommonConfig CommonCfg }
	cfgApp := config.NewConfigToml(confStr, &common_cfg)

	CommonConfig = common_cfg.CommonConfig

	if cfgApp == nil {
		logs.Critical("Config Read Error\n")
		logs.Close()
		os.Exit(1)
	}
}
func main() {
	loadConfig("config.toml")
	logs.Warn("Config: %v", CommonConfig)
	code_file := flag.String("f", "Code.xlsx", "Code.xlsx")
	flag.Parse()

	defer logs.Close()
	excelFileName := *code_file
	fileSuffix := path.Ext(excelFileName)
	filenameOnly := strings.TrimSuffix(excelFileName, fileSuffix)
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		panic(err)
	}
	codeSheet, ok := xlFile.Sheet["Code"]
	if !ok {
		logs.Error("Code Sheet Cannot Found!")
		os.Exit(1)
		return
	}

	CodeDatas := []CodeData{}

	for ridx, row := range codeSheet.Rows {
		if ridx >= 4 && len(row.Cells) > 5 {
			data := make([]string, 0, 64)
			for cidx, cell := range row.Cells {
				s, err := cell.String()
				if err != nil {
					logs.Error("cell err by %d,%d in %s", ridx, cidx, err.Error())
					os.Exit(1)
					return
				}
				data = append(data, s)
			}
			logs.Info("row %2d -- %v", ridx, data)

			if data[0] != "" {
				c := CodeData{}
				c.FromXlsx(row.Cells[:])
				logs.Info("data %v", c)
				CodeDatas = append(CodeDatas, c)
			}
		}
	}

	all := make([]byte, 0, 320*10000)
	for _, data := range CodeDatas {
		var n []string
		if CommonConfig.IsToDynamoDB {
			n = AddDynamoDB(
				data.BID,
				data.TID,
				data.Max,
				data.IsLimit,
				false,
				data.TimeBegin,
				data.TimeEnd,
				data.Title,
				data.ItemIDs[:], data.ItemCounts[:])
		} else {
			n = AddMongoDB(
				data.BID,
				data.TID,
				data.Max,
				data.IsLimit,
				false,
				data.TimeBegin,
				data.TimeEnd,
				data.Title,
				data.ItemIDs[:], data.ItemCounts[:])
		}

		for _, ns := range n {
			all = append(all, []byte(ns)...)
		}
	}
	resf := fmt.Sprintf("%s.csv", filenameOnly)
	ioutil.WriteFile(resf, all, 0666)
}

type CodeData struct {
	BID        int64
	TID        int64
	Max        int64
	IsLimit    bool
	TimeBegin  int64
	TimeEnd    int64
	Title      string
	ItemIDs    []string
	ItemCounts []uint32
}

func toTime(s string, err error) (int64, error) {
	if err != nil {
		return 0, err
	} else {
		t, err := time.ParseInLocation("2006\\-01\\-02", s, local)
		if err != nil {
			return 0, err
		}
		return t.Unix(), nil
	}
}

func (c *CodeData) FromXlsx(cells []*xlsx.Cell) {
	var err error

	if c.BID, err = cells[0].Int64(); err != nil {
		panic(err)
	}
	if c.TID, err = cells[1].Int64(); err != nil {
		panic(err)
	}
	if c.Title, err = cells[2].String(); err != nil {
		panic(err)
	}
	if c.TimeBegin, err = toTime(cells[3].String()); err != nil {
		panic(err)
	}
	if c.TimeEnd, err = toTime(cells[4].String()); err != nil {
		panic(err)
	}
	if class, err := cells[5].Int(); err != nil {
		panic(err)
	} else {
		if class == 2 {
			c.IsLimit = true
		}
	}
	if c.Max, err = cells[6].Int64(); err != nil {
		panic(err)
	}

	c.ItemIDs = make([]string, 0, 16)
	c.ItemCounts = make([]uint32, 0, 16)
	for i := 7; i < 19 && i < len(cells); i += 2 {
		var (
			id    string
			count int
			err   error
		)
		if id, err = cells[i].String(); err != nil {
			panic(err)
		}
		if id == "" {
			continue
		}
		if count, err = cells[i+1].Int(); err != nil {
			panic(err)
		}
		if id != "" {
			if count <= 0 {
				panic(errors.New("count <= 0"))
			}
			c.ItemIDs = append(c.ItemIDs, id)
			c.ItemCounts = append(c.ItemCounts, uint32(count))
		}
	}
}
