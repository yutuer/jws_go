package gamedata

import (
	"fmt"
	"strconv"

	"github.com/golang/protobuf/proto"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
)

const (
	FteConditionRoleOpenTypStage  uint32 = 1
	FteConditionRoleOpenTypCorpLv uint32 = 0
)

type fteCondRole struct {
	Typ  uint32 // 开放的方式 	1=通过关卡	0=战队等级
	Pstr string // 开放的条件 string
	Pint int    // 开放的条件 int
}

// TODO By FanYang 验证所有角色的解锁条件都是有效的
var (
	gdFteConditionRole [AVATAR_NUM_CURR]fteCondRole
	gdEquipJade        map[uint32]map[int]uint32
)

func loadFteConditionRoleConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.CONDITIONROLE_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	item_data := ar.GetItems()
	for _, a := range item_data {
		roleID := int(a.GetRoleID())
		if roleID < 0 || roleID >= AVATAR_NUM_CURR {
			panic(fmt.Errorf("loadFteConditionRoleConfig roleID Err By %d %v", roleID, a.GetTipIDS()))
		}
		typ := a.GetConditionType()
		gdFteConditionRole[roleID].Pstr = a.GetConditionValue()
		if typ == FteConditionRoleOpenTypCorpLv {
			f, err := strconv.ParseFloat(a.GetConditionValue(), 32)
			gdFteConditionRole[roleID].Pint = int(f)
			if err != nil {
				panic(err)
			}
		}
		gdFteConditionRole[roleID].Typ = typ
	}

	//logs.Trace("gdEquipStarData %v", gdEquipStarData)
}

func GetAvatarOpenCondData(avatarID int) *fteCondRole {
	if avatarID < 0 || avatarID >= len(gdFteConditionRole) {
		return nil
	}

	return &gdFteConditionRole[avatarID]
}

func loadJadeConditionConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.JADECONDITON_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	item_data := ar.GetItems()

	gdEquipJade = make(map[uint32]map[int]uint32, len(item_data))
	for _, a := range item_data {
		s := make(map[int]uint32, len(a.GetPositon_Table()))
		for _, t := range a.GetPositon_Table() {
			s[GetJadeSlot(t.GetJadeType())] = t.GetPositonLevel()
		}
		gdEquipJade[a.GetItemCondition()] = s
	}
}

func GetEquipJadeUnlockLvl(equipSlot uint32, jadeSlot int) (bool, uint32) {
	m, ok := gdEquipJade[equipSlot]
	if ok {
		return true, m[jadeSlot]
	}
	return false, 0
}
