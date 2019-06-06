package hot_activities

import (
	"fmt"
	"sort"
	"strconv"

	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/tools/dataChecker/utils"
)

// CheckSubAcitivityTimeRange 检查子活动时间范围是否在父活动之内
// 如果时间类型不同，会直接报错
func (haData *HotActivitiesData) CheckSubAcitivityTimeRange() (wrong []*ProtobufGen.HotActivityTime, ok bool) {
	wrong = make([]*ProtobufGen.HotActivityTime, 0)
	var cBegin, cEnd, pBegin, pEnd int64

	// Parent Activity
	for pid, slice := range haData.subActivities {
		pData := dataMap[pid]
		if pData.GetTimeType() == 1 {
			pBegin = HotActivityTimeConvert(pData.GetStartTime())
			pEnd = HotActivityTimeConvert(pData.GetEndTime())
		}

		// Children Activities
		for _, cData := range slice {
			// 时间类型不一致，服务器分组id不一致，都报错
			if cData.GetTimeType() != pData.GetTimeType() || cData.GetServerGroupID() != pData.GetServerGroupID() {
				wrong = append(wrong, cData)
				continue
			}

			if pData.GetTimeType() == 1 {
				cBegin = HotActivityTimeConvert(cData.GetStartTime())
				cEnd = HotActivityTimeConvert(cData.GetEndTime())
			}

			if cBegin < pBegin || cEnd > pEnd {
				wrong = append(wrong, cData)
			}
		}
	}

	ok = len(wrong) == 0

	return
}

// Proto2Timestamp 返回按时间排序的[]HotActivity
func Proto2Timestamp(d []*ProtobufGen.HotActivityTime) (hot []HotActivity) {
	for _, hotData := range d {
		h := new(HotActivity)
		h.Source = hotData
		if hotData.GetTimeType() == 1 {
			h.starttimestamp = HotActivityTimeConvert(hotData.GetStartTime())
			h.endtimestamp = HotActivityTimeConvert(hotData.GetEndTime())
		} else if hotData.GetTimeType() == 2 {
			start, err1 := strconv.ParseFloat(hotData.GetStartTime(), 32)
			end, err2 := strconv.ParseFloat(hotData.GetEndTime(), 32)
			if err1 != nil || err2 != nil {
				// 之前已经检查过一次时间了，这里略过
				continue
			}
			h.starttimestamp = int64(start)
			h.endtimestamp = int64(end)
		}

		hot = append(hot, *h)
	}

	// 先排序结束时间再排序开始时间，保证开始时间相同时，结束时间早的在前
	sort.Slice(hot, func(i, j int) bool { return hot[i].endtimestamp < hot[j].endtimestamp })
	sort.Slice(hot, func(i, j int) bool { return hot[i].starttimestamp < hot[j].starttimestamp })

	return
}

// CheckHotActivityTimeRange 检查是否有重合区间，返回具有重合时间的集合
func CheckHotActivityTimeRange(d []*ProtobufGen.HotActivityTime) ([][]*ProtobufGen.HotActivityTime, bool) {
	h := Proto2Timestamp(d)
	n := len(h)
	if n < 2 {
		return nil, true
	}

	// 记录有重合的组
	overlapGroups := make([][]*ProtobufGen.HotActivityTime, 0)

	for i := 0; i < n-1; i++ {
		data := h[i]
		overlap := []*ProtobufGen.HotActivityTime{data.Source}

		for j := i + 1; j < n; j++ {
			curData := h[j]

			if curData.starttimestamp > data.endtimestamp {
				i = j - 1
				break
			} else if curData.starttimestamp >= data.starttimestamp {
				overlap = append(overlap, curData.Source)
			}
		}

		if len(overlap) > 1 {
			overlapGroups = append(overlapGroups, overlap)
		}
	}

	return overlapGroups, len(overlapGroups) == 0
}

// RecordActivityTimeError 处理时间
func RecordActivityError(info string, checkType int, errType int, errData []*ProtobufGen.HotActivityTime) {
	logs := make([]string, 0, len(errData))

	switch checkType {
	case utils.IS_TIME_CORRECT, utils.IS_TIME_RANGE_CORRECT:
		for _, eData := range errData {
			log := fmt.Sprintf("活动ID: %v, 开始时间:%v 结束时间:%v", eData.GetActivityID(), eData.GetStartTime(), eData.GetEndTime())
			logs = append(logs, log)
		}
	case utils.IS_SERVER_GROUP_OVERLAP:
		for _, eData := range errData {
			log := fmt.Sprintf("活动ID: %v, 类型:%v GID:%v", eData.GetActivityID(), eData.GetActivityType(), eData.GetServerGroupID())
			logs = append(logs, log)
		}
	}
	reporter.Record(-1, info, checkType, errType, logs)
}

// IsTimeAvailable 检查时间是否正确以及开始时间是否小于结束时间
func IsTimeAvailable(h *ProtobufGen.HotActivityTime) (ok bool) {
	startStr := h.GetStartTime()
	endStr := h.GetEndTime()

	switch h.GetTimeType() {
	case 1:
		start := HotActivityTimeConvert(startStr)
		end := HotActivityTimeConvert(endStr)
		if start != 0 && end != 0 {
			ok = start < end
		}
	case 2:
		// 有个坑，string转换后的数字是"2.0"，不能用Atoi
		start, err1 := strconv.ParseFloat(startStr, 32)
		end, err2 := strconv.ParseFloat(endStr, 32)
		if err1 != nil || err2 != nil {
			return
		}
		ok = start < end
	}

	return
}

