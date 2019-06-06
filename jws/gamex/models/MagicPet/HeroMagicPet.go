package MagicPet

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

type HeroMagicPets struct {
	Pets []HeroMagicPet `json:"pets"`
}

type HeroMagicPet struct {
	Lev                                 uint32     `json:"lev"`     //等级
	Star                                uint32     `json:"star"`    //星级
	CompreTalent                        int32      `json:"c_t"`     //综合资质
	CasualCompreTalent                  int32      `json:"cc_t"`    //临时综合
	Talents                             []Talent   `json:"ts"`      //具体资质
	CasualTalents                       []Talent   `json:"c_ts"`    //临时资质
	StarCountTimes                      uint32     `json:"s_ct"`    //升星次数
	SpecialChangeCountTimes             uint32     `json:"sc_ct"`   //道具洗练次数
	SpecialSpecialSection               [][]uint32 `json:"sss"`     //道具洗练时随机产生的特殊区间,每一个特殊区间都有一个[N]*uint32，代表N个进入特殊区间的次数。洗练时如果发现这个变量为nil要开辟空间。
	NormalChangeCountTimes              uint32     `json:"nc_ct"`   //普通洗练次数
	NormalSpecialSection                [][]uint32 `json:"nss"`     //普通洗练时随机产生的特殊区间,每一个特殊区间都有一个[N]*uint32，代表N个进入特殊区间的次数。洗练时如果发现这个变量为nil要开辟空间。
	IsNotFirstTimeChangeMagicPetTalents bool       `json:"inftcmp"` //是否是第一次洗练灵宠之前，默认false表示是第一次，true表示不是第一次
	IsNotFirstTime                      bool       `json:"inft"`    //是否是降星后的结果,false表示是第一次，true表示不是第一次
}

func (hmp *HeroMagicPets) GetPets() []HeroMagicPet {
	if hmp.Pets == nil {
		hmp.Pets = append(hmp.Pets, HeroMagicPet{})
	}
	return hmp.Pets
}

//资质
type Talent struct {
	Type  uint32 //类型
	Value uint32 //数值
}

func (t Talent) GetType() uint32 {
	return t.Type
}
func (t Talent) GetValue() uint32 {
	return t.Value
}

func (hmp *HeroMagicPet) GetSpecialSpecialSection() [][]uint32 {
	if hmp.SpecialSpecialSection == nil {
		hmp.SpecialSpecialSection = make([][]uint32, len(gamedata.PetAptitudeSpecialSpecial))
	}
	return hmp.SpecialSpecialSection
}
func (hmp *HeroMagicPet) GetNormalSpecialSection() [][]uint32 {
	if hmp.NormalSpecialSection == nil {
		hmp.NormalSpecialSection = make([][]uint32, len(gamedata.PetAptitudeNormalSpecial))
	}
	return hmp.NormalSpecialSection
}

func (hmp *HeroMagicPet) GetTalents() []Talent {
	if hmp.Talents == nil {
		attributesNum := gamedata.GetMagicPetConfig().GetAttributeAmount()
		hmp.Talents = make([]Talent, attributesNum)
		for i := uint32(0); i < attributesNum; i++ {
			//一攻2防2血
			hmp.Talents[i] = Talent{(i/2 + 1) % 3, 0}
		}
	}
	return hmp.Talents
}
func (hmp *HeroMagicPet) GetCasualTalents() []Talent {
	if hmp.CasualTalents == nil {
		attributesNum := gamedata.GetMagicPetConfig().GetAttributeAmount()
		hmp.CasualTalents = make([]Talent, attributesNum)
		for i := uint32(0); i < attributesNum; i++ {
			//一攻2防2血
			hmp.CasualTalents[i] = Talent{(i/2 + 1) % 3, 0}
		}
	}
	return hmp.CasualTalents
}
