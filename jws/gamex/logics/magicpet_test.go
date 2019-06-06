package logics

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/MagicPet"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"

	"github.com/stretchr/testify/assert"
)

type UtilCountTimes struct {
	countTimes uint32
}

var (
	aptitudes []*ProtobufGen.PETAPTITUDE
	pop       [][]uint32
	rander    *rand.Rand
)

func init() {
	rander = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func TestInSpecialSection(t *testing.T) {
	var (
		AptitudeID     uint32 = 1
		IntervalEnd    uint32 = 5
		IntervalWeight uint32 = 10
		NForSpecial    uint32 = 4
		MForSpecial    uint32 = 10
		Firstcontrol   uint32 = 50
	)
	var (
		AptitudeID1     uint32 = 2
		IntervalEnd1    uint32 = 10
		IntervalWeight1 uint32 = 15
		NForSpecial1    uint32 = 4
		MForSpecial1    uint32 = 10
		Firstcontrol1   uint32 = 50
	)
	aptitudes = make([]*ProtobufGen.PETAPTITUDE, 0, 3)
	pop = make([][]uint32, 2, 3)
	aptitudes = append(aptitudes, &ProtobufGen.PETAPTITUDE{AptitudeID: &AptitudeID, IntervalEnd: &IntervalEnd, IntervalWeight: &IntervalWeight, NForSpecial: &NForSpecial, MForSpecial: &MForSpecial, Firstcontrol: &Firstcontrol})
	pop[0] = append(pop[0], 0, 1, 5, 9)
	aptitudes = append(aptitudes, &ProtobufGen.PETAPTITUDE{AptitudeID: &AptitudeID1, IntervalEnd: &IntervalEnd1, IntervalWeight: &IntervalWeight1, NForSpecial: &NForSpecial1, MForSpecial: &MForSpecial1, Firstcontrol: &Firstcontrol1})
	pop[1] = append(pop[1], 1, 3, 5, 2)

	resultsection, _ := isSpecialSection(104, aptitudes, pop, rander)

	fmt.Println(resultsection.GetAptitudeID())

}

func TestAccount_SetStateOfShowMagicPet(t *testing.T) {
	p := new(Account)
	p.Account = account.Debuger.GetNewAccount()

	req := new(reqMsgSetStateOfShowMagicPet)
	req.ReqStateOfShowMagicPet = true
	resp := new(rspMsgSetStateOfShowMagicPet)

	// 失败：未达到开启条件
	assert.Equal(t, uint32(errCode.CommonConditionFalse), p.SetStateOfShowMagicPetHandler(req, resp))

	// 设为满级并解锁所有关卡
	p.Account.Profile.CorpInf.Level = gamedata.GetCommonCfg().GetCorpLevelUpperLimit()
	p.DebugUnlockAllLevel()

	// 成功：显示
	assert.Equal(t, uint32(0), p.SetStateOfShowMagicPetHandler(req, resp))
	assert.False(t, p.Account.Profile.Hero.IsNotShowMagicPet)

	// 成功：不显示
	req.ReqStateOfShowMagicPet = false
	assert.Equal(t, uint32(0), p.SetStateOfShowMagicPetHandler(req, resp))
	assert.True(t, p.Account.Profile.Hero.IsNotShowMagicPet)
}

func TestAccount_ShowMagicPet(t *testing.T) {
	p := new(Account)
	p.Account = account.Debuger.GetNewAccount()

	req := new(reqMsgShowMagicPet)
	resp := new(rspMsgShowMagicPet)

	// 失败：未达到开启条件
	assert.Equal(t, uint32(errCode.CommonConditionFalse), p.ShowMagicPetHandler(req, resp))

	// 设为满级并解锁所有关卡
	p.Account.Profile.CorpInf.Level = gamedata.GetCommonCfg().GetCorpLevelUpperLimit()
	p.DebugUnlockAllLevel()

	// 成功：
	assert.Equal(t, uint32(0), p.ShowMagicPetHandler(req, resp))
	assert.Equal(t, account.AVATAR_NUM_MAX, len(resp.HeroMagicPetsInfo))
}

func TestAccount_MagicPetLevUp(t *testing.T) {
	p := new(Account)
	p.Account = account.Debuger.GetNewAccount()

	var (
		heroId int64  = 16
		level  uint32 = 50
	)

	pet := new(MagicPet.HeroMagicPet)
	p.Account.Profile.Hero.HeroMagicPets[heroId].Pets = []MagicPet.HeroMagicPet{*pet}
	pet = &p.Account.Profile.Hero.HeroMagicPets[heroId].Pets[0]
	heroLevel := &p.Account.Profile.Hero.HeroLevel[heroId]

	req := new(reqMsgMagicPetLevUp)
	resp := new(rspMsgMagicPetLevUp)

	// 失败：未达到开启条件
	assert.Equal(t, uint32(errCode.CommonConditionFalse), p.MagicPetLevUpHandler(req, resp))

	// 解锁
	p.Account.Profile.CorpInf.Level = gamedata.GetCommonCfg().GetCorpLevelUpperLimit()
	p.DebugUnlockAllLevel()

	req.HeroID = account.AVATAR_NUM_MAX

	// 失败：HeroID非法
	assert.Equal(t, uint32(errCode.CommonInvalidParam), p.MagicPetLevUpHandler(req, resp))

	req.HeroID = heroId
	*heroLevel = level >> 1
	pet.Lev = level

	// 失败：Hero等级小于Pet
	assert.Equal(t, uint32(errCode.CommonConditionFalse), p.MagicPetLevUpHandler(req, resp))

	*heroLevel = level

	// 失败：等于
	assert.Equal(t, uint32(errCode.CommonConditionFalse), p.MagicPetLevUpHandler(req, resp))

	*heroLevel = gamedata.GetCommonCfg().GetCorpLevelUpperLimit()
	// TODO: 目前没有办法获取Pet等级上限值，手动填100；如果将来等级上限提高了，需要改这个值
	pet.Lev = 100

	// 失败：Pet等级达到上限
	assert.Equal(t, uint32(errCode.CommonMaxLimit), p.MagicPetLevUpHandler(req, resp))

	pet.Lev = level

	// 失败：通货不足
	assert.Equal(t, uint32(errCode.CommonLessMoney), p.MagicPetLevUpHandler(req, resp))

	// VI_PET_LEVEL type:32
	p.Profile.GetSC().AddSC(32, 99999999, "Magic Pet Unit Test")

	// 成功：0 -> 1
	pet.Lev = 0
	assert.Equal(t, uint32(0), p.MagicPetLevUpHandler(req, resp))
	assert.Equal(t, uint32(1), pet.Lev)

	// 成功：1 -> 2
	assert.Equal(t, uint32(0), p.MagicPetLevUpHandler(req, resp))
	assert.Equal(t, uint32(2), pet.Lev)

	// 成功：98 -> 99
	pet.Lev = 98
	assert.Equal(t, uint32(0), p.MagicPetLevUpHandler(req, resp))
	assert.Equal(t, uint32(99), pet.Lev)

	// 成功：99 -> 100
	assert.Equal(t, uint32(0), p.MagicPetLevUpHandler(req, resp))
	assert.Equal(t, uint32(100), pet.Lev)
}

func TestAccount_MagicPetStarUp(t *testing.T) {
	p := new(Account)
	p.Account = account.Debuger.GetNewAccount()

	var (
		heroId      int64  = 32
		unlockLevel uint32 = gamedata.GetMagicPetConfig().GetStarCondition()
	)

	pet := new(MagicPet.HeroMagicPet)
	p.Account.Profile.Hero.HeroMagicPets[heroId].Pets = []MagicPet.HeroMagicPet{*pet}
	pet = &p.Account.Profile.Hero.HeroMagicPets[heroId].Pets[0]

	req := new(reqMsgMagicPetStarUp)
	resp := new(rspMsgMagicPetStarUp)

	// 失败：未达到开启条件
	assert.Equal(t, uint32(errCode.CommonConditionFalse), p.MagicPetStarUpHandler(req, resp))

	// 解锁
	p.Account.Profile.CorpInf.Level = gamedata.GetCommonCfg().GetCorpLevelUpperLimit()
	p.DebugUnlockAllLevel()

	req.HeroID = account.AVATAR_NUM_MAX

	// 失败：HeroID非法
	assert.Equal(t, uint32(errCode.CommonInvalidParam), p.MagicPetStarUpHandler(req, resp))

	req.HeroID = heroId
	pet.Lev = unlockLevel - 1

	// 失败：没有解锁魔宠升星
	assert.Equal(t, uint32(errCode.CommonConditionFalse), p.MagicPetStarUpHandler(req, resp))

	pet.Lev += 1
	pet.Star = gamedata.GetMaxStar()

	// 失败：满星
	assert.Equal(t, uint32(errCode.CommonMaxLimit), p.MagicPetStarUpHandler(req, resp))

	// TODO: 下面的用例都是根据gamedata表决定的，如果策划改表则可能需要重新赋值
	// 22 概率是100%，方便查看结果
	pet.Star = 21

	// 失败：缺材料
	assert.Equal(t, uint32(errCode.CommonLessMoney), p.MagicPetStarUpHandler(req, resp))

	// VI_PET_STAR: 33, VI_PET_STAR2: 34 卡点材料2-6
	p.Profile.GetSC().AddSC(33, 99999999, "Magic Pet Unit Test")
	p.Profile.GetSC().AddSC(34, 99999999, "Magic Pet Unit Test")
	for i := 2; i < 7; i++ {
		p.Account.BagProfile.StackBag.Add(
			*gamedata.NewBagItemData(),
			fmt.Sprintf("MAT_PET_BREAK%d", i),
			9999,
			p.Account.AccountID.String(),
			p.Account.GetRand(),
			time.Now().UnixNano(),
		)
	}

	// 成功：升星成功，走概率
	assert.Equal(t, uint32(0), p.MagicPetStarUpHandler(req, resp))
	assert.Equal(t, uint32(22), pet.Star)
	assert.False(t, pet.IsNotFirstTime)

	pet.Star = 64
	pet.StarCountTimes = 17

	// 成功：升星失败，未达到首次暗控，掉星
	assert.Equal(t, uint32(0), p.MagicPetStarUpHandler(req, resp))
	assert.Equal(t, uint32(63), pet.Star)
	assert.True(t, pet.IsNotFirstTime)

	pet.Star = 64
	pet.StarCountTimes = 17
	pet.IsNotFirstTime = false
	req.Special = true

	// 成功：升星失败，未达到首次暗控，使用道具不掉星
	assert.Equal(t, uint32(64), pet.Star)
	assert.NotEqual(t, uint32(0), pet.StarCountTimes)
	assert.False(t, pet.IsNotFirstTime)

	pet.Star = 64
	pet.StarCountTimes = 36

	// 成功：升星成功，达到保底暗控
	assert.Equal(t, uint32(0), p.MagicPetStarUpHandler(req, resp))
	assert.Equal(t, uint32(65), pet.Star)
}

func TestAccount_MagicPetChangeTalent(t *testing.T) {
	p := new(Account)
	p.Account = account.Debuger.GetNewAccount()

	var (
		heroId           int64  = 8
		unlockStar       uint32 = gamedata.GetMagicPetConfig().GetAptitudeCondition()
		testNormalCount  uint32 = 15
		testSpecialCount uint32 = 31
	)

	pet := new(MagicPet.HeroMagicPet)
	p.Account.Profile.Hero.HeroMagicPets[heroId].Pets = []MagicPet.HeroMagicPet{*pet}
	pet = &p.Account.Profile.Hero.HeroMagicPets[heroId].Pets[0]

	req := new(reqMsgMagicPetChangeTalent)
	resp := new(rspMsgMagicPetChangeTalent)

	// 失败：未达到开启条件
	assert.Equal(t, uint32(errCode.CommonConditionFalse), p.MagicPetChangeTalentHandler(req, resp))

	// 解锁
	p.Account.Profile.CorpInf.Level = gamedata.GetCommonCfg().GetCorpLevelUpperLimit()
	p.DebugUnlockAllLevel()

	req.HeroID = account.AVATAR_NUM_MAX

	// 失败：HeroID非法
	assert.Equal(t, uint32(errCode.CommonInvalidParam), p.MagicPetChangeTalentHandler(req, resp))

	req.HeroID = heroId
	pet.Star = unlockStar - 1

	// 失败：没有解锁魔宠资质
	assert.Equal(t, uint32(errCode.CommonConditionFalse), p.MagicPetChangeTalentHandler(req, resp))

	pet.Star += 1

	// 失败：材料不足
	assert.Equal(t, uint32(errCode.CommonLessMoney), p.MagicPetChangeTalentHandler(req, resp))

	// SC_PET_APTITUDE:35 SC_PET_APTITUDE2:36
	p.Profile.GetSC().AddSC(35, 99999999, "Magic Pet Unit Test")
	req.Special = true

	// 失败：高级材料不足
	assert.Equal(t, uint32(errCode.CommonLessMoney), p.MagicPetChangeTalentHandler(req, resp))

	p.Profile.GetSC().AddSC(36, 99999999, "Magic Pet Unit Test")
	req.Special = false
	pet.NormalChangeCountTimes = testNormalCount
	pet.SpecialChangeCountTimes = testSpecialCount

	// 成功：正常
	assert.Equal(t, uint32(0), p.MagicPetChangeTalentHandler(req, resp))
	assert.Equal(t, testNormalCount+1, pet.NormalChangeCountTimes)
	assert.Equal(t, testSpecialCount, pet.SpecialChangeCountTimes)

	req.Special = true
	pet.NormalChangeCountTimes = testNormalCount
	pet.SpecialChangeCountTimes = testSpecialCount

	// 成功：道具
	assert.Equal(t, uint32(0), p.MagicPetChangeTalentHandler(req, resp))
	assert.Equal(t, testNormalCount, pet.NormalChangeCountTimes)
	assert.Equal(t, testSpecialCount+1, pet.SpecialChangeCountTimes)
}

func TestAccount_MagicPetSaveTalent(t *testing.T) {
	p := new(Account)
	p.Account = account.Debuger.GetNewAccount()

	var (
		heroId    int64 = 4
		talentVal int32 = 5
	)

	pet := new(MagicPet.HeroMagicPet)
	pet.Talents = []MagicPet.Talent{
		{0, 100},
		{1, 200},
		{1, 200},
		{2, 400},
		{2, 400},
	}
	pet.CasualTalents = []MagicPet.Talent{
		{0, 1000},
		{0, 1000},
		{0, 1000},
		{0, 1000},
		{0, 1000},
	}
	p.Account.Profile.Hero.HeroMagicPets[heroId].Pets = []MagicPet.HeroMagicPet{*pet}
	pet = &p.Account.Profile.Hero.HeroMagicPets[heroId].Pets[0]

	req := new(reqMsgMagicPetSaveTalent)
	resp := new(rspMsgMagicPetSaveTalent)

	// 失败：未达到开启条件
	assert.Equal(t, uint32(errCode.CommonConditionFalse), p.MagicPetSaveTalentHandler(req, resp))

	// 解锁
	p.Account.Profile.CorpInf.Level = gamedata.GetCommonCfg().GetCorpLevelUpperLimit()
	p.DebugUnlockAllLevel()

	req.HeroID = account.AVATAR_NUM_MAX

	// 失败：HeroID非法
	assert.Equal(t, uint32(errCode.CommonInvalidParam), p.MagicPetSaveTalentHandler(req, resp))

	req.HeroID = heroId
	pet.Star = gamedata.GetMaxStar()
	pet.CasualCompreTalent = talentVal

	// 成功：保存且数据一致
	assert.Equal(t, uint32(0), p.MagicPetSaveTalentHandler(req, resp))
	assert.EqualValues(t, pet.CompreTalent, resp.CompreTalent)             // int32 <-> int64
	assert.EqualValues(t, pet.CasualCompreTalent, resp.CasualCompreTalent) // int32 <-> int64
	assert.Equal(t, len(pet.Talents), len(resp.Talents))
	assert.Equal(t, len(pet.CasualTalents), len(resp.CasualTalents))
}

// BenchmarkAccount_MagicPetChangeTalent 道具洗练性能测试
func BenchmarkAccount_MagicPetChangeTalent(b *testing.B) {
	// 建账号
	p := new(Account)
	p.Account = account.Debuger.GetNewAccount()

	heroId := int64(3)
	pet := new(MagicPet.HeroMagicPet)
	p.Account.Profile.Hero.HeroMagicPets[heroId].Pets = []MagicPet.HeroMagicPet{*pet}
	pet = &p.Account.Profile.Hero.HeroMagicPets[heroId].Pets[0]

	// 解锁MagicPet
	p.Account.Profile.CorpInf.Level = gamedata.GetCommonCfg().GetCorpLevelUpperLimit()
	p.DebugUnlockAllLevel()

	// 解锁洗练
	pet.Star = gamedata.GetMagicPetConfig().GetAptitudeCondition()

	// 设置为道具洗练
	req := new(reqMsgMagicPetChangeTalent)
	req.HeroID = heroId
	req.Special = true
	resp := new(rspMsgMagicPetChangeTalent)

	// 加道具
	p.Profile.GetSC().AddSC(35, math.MaxInt64, "SC_PET_APTITUDE:35")
	p.Profile.GetSC().AddSC(36, math.MaxInt64, "SC_PET_APTITUDE:36")

	// 关log以免影响性能
	logs.Close()

	for i := 0; i < b.N; i++ {
		p.MagicPetChangeTalentHandler(req, resp)
	}
}

func TestNoSort(t *testing.T) {
	help := []uint32{1, 2, 3, 4, 5}

	hits := make(map[uint32]int)

	for i := 0; i < 1000; i++ {
		noSort(help, rander)
		hits[help[0]] += 1
	}

	// 确定shuffle有效，每个元素都在第一位出现过
	assert.Equal(t, len(help), len(hits))
}

// TestSpecialGroupHits 计算100万次道具洗练中，各个特殊组的命中  -=大写函数字母以执行用例=-
func TestspecialGroupHits(t *testing.T) {
	rations := make(map[*ProtobufGen.PETAPTITUDE]int) // [AptitudeID]Count
	pet := new(MagicPet.HeroMagicPet)

	logs.Close()

	for ; pet.SpecialChangeCountTimes < 1000000; pet.SpecialChangeCountTimes++ {
		section := choseSection(pet.SpecialChangeCountTimes,
			gamedata.SpecialNormalWeight,
			gamedata.PetAptitudeSpecialSpecial,
			gamedata.PetAptitudeSpecialNormal,
			pet.GetSpecialSpecialSection(), rander)
		rations[section] += 1
	}

	rationLogs := make([]string, 0, len(rations))

	for section, count := range rations {
		rationLogs = append(rationLogs, fmt.Sprintf("ID:%d, 区间:%d-%d, 次数:%d",
			section.GetAptitudeID(),
			section.GetIntervalStart(),
			section.GetIntervalEnd(),
			count))
	}

	sort.Strings(rationLogs)

	for _, rationLog := range rationLogs {
		fmt.Println(rationLog)
	}
}

// TestNormalGroupHits 计算100万次普通洗练中，各个特殊组的命中  -=大写函数字母以执行用例=-
func TestnormalGroupHits(t *testing.T) {
	rations := make(map[*ProtobufGen.PETAPTITUDE]int) // [AptitudeID]Count
	pet := new(MagicPet.HeroMagicPet)

	logs.Close()

	for ; pet.NormalChangeCountTimes < 1000000; pet.NormalChangeCountTimes++ {
		section := choseSection(pet.NormalChangeCountTimes,
			gamedata.NormalNormalWeight,
			gamedata.PetAptitudeNormalSpecial,
			gamedata.PetAptitudeNormalNormal,
			pet.GetNormalSpecialSection(), rander)
		rations[section] += 1
	}

	rationLogs := make([]string, 0, len(rations))

	for section, count := range rations {
		rationLogs = append(rationLogs, fmt.Sprintf("ID:%d, 区间:%d-%d, 次数:%d",
			section.GetAptitudeID(),
			section.GetIntervalStart(),
			section.GetIntervalEnd(),
			count))
	}

	sort.Strings(rationLogs)

	for _, rationLog := range rationLogs {
		fmt.Println(rationLog)
	}
}

// TestStarUpHits 计算1万次魔宠升级到满星时，每级平均升级次数   -=大写函数字母以执行用例=-
func TeststarUpHits(t *testing.T) {
	rations := make(map[uint32]uint32) // [pet.Star] Count

	logs.Close()

	// 建账号
	p := new(Account)
	p.Account = account.Debuger.GetNewAccount()

	heroId := int64(19)
	pet := new(MagicPet.HeroMagicPet)
	pet.Lev = 100
	p.Account.Profile.Hero.HeroMagicPets[heroId].Pets = []MagicPet.HeroMagicPet{*pet}
	pet = &p.Account.Profile.Hero.HeroMagicPets[heroId].Pets[0]

	// 解锁MagicPet
	p.Account.Profile.CorpInf.Level = gamedata.GetCommonCfg().GetCorpLevelUpperLimit()
	p.DebugUnlockAllLevel()

	// 给材料
	p.Profile.GetSC().AddSC(33, math.MaxInt64, "Magic Pet Unit Test")
	p.Profile.GetSC().AddSC(34, math.MaxInt64, "Magic Pet Unit Test")

	for i := 2; i < 7; i++ {
		p.Account.BagProfile.StackBag.Add(
			*gamedata.NewBagItemData(),
			fmt.Sprintf("MAT_PET_BREAK%d", i),
			math.MaxInt32,
			p.Account.AccountID.String(),
			p.Account.GetRand(),
			time.Now().UnixNano(),
		)
	}

	req := new(reqMsgMagicPetStarUp)
	resp := new(rspMsgMagicPetStarUp)

	// 不用道具只能到20，不用试了
	req.HeroID = heroId
	req.Special = true

	// 测试次数
	for i := 0; i < 10000; i++ {
		// 初始化pet信息
		pet.Star = 0
		pet.StarCountTimes = 0
		pet.IsNotFirstTime = false

		// 洗练直到满星
		for pet.Star != gamedata.GetMaxStar() {
			rations[pet.Star] += 1
			r := p.MagicPetStarUpHandler(req, resp)
			if r != 0 {
				t.Fatal("Failed!")
			}
		}
	}

	// 打印结果
	starUpLogs := make([]string, 0)
	for starNum, count := range rations {
		starUpLogs = append(starUpLogs, fmt.Sprintf("星级%3d : 次数%4.4f", starNum, float32(count)/10000))
	}

	sort.Strings(starUpLogs)

	for _, starUplog := range starUpLogs {
		fmt.Println(starUplog)
	}
}
