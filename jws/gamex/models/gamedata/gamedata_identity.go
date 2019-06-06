package gamedata

import (
	"github.com/golang/protobuf/proto"
	"strings"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// identity 数据

var (
	gdIdentity     map[string]*ProtobufGen.IDENTITY
	gdCounterSkill map[string]*ProtobufGen.COUNTERSKILL
	gdPassiveSkill map[string]*ProtobufGen.PASSIVESKILL
	gdTriggerSkill map[string]*ProtobufGen.TRIGGERSKILL

	gdIdentityData     map[string]*IdentityData
	gdCounterSkillData map[string]*CounterSkillData
	gdPassiveSkillData map[string]*PassiveSkillData
	gdTriggerSkillData map[string]*TriggerSkillData
	gdSkillId2Cost     map[string]*CostData
	gdSkillId2IDID     map[string]string
)

type IdentityData struct {
	ProtobufGen.IDENTITY
	PassiveSkill *PassiveSkillData
	CounterSkill *CounterSkillData
	TriggerSkill *TriggerSkillData
}

type PassiveSkillData struct {
	ProtobufGen.PASSIVESKILL
}

type CounterSkillData struct {
	ProtobufGen.COUNTERSKILL
}

type TriggerSkillData struct {
	ProtobufGen.TRIGGERSKILL
}

type SkillCost struct {
	ActivationItemId    int
	ActivationItemCount int
}

//所有英雄的所有技能
type HeroSkills struct {
	PassiveSkill map[string]struct{}
	CounterSkill map[string]struct{}
	TriggerSkill map[string]struct{}
}

const (
	IdentityPropTypeCamp       = "Camp"
	IdentityPropTypeGender     = "Gender"
	IdentityPropTypeCombatType = "CombatType"
	IdentityPropTypeWeaponType = "WeaponType"
)

var p HeroSkills

/*
Camp - 阵营；wei=魏，shu=蜀，wu=吴，qun=群
Gender - 性别; male=男，female=女
CombatType - 物法；wuli=物理，fashu=法术
WeaponType - 兵种；暂时留空，先填long
*/
func (c *CounterSkillData) IfTakeEffect(i *IdentityData) bool {
	switch c.GetTargetProp() {
	case IdentityPropTypeCamp:
		return c.GetTargetPropValue() == i.GetCamp()
	case IdentityPropTypeGender:
		return c.GetTargetPropValue() == i.GetGender()
	case IdentityPropTypeCombatType:
		return c.GetTargetPropValue() == i.GetCombatType()
	case IdentityPropTypeWeaponType:
		return c.GetTargetPropValue() == i.GetWeaponType()
	default:
		return false
	}
}

// GetIdentityData
func GetIdentityData(idId string) *IdentityData {
	r, ok := gdIdentityData[idId]
	if !ok || r == nil {
		logs.Error(
			"GetIdentity Err By %s",
			idId)
		return nil
	}
	return r
}

// GetCounterSkillData
func GetCounterSkillData(counterSkillId string) *CounterSkillData {
	r, ok := gdCounterSkillData[counterSkillId]
	if !ok || r == nil {
		logs.Error(
			"GetCounterSkill Err By %s",
			counterSkillId)
		return nil
	}
	return r
}

// GetPassiveSkillData
func GetPassiveSkillData(passiveSkillId string) *PassiveSkillData {
	r, ok := gdPassiveSkillData[passiveSkillId]
	if !ok || r == nil {
		logs.Error(
			"GetPassiveSkillData Err By %s",
			passiveSkillId)
		return nil
	}
	return r
}

// GetTriggerSkillData
//TODO

//GetCostData

func GetSkillCostData(skillid string) *CostData {
	return gdSkillId2Cost[skillid]
}

func GetIdIdBySkillID(skillid string) string {
	return gdSkillId2IDID[skillid]
}

func loadIdentityData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.IDENTITY_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	data := ar.GetItems()
	gdIdentity = make(map[string]*ProtobufGen.IDENTITY, len(data))
	gdSkillId2IDID = make(map[string]string, 64)
	for _, c := range data {
		gdIdentity[c.GetIdid()] = c

		//整理 Skillid2IDID
		if c.GetCounterSkillId() != "" {
			for _, s := range strings.Split(c.GetCounterSkillId(), ",") {
				gdSkillId2IDID[s] = c.GetIdid()
			}
		} else if c.GetTriggerSkillId() != "" {
			for _, x := range strings.Split(c.GetTriggerSkillId(), ",") {
				gdSkillId2IDID[x] = c.GetIdid()
			}

		} else if c.GetPassiveSkillId() != "" {
			for _, f := range strings.Split(c.GetPassiveSkillId(), ",") {
				gdSkillId2IDID[f] = c.GetIdid()
			}
		}

	}
}

func loadCounterSkillData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.COUNTERSKILL_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	data := ar.GetItems()
	gdCounterSkill = make(map[string]*ProtobufGen.COUNTERSKILL, len(data))

	for _, c := range data {
		gdCounterSkill[c.GetCounterSkillId()] = c
	}
}

