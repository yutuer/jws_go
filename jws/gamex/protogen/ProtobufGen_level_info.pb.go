// Code generated by protoc-gen-go.
// source: ProtobufGen_level_info.proto
// DO NOT EDIT!

package ProtobufGen

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type LEVEL_INFO struct {
	// * 关卡ID
	LevelID *string `protobuf:"bytes,1,req,name=levelID,def=" json:"levelID,omitempty"`
	// * 关卡简称
	ChapterShorthand *string `protobuf:"bytes,2,opt,name=chapterShorthand,def=" json:"chapterShorthand,omitempty"`
	// * 关卡所属章节
	ChapterID *string `protobuf:"bytes,3,opt,name=chapterID,def=" json:"chapterID,omitempty"`
	// * 0:测试，1:主线，2:精英，3:金币关，4:BOSS战，5:精铁关，6:PVP，7:兵临城下，8:天命关，9:爬塔，10:组队BOSS，11:小关卡，12:切磋，13：地狱,14:军团boss，15：单人烽火燎原。16：多人烽火燎原。17：比武场.18:体验关.19:远征.20军团战。21节日boss活动关。22.出奇制胜【屠】。23.出奇制胜【斩】。24.出奇制胜【护】。25.出奇制胜【士】。26.9V9竞技。27.押运粮草。28.世界boss。29.组队boss。
	LevelType *int32 `protobuf:"varint,4,opt,name=levelType,def=0" json:"levelType,omitempty"`
	// * 地形枚举
	TerrainType *int32 `protobuf:"varint,5,opt,name=terrainType,def=0" json:"terrainType,omitempty"`
	// * 体力消耗
	Energy *int32 `protobuf:"varint,6,opt,name=energy,def=0" json:"energy,omitempty"`
	// * 消耗包子
	HighEnergy *int32 `protobuf:"varint,38,opt,def=0" json:"HighEnergy,omitempty"`
	// * 扫荡消耗的道具【日本版本专属】
	SweepItem *string `protobuf:"bytes,43,opt,def=" json:"SweepItem,omitempty"`
	// * 扫荡道具数量【日本版本专属】
	SweepItemCost *uint32 `protobuf:"varint,44,opt,def=0" json:"SweepItemCost,omitempty"`
	// * 每日最大访问
	MaxDailyAccess *int32 `protobuf:"varint,7,opt,name=maxDailyAccess,def=0" json:"maxDailyAccess,omitempty"`
	// * 0:正常,1:限时
	FinishType *int32 `protobuf:"varint,8,opt,name=finishType,def=0" json:"finishType,omitempty"`
	// * 限时
	TimeLimit *int32 `protobuf:"varint,9,opt,name=timeLimit,def=0" json:"timeLimit,omitempty"`
	// * 加星时间
	TimeGoal *int32 `protobuf:"varint,10,opt,name=timeGoal,def=0" json:"timeGoal,omitempty"`
	// * 加星生命比例
	HpGoal *float32 `protobuf:"fixed32,11,opt,name=hpGoal,def=0" json:"hpGoal,omitempty"`
	// * 前续关卡ID列表
	PreLevelID *string `protobuf:"bytes,12,opt,name=preLevelID,def=" json:"preLevelID,omitempty"`
	// * 队伍等级限制
	TeamRequirement *int32 `protobuf:"varint,13,opt,name=teamRequirement,def=0" json:"teamRequirement,omitempty"`
	// * 角色等级限制
	LevelRequirement *int32 `protobuf:"varint,14,opt,name=levelRequirement,def=0" json:"levelRequirement,omitempty"`
	// * 角色专属
	RoleOnly *int32 `protobuf:"varint,15,opt,name=roleOnly,def=-1" json:"roleOnly,omitempty"`
	// * 限制体验的武将
	TasteHeroID *int32 `protobuf:"varint,39,opt,name=tasteHeroID,def=0" json:"tasteHeroID,omitempty"`
	// * 关卡排序
	LevelIndex *int32 `protobuf:"varint,16,opt,name=levelIndex,def=0" json:"levelIndex,omitempty"`
	// * 精英关卡排序
	EliteLevelIndex *int32 `protobuf:"varint,17,opt,name=eliteLevelIndex,def=0" json:"eliteLevelIndex,omitempty"`
	// * 地狱关卡排序
	HardLevelIndex *int32 `protobuf:"varint,18,opt,def=0" json:"HardLevelIndex,omitempty"`
	// * 关卡城镇
	LevelTown *string `protobuf:"bytes,19,opt,name=levelTown,def=" json:"levelTown,omitempty"`
	// * 玩法ID
	GameModeID *int32 `protobuf:"varint,20,opt,def=0" json:"GameModeID,omitempty"`
	// * 章节名称
	ChapterNameIDS *string `protobuf:"bytes,21,opt,def=" json:"ChapterNameIDS,omitempty"`
	// * 关卡名称
	LevelNameIDS *string `protobuf:"bytes,22,opt,name=levelNameIDS,def=" json:"levelNameIDS,omitempty"`
	// * 关卡地名
	PlaceName *string `protobuf:"bytes,23,opt,def=" json:"PlaceName,omitempty"`
	// * 场景名称
	SceneID *string `protobuf:"bytes,24,opt,def=" json:"SceneID,omitempty"`
	// * 场景文件夹
	SceneFile *string `protobuf:"bytes,25,opt,def=" json:"SceneFile,omitempty"`
	// * 缩略图名
	SmallBgID *string `protobuf:"bytes,41,opt,def=" json:"SmallBgID,omitempty"`
	// * 户型图名
	PreviewID *string `protobuf:"bytes,42,opt,def=" json:"PreviewID,omitempty"`
	// * 关卡小地图
	Radar *string `protobuf:"bytes,26,opt,def=" json:"Radar,omitempty"`
	// * 关卡预览图
	PreviewImg *string `protobuf:"bytes,27,opt,def=" json:"PreviewImg,omitempty"`
	// * 加载的loadingtips，空着表示走全局模式；单独使用则用英文逗号隔开，随机抽
	LoadingTipsID *string `protobuf:"bytes,40,opt,def=" json:"LoadingTipsID,omitempty"`
	// * 场景加载图
	SceneLoading *string `protobuf:"bytes,28,opt,def=" json:"SceneLoading,omitempty"`
	// * 未通关情况加载图
	FirstAccessLoading *string `protobuf:"bytes,29,opt,def=" json:"FirstAccessLoading,omitempty"`
	// * 首次通关回城加载图
	FirstPassLoading *string `protobuf:"bytes,30,opt,def=" json:"FirstPassLoading,omitempty"`
	// * 关卡重置次数购买规则
	LevelPurchasePolicy *string `protobuf:"bytes,31,opt,def=" json:"LevelPurchasePolicy,omitempty"`
	// * 实际推荐战力
	LevelGS *uint32 `protobuf:"varint,32,opt,def=0" json:"LevelGS,omitempty"`
	// * 显示用推荐战力
	DisplayGS *uint32 `protobuf:"varint,33,opt,def=0" json:"DisplayGS,omitempty"`
	// * 关卡展示BOSS
	EnemyBoss *string `protobuf:"bytes,34,opt,def=" json:"EnemyBoss,omitempty"`
	// * 为最高通关关卡时，战斗结束是否回城，默认回战役界面，1=回到城镇
	BackToTown *uint32 `protobuf:"varint,35,opt,def=0" json:"BackToTown,omitempty"`
	// * 解锁名将ID
	UnlockHeroID      *string                 `protobuf:"bytes,36,opt,def=" json:"UnlockHeroID,omitempty"`
	DropItem_Template []*LEVEL_INFO_DropItems `protobuf:"bytes,37,rep" json:"DropItem_Template,omitempty"`
	XXX_unrecognized  []byte                  `json:"-"`
}

