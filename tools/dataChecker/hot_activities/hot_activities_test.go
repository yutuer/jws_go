package hot_activities

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/tools/dataChecker/utils"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

// debugReloadHotActivityData  Mock路径后，重新加载一遍数据
func debugReloadHotActivityData() {
	hat := GetHotActivityTime()
	dataMap = make(map[uint32]*ProtobufGen.HotActivityTime, len(hat))
	for _, d := range hat {
		dataMap[*d.ActivityID] = d
	}
}

func TestNewHotActivitiesData(t *testing.T) {
	haData := NewHotActivitiesData()

	assert.NotNil(t, haData.limitHeroGachaBySid)
	assert.Empty(t, haData.limitHeroGachaBySid)
	assert.NotNil(t, haData.limitHeroGachaByAid)
	assert.Empty(t, haData.limitHeroGachaByAid)
	assert.NotNil(t, haData.icActivityBySid)
	assert.Empty(t, haData.icActivityBySid)
	assert.NotNil(t, haData.icActivityByAType)
	assert.Empty(t, haData.icActivityByAType)
	assert.NotNil(t, haData.subActivities)
	assert.Empty(t, haData.subActivities)
}

func TestgetHotActivityTime(t *testing.T) {
	hats := GetHotActivityTime()
	for _, hat := range hats {
		t.Logf("HotActivity: ", hat)
	}
}

func TestgetAllServerGroup(t *testing.T) {
	asg := GetAllServerGroup()
	for _, sg := range asg {
		t.Logf("AllServerGroup: ", sg)
	}
}

func TestgetHAServerGroup(t *testing.T) {
	HAsg := GetHAServerGroup()
	for _, sg := range HAsg {
		t.Logf("HAServerGroup: ", sg)
	}
}

func TestgetSGActivity(t *testing.T) {
	HAsg := GetSGActivity()
	for _, sg := range HAsg {
		t.Logf("SGActivity: ", sg)
	}
}

func TestIsIncompatibleActivity(t *testing.T) {
	assert.True(t, isIncompatibleActivity(10001))
	assert.False(t, isIncompatibleActivity(1001))
}

func TestHotActivityTimeConvert(t *testing.T) {
	t.Run("Right1", func(t *testing.T) {
		timeString := "20170801_13:00"
		timestamp := HotActivityTimeConvert(timeString)
		assert.Equal(t, int64(1501592400), timestamp)
	})

	t.Run("Wrong1", func(t *testing.T) {
		timeString := "20170801-13:00"
		timestamp := HotActivityTimeConvert(timeString)
		assert.Equal(t, int64(0), timestamp)
	})

	t.Run("Wrong2", func(t *testing.T) {
		timeString := "20170931-00:00"
		timestamp := HotActivityTimeConvert(timeString)
		assert.Equal(t, int64(0), timestamp)
	})
}

// debugSetHotActivityData 生成默认开启的HotActivity数据
func debugSetHotActivityData(aId uint32, aType uint32, serverGroupId uint32) *ProtobufGen.HotActivityTime {
	aValid := uint32(1)
	timeType1 := uint32(1)
	aTitle := fmt.Sprintf("Title%d-%d-%d", aId, aType, serverGroupId)

	h := &ProtobufGen.HotActivityTime{
		ActivityID:     &aId,
		ActivityType:   &aType,
		ActivityPID:    new(uint32),
		ActivityValid:  &aValid,
		ActivityTitle:  &aTitle,
		TimeType:       &timeType1,
		StartTime:      new(string),
		EndTime:        new(string),
		Duration:       new(uint32),
		ServerGroupID:  &serverGroupId,
		ChannelGroupID: new(uint32),
		TeleID:         new(string),
		ConditionID:    new(uint32),
		HotActivity:    new(uint32),
		TabIDS:         new(string),
	}

	return h
}

// debugSetHotActivityTime 设置活动时间
func debugSetHotActivityTime(h *ProtobufGen.HotActivityTime, timeType uint32, startTime string, endTime string) {
	h.TimeType = &timeType
	h.StartTime = &startTime
	h.EndTime = &endTime
}

func TestHotActivitiesData_parseByActivityType(t *testing.T) {
	haData := NewHotActivitiesData()
	h := debugSetHotActivityData(0, 15, 0)

	haData.parseByActivityType(h)

	_, ok := haData.icActivityByAType[15]
	assert.True(t, ok)
	assert.Equal(t, 1, len(haData.icActivityByAType))
}

