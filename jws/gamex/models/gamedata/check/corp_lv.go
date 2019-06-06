package check

/*
	对所有以战队等级为索引的数据进行检查
	保证在可用等级范围之内每个等级都有正确的返回, 否之Panic
*/

type AssertCorpLvFunc func(lv int) bool

var maxCorpLv int = -1

func InitByCorpData(max int) {
	maxCorpLv = max
}

func ByCorpLv(dataName string, assertFunc AssertCorpLvFunc) {
	if assertFunc == nil {
		panicf("ByCorpLv Err No Func By %s", dataName)
	}

	if maxCorpLv < 0 {
		panicf("ByCorpLv Err No Corp Max Lv")
	}

	for i := 1; i < maxCorpLv; i++ {
		if !assertFunc(i) {
			panicf("ByCorpLv Err %d By %s", i, dataName)
		}
	}
}

// 对于一个数组,从1到N都又有效地值
func ByArrat1ToN(dataName string, n int, assertFunc AssertCorpLvFunc) {
	if assertFunc == nil {
		panicf("ByArrat1ToN Err No Func By %s", dataName)
	}

	if maxCorpLv < 0 {
		panicf("ByArrat1ToN Err No Corp Max Lv")
	}

	for i := 1; i < n; i++ {
		if !assertFunc(i) {
			panicf("ByArrat1ToN Err %d By %s", i, dataName)
		}
	}
}