// CheckLimitHeroGachaServerGroup 检查同一个活动的服务器分组里，是否有重合的服务器id
func CheckLimitHeroGachaServerGroup(s map[uint32][]uint32) (wrong []string, ok bool) {
	serverGroupData := GetAllServerGroup()

	for aid, sids := range s {
		serverPool := make(map[uint32]bool)

		for _, sid := range sids {
			// 遍历SEVERGROUP中，GroupID与当前ServerGroupID一致的服务器ID
			for _, sgd := range serverGroupData {
				if sgd.GetGroupID() == sid {
					serverId := sgd.GetSID()

					// 如果存在说明有服务器重合，报错
					if _, ok := serverPool[serverId]; ok {
						wrong = append(wrong, fmt.Sprintf("活动ID:%d, 分组ID:%d, 服务器ID:%d", aid, serverId, sid))
					} else {
						serverPool[serverId] = true
					}
				}
			}
		}
	}

	ok = len(wrong) == 0

	return
}

func IsContainsSameServerID(h []*ProtobufGen.HotActivityTime) (wrong []string, ok bool) {
	haServerGroup := GetHAServerGroup()
	serverPool := make(map[uint32]uint32)

	for _, hotData := range h {
		// 获取活动的ServerGroupID
		serverGroupID := hotData.GetServerGroupID()

		for _, haServerData := range haServerGroup {
			if serverGroupID == haServerData.GetServerGroupID() {
				// 获得这行下的分组Slice
				serverGroups := haServerData.GetAccCon_Table()

				for _, serverGroup := range serverGroups {
					// 遍历每组server
					begin := serverGroup.GetServerGroupValue1()
					end := serverGroup.GetServerGroupValue2()

					for sid := begin; sid < end+1; sid++ {
						if conflict, ok := serverPool[sid]; ok {
							info := fmt.Sprintf("活动ID:%d, 分组ID:%d, 服务器ID:%d, 冲突ID:%v",
								hotData.GetActivityID(), hotData.GetServerGroupID(), sid, conflict)
							wrong = append(wrong, info)
							continue
						} else {
							serverPool[sid] = hotData.GetActivityID()
						}
					}
				}
			}
		}
	}

	// TODO: 我艹这循环了几遍……

	ok = len(wrong) == 0

	return
}

// CheckHotActivityData 返回错误以及是否ok 这是main调用的执行功能的检查函数
func CheckHotActivityData() (errorCount int) {
	haData := NewHotActivitiesData()
	haData.ParseAll()

	// 限时神将 - 检查每个服务器分组
	for sid, dataGroup := range haData.limitHeroGachaBySid {
		errDataGroup, ok := CheckHotActivityTimeRange(dataGroup)
		if !ok {
			for _, errData := range errDataGroup {
				RecordActivityError(fmt.Sprintf("限时神将时间:%v ", sid),
					utils.IS_TIME_RANGE_CORRECT,
					utils.HOT_ACTIVITY_TIME_RANGE_INVALID,
					errData)
				errorCount += len(errData)
			}
		}
	}

	// 限时神将 - 检查每个活动的服务器是否有重复
	errLog, ok := CheckLimitHeroGachaServerGroup(haData.limitHeroGachaByAid)
	if !ok {
		reporter.Record(-1,
			"限时神将分组:",
			utils.IS_SERVER_GROUP_OVERLAP,
			utils.HOT_ACTIVITY_SEVER_ID_INVALID,
			errLog)
		errorCount += len(errLog)
	}

	// 子活动检查
	errData, ok := haData.CheckSubAcitivityTimeRange()
	if !ok {
		RecordActivityError("子活动检查", utils.IS_TIME_RANGE_CORRECT, utils.HOT_ACTIVITY_TIME_RANGE_INVALID, errData)
		errorCount += len(errData)
	}

	// 其他活动检查 - 同一个服务器组，同一种类型的活动时间是否有覆盖
	for sid, data := range haData.icActivityBySid {
		for aTypeId, d := range data {
			errDataGroup, ok := CheckHotActivityTimeRange(d)
			if !ok {
				for _, errData := range errDataGroup {
					RecordActivityError(fmt.Sprintf("服务器分组:%v 活动类型:%v ", sid, aTypeId),
						utils.IS_TIME_RANGE_CORRECT,
						utils.HOT_ACTIVITY_TIME_INVALID,
						errData)
					errorCount += len(errData)
				}
			}
		}
	}

	// 其他活动检查 - 同一种活动类型，时间有覆盖时，服务器组包含的服务器是否有覆盖
	for typeId, h := range haData.icActivityByAType {
		overlaps, ok := CheckHotActivityTimeRange(h)
		// 时间有重合
		for _, overlap := range overlaps {
			if !ok {
				// ServerID有重合
				errSidInfo, ok := IsContainsSameServerID(overlap)
				if !ok {
					reporter.Record(-1,
						fmt.Sprintf("活动类型:%v", typeId),
						utils.IS_SERVER_GROUP_OVERLAP,
						utils.HOT_ACTIVITY_SEVER_ID_INVALID,
						errSidInfo)
					errorCount += len(errData)
				}
			}
		}
	}

	return
}

// Report 输入格式化的错误信息到指定目录
func Report() {
	reporter.Report()
}
