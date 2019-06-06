package jade

import "vcs.taiyouxi.net/jws/gamex/protogen"

// 为了支持宝石的自动装备，将宝石放在堆heap中，每次取出等级最大，经验最大的宝石，进行镶嵌

type JadeItem struct {
	Id      uint32
	Cfg     *ProtobufGen.Item
	JadeLvl int32
	Exp     uint32
}

type JadeHeap []*JadeItem

func (pq JadeHeap) Len() int { return len(pq) }

func (pq JadeHeap) Less(i, j int) bool {
	if pq[i].JadeLvl != pq[j].JadeLvl {
		return pq[i].JadeLvl > pq[j].JadeLvl
	} else {
		return pq[i].Exp > pq[j].Exp
	}
}

func (pq JadeHeap) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *JadeHeap) Push(x interface{}) {
	item := x.(*JadeItem)
	*pq = append(*pq, item)
}

func (pq *JadeHeap) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}