func (m *LEVEL_INFO) Reset()         { *m = LEVEL_INFO{} }
func (m *LEVEL_INFO) String() string { return proto.CompactTextString(m) }
func (*LEVEL_INFO) ProtoMessage()    {}

const Default_LEVEL_INFO_LevelType int32 = 0
const Default_LEVEL_INFO_TerrainType int32 = 0
const Default_LEVEL_INFO_Energy int32 = 0
const Default_LEVEL_INFO_HighEnergy int32 = 0
const Default_LEVEL_INFO_SweepItemCost uint32 = 0
const Default_LEVEL_INFO_MaxDailyAccess int32 = 0
const Default_LEVEL_INFO_FinishType int32 = 0
const Default_LEVEL_INFO_TimeLimit int32 = 0
const Default_LEVEL_INFO_TimeGoal int32 = 0
const Default_LEVEL_INFO_HpGoal float32 = 0
const Default_LEVEL_INFO_TeamRequirement int32 = 0
const Default_LEVEL_INFO_LevelRequirement int32 = 0
const Default_LEVEL_INFO_RoleOnly int32 = -1
const Default_LEVEL_INFO_TasteHeroID int32 = 0
const Default_LEVEL_INFO_LevelIndex int32 = 0
const Default_LEVEL_INFO_EliteLevelIndex int32 = 0
const Default_LEVEL_INFO_HardLevelIndex int32 = 0
const Default_LEVEL_INFO_GameModeID int32 = 0
const Default_LEVEL_INFO_LevelGS uint32 = 0
const Default_LEVEL_INFO_DisplayGS uint32 = 0
const Default_LEVEL_INFO_BackToTown uint32 = 0