func TestHotActivitiesData_parseByServerGroup(t *testing.T) {
	haData := NewHotActivitiesData()
	h := debugSetHotActivityData(0, 0, 2001)

	haData.parseByServerGroup(h)

	_, ok := haData.icActivityBySid[2001]
	assert.True(t, ok)
	assert.Equal(t, 1, len(haData.icActivityBySid))
}

func TestHotActivitiesData_ParseicActivityBySid(t *testing.T) {

	// 正常流程
	t.Run("限时神将", func(t *testing.T) {
		haData := NewHotActivitiesData()
		h := debugSetHotActivityData(100025, 1001, 0)

		haData.ParseicActivity(h)

		assert.Equal(t, 1, len(haData.limitHeroGachaBySid))
	})

	t.Run("受限制活动", func(t *testing.T) {
		haData := NewHotActivitiesData()
		h := debugSetHotActivityData(100026, 64, 0)
		haData.ParseicActivity(h)

		assert.Equal(t, 1, len(haData.icActivityBySid))
		assert.Equal(t, 1, len(haData.icActivityByAType))
	})

	t.Run("非受限制活动", func(t *testing.T) {
		haData := NewHotActivitiesData()
		h := debugSetHotActivityData(100026, 26, 0)
		haData.ParseicActivity(h)

		assert.Empty(t, len(haData.limitHeroGachaBySid))
		assert.Empty(t, len(haData.icActivityBySid))
		assert.Empty(t, len(haData.icActivityByAType))
	})

	// 异常流程

	// 遍历
	t.Run("受限组新增和累加", func(t *testing.T) {
		haData := NewHotActivitiesData()

		t.Run("ActType = 64, GroupID = 233", func(t *testing.T) {
			h := debugSetHotActivityData(100231, 64, 233)
			haData.ParseicActivity(h)

			assert.Empty(t, len(haData.limitHeroGachaBySid))
			assert.Equal(t, 1, len(haData.icActivityBySid))
			assert.Equal(t, 1, len(haData.icActivityByAType))
		})

		t.Run("ActType = 74, GroupID = 233", func(t *testing.T) {
			h := debugSetHotActivityData(100232, 74, 233)
			haData.ParseicActivity(h)

			assert.Empty(t, len(haData.limitHeroGachaBySid))
			assert.Equal(t, 1, len(haData.icActivityBySid))
			assert.Equal(t, 2, len(haData.icActivityByAType))
		})

		t.Run("limitHeroGachaBySid", func(t *testing.T) {
			h := debugSetHotActivityData(102222, 1003, 0)
			haData.ParseicActivity(h)

			//assert.Equal(t, 1, len(haData.limitHeroGachaBySid))
			assert.Equal(t, 1, len(haData.icActivityBySid))
			assert.Equal(t, 2, len(haData.icActivityByAType))
		})

		t.Run("Should Pass 1", func(t *testing.T) {
			h := debugSetHotActivityData(100239, 174, 666)
			haData.ParseicActivity(h)

			//assert.Equal(t, 1, len(haData.limitHeroGachaBySid))
			assert.Equal(t, 1, len(haData.icActivityBySid))
			assert.Equal(t, 2, len(haData.icActivityByAType))
		})

		t.Run("ActType = 74, GroupID = 666", func(t *testing.T) {
			h := debugSetHotActivityData(100233, 74, 666)
			haData.ParseicActivity(h)

			//assert.Equal(t, 1, len(haData.limitHeroGachaBySid))
			assert.Equal(t, 2, len(haData.icActivityBySid))
			assert.Equal(t, 2, len(haData.icActivityByAType))
		})

		t.Run("ActType = 64, GroupID = 666", func(t *testing.T) {
			h := debugSetHotActivityData(100234, 64, 666)
			haData.ParseicActivity(h)

			//assert.Equal(t, 1, len(haData.limitHeroGachaBySid))
			assert.Equal(t, 2, len(haData.icActivityBySid))
			assert.Equal(t, 2, len(haData.icActivityByAType))
		})

		t.Run("limitHeroGachaBySid2", func(t *testing.T) {
			h := debugSetHotActivityData(102224, 1001, 0)
			haData.ParseicActivity(h)

			//assert.Equal(t, 2, len(haData.limitHeroGachaBySid))
			assert.Equal(t, 2, len(haData.icActivityBySid))
			assert.Equal(t, 2, len(haData.icActivityByAType))
		})
	})
}

