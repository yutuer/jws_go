package gamedata

type PriceData struct {
	PriceTyp   string
	PriceCount uint32
	PriceData  string
	Cost       CostData
}

func (p *PriceData) AddItem(itemId string, count uint32) {
	p.PriceTyp = itemId
	p.PriceCount = count
	p.Cost.AddItem(itemId, count)
}

func (p *PriceData) Gives() *CostData {
	return &p.Cost
}

func (p *PriceData) AddItemWithData(itemId string, data BagItemData, count uint32) {
	p.PriceTyp = itemId
	p.PriceCount = count
	p.PriceData = data.ToDataStr()
	p.Cost.AddItemWithData(itemId, data, count)
}

func NewPriceData(itemId string, count uint32) *PriceData {
	res := &PriceData{}
	res.PriceTyp = itemId
	res.PriceCount = count
	res.Cost.AddItem(itemId, count)
	return res
}

type PriceDatas struct {
	CostData2Client
	Cost CostData
}

func NewPriceDatas(cap int) PriceDatas {
	res := PriceDatas{}
	res.Init2Client(cap)
	return res
}

func (p *PriceDatas) Gives() *CostData {
	return &p.Cost
}

func (p *PriceDatas) AddItem(itemId string, count uint32) {
	if count == 0 {
		return
	}
	is_2_sc, itemID, countc := IsItemTreasurebox(itemId)
	if is_2_sc {
		//logs.Warn("AddItemWithData %v-%v %v-%v", itemId, count, itemID, countc)
		p.AddItem2Client(itemID, countc*count)
		p.Cost.AddItem(itemID, countc*count)
	} else {
		p.AddItem2Client(itemId, count)
		p.Cost.AddItem(itemId, count)
	}
}

func (p *PriceDatas) AddItemWithData(itemId string, data BagItemData, count uint32) {
	if count == 0 {
		return
	}
	// Virtual Item
	is_2_sc, itemID, countc := IsItemTreasurebox(itemId)
	if is_2_sc {
		//logs.Warn("AddItemWithData %v-%v %v-%v", itemId, count, itemID, countc)
		p.AddItemWithData2Client(itemID, data, countc*count)
		p.Cost.AddItemWithData(itemID, data, countc*count)
	} else {
		p.AddItemWithData2Client(itemId, data, count)
		p.Cost.AddItemWithData(itemId, data, count)
	}
}

func (p *PriceDatas) AddOther(o *PriceDatas) {
	if o == nil {
		return
	}
	p.AddOther2Client(o)
	p.Cost.AddGroup(&(o.Cost))
}

type PriceDataSet struct {
	Datas []PriceDatas
}

func NewPriceDataSet(cap int) *PriceDataSet {
	n := &PriceDataSet{}
	return n.Init(cap)
}

func (p *PriceDataSet) Init(cap int) *PriceDataSet {
	p.Datas = make([]PriceDatas, 0, cap)
	return p
}

func (p *PriceDataSet) AppendDatas(d []PriceDatas) *PriceDataSet {
	for _, ds := range d {
		p.Datas = append(p.Datas, ds)
	}
	return p
}

func (p *PriceDataSet) AppendData(d PriceDatas) *PriceDataSet {
	p.Datas = append(p.Datas, d)
	return p
}

func (p *PriceDataSet) AppendOther(d *PriceDataSet) *PriceDataSet {
	return p.AppendDatas(d.Datas)
}

func (p *PriceDataSet) Mk2One() *PriceDatas {
	res := NewPriceDatas(len(p.Datas)*3 + 1)
	for i := 0; i < len(p.Datas); i++ {
		res.AddOther(&(p.Datas[i]))
	}
	return &res
}