func (m *LEVEL_INFO) GetLevelID() string {
	if m != nil && m.LevelID != nil {
		return *m.LevelID
	}
	return ""
}

func (m *LEVEL_INFO) GetChapterShorthand() string {
	if m != nil && m.ChapterShorthand != nil {
		return *m.ChapterShorthand
	}
	return ""
}

func (m *LEVEL_INFO) GetChapterID() string {
	if m != nil && m.ChapterID != nil {
		return *m.ChapterID
	}
	return ""
}

func (m *LEVEL_INFO) GetLevelType() int32 {
	if m != nil && m.LevelType != nil {
		return *m.LevelType
	}
	return Default_LEVEL_INFO_LevelType
}

func (m *LEVEL_INFO) GetTerrainType() int32 {
	if m != nil && m.TerrainType != nil {
		return *m.TerrainType
	}
	return Default_LEVEL_INFO_TerrainType
}

func (m *LEVEL_INFO) GetEnergy() int32 {
	if m != nil && m.Energy != nil {
		return *m.Energy
	}
	return Default_LEVEL_INFO_Energy
}

func (m *LEVEL_INFO) GetHighEnergy() int32 {
	if m != nil && m.HighEnergy != nil {
		return *m.HighEnergy
	}
	return Default_LEVEL_INFO_HighEnergy
}

func (m *LEVEL_INFO) GetSweepItem() string {
	if m != nil && m.SweepItem != nil {
		return *m.SweepItem
	}
	return ""
}

func (m *LEVEL_INFO) GetSweepItemCost() uint32 {
	if m != nil && m.SweepItemCost != nil {
		return *m.SweepItemCost
	}
	return Default_LEVEL_INFO_SweepItemCost
}

func (m *LEVEL_INFO) GetMaxDailyAccess() int32 {
	if m != nil && m.MaxDailyAccess != nil {
		return *m.MaxDailyAccess
	}
	return Default_LEVEL_INFO_MaxDailyAccess
}