func TestCheckHotActivityTimeRange(t *testing.T) {
	time1String := "20161216_05:00"
	time2String := "20161216_13:00"
	time3String := "20161216_21:00"
	time4String := "20161217_05:00"
	time5String := "20161217_13:00"
	time6String := "20161217_21:00"

	// [1 [2]]
	t.Run("有重合1", func(t *testing.T) {
		d1 := debugSetHotActivityData(0, 0, 0)
		d2 := debugSetHotActivityData(0, 0, 0)

		d1.StartTime = &time1String
		d1.EndTime = &time4String

		d2.StartTime = &time2String
		d2.EndTime = &time3String

		d := []*ProtobufGen.HotActivityTime{d1, d2}

		r, ok := CheckHotActivityTimeRange(d)

		assert.Equal(t, 1, len(r))
		assert.Equal(t, 2, len(r[0]))
		assert.False(t, ok)
	})

	// [1 [2 [3]]]
	t.Run("有重合2", func(t *testing.T) {
		d1 := debugSetHotActivityData(0, 0, 0)
		d2 := debugSetHotActivityData(0, 0, 0)
		d3 := debugSetHotActivityData(0, 0, 0)

		debugSetHotActivityTime(d1, 1, time1String, time6String)
		debugSetHotActivityTime(d2, 1, time2String, time5String)
		debugSetHotActivityTime(d3, 1, time3String, time4String)

		d := []*ProtobufGen.HotActivityTime{d1, d2, d3}

		r, ok := CheckHotActivityTimeRange(d)

		assert.Equal(t, 2, len(r))
		assert.Equal(t, 3, len(r[0]))
		assert.Equal(t, 2, len(r[1]))
		assert.False(t, ok)
	})

	// [1 [2] [3]]
	t.Run("有重合3", func(t *testing.T) {
		d1 := debugSetHotActivityData(0, 0, 0)
		d2 := debugSetHotActivityData(0, 0, 0)
		d3 := debugSetHotActivityData(0, 0, 0)

		debugSetHotActivityTime(d1, 1, time1String, time6String)
		debugSetHotActivityTime(d2, 1, time2String, time3String)
		debugSetHotActivityTime(d3, 1, time4String, time5String)

		d := []*ProtobufGen.HotActivityTime{d1, d2, d3}

		r, ok := CheckHotActivityTimeRange(d)

		assert.Equal(t, 1, len(r))
		assert.Equal(t, 3, len(r[0]))
		assert.False(t, ok)
	})

	// [ 1 [] 2 [] 3 ]
	t.Run("有重合4", func(t *testing.T) {
		d1 := debugSetHotActivityData(0, 0, 0)
		d2 := debugSetHotActivityData(0, 0, 0)
		d3 := debugSetHotActivityData(0, 0, 0)

		debugSetHotActivityTime(d1, 1, time1String, time4String)
		debugSetHotActivityTime(d2, 1, time2String, time5String)
		debugSetHotActivityTime(d3, 1, time3String, time6String)

		d := []*ProtobufGen.HotActivityTime{d1, d2, d3}

		r, ok := CheckHotActivityTimeRange(d)

		assert.Equal(t, 2, len(r))
		assert.Equal(t, 3, len(r[0]))
		assert.Equal(t, 2, len(r[1]))
		assert.False(t, ok)
	})

	t.Run("无重合", func(t *testing.T) {
		d1 := debugSetHotActivityData(0, 0, 0)
		d2 := debugSetHotActivityData(0, 0, 0)
		d3 := debugSetHotActivityData(0, 0, 0)

		debugSetHotActivityTime(d1, 1, time1String, time2String)
		debugSetHotActivityTime(d2, 1, time3String, time4String)
		debugSetHotActivityTime(d3, 1, time5String, time6String)

		d := []*ProtobufGen.HotActivityTime{d1, d2, d3}

		r, ok := CheckHotActivityTimeRange(d)

		assert.Empty(t, r)
		assert.True(t, ok)
	})
}

