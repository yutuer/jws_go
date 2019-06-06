package gamedata

type ICostData2Client interface {
	IsNotEmpty() bool
	AddItem2Client(itemId string, count uint32)
	AddItemWithData2Client(itemId string, data BagItemData, count uint32)
	AddOther2Client(o ICostData2Client)
	GetItem(i int) (bool, string, uint32, string, BagItemData)
	Len() int
}

type CostData2Client struct {
	Item2Client       []string
	Count2Client      []uint32
	Data2Client       []string
	DataStruct2Client []BagItemData
	AllSc0            uint32
	AllCorpXp         uint32
	EachHeroXp        uint32
}

func (p *CostData2Client) GetItem(i int) (bool, string, uint32, string, BagItemData) {
	if i < 0 || i > p.Len() {
		return false, "", 0, "", BagItemData{}
	}
	return true,
		p.Item2Client[i],
		p.Count2Client[i],
		p.Data2Client[i],
		p.DataStruct2Client[i]
}

func (p *CostData2Client) Len() int {
	return len(p.Item2Client)
}

func (p *CostData2Client) Init2Client(cap int) {
	p.tryInitData2Client(cap)
}

func (p *CostData2Client) tryInitData2Client(cap int) {
	if p.Item2Client == nil {
		p.Item2Client = make([]string, 0, cap)
		p.Count2Client = make([]uint32, 0, cap)
		p.Data2Client = make([]string, 0, cap)
		p.DataStruct2Client = make([]BagItemData, 0, cap)
	}
}

func (p *CostData2Client) IsNotEmpty() bool {
	for _, c := range p.Count2Client {
		if c > 0 {
			return true
		}
	}
	return false
}

func (p *CostData2Client) AddItem2Client(itemId string, count uint32) {
	p.AddItemWithData2Client(itemId, BagItemData{}, count)
}

func (p *CostData2Client) AddItemWithData2Client(itemId string, data BagItemData, count uint32) {
	p.tryInitData2Client(32)
	p.Item2Client = append(p.Item2Client, itemId)
	p.Count2Client = append(p.Count2Client, count)
	p.Data2Client = append(p.Data2Client, data.ToDataStr())
	p.DataStruct2Client = append(p.DataStruct2Client, data)

	switch itemId {
	case VI_Sc0:
		p.AllSc0 += count
	case VI_CorpXP:
		p.AllCorpXp += count
	case VI_XP:
		p.EachHeroXp += count
	}
}

func (p *CostData2Client) AddOther2Client(o ICostData2Client) {
	if o == nil {
		return
	}
	for i := 0; i < o.Len(); i++ {
		ok, it, c, _, ds := o.GetItem(i)
		if ok {
			p.AddItemWithData2Client(it, ds, c)
		}
	}
}