func (m *LEVEL_INFO) GetFinishType() int32 {
	if m != nil && m.FinishType != nil {
		return *m.FinishType
	}
	return Default_LEVEL_INFO_FinishType
}

func (m *LEVEL_INFO) GetTimeLimit() int32 {
	if m != nil && m.TimeLimit != nil {
		return *m.TimeLimit
	}
	return Default_LEVEL_INFO_TimeLimit
}

func (m *LEVEL_INFO) GetTimeGoal() int32 {
	if m != nil && m.TimeGoal != nil {
		return *m.TimeGoal
	}
	return Default_LEVEL_INFO_TimeGoal
}

func (m *LEVEL_INFO) GetHpGoal() float32 {
	if m != nil && m.HpGoal != nil {
		return *m.HpGoal
	}
	return Default_LEVEL_INFO_HpGoal
}

func (m *LEVEL_INFO) GetPreLevelID() string {
	if m != nil && m.PreLevelID != nil {
		return *m.PreLevelID
	}
	return ""
}

func (m *LEVEL_INFO) GetTeamRequirement() int32 {
	if m != nil && m.TeamRequirement != nil {
		return *m.TeamRequirement
	}
	return Default_LEVEL_INFO_TeamRequirement
}

func (m *LEVEL_INFO) GetLevelRequirement() int32 {
	if m != nil && m.LevelRequirement != nil {
		return *m.LevelRequirement
	}
	return Default_LEVEL_INFO_LevelRequirement
}

func (m *LEVEL_INFO) GetRoleOnly() int32 {
	if m != nil && m.RoleOnly != nil {
		return *m.RoleOnly
	}
	return Default_LEVEL_INFO_RoleOnly
}

func (m *LEVEL_INFO) GetTasteHeroID() int32 {
	if m != nil && m.TasteHeroID != nil {
		return *m.TasteHeroID
	}
	return Default_LEVEL_INFO_TasteHeroID
}

func (m *LEVEL_INFO) GetLevelIndex() int32 {
	if m != nil && m.LevelIndex != nil {
		return *m.LevelIndex
	}
	return Default_LEVEL_INFO_LevelIndex
}

func (m *LEVEL_INFO) GetEliteLevelIndex() int32 {
	if m != nil && m.EliteLevelIndex != nil {
		return *m.EliteLevelIndex
	}
	return Default_LEVEL_INFO_EliteLevelIndex
}

func (m *LEVEL_INFO) GetHardLevelIndex() int32 {
	if m != nil && m.HardLevelIndex != nil {
		return *m.HardLevelIndex
	}
	return Default_LEVEL_INFO_HardLevelIndex
}

func (m *LEVEL_INFO) GetLevelTown() string {
	if m != nil && m.LevelTown != nil {
		return *m.LevelTown
	}
	return ""
}

func (m *LEVEL_INFO) GetGameModeID() int32 {
	if m != nil && m.GameModeID != nil {
		return *m.GameModeID
	}
	return Default_LEVEL_INFO_GameModeID
}

func (m *LEVEL_INFO) GetChapterNameIDS() string {
	if m != nil && m.ChapterNameIDS != nil {
		return *m.ChapterNameIDS
	}
	return ""
}

func (m *LEVEL_INFO) GetLevelNameIDS() string {
	if m != nil && m.LevelNameIDS != nil {
		return *m.LevelNameIDS
	}
	return ""
}

func (m *LEVEL_INFO) GetPlaceName() string {
	if m != nil && m.PlaceName != nil {
		return *m.PlaceName
	}
	return ""
}

func (m *LEVEL_INFO) GetSceneID() string {
	if m != nil && m.SceneID != nil {
		return *m.SceneID
	}
	return ""
}

func (m *LEVEL_INFO) GetSceneFile() string {
	if m != nil && m.SceneFile != nil {
		return *m.SceneFile
	}
	return ""
}

func (m *LEVEL_INFO) GetSmallBgID() string {
	if m != nil && m.SmallBgID != nil {
		return *m.SmallBgID
	}
	return ""
}

