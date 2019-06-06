package gvg

type GVGSortItem struct {
	IntKey int
	StrKey string
	StrVal string
	IntVal int64
}

type GVGSortArray []*GVGSortItem

func (gsa GVGSortArray) Len() int {
	return len(gsa)
}

func (gsa GVGSortArray) Less(i, j int) bool {

	if gsa[i].StrVal == "" && gsa[j].StrVal == "" {
		if convertTrueScore(gsa[i].IntVal) == convertTrueScore(gsa[j].IntVal) {
			return gsa[i].IntVal < gsa[j].IntVal
		}
		if gsa[i].IntVal == gsa[j].IntVal {
			return gsa[i].StrKey > gsa[j].StrKey
		}
		return gsa[i].IntVal > gsa[j].IntVal
	} else if gsa[i].IntVal == 0 && gsa[j].IntVal == 0 {
		return gsa[i].StrVal > gsa[j].StrVal
	}
	return false
}

func (gsa GVGSortArray) Swap(i, j int) {
	tmp := gsa[i]
	gsa[i] = gsa[j]
	gsa[j] = tmp
}
