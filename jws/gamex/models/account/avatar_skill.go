package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type AvatarSkill struct {
	Skills           [AVATAR_NUM_MAX][AVATAR_SKILL_MAX]uint32 `json:"s"`
	SkillPractices   [CORP_SKILLPRACTICE_MAX]uint32           `json:"sp"`
	isHasSkillUnlock bool
}

func (p *AvatarSkill) IsHasSkillUnlockThenNeedSync() bool {
	return p.isHasSkillUnlock
}

func (p *AvatarSkill) SetSkillUnlockHasSync() {
	p.isHasSkillUnlock = false
}

func (p *AvatarSkill) AddSkill(avatar_id int, skill_idx int, lv_add uint32) bool {
	logs.Trace("Avatar AddSkill %d %d", avatar_id, lv_add)
	if avatar_id >= len(p.Skills) {
		logs.Error("No Avatar Type %d", avatar_id)
		return false
	}

	if skill_idx >= len(p.Skills[avatar_id]) {
		logs.Error("No AvatarSkill Type %d", skill_idx)
		return false
	}

	p.Skills[avatar_id][skill_idx] += lv_add
	return true
}

func (p *AvatarSkill) AddSkillNoInit(avatar_id int, skill_idx int, lv_add uint32) bool {
	if avatar_id >= len(p.Skills) {
		logs.Error("No Avatar Type %d", avatar_id)
		return false
	}

	if skill_idx >= len(p.Skills[avatar_id]) {
		logs.Error("No AvatarSkill Type %d", skill_idx)
		return false
	}

	if p.Skills[avatar_id][skill_idx] <= 0 {
		p.Skills[avatar_id][skill_idx] += lv_add
	}
	return true
}

func (p *AvatarSkill) SetSkillLv(avatar_id int, skill_idx int, lv uint32) bool {
	logs.Trace("Avatar %d SetSkillLv %d", avatar_id, lv)
	if avatar_id >= len(p.Skills) {
		logs.Error("No Avatar Type %d", avatar_id)
		return false
	}

	if skill_idx >= len(p.Skills[avatar_id]) {
		logs.Error("No AvatarSkill Type %d", skill_idx)
		return false
	}

	p.Skills[avatar_id][skill_idx] = lv
	return true
}

func (p *AvatarSkill) UnlockSkill(avatar_id int, skill_idx int) bool {
	logs.Trace("Avatar UnlockSkill %d %d", avatar_id, skill_idx)
	p.isHasSkillUnlock = true
	return p.AddSkill(avatar_id, skill_idx, 1)
}

func (p *AvatarSkill) Get(avatar_id int, skill_idx int) uint32 {
	if avatar_id >= len(p.Skills) {
		logs.Error("No Avatar Type %d", avatar_id)
		return 0
	}

	if skill_idx >= len(p.Skills[avatar_id]) {
		logs.Error("No AvatarSkill Type %d", skill_idx)
		return 0
	}

	return p.Skills[avatar_id][skill_idx]
}

func (p *AvatarSkill) GetByAvatar(avatar_id int) []uint32 {
	if avatar_id >= len(p.Skills) {
		logs.Error("No Avatar Type %d", avatar_id)
		return []uint32{}[:]
	}

	return p.Skills[avatar_id][:]
}

func (p *AvatarSkill) GetAll() ([]uint32, int) {
	lvs := make([]uint32, 0, AVATAR_NUM_MAX*AVATAR_SKILL_MAX)
	for i := 0; i < len(p.Skills); i++ {
		for j := 0; j < len(p.Skills[i]); j++ {
			lvs = append(lvs, p.Skills[i][j])
		}
	}
	return lvs[:], AVATAR_SKILL_MAX
}

func (p *AvatarSkill) AddPracticeLevel(idx int) bool {
	logs.Trace("Corp %d AddSkill %d", idx, 1)
	if idx >= len(p.SkillPractices) {
		logs.Error("No Avatar Type %d", idx)
		return false
	}

	data := gamedata.GetSkillPracticeLevelInfo(idx)
	if data == nil {
		return false
	}

	p.SkillPractices[idx] += 1

	toLv := int(p.SkillPractices[idx])

	// 因为客户端需要分别取各个角色的技能等级, 所以这里保留原来分角色的结构,
	// 技能修炼等级升级时更新对应的技能等级值
	if toLv < len(data.RelationSkills) && toLv < len(data.SkillLv) {
		rs := data.RelationSkills[toLv]
		for _, skill := range rs {
			p.SetSkillLv(
				skill.AvatarId,
				skill.SkillId,
				data.SkillLv[toLv])
		}
	}
	return true
}

func (p *AvatarSkill) GetPracticeLevel() []uint32 {
	return p.SkillPractices[:]
}

func (p *AvatarSkill) OnAfterLogin() {
	unlockSkills := gamedata.GetLevelUnlockSkills(1)
	if unlockSkills != nil && len(unlockSkills) > 0 {
		for _, skill := range unlockSkills {
			p.AddSkillNoInit(skill.AvatarId, skill.SkillId, 1)
		}
	}
}
