package gamedata

import (
	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type skillLvData struct {
	CostToThisLv []CostData
	SkillGS      []uint32
	AvatarLvNeed []uint32
	SkillId      string
	CDTime       []float32
}

func (s *skillLvData) AddData(data *ProtobufGen.SKILLUPGRADE) {
	if s.CostToThisLv == nil {
		s.CostToThisLv = make([]CostData, 64, 128)
	}

	if s.SkillGS == nil {
		s.SkillGS = make([]uint32, 64, 128)
	}

	if s.AvatarLvNeed == nil {
		s.AvatarLvNeed = make([]uint32, 64, 128)
	}

	if s.CDTime == nil {
		s.CDTime = make([]float32, 64, 128)
	}

	lv := int(data.GetSkillLevel())
	for lv >= len(s.CostToThisLv) {
		s.CostToThisLv = append(s.CostToThisLv, CostData{})
		s.SkillGS = append(s.SkillGS, 0)
		s.AvatarLvNeed = append(s.AvatarLvNeed, 0)
		s.CDTime = append(s.CDTime, 0)
	}

	s.SkillGS[lv] = data.GetSkillGS()
	s.AvatarLvNeed[lv] = data.GetUnlockLevel()
	s.CostToThisLv[lv].AddItem(data.GetSkillCoin(), data.GetSkillCost())
	s.SkillId = data.GetSkillID()
	s.CDTime[lv] = data.GetCDTime()

}

type avatarSkillInfo struct {
	AvatarId int
	SkillId  int
}

var (
	gdSkillLevelInfo   [AVATAR_NUM_MAX][AVATAR_SKILL_MAX]skillLvData
	gdUnlockSkillLevel [][]avatarSkillInfo
)

func addUnlockSkillLevel(lv, avatar_id, skill_id int) {
	if gdUnlockSkillLevel == nil {
		gdUnlockSkillLevel = make([][]avatarSkillInfo, 0, 64) // 只有前面的等级会解锁
	}

	for len(gdUnlockSkillLevel) <= lv {
		gdUnlockSkillLevel = append(gdUnlockSkillLevel, []avatarSkillInfo{})
	}

	if gdUnlockSkillLevel[lv] == nil {
		gdUnlockSkillLevel[lv] = make([]avatarSkillInfo, 0, 32) // 同一级解锁的也不多
	}

	gdUnlockSkillLevel[lv] = append(gdUnlockSkillLevel[lv],
		avatarSkillInfo{avatar_id, skill_id})
}

func loadSkillLevelInfo(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.SKILLUPGRADE_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()

	for _, c := range lv_data {
		idx := c.GetSkillIndex()
		//
		// 注意 表中 配置的格式是 avatarId * 100 + skillid + 1
		// avatarId 从零开始
		avatar_id := idx / 100
		skill_idx := idx%100 - 1

		if avatar_id >= AVATAR_NUM_MAX {
			logs.Error("avatar_id error by lv %d", avatar_id)
			continue
		}

		if skill_idx >= AVATAR_SKILL_MAX {
			logs.Error("skill_idx error by lv %d", skill_idx)
			continue
		}

		gdSkillLevelInfo[avatar_id][skill_idx].AddData(c)
	}

	// 填充解锁索引数据
	for avatar := 0; avatar < AVATAR_NUM_CURR; avatar++ {
		for skill := 0; skill < AVATAR_SKILL_MAX; skill++ {
			info := gdSkillLevelInfo[avatar][skill]
			if info.AvatarLvNeed != nil && len(info.AvatarLvNeed) >= 2 {
				addUnlockSkillLevel(int(info.AvatarLvNeed[1]), avatar, skill)
			}
		}
	}

	//logs.Trace("addUnlockSkillLevel %v", gdUnlockSkillLevel)

	//logs.Trace("gdSkillLevelInfo %v", gdSkillLevelInfo)
}

func GetLevelUnlockSkills(lv int) []avatarSkillInfo {
	if lv < 0 || lv >= len(gdUnlockSkillLevel) {
		return nil
	}
	infos := gdUnlockSkillLevel[lv]
	if infos == nil {
		return nil
	} else {
		return infos[:]
	}
}

func GetSkillLevelConfig(avatar_id int, skill_idx int) *skillLvData {
	if avatar_id >= AVATAR_NUM_MAX {
		logs.Error("No Avatar Type %d", avatar_id)
		return nil
	}

	if skill_idx >= AVATAR_SKILL_MAX {
		logs.Error("No AvatarSkill Type %d", skill_idx)
		return nil
	}
	return &gdSkillLevelInfo[avatar_id][skill_idx]
}

// 新的技能修炼等级逻辑

type skillPracticeLvData struct {
	CostToThisLv []CostData
	Idx          uint32
	ATK          []uint32
	DEF          []uint32
	HP           []uint32

	SkillLv        []uint32
	RelationSkills [][]avatarSkillInfo
}

func (s *skillPracticeLvData) AddData(data *ProtobufGen.SKILLPRACTICE) {
	if s.CostToThisLv == nil {
		s.CostToThisLv = make([]CostData, 0, 200)
	}

	if s.ATK == nil {
		s.ATK = make([]uint32, 0, 200)
	}
	if s.DEF == nil {
		s.DEF = make([]uint32, 0, 200)
	}
	if s.HP == nil {
		s.HP = make([]uint32, 0, 200)
	}

	if s.SkillLv == nil {
		s.SkillLv = make([]uint32, 0, 200)
	}

	if s.RelationSkills == nil {
		s.RelationSkills = make([][]avatarSkillInfo, 0, 200)
	}

	lv := int(data.GetPracticeLevel())
	for lv >= len(s.CostToThisLv) {
		s.CostToThisLv = append(s.CostToThisLv, CostData{})
		s.ATK = append(s.ATK, 0)
		s.DEF = append(s.DEF, 0)
		s.HP = append(s.HP, 0)
		s.SkillLv = append(s.SkillLv, 0)
		s.RelationSkills = append(s.RelationSkills, make([]avatarSkillInfo, 0, 8)[:])
	}

	s.ATK[lv] = data.GetATK()
	s.DEF[lv] = data.GetDEF()
	s.HP[lv] = data.GetHP()
	s.SkillLv[lv] = data.GetSkillLevel()
	s.CostToThisLv[lv].AddItem(data.GetCoinItem(), data.GetCoinCount())
	s.Idx = data.GetPracticeID() - 1
	rs := data.GetRelationSkill_Template()
	for _, sid := range rs {
		skillIDx := sid.GetSkillIndex()
		//
		// 注意 表中 配置的格式是 avatarId * 100 + skillid + 1
		// avatarId 从零开始
		avatar_id := skillIDx / 100
		skill_idx := skillIDx%100 - 1

		if avatar_id >= AVATAR_NUM_MAX {
			logs.Error("avatar_id error by lv %d", avatar_id)
			continue
		}

		if skill_idx >= AVATAR_SKILL_MAX {
			logs.Error("skill_idx error by lv %d", skill_idx)
			continue
		}

		s.RelationSkills[lv] = append(s.RelationSkills[lv], avatarSkillInfo{
			AvatarId: int(avatar_id),
			SkillId:  int(skill_idx),
		})
	}

}

var (
	gdSkillPracticeLevelInfo [CORP_SKILLPRACTICE_MAX]skillPracticeLvData
)

func loadSkillPracticeLevelInfo(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	lv_ar := &ProtobufGen.SKILLPRACTICE_ARRAY{}
	err = proto.Unmarshal(buffer, lv_ar)
	errcheck(err)

	lv_data := lv_ar.GetItems()

	for _, c := range lv_data {
		idx := c.GetPracticeID()
		gdSkillPracticeLevelInfo[idx-1].AddData(c)
	}

	logs.Trace("gdSkillPracticeLevelInfo %v", gdSkillPracticeLevelInfo)
}

func GetSkillPracticeLevelInfo(idx int) *skillPracticeLvData {
	if idx < 0 || idx >= len(gdSkillPracticeLevelInfo) {
		return nil
	}
	return &gdSkillPracticeLevelInfo[idx]
}