func (m *LEVEL_INFO) GetPreviewID() string {
	if m != nil && m.PreviewID != nil {
		return *m.PreviewID
	}
	return ""
}

func (m *LEVEL_INFO) GetRadar() string {
	if m != nil && m.Radar != nil {
		return *m.Radar
	}
	return ""
}

func (m *LEVEL_INFO) GetPreviewImg() string {
	if m != nil && m.PreviewImg != nil {
		return *m.PreviewImg
	}
	return ""
}

func (m *LEVEL_INFO) GetLoadingTipsID() string {
	if m != nil && m.LoadingTipsID != nil {
		return *m.LoadingTipsID
	}
	return ""
}

func (m *LEVEL_INFO) GetSceneLoading() string {
	if m != nil && m.SceneLoading != nil {
		return *m.SceneLoading
	}
	return ""
}

func (m *LEVEL_INFO) GetFirstAccessLoading() string {
	if m != nil && m.FirstAccessLoading != nil {
		return *m.FirstAccessLoading
	}
	return ""
}

func (m *LEVEL_INFO) GetFirstPassLoading() string {
	if m != nil && m.FirstPassLoading != nil {
		return *m.FirstPassLoading
	}
	return ""
}

func (m *LEVEL_INFO) GetLevelPurchasePolicy() string {
	if m != nil && m.LevelPurchasePolicy != nil {
		return *m.LevelPurchasePolicy
	}
	return ""
}

func (m *LEVEL_INFO) GetLevelGS() uint32 {
	if m != nil && m.LevelGS != nil {
		return *m.LevelGS
	}
	return Default_LEVEL_INFO_LevelGS
}

func (m *LEVEL_INFO) GetDisplayGS() uint32 {
	if m != nil && m.DisplayGS != nil {
		return *m.DisplayGS
	}
	return Default_LEVEL_INFO_DisplayGS
}

func (m *LEVEL_INFO) GetEnemyBoss() string {
	if m != nil && m.EnemyBoss != nil {
		return *m.EnemyBoss
	}
	return ""
}

func (m *LEVEL_INFO) GetBackToTown() uint32 {
	if m != nil && m.BackToTown != nil {
		return *m.BackToTown
	}
	return Default_LEVEL_INFO_BackToTown
}

func (m *LEVEL_INFO) GetUnlockHeroID() string {
	if m != nil && m.UnlockHeroID != nil {
		return *m.UnlockHeroID
	}
	return ""
}

func (m *LEVEL_INFO) GetDropItem_Template() []*LEVEL_INFO_DropItems {
	if m != nil {
		return m.DropItem_Template
	}
	return nil
}

type LEVEL_INFO_DropItems struct {
	// * 物品ID
	DropItemID       *string `protobuf:"bytes,1,opt,def=" json:"DropItemID,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *LEVEL_INFO_DropItems) Reset()         { *m = LEVEL_INFO_DropItems{} }
func (m *LEVEL_INFO_DropItems) String() string { return proto.CompactTextString(m) }
func (*LEVEL_INFO_DropItems) ProtoMessage()    {}

func (m *LEVEL_INFO_DropItems) GetDropItemID() string {
	if m != nil && m.DropItemID != nil {
		return *m.DropItemID
	}
	return ""
}

type LEVEL_INFO_ARRAY struct {
	Items            []*LEVEL_INFO `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_unrecognized []byte        `json:"-"`
}

func (m *LEVEL_INFO_ARRAY) Reset()         { *m = LEVEL_INFO_ARRAY{} }
func (m *LEVEL_INFO_ARRAY) String() string { return proto.CompactTextString(m) }
func (*LEVEL_INFO_ARRAY) ProtoMessage()    {}

func (m *LEVEL_INFO_ARRAY) GetItems() []*LEVEL_INFO {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
}