func TestIsTimeAvailable(t *testing.T) {
	time1String := "20161216_05:00"
	time2String := "20161216_13:00"
	time3String := "0.0"
	time4String := "2.0"

	t.Run("类型1-大于", func(t *testing.T) {
		p := debugSetHotActivityData(23333, 0, 0)
		debugSetHotActivityTime(p, 1, time1String, time2String)

		assert.True(t, IsTimeAvailable(p))
	})

	t.Run("类型1-等于", func(t *testing.T) {
		p := debugSetHotActivityData(23333, 0, 0)
		debugSetHotActivityTime(p, 1, time1String, time1String)

		assert.False(t, IsTimeAvailable(p))
	})

	t.Run("类型1-小于", func(t *testing.T) {
		p := debugSetHotActivityData(23333, 0, 0)
		debugSetHotActivityTime(p, 1, time2String, time1String)

		assert.False(t, IsTimeAvailable(p))
	})

	t.Run("类型2-大于", func(t *testing.T) {
		p := debugSetHotActivityData(23333, 0, 0)
		debugSetHotActivityTime(p, 2, time3String, time4String)

		assert.True(t, IsTimeAvailable(p))
	})

	t.Run("类型2-等于", func(t *testing.T) {
		p := debugSetHotActivityData(23333, 0, 0)
		debugSetHotActivityTime(p, 2, time3String, time3String)

		assert.False(t, IsTimeAvailable(p))
	})

	t.Run("类型2-小于", func(t *testing.T) {
		p := debugSetHotActivityData(23333, 0, 0)
		debugSetHotActivityTime(p, 2, time4String, time3String)

		assert.False(t, IsTimeAvailable(p))
	})

	t.Run("类型不一致1", func(t *testing.T) {
		p := debugSetHotActivityData(23333, 0, 0)
		debugSetHotActivityTime(p, 1, time1String, time3String)

		assert.False(t, IsTimeAvailable(p))
	})

	t.Run("类型不一致2", func(t *testing.T) {
		p := debugSetHotActivityData(23333, 0, 0)
		debugSetHotActivityTime(p, 2, time3String, time2String)

		assert.False(t, IsTimeAvailable(p))
	})
}

func TestHotActivitiesData_CheckSubAcitivityTimeRange(t *testing.T) {
	time1String := "20161216_05:00"
	time2String := "20161216_13:00"
	time3String := "20161216_21:00"
	time4String := "20161217_05:00"
	time5String := "20161217_13:00"
	time6String := "20161217_21:00"

	p := debugSetHotActivityData(23333, 0, 0)
	debugSetHotActivityTime(p, 1, time1String, time4String)

	// Mock
	dataMap[*p.ActivityID] = p

	// 正常流程
	t.Run("包含于", func(t *testing.T) {
		d := NewHotActivitiesData()

		c1 := debugSetHotActivityData(1, 0, 0)
		c2 := debugSetHotActivityData(2, 0, 0)

		c1.ActivityPID, c2.ActivityPID = p.ActivityID, p.ActivityID

		debugSetHotActivityTime(c1, 1, time1String, time4String)
		debugSetHotActivityTime(c2, 1, time2String, time3String)

		m := make(map[uint32][]*ProtobufGen.HotActivityTime)
		m[*p.ActivityID] = []*ProtobufGen.HotActivityTime{c1, c2}
		d.subActivities = m

		r, ok := d.CheckSubAcitivityTimeRange()

		assert.Equal(t, 0, len(r))
		assert.True(t, ok)
	})

	t.Run("相交", func(t *testing.T) {
		d := NewHotActivitiesData()

		c1 := debugSetHotActivityData(1, 0, 0)
		c2 := debugSetHotActivityData(2, 0, 0)

		c1.ActivityPID = p.ActivityID
		c2.ActivityPID = p.ActivityID

		debugSetHotActivityTime(c1, 1, time2String, time5String)
		debugSetHotActivityTime(c2, 1, time3String, time6String)

		m := make(map[uint32][]*ProtobufGen.HotActivityTime)
		m[*p.ActivityID] = []*ProtobufGen.HotActivityTime{c1, c2}
		d.subActivities = m

		r, ok := d.CheckSubAcitivityTimeRange()

		assert.Equal(t, 2, len(r))
		assert.Equal(t, uint32(1), r[0].GetActivityID())
		assert.False(t, ok)
	})

	t.Run("不相交", func(t *testing.T) {
		d := NewHotActivitiesData()

		c1 := debugSetHotActivityData(1, 0, 0)
		c2 := debugSetHotActivityData(2, 0, 0)

		c1.ActivityPID = p.ActivityID
		c2.ActivityPID = p.ActivityID

		debugSetHotActivityTime(c1, 1, time5String, time6String)
		debugSetHotActivityTime(c2, 1, time1String, time6String)

		m := make(map[uint32][]*ProtobufGen.HotActivityTime)
		m[*p.ActivityID] = []*ProtobufGen.HotActivityTime{c1, c2}
		d.subActivities = m

		r, ok := d.CheckSubAcitivityTimeRange()

		assert.Equal(t, 2, len(r))
		assert.Equal(t, uint32(1), r[0].GetActivityID())
		assert.False(t, ok)
	})

	t.Run("时间类型不同", func(t *testing.T) {
		d := NewHotActivitiesData()

		c1 := debugSetHotActivityData(1, 0, 0)

		c1.ActivityPID = p.ActivityID
		debugSetHotActivityTime(c1, 2, time5String, time6String)

		m := make(map[uint32][]*ProtobufGen.HotActivityTime)
		m[*p.ActivityID] = []*ProtobufGen.HotActivityTime{c1}
		d.subActivities = m

		r, ok := d.CheckSubAcitivityTimeRange()

		assert.Equal(t, 1, len(r))
		assert.Equal(t, uint32(1), r[0].GetActivityID())
		assert.False(t, ok)
	})
}

