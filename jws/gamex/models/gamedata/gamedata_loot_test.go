package gamedata

//
//import (
//	"testing"
//)
//
//func initTest() {
//	ts := make(map[string]lootTemplate)
//	igs := make(map[string]lootItemGroup)
//	ts["t1"] = lootTemplate{[]lootTItem{lootTItem{"tt1", 30}, lootTItem{"tt2", 40}}}
//
//	igs["tt1"] = lootItemGroup{[]lootIGItem{lootIGItem{ItemIdx_t(11), 20, 11, 22}, lootIGItem{ItemIdx_t(12), 50, 71, 82}}}
//	igs["tt2"] = lootItemGroup{[]lootIGItem{lootIGItem{ItemIdx_t(21), 50, 11, 22}, lootIGItem{ItemIdx_t(22), 50, 71, 82}}}
//
//	loadIGRandArrayImp(igs)
//	loadTRandArrayImp(ts)
//}
//
//func TestRandArray(t *testing.T) {
//	initTest()
//
//	for i, info := range gdLootItemGroupRandArray["tt1"] {
//		t.Logf(" %d - %d %s %s", i, info.GetID(), info, &gdLootItemGroupRandArray["tt1"][i])
//	}
//}
//
//func TestItemGroupRand(t *testing.T) {
//	initTest()
//
//	info := make(map[ItemIdx_t]int32)
//	for j := 0; j < 100000; j++ {
//		_, w := LootItemGroupRandSelect("tt2")
//		var id ItemIdx_t = ItemIdx_t(0)
//		if w != nil {
//			id = w.GetID()
//		}
//		_, ok := info[id]
//		if ok {
//			info[id]++
//		} else {
//			info[id] = 1
//		}
//
//	}
//	t.Logf("%s", info)
//}
//
//func TestTemplateRand(t *testing.T) {
//	initTest()
//
//	info := make(map[string]int32)
//	for j := 0; j < 100000; j++ {
//		_, w := LootTemplateRandSelect(0, "t1")
//		id := ""
//		if w != nil {
//			id = w.GetID()
//		}
//		_, ok := info[id]
//		if ok {
//			info[id]++
//		} else {
//			info[id] = 1
//		}
//
//	}
//	t.Logf("%s", info)
//}
