package main

//Gen 批量生成兑换码
func Gen(batchID, typID, count int64, isNoLimit, isRandom bool) []string {
	if count >= 100000 {
		return nil
	}
	res := make([]string, 0, count)
	var c int64
	for ; c < count; c++ {
		id := mkGiftCodeNum(batchID, typID, c, isNoLimit, isRandom)
		res = append(res, toNumInStr(id))
	}
	return res
}

type RedeemCodeValues struct {
	BatchID string   `json:"bid"`
	GroupID string   `json:"gid"`
	ItemIDs []string `json:"items"`
	Counts  []uint32 `json:"counts"`
	Title   string   `json:"title"`
}

/*
	DoneBy string  用户gid:sid:uuid
	Bind   string  gid, gid:sid,
	State  int     New|Done|OutdatedDone
	Begin  int     开始时间
	End    int     结束时间
	Value  string  Json内容 {
		BatchID string   `json:"bid"`
		GroupID string   `json:"gid"`
		ItemIDs []string `json:"items"`
		Counts  []uint32 `json:"counts"`
	}
*/