func TestCheckHotActivityData(t *testing.T) {
	// Mock
	utils.DebugSetGamedataDir("tools/dataChecker/test/hot_activities_test")
	f := debugConstructTestData()
	f2 := debugConstrucHAServerGroupData()
	debugReloadHotActivityData()

	t.Run("Check it now", func(t *testing.T) {
		n := CheckHotActivityData()
		assert.Equal(t, 11, n)
		assert.Equal(t, 5, len(reporter.Unexceptions))

		//Report()
	})

	// 恢复并删除生成的testdata
	utils.DebugSetGamedataDir("jws/gamex/conf/data")
	debugReloadHotActivityData()
	os.Remove(f)
	os.Remove(f2)
	//reporter.DebugRemoveLastlog()
}

// debugConstructTestData 构建测试数据并写入到指定的data目录
func debugConstructTestData() (testFilename string) {
	timeString1 := "20170805_05:00"
	timeString2 := "20170805_13:00"
	timeString3 := "20170805_21:00"
	timeString4 := "10.0"
	timeString5 := "24.0"

	// 活动类型时间重合
	a1 := debugSetHotActivityData(1, 50, 0)
	debugSetHotActivityTime(a1, 1, timeString1, timeString2)

	a2 := debugSetHotActivityData(2, 50, 0)
	debugSetHotActivityTime(a2, 1, timeString1, timeString2)

	// 子类型时间不重合
	pID := uint32(3)
	a3 := debugSetHotActivityData(3, 66, 2)
	debugSetHotActivityTime(a3, 1, timeString1, timeString2)

	a4 := debugSetHotActivityData(4, 68, 2)
	a4.ActivityPID = &pID
	debugSetHotActivityTime(a4, 1, timeString2, timeString3)

	// 活动时间错误
	a5 := debugSetHotActivityData(5, 82, 3)
	debugSetHotActivityTime(a5, 1, timeString2, timeString1)

	// 活动时间错误 - 子类型
	a6 := debugSetHotActivityData(6, 69, 1)
	a6.ActivityPID = &pID
	debugSetHotActivityTime(a6, 1, timeString3, timeString2)

	// 服务器分组错误 - 子类型
	a7 := debugSetHotActivityData(7, 70, 2)
	a7.ActivityPID = &pID
	debugSetHotActivityTime(a7, 1, timeString2, timeString3)

	// 时间类型不匹配 - 子类型
	a8 := debugSetHotActivityData(8, 71, 1)
	a8.ActivityPID = &pID
	debugSetHotActivityTime(a8, 2, timeString4, timeString5)

	// 活动分组ServerID重合
	a9 := debugSetHotActivityData(9, 78, 2)
	debugSetHotActivityTime(a9, 1, timeString1, timeString2)

	a10 := debugSetHotActivityData(10, 78, 3)
	debugSetHotActivityTime(a10, 1, timeString1, timeString2)

	// 构建
	HATs := &ProtobufGen.HotActivityTime_ARRAY{
		Items: []*ProtobufGen.HotActivityTime{a1, a2, a3, a4, a5, a6, a7, a8, a9, a10},
	}

	buff, err := proto.Marshal(HATs)
	if err != nil {
		panic(err)
	}

	testFilename = filepath.Join(utils.GetVCSRootPath(), "tools/dataChecker/test/hot_activities_test", "hotactivitytime.data")
	utils.WriteBuff2Bin(testFilename, buff)
	return
}

