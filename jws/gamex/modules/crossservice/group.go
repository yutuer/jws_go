package crossservice

var listGroupHandle = []func(uint) uint32{}

//RegGroupHandle ..
func RegGroupHandle(f func(uint) uint32) {
	listGroupHandle = append(listGroupHandle, f)
}

func getGroupIDs(sid uint) []uint32 {
	l := []uint32{}
	for _, f := range listGroupHandle {
		l = append(l, f(sid))
	}
	return l
}