func loadPassiveSkillData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.PASSIVESKILL_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	data := ar.GetItems()
	gdPassiveSkill = make(map[string]*ProtobufGen.PASSIVESKILL, len(data))

	for _, c := range data {
		gdPassiveSkill[c.GetPassiveSkillId()] = c
	}
}

func loadTriggerSkillData(filepath string) {
	buffer, err := loadBin(filepath)
	panicIfErr(err)

	ar := new(ProtobufGen.TRIGGERSKILL_ARRAY)
	panicIfErr(proto.Unmarshal(buffer, ar))

	data := ar.GetItems()
	gdTriggerSkill = make(map[string]*ProtobufGen.TRIGGERSKILL, len(data))

	for _, c := range data {
		gdTriggerSkill[c.GetSkillID()] = c
	}

}

func mkIdentityDatas(loadFunc func(dfilepath string, loadfunc func(string))) {
	loadFunc("identity.data", loadIdentityData)
	loadFunc("counterskill.data", loadCounterSkillData)
	loadFunc("passiveskill.data", loadPassiveSkillData)
	loadFunc("triggerskill.data", loadTriggerSkillData)

	gdPassiveSkillData = make(
		map[string]*PassiveSkillData,
		len(gdPassiveSkill))
	for id, data := range gdPassiveSkill {
		n := new(PassiveSkillData)
		n.PASSIVESKILL = *data
		gdPassiveSkillData[id] = n
	}

	gdCounterSkillData = make(
		map[string]*CounterSkillData,
		len(gdCounterSkill))
	for id, data := range gdCounterSkill {
		n := new(CounterSkillData)
		n.COUNTERSKILL = *data
		gdCounterSkillData[id] = n
	}

	gdTriggerSkillData = make(map[string]*TriggerSkillData,
		len(gdTriggerSkill))
	for id, data := range gdTriggerSkill {
		n := new(TriggerSkillData)
		n.TRIGGERSKILL = *data
		gdTriggerSkillData[id] = n
	}

	gdIdentityData = make(
		map[string]*IdentityData,
		len(gdIdentity))
	for id, data := range gdIdentity {
		n := new(IdentityData)
		n.IDENTITY = *data
		nPassiveSkill, ok := gdPassiveSkillData[n.GetPassiveSkillId()]
		if ok && nPassiveSkill != nil {
			n.PassiveSkill = nPassiveSkill
		}
		nCounterSkill, ok := gdCounterSkillData[n.GetCounterSkillId()]
		if ok && nCounterSkill != nil {
			n.CounterSkill = nCounterSkill
		}
		nTriggerSkill, ok := gdTriggerSkillData[n.GetTriggerSkillId()]
		if ok && nTriggerSkill != nil {
			n.TriggerSkill = nTriggerSkill
		}
		gdIdentityData[id] = n
	}

	logs.Trace("gdIdentityData %v", gdIdentityData)
	logs.Trace("gdCounterSkill %v", gdCounterSkillData)
	logs.Trace("gdPassiveSkill %v", gdPassiveSkillData)

	//整理SkillCost
	gdSkillId2Cost = make(map[string]*CostData, len(gdCounterSkill)+len(gdPassiveSkill)+len(gdTriggerSkill))
	p.CounterSkill = make(map[string]struct{}, 100)
	p.PassiveSkill = make(map[string]struct{}, 100)
	p.TriggerSkill = make(map[string]struct{}, 100)
	for k, v := range gdCounterSkill {
		data := &CostData{}
		data.AddItem(v.GetActivationItemId(), uint32(v.GetActivationItemCount()))
		gdSkillId2Cost[k] = data
		p.CounterSkill[v.GetCounterSkillId()] = struct{}{}
	}
	for k, v := range gdPassiveSkill {
		data := &CostData{}
		data.AddItem(v.GetActivationItemId(), uint32(v.GetActivationItemCount()))
		gdSkillId2Cost[k] = data
		p.PassiveSkill[v.GetPassiveSkillId()] = struct{}{}
	}
	for k, v := range gdTriggerSkill {
		data := &CostData{}
		data.AddItem(v.GetActivationItemId(), uint32(v.GetActivationItemCount()))
		gdSkillId2Cost[k] = data
		p.TriggerSkill[v.GetSkillID()] = struct{}{}
	}

}

// 通过skill第查看技能属于哪种,返回1属于Triggerskill,返回2属于Passiveskill,返回3属于Counterskill
const (
	_      = iota
	Tkill  = "Tkill"
	Pkill  = "Pkill"
	Ckill  = "Ckill"
	NOkill = "NOkill"
)

func GetWhichSkillBySkillId(skillid string) string {
	_, ok := p.TriggerSkill[skillid]
	if ok {
		return Tkill
	}

	_, ok = p.PassiveSkill[skillid]
	if ok {
		return Pkill
	}
	_, ok = p.CounterSkill[skillid]
	if ok {
		return Ckill
	}

	return NOkill
}

func GetpskillByidid(idid string) string {
	result := gdIdentity[idid].GetPassiveSkillId()
	return result

}

func GetcskillByidid(idid string) string {
	result := gdIdentity[idid].GetCounterSkillId()
	return result

}

func GettskillByidid(idid string) string {
	result := gdIdentity[idid].GetTriggerSkillId()
	return result
}