func TestProto2Timestamp(t *testing.T) {
	d1 := debugSetHotActivityData(1, 11, 0)
	debugSetHotActivityTime(d1, 1, "20170805_05:00", "20170805_21:00")

	d2 := debugSetHotActivityData(2, 12, 0)
	debugSetHotActivityTime(d2, 1, "20170805_03:00", "20170806_23:00")

	d3 := debugSetHotActivityData(3, 13, 0)
	debugSetHotActivityTime(d3, 1, "20170805_03:00", "20170805_23:00")

	d4 := debugSetHotActivityData(4, 14, 0)
	debugSetHotActivityTime(d4, 2, "2.0", "7.0")

	d5 := debugSetHotActivityData(5, 15, 0)
	debugSetHotActivityTime(d5, 2, "1.0", "20.0")

	d6 := debugSetHotActivityData(6, 16, 0)
	debugSetHotActivityTime(d6, 2, "1.0", "9.0")

	d := []*ProtobufGen.HotActivityTime{d1, d2, d3, d4, d5, d6}

	h := Proto2Timestamp(d)

	// 结果应该为 [6, 5, 4, 3, 2, 1]
	for i, hot := range h {
		assert.Equal(t, uint32(6-i), hot.Source.GetActivityID())
	}
}

func debugGetSGACTIVITY(groupId uint32, hotActivityId []uint32) *ProtobufGen.SGACTIVITY {
	d := &ProtobufGen.SGACTIVITY{
		GroupID:       &groupId,
		HotActivityID: hotActivityId,
	}

	return d
}

func debugGetSEVERGROUP(sid uint32, groupId uint32) *ProtobufGen.SEVERGROUP {
	d := &ProtobufGen.SEVERGROUP{
		SID:              &sid,
		GroupID:          &groupId,
		WspvpGroupID:     new(uint32),
		WspvpBot:         new(uint32),
		RobCropsGroupID:  new(uint32),
		Sbatch:           new(uint32),
		EffectiveTime:    new(string),
		WorldBossGroupID: new(uint32),
		HGRHotID:         new(uint32),
		TeamBossGroupID:  new(uint32),
	}

	return d
}

func debugGetSERVERGROUP(groupId uint32, g1Start uint32, g1End uint32, g2Start uint32, g2End uint32) *ProtobufGen.SERVERGROUP {
	d := &ProtobufGen.SERVERGROUP{
		ServerGroupID:      &groupId,
		ServerGroupType:    new(uint32),
		ServerGroupSubType: new(uint32),
	}

	group1 := &ProtobufGen.SERVERGROUP_AcceptCondition{
		ServerGroupValue1: &g1Start,
		ServerGroupValue2: &g1End,
	}

	group2 := &ProtobufGen.SERVERGROUP_AcceptCondition{
		ServerGroupValue1: &g2Start,
		ServerGroupValue2: &g2End,
	}

	d.AccCon_Table = []*ProtobufGen.SERVERGROUP_AcceptCondition{group1, group2}

	return d
}

func debugConstrucHAServerGroupData() (filename string) {
	d0 := debugGetSERVERGROUP(0, 3, 3, 6, 6)
	d1 := debugGetSERVERGROUP(1, 1, 3, 5, 9)
	d2 := debugGetSERVERGROUP(2, 3, 5, 11, 12)
	d3 := debugGetSERVERGROUP(3, 6, 6, 11, 20)

	d := &ProtobufGen.SERVERGROUP_ARRAY{
		Items: []*ProtobufGen.SERVERGROUP{d0, d1, d2, d3},
	}

	buff, err := proto.Marshal(d)
	if err != nil {
		panic(err)
	}

	filename = filepath.Join(utils.GetVCSRootPath(), "tools/dataChecker/test/hot_activities_test", "servergroup.data")
	utils.WriteBuff2Bin(filename, buff)

	return
}

/*
func debugConstructLimitHeroGachaTestData() {
	sga1 := debugGetSGACTIVITY(0, []uint32{100001, 100002, 100003})
	sga2 := debugGetSGACTIVITY(1, []uint32{100001, 100003})
	sga3 := debugGetSGACTIVITY(2, []uint32{100002})
	sga3 := debugGetSGACTIVITY(3, []uint32{100002})

	sg1 := debugGetSEVERGROUP(0, 11)
	sg2 := debugGetSEVERGROUP(0, 12)
}

func TestCheckLimitHeroGachaServerGroup(t *testing.T) {
	haData := NewHotActivitiesData()

	r, ok := CheckLimitHeroGachaServerGroup(haData.limitHeroGachaByAid)
	assert.True(t, ok)
	assert.Empty(t, r)
}
*/
