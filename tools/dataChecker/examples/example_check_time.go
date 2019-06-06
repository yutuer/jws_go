package examples

import (
	"time"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

// HotActivityTimeConvert 将"yyyymmdd_hh:mm"转换成timestamp
func HotActivityTimeConvert(hatTime string) int64 {
	timeForm := "20060102_15:04"
	t, err := time.Parse(timeForm, hatTime)
	if err != nil {
		return 0
	}

	return t.Unix()
}

// CheckSubAcitivityTimeRange 检查子活动时间范围是否在父活动之内；如果时间类型不同，会直接报错
func CheckSubAcitivityTimeRange(p *ProtobufGen.HotActivityTime, c []*ProtobufGen.HotActivityTime) (wrong []*ProtobufGen.HotActivityTime, ok bool) {
	wrong = make([]*ProtobufGen.HotActivityTime, 0)
	var cBegin, cEnd, pBegin, pEnd int64

	// Parent Activity
	if p.GetTimeType() == 1 {
		pBegin = HotActivityTimeConvert(p.GetStartTime())
		pEnd = HotActivityTimeConvert(p.GetEndTime())
	}

	if pBegin > pEnd {
		wrong = append(wrong, p)
		ok = false
		return
	}

	// Children Activities
	for _, cData := range c {
		if cData.GetTimeType() != p.GetTimeType() {
			wrong = append(wrong, cData)
			continue
		}

		if p.GetTimeType() == 1 {
			cBegin = HotActivityTimeConvert(cData.GetStartTime())
			cEnd = HotActivityTimeConvert(cData.GetEndTime())
		}

		if cBegin < pBegin || cEnd > pEnd {
			wrong = append(wrong, cData)
		}
	}

	ok = len(wrong) == 0

	return
}

// CheckHotActivityTimeRange 检查每组数据时间是否有重合区间，返回具有重合时间的集合
func CheckHotActivityTimeRange(d []*ProtobufGen.HotActivityTime) ([][]*ProtobufGen.HotActivityTime, bool) {
	n := len(d)
	if n < 2 {
		return nil, true
	}

	// 记录有重合的组
	overlapGroups := make([][]*ProtobufGen.HotActivityTime, 0)

	for i := 0; i < n-1; i++ {
		h := d[i]
		overlap := []*ProtobufGen.HotActivityTime{h}
		begin := HotActivityTimeConvert(*h.StartTime)
		end := HotActivityTimeConvert(*h.EndTime)

		for j := i + 1; j < n; j++ {
			curH := d[j]
			curBegin := HotActivityTimeConvert(*curH.StartTime)

			if curBegin > end {
				i = j - 1
				break
			} else if curBegin > begin {
				overlap = append(overlap, curH)
			}
		}

		if len(overlap) > 1 {
			overlapGroups = append(overlapGroups, overlap)
		}
	}

	return overlapGroups, len(overlapGroups) == 0
}
