package stage_star

const (
	Star000 = iota
	Star001
	Star010
	Star011
	Star100
	Star101
	Star110
	Star111
	StarCount
)

func GetStarCount(i int32) int32 {
	switch i {
	case Star000:
		return 0
	case Star001:
		return 1
	case Star010:
		return 1
	case Star011:
		return 2
	case Star100:
		return 1
	case Star101:
		return 2
	case Star110:
		return 2
	case Star111:
		return 3
	default:
		return 0
	}
}

func AddStar(o, n int32) int32 {
	if GetStarCount(n) >= GetStarCount(o) {
		return n
	}
	//logs.Trace("on %b %b", o, n)
	return o
}
