package gamedata

type givesData struct {
	Data   CostData
	Ids    []string
	Counts []uint32
}

func (g *givesData) GetData() *CostData {
	return &g.Data
}

func (g *givesData) AddItem(item_id string, count uint32) {
	if count == 0 {
		return
	}

	if g.Ids == nil {
		g.Ids = make([]string, 0, 8)
		g.Counts = make([]uint32, 0, 8)
	}
	g.Data.AddItem(item_id, count)
	if item_id == VI_Hc_Buy || item_id == VI_Hc_Compensate || item_id == VI_Hc_Give {
		item_id = VI_Hc
	}
	g.Ids = append(g.Ids, item_id)
	g.Counts = append(g.Counts, count)
}
