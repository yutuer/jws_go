package hot_activities

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/tools/dataChecker/utils"

	"github.com/golang/protobuf/proto"
)

var (
	reporter *utils.Reporter
	dataMap  map[uint32]*ProtobufGen.HotActivityTime
)

func init() {
	hat := GetHotActivityTime()
	dataMap = make(map[uint32]*ProtobufGen.HotActivityTime, len(hat))
	for _, d := range hat {
		dataMap[*d.ActivityID] = d
	}
	reporter = utils.NewReporter()
}

type HotActivity struct {
	Source         *ProtobufGen.HotActivityTime
	starttimestamp int64
	endtimestamp   int64
}

type HotActivitiesData struct {
	limitHeroGachaByAid map[uint32][]uint32                                  // [AcitivtyID] []ServerGroupID
	limitHeroGachaBySid map[uint32][]*ProtobufGen.HotActivityTime            // [ServerGroupID] []*Activity
	icActivityBySid     map[uint32]map[uint32][]*ProtobufGen.HotActivityTime // [ServerGroupID][Type][]*Activity
	icActivityByAType   map[uint32][]*ProtobufGen.HotActivityTime            // [TypeID] []*Activity
	subActivities       map[uint32][]*ProtobufGen.HotActivityTime            // [PActivityID] []*Activity
}

//  不能重复开启的活动列表，强烈建议之后配置成表格
var incompitableActivityTypes = []uint32{
	50,    //开玉石
	60,    //招财进宝
	64,    //军团红包
	65,    //无双宝库
	66,    //开服排行榜总榜
	67,    //神兽等级榜
	68,    //宝石积分榜
	69,    //装备星级榜
	70,    //主将星级榜
	71,    //战队等级榜
	72,    //战力等级榜
	73,    //副本掉落
	74,    //兑换商店
	75,    //主将巡礼
	76,    //神兵再临
	77,    //星图排行榜
	78,    //羁绊排行榜
	79,    //幻甲星级榜
	80,    //神兵积分榜
	81,    //无双战力榜
	82,    //第二周排行榜总榜
	83,    //第二周宝石积分榜
	10001, //名将投资
	20001, //打年兽
}

// NewHotActivitiesData 返回一个初始化的HotActivitiesData指针
func NewHotActivitiesData() *HotActivitiesData {
	h := &HotActivitiesData{}
	h.limitHeroGachaByAid = make(map[uint32][]uint32)
	h.limitHeroGachaBySid = make(map[uint32][]*ProtobufGen.HotActivityTime)
	h.icActivityBySid = make(map[uint32]map[uint32][]*ProtobufGen.HotActivityTime)
	h.icActivityByAType = make(map[uint32][]*ProtobufGen.HotActivityTime)
	h.subActivities = make(map[uint32][]*ProtobufGen.HotActivityTime)

	return h
}

// GetHotActivityTime 从data文件中读取并生成HotActivityTime
func GetHotActivityTime() []*ProtobufGen.HotActivityTime {
	HATFilename := utils.GetDataFileFullPath("hotactivitytime")
	buff, err := utils.LoadBin2Buff(HATFilename)
	if err != nil {
		panic(err)
	}

	HATs := &ProtobufGen.HotActivityTime_ARRAY{}
	err = proto.Unmarshal(buff, HATs)
	if err != nil {
		panic(err)
	}

	return HATs.Items
}

// GetAllServerGroup 获取所有跨服活动的服务器分组
func GetAllServerGroup() []*ProtobufGen.SEVERGROUP {
	allServerGroupFilename := utils.GetDataFileFullPath("severgroup") // 没有拼写错误，就是少个r
	buff, err := utils.LoadBin2Buff(allServerGroupFilename)
	if err != nil {
		panic(err)
	}

	allServerGroup := &ProtobufGen.SEVERGROUP_ARRAY{}
	err = proto.Unmarshal(buff, allServerGroup)
	if err != nil {
		panic(err)
	}

	return allServerGroup.Items
}

// GetHAServerGroup 获取HotActivity跨服活动的服务器分组
func GetHAServerGroup() []*ProtobufGen.SERVERGROUP {
	HAServerGroupFilename := utils.GetDataFileFullPath("servergroup")
	buff, err := utils.LoadBin2Buff(HAServerGroupFilename)
	if err != nil {
		panic(err)
	}

	HAServerGroup := &ProtobufGen.SERVERGROUP_ARRAY{}
	err = proto.Unmarshal(buff, HAServerGroup)
	if err != nil {
		panic(err)
	}

	return HAServerGroup.Items
}

// GetSGActivity 获取限时神将专属的分组
func GetSGActivity() []*ProtobufGen.SGACTIVITY {
	GActivityFilename := utils.GetDataFileFullPath("sgactivity")
	buff, err := utils.LoadBin2Buff(GActivityFilename)
	if err != nil {
		panic(err)
	}

	GActivity := &ProtobufGen.SGACTIVITY_ARRAY{}
	err = proto.Unmarshal(buff, GActivity)
	if err != nil {
		panic(err)
	}

	return GActivity.Items
}

// parseByServerGroup 按ServerGroup对活动进行分类
func (haData *HotActivitiesData) parseByServerGroup(h *ProtobufGen.HotActivityTime) {
	serverGroupId := h.GetServerGroupID()
	typeId := h.GetActivityType()

	if _, ok := haData.icActivityBySid[serverGroupId]; !ok {
		haData.icActivityBySid[*h.ServerGroupID] = make(map[uint32][]*ProtobufGen.HotActivityTime)
		haData.icActivityBySid[serverGroupId][typeId] = make([]*ProtobufGen.HotActivityTime, 0)
	} else {
		if _, ok := haData.icActivityBySid[serverGroupId][typeId]; !ok {
			haData.icActivityBySid[serverGroupId][typeId] = make([]*ProtobufGen.HotActivityTime, 0)
		}
	}

	haData.icActivityBySid[serverGroupId][typeId] = append(haData.icActivityBySid[serverGroupId][typeId], h)
}

// parseByActivityType 按ActivityType进行分类
func (haData *HotActivitiesData) parseByActivityType(h *ProtobufGen.HotActivityTime) {
	if _, ok := haData.icActivityByAType[*h.ActivityType]; ok {
		haData.icActivityByAType[*h.ActivityType] = append(haData.icActivityByAType[*h.ActivityType], h)
	} else {
		haData.icActivityByAType[*h.ActivityType] = []*ProtobufGen.HotActivityTime{h}
	}
}

// parseSubActivity 按ActivityPID分类
func (haData *HotActivitiesData) parseSubActivity(h *ProtobufGen.HotActivityTime) {
	if _, ok := haData.subActivities[*h.ActivityPID]; ok {
		haData.subActivities[*h.ActivityPID] = append(haData.subActivities[*h.ActivityPID], h)
	} else {
		haData.subActivities[*h.ActivityPID] = []*ProtobufGen.HotActivityTime{h}
	}
}

// parseLimitHeroGachaBySId 导入限时神将数据 [ActivityId] []ServerGroupIds
func (haData *HotActivitiesData) parseLimitHeroGachaByAId(h *ProtobufGen.HotActivityTime) {
	sIds := make([]uint32, 0)
	limitHeroGachaGroupDatas := GetSGActivity()

	for _, d := range limitHeroGachaGroupDatas {
		// 按ActivityID导入
		for _, id := range d.HotActivityID {
			if *h.ActivityID == id {
				sIds = append(sIds, *d.GroupID)
			}
		}
	}

	if len(sIds) != 0 {
		haData.limitHeroGachaByAid[*h.ActivityID] = sIds
	}
}

// parseLimitHeroGachaBySId 导入限时神将数据 [ServerGroup] Data
func (haData *HotActivitiesData) parseLimitHeroGachaBySId(h *ProtobufGen.HotActivityTime) {
	hasServerGroupId := false
	limitHeroGachaGroupDatas := GetSGActivity()

	for _, lgd := range limitHeroGachaGroupDatas {
		// 按ServerGroupID导入
		for _, aId := range lgd.HotActivityID {
			if *h.ActivityID == aId {
				hasServerGroupId = true
				if s, ok := haData.limitHeroGachaBySid[*lgd.GroupID]; ok {
					haData.limitHeroGachaBySid[*lgd.GroupID] = append(s, h)
				} else {
					haData.limitHeroGachaBySid[*lgd.GroupID] = []*ProtobufGen.HotActivityTime{h}
				}
			}
		}
	}

	if !hasServerGroupId {
		// 限时神将配置的服务器分组ID不正确
	}
}

// ParseAll 处理所有数据
func (haData *HotActivitiesData) ParseAll() {
	hotActivityTimes := GetHotActivityTime()
	errGroup := make([]*ProtobufGen.HotActivityTime, 0)

	for _, h := range hotActivityTimes {
		// 只检查开启的活动
		if h.GetActivityValid() == 1 {
			if IsTimeAvailable(h) {
				haData.ParseicActivity(h)
			} else {
				errGroup = append(errGroup, h)
			}
		}
	}

	if len(errGroup) > 1 {
		RecordActivityError(" 时间检查 ",
			utils.IS_TIME_CORRECT,
			utils.HOT_ACTIVITY_TIME_INVALID,
			errGroup,
		)
	}
}

// ParseicActivityBySid 检查活动是否属于需要检查的类型
func (haData *HotActivitiesData) ParseicActivity(h *ProtobufGen.HotActivityTime) {
	if 1000 < *h.ActivityType && *h.ActivityType < 10000 { // 限时神将
		haData.parseLimitHeroGachaByAId(h)
		haData.parseLimitHeroGachaBySId(h)
	} else if isIncompatibleActivity(*h.ActivityType) {
		if h.GetActivityPID() != 0 {
			haData.parseSubActivity(h)
		}
		haData.parseByServerGroup(h)
		haData.parseByActivityType(h)
	}
}

// isIncompatibleActivity 判断当前活动Id是否属于不能重复的id
func isIncompatibleActivity(activityType uint32) bool {
	for _, atype := range incompitableActivityTypes {
		if activityType == atype {
			return true
		}
	}

	return false
}

// HotActivityTimeConvert 将"yyyymmdd_hh:mm"转换成timestamp
func HotActivityTimeConvert(hatTime string) int64 {
	timeForm := "20060102_15:04"
	t, err := time.Parse(timeForm, hatTime)
	if err != nil {
		return 0
	}

	return t.Unix()
}
