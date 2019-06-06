package logics

import (
	"math/rand"
	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/MagicPet"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// SetStateOfShowMagicPet : 设置灵宠显示状态
func (p *Account) SetStateOfShowMagicPetHandler(req *reqMsgSetStateOfShowMagicPet, resp *rspMsgSetStateOfShowMagicPet) uint32 {
	//先判断是否开启了灵宠
	if !account.CondCheck(gamedata.Mod_MagicPet, p.Account) {
		return errCode.CommonConditionFalse
	}
	p.Profile.GetHero().IsNotShowMagicPet = !req.ReqStateOfShowMagicPet

	resp.ResStateOfShowMagicPet = !p.Profile.GetHero().IsNotShowMagicPet
	return 0
}

// ShowMagicPet : 传输英雄灵宠信息
func (p *Account) ShowMagicPetHandler(req *reqMsgShowMagicPet, resp *rspMsgShowMagicPet) uint32 {
	//先判断是否开启了灵宠
	if !account.CondCheck(gamedata.Mod_MagicPet, p.Account) {
		return errCode.CommonConditionFalse
	}
	hero := p.Profile.GetHero()
	propData := hero.HeroMagicPets
	resp.HeroMagicPetsInfo = make([][]byte, 0, helper.AVATAR_NUM_CURR)
	for i, item := range propData {
		resp.HeroMagicPetsInfo = append(resp.HeroMagicPetsInfo, encode(p.getMagicPetInfo(i, item)))
	}
	return 0
}

//获取一个英雄灵宠信息，index为id，info为玩家信息某一武将对应灵宠信息
func (p *Account) getMagicPetInfo(index int, info MagicPet.HeroMagicPets) *HeroMagicPetInfo {
	MagicPetInfo := &HeroMagicPetInfo{}
	pet := &info.GetPets()[0]
	//获取属性条数
	talentNums := gamedata.GetMagicPetConfig().GetAttributeAmount()
	//参数置入
	MagicPetInfo.HeroID = int64(index)
	MagicPetInfo.PetStar = int64(pet.Star)
	MagicPetInfo.PetLev = int64(pet.Lev)
	//资质和临时资质各有talentNums条
	aptitudes := make([][]byte, 0, talentNums)
	casual_aptitudes := make([][]byte, 0, talentNums)
	talent := pet.GetTalents()
	casual_talents := pet.GetCasualTalents()
	for j := uint32(0); j < uint32(len(talent)); j++ {
		aptitudes = append(aptitudes, encode(MagicPetAptitude{int64(talent[j].Type), int64(talent[j].Value)}))
		casual_aptitudes = append(casual_aptitudes, encode(MagicPetAptitude{int64(casual_talents[j].Type), int64(casual_talents[j].Value)}))
	}
	MagicPetInfo.PetAptitudes = aptitudes
	MagicPetInfo.CasualPetAptitude = casual_aptitudes
	MagicPetInfo.PetCompreTalent = int64(pet.CompreTalent)
	if !pet.IsNotFirstTimeChangeMagicPetTalents {
		MagicPetInfo.CasualCompreTalent = -1
	} else {
		MagicPetInfo.CasualCompreTalent = int64(pet.CasualCompreTalent)
	}
	MagicPetInfo.ShowMagicPet = !p.Profile.GetHero().IsNotShowMagicPet
	return MagicPetInfo
}

// MagicPetLevUp : 英雄灵宠升级
func (p *Account) MagicPetLevUpHandler(req *reqMsgMagicPetLevUp, resp *rspMsgMagicPetLevUp) uint32 {
	//先判断是否开启了灵宠
	if !account.CondCheck(gamedata.Mod_MagicPet, p.Account) {
		return errCode.CommonConditionFalse
	}
	hero := p.Profile.GetHero()
	if req.HeroID < 0 || int(req.HeroID) >= len(hero.HeroLevel) {
		logs.Warn("CommonInvalidParam Hero_ID")
		return errCode.CommonInvalidParam
	}

	pets := &hero.HeroMagicPets[req.HeroID]
	//目前每个武将只有一个灵宠，所以只判断pets[0]
	pet := &pets.GetPets()[0]
	lev := pet.Lev
	heroLev := hero.HeroLevel[req.HeroID]

	if lev >= heroLev {
		logs.Warn("MagicPetLevUpHandler MagicPetLev:%d is equal to Hero:%d", lev, heroLev)
		return errCode.CommonConditionFalse
	}
	costData := &gamedata.CostData{}
	if gamedata.GetMagicPetLvInfo(pet.Lev+1) == nil {
		return errCode.CommonMaxLimit
	}
	for _, data := range gamedata.GetMagicPetLvInfo(pet.Lev + 1).Material_Table {
		costData.AddItem(data.GetMaterialID(), data.GetMaterialAmount())
	}
	if !account.CostBySync(p.Account, costData, resp, "LevelUpPet") {
		return errCode.CommonLessMoney
	}
	// gs change
	p.Profile.GetData().SetNeedCheckMaxGS()
	resp.OnChangeMagicPetInfo()
	//升级
	pet.Lev++
	resp.HeroID = req.HeroID
	resp.Level = int64(pet.Lev)

	//埋点
	logiclog.LogHeroMagicPetLevUp(
		p.AccountID.String(),
		int(req.HeroID),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, 0,
		int(pet.Lev-1),
		int(pet.Lev),
		p.Profile.GetData().GetCurrGS(p.Account),
		int(p.Profile.GetVip().V),
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	return 0
}

// MagicPetStarUp : 英雄灵宠升星
func (p *Account) MagicPetStarUpHandler(req *reqMsgMagicPetStarUp, resp *rspMsgMagicPetStarUp) uint32 {
	//先判断是否开启了灵宠
	if !account.CondCheck(gamedata.Mod_MagicPet, p.Account) {
		return errCode.CommonConditionFalse
	}
	hero := p.Profile.GetHero()
	if req.HeroID < 0 || int(req.HeroID) >= len(hero.HeroLevel) {
		logs.Warn("CommonInvalidParam Hero_ID")
		return errCode.CommonInvalidParam
	}

	resp.HeroID = req.HeroID

	pets := &hero.HeroMagicPets[req.HeroID]
	special := req.Special
	//目前每个武将只有一个灵宠，所以只判断pets[0]
	pet := &pets.GetPets()[0]

	if pet.Lev < gamedata.GetMagicPetConfig().GetStarCondition() {
		return errCode.CommonConditionFalse
	}

	//如果星级达到最大，返回错误
	if pet.Star >= gamedata.GetMaxStar() {
		logs.Warn("MagicPetStarlevUpHandler MagicPetStarLev:%d is max", pet.Star)
		return errCode.CommonMaxLimit
	}

	costData := &gamedata.CostData{}

	if gamedata.GetMagicPetStarInfo(pet.Star+1) == nil {
		return errCode.CommonMaxLimit
	}

	for _, data := range gamedata.GetMagicPetStarInfo(pet.Star + 1).Material_Table {
		if !special && data.GetMaterialID() == "VI_PET_STAR2" {
			continue
		}
		costData.AddItem(data.GetMaterialID(), data.GetMaterialAmount())
	}
	if !account.CostBySync(p.Account, costData, resp, "StarUpPet") {
		return errCode.CommonLessMoney
	}
	starInfo := gamedata.GetStar(pet.Star + 1)

	iSSuccess, _ := judgeResultAndIfControl(pet, starInfo, p.Account.GetRand())

	if iSSuccess {
		starUpSuccess(pet)
	} else {
		starUpFail(pet, special)
	}

	resp.Star = int64(pet.Star)
	p.Profile.GetData().SetNeedCheckMaxGS()
	resp.OnChangeMagicPetInfo()

	//埋点

	logiclog.LogHeroMagicPetStarUp(p.AccountID.String(),
		int(req.HeroID),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, 0,
		int(pet.Star-1),
		int(pet.Star),
		p.Profile.GetData().GetCurrGS(p.Account),
		int(p.Profile.GetVip().V),
		req.Special,
		iSSuccess,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	return 0
}

//判断是否成功并返回是否成功，是否是暗控结果。
func judgeResultAndIfControl(pet *MagicPet.HeroMagicPet, starInfo *ProtobufGen.PETSTAR, rander *rand.Rand) (bool, bool) {
	//低于首次暗控，必失败
	if !pet.IsNotFirstTime && pet.StarCountTimes < starInfo.GetFirstcontrol() {
		return false, true
	}
	//达到保底暗控，必成功
	if pet.StarCountTimes >= starInfo.GetLastcontrol()-1 {
		return true, true
	}
	//不进行暗控，看脸

	//升星失败
	if !randomResult(starInfo.GetProbability(), rander) {
		return false, false
	}
	//升星成功
	return true, false
}

//传入一个概率，生成结果。
func randomResult(probability float32, rander *rand.Rand) bool {
	rnd := rander.Float32()
	return rnd < probability
}

//升星成功时的逻辑
func starUpSuccess(pet *MagicPet.HeroMagicPet) {
	pet.StarCountTimes = 0
	pet.IsNotFirstTime = false
	pet.Star++
}

//升星失败时的逻辑
func starUpFail(pet *MagicPet.HeroMagicPet, special bool) {
	pet.StarCountTimes++
	//失败会掉星并且没有用保星符
	if gamedata.GetStar(pet.Star+1).GetStarType() == 2 && !special {
		pet.Star--
		pet.IsNotFirstTime = true
		pet.StarCountTimes = 0
	}
}

// MagicPetChangeTalent : 英雄灵宠洗练
func (p *Account) MagicPetChangeTalentHandler(req *reqMsgMagicPetChangeTalent, resp *rspMsgMagicPetChangeTalent) uint32 {
	//先判断是否开启了灵宠
	if !account.CondCheck(gamedata.Mod_MagicPet, p.Account) {
		return errCode.CommonConditionFalse
	}
	hero := p.Profile.GetHero()
	if req.HeroID < 0 || int(req.HeroID) >= len(hero.HeroLevel) {
		logs.Warn("CommonInvalidParam Hero_ID")
		return errCode.CommonInvalidParam
	}
	resp.HeroID = req.HeroID
	pets := &hero.HeroMagicPets[req.HeroID]
	special := req.Special
	//目前每个武将只有一个灵宠，所以只判断pets[0]
	pet := &pets.GetPets()[0]

	if pet.Star < gamedata.GetMagicPetConfig().GetAptitudeCondition() {
		return errCode.CommonConditionFalse
	}

	pet.IsNotFirstTimeChangeMagicPetTalents = true
	//综合资质
	resp.CompreTalent = int64(pet.CompreTalent)
	//资质
	resp.Talents = make([][]byte, 0)
	for _, data := range pet.GetTalents() {
		resp.Talents = append(resp.Talents, encode(MagicPetAptitude{int64(data.Type), int64(data.Value)}))
	}

	costData := &gamedata.CostData{}

	costData.AddItem(gamedata.GetMagicPetConfig().GetMaterialID(), gamedata.GetMagicPetConfig().GetMaterialAmount())
	if special {
		costData.AddItem(gamedata.GetMagicPetConfig().GetMaterialID2(), gamedata.GetMagicPetConfig().GetMaterialAmount2())
	}

	if !account.CostBySync(p.Account, costData, resp, "ChangePetTalents") {
		return errCode.CommonLessMoney
	}

	//确定综合资质

	//最后进入的区间
	section := selectSection(special, pet, p.Account.GetRand())
	//确定区间范围内的一个值
	//这个值是发送给客户端的，必须存在
	value := sureValue(section, p.Account.GetRand())
	//这个值是用于服务器计算资质的，必须存在
	talentValue := float32(value) * gamedata.GetMagicPetConfig().GetMinimumUnit()
	//临时综合资质
	resp.CasualCompreTalent = int64(value)
	//为埋点准备，洗练之前的临时综合资质
	saveCasualCompreTalent := pet.CasualCompreTalent
	pet.CasualCompreTalent = int32(value)
	//临时资质数值
	//先确定平均单个属性资质
	preTalentsValue := talentValue / float32(gamedata.GetMagicPetConfig().GetRandomMulAptitude()) * float32(gamedata.GetMagicPetConfig().GetRandomAptitude())
	//生成资质,talentsNum为资质数量
	talentsNum := gamedata.GetMagicPetConfig().GetAttributeAmount()
	myCasualTalentsValue := generateTalentsValues(preTalentsValue, talentsNum, p.Account.GetRand())
	//埋点准备数据，洗练之前的灵宠详细资质
	beforeSaveCasualTalents := make([]logiclog.TalentInterface, len(pet.GetTalents()))
	for i, v := range pet.GetCasualTalents() {
		beforeSaveCasualTalents[i] = v
	}
	//生成资质
	pet.CasualTalents = generateTalentsTypes(pet.GetCasualTalents(), myCasualTalentsValue, p.Account.GetRand())

	resp.CasualTalents = make([][]byte, talentsNum)
	//埋点准备数据，洗练之后的灵宠详细资质及客户端返回值
	endSaveCasualTalents := make([]logiclog.TalentInterface, len(pet.GetTalents()))
	for i, data := range pet.GetCasualTalents() {
		resp.CasualTalents[i] = encode(MagicPetAptitude{int64(data.Type), int64(data.Value)})
		endSaveCasualTalents[i] = data
	}
	//埋点
	logiclog.LogHeroMagicPetTalent(p.AccountID.String(),
		int(req.HeroID),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, 0,
		int(saveCasualCompreTalent),
		int(pet.CasualCompreTalent),
		beforeSaveCasualTalents,
		endSaveCasualTalents,
		int(p.Profile.GetVip().V),
		req.Special,
		int(pet.SpecialChangeCountTimes),
		int(pet.NormalChangeCountTimes),
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")
	return 0
}

//选择进入的区间
func selectSection(special bool, pet *MagicPet.HeroMagicPet, rander *rand.Rand) *ProtobufGen.PETAPTITUDE {
	//首先判断是否使用道具，使用道具时查使用道具的表，不使用道具时查不使用道具的表
	if special {
		//使用道具时
		pet.SpecialChangeCountTimes++
		logs.Trace("It is %d times specialChange", pet.SpecialChangeCountTimes)
		return choseSection(pet.SpecialChangeCountTimes, gamedata.SpecialNormalWeight,
			gamedata.PetAptitudeSpecialSpecial, gamedata.PetAptitudeSpecialNormal, pet.GetSpecialSpecialSection(), rander)
	} else {
		//不使用道具时
		pet.NormalChangeCountTimes++
		logs.Trace("It is %d times normalChange", pet.NormalChangeCountTimes)
		return choseSection(pet.NormalChangeCountTimes, gamedata.NormalNormalWeight,
			gamedata.PetAptitudeNormalSpecial, gamedata.PetAptitudeNormalNormal, pet.GetNormalSpecialSection(), rander)
	}
}

//确定区间范围内的一个值
func sureValue(section *ProtobufGen.PETAPTITUDE, rander *rand.Rand) uint32 {
	if section.GetIntervalStart() >= section.GetIntervalEnd() {
		return section.GetIntervalEnd()
	} else {
		return uint32(rander.Int31n(int32(section.GetIntervalEnd()-section.GetIntervalStart()))) + section.GetIntervalStart()
	}
}
func generateTalentsValues(preTalentsValue float32, talentsNum uint32, rander *rand.Rand) []uint32 {
	//定义X值：
	X := 100
	if preTalentsValue > 0.5*float32(gamedata.GetMagicPetConfig().GetRandomAptitude()) {
		X = int((float32(gamedata.GetMagicPetConfig().GetRandomAptitude()) - preTalentsValue) / preTalentsValue * 100)
	}
	//在[0,X]取a，b
	var a int32
	var b int32
	if X == 0 {
		a = 0
		b = 0
	} else {
		a = rander.Int31n(int32(X))
		b = rander.Int31n(int32(X))
	}

	//生成资质数值

	var myCasualTalentsValue []uint32
	myCasualTalentsValue = make([]uint32, 0, talentsNum)
	myCasualTalentsValue = append(myCasualTalentsValue, uint32(preTalentsValue))
	myCasualTalentsValue = append(myCasualTalentsValue, uint32(preTalentsValue)+uint32(preTalentsValue*float32(a)/100))
	myCasualTalentsValue = append(myCasualTalentsValue, uint32(preTalentsValue)-uint32(preTalentsValue*float32(a)/100))
	myCasualTalentsValue = append(myCasualTalentsValue, uint32(preTalentsValue)+uint32(preTalentsValue*float32(b)/100))
	myCasualTalentsValue = append(myCasualTalentsValue, uint32(preTalentsValue)-uint32(preTalentsValue*float32(b)/100))

	noSort(myCasualTalentsValue, rander)
	return myCasualTalentsValue
}
func generateTalentsTypes(casualTalents []MagicPet.Talent, myCasualTalentsValue []uint32, rander *rand.Rand) []MagicPet.Talent {
	//countA表示攻出现次数，countB表示防出现次数，countC表示血出现次数。
	var countA uint32
	var countB uint32
	var countC uint32

	for i := uint32(0); i < uint32(len(casualTalents)); i++ {
		sumweight := gamedata.GetTypeAptitude(countA).GetATKWeight() + gamedata.GetTypeAptitude(countB).GetDEFWeight() + gamedata.GetTypeAptitude(countC).GetHPWeight()
		randNum := rander.Int31n(int32(sumweight))
		randNum -= int32(gamedata.GetTypeAptitude(countA).GetATKWeight())
		if randNum <= 0 {
			countA++
			casualTalents[i] = MagicPet.Talent{0, myCasualTalentsValue[i]}
			continue
		}
		randNum -= int32(gamedata.GetTypeAptitude(countB).GetDEFWeight())
		if randNum <= 0 {
			countB++
			casualTalents[i] = MagicPet.Talent{1, myCasualTalentsValue[i]}
			continue
		}
		randNum -= int32(gamedata.GetTypeAptitude(countC).GetHPWeight())
		if randNum <= 0 {
			countC++
			casualTalents[i] = MagicPet.Talent{2, myCasualTalentsValue[i]}
			continue
		}
	}
	return casualTalents
}

//洗牌算法，传入一个切片，将切片内的顺序打乱
func noSort(info []uint32, rander *rand.Rand) {
	for i := 0; i < len(info); i++ {
		randnum := rander.Int31n(int32(len(info)))
		info[i], info[randnum] = info[randnum], info[i]
	}
}

//选择区间
func choseSection(countTimes, weight uint32, specialaptitudes, normalaptitude []*ProtobufGen.PETAPTITUDE, pop [][]uint32, rander *rand.Rand) *ProtobufGen.PETAPTITUDE {
	//判断是否进入了特殊组
	petAptitudeSection, ok := isSpecialSection(countTimes, specialaptitudes, pop, rander)
	if ok {
		//进入特殊组时，直接将取得的区间返回
		return petAptitudeSection
	} else {
		//未进入特殊组时，用inNormalSection去随机获取区间
		return isNormalSection(weight, normalaptitude, rander)
	}
}

//在普通区间中根据配置的权重随机选择区间
func isNormalSection(weight uint32, aptitudes []*ProtobufGen.PETAPTITUDE, rander *rand.Rand) *ProtobufGen.PETAPTITUDE {
	randomNum := rander.Int31n(int32(weight))
	var i int
	for i = range aptitudes {
		randomNum -= int32(aptitudes[i].GetIntervalWeight())
		if randomNum <= 0 {
			break
		}
	}
	return aptitudes[i]
}

//判断是否符合进入某一特殊区间的条件，如果符合返回此区间，和true，存在多个特殊区间时，返回边界终止值最高的那个。
func isSpecialSection(countTimes uint32, aptitudes []*ProtobufGen.PETAPTITUDE, pop [][]uint32, rander *rand.Rand) (result *ProtobufGen.PETAPTITUDE, flag bool) {
	for i, data := range aptitudes {
		//判断是否满足特殊组要求,满足要求后，判断是否要替换原本的特殊组
		if i >= len(pop) {
			logs.Error("error magicpetAptitude.data")
			continue
		}
		var judge bool
		pop[i], judge = judgeMeetReq(countTimes, data, pop[i], rander)
		if judge && (result == nil || result.GetIntervalEnd() < data.GetIntervalEnd()) {
			result = data
			flag = true
		}
	}
	return result, flag
}

//判断是否满足进入特殊区间的条件
func judgeMeetReq(countTimes uint32, section *ProtobufGen.PETAPTITUDE, randpop []uint32, rander *rand.Rand) ([]uint32, bool) {
	if countTimes < section.GetFirstcontrol() {
		return randpop, false
	}

	target := (countTimes - section.GetFirstcontrol()) % section.GetMForSpecial()
	if target == 0 { //第一次进入，生成随机值
		N := section.GetNForSpecial()
		randpop = make([]uint32, N)
		for i := uint32(0); i < N; i++ {
			if i >= uint32(len(randpop)) {
				logs.Error("error magicpetAptitude.data")
				continue
			}
			flag := false
			for !flag {
				flag = true
				randpop[i] = uint32(rander.Int31n(int32(section.GetMForSpecial())))
				//检测重复，如果出现重复则重新随机
				for j := uint32(0); j < i; j++ {
					if randpop[i] == randpop[j] || randpop[i] == 0 {
						flag = false
					}
				}
			}
		}
	}
	if countTimes == section.GetFirstcontrol() {
		return randpop, true
	}
	for i := range randpop {
		if randpop[i] == target {
			return randpop, true
		}
	}
	return randpop, false
}

// MagicPetSaveTalent : 英雄灵宠洗练保存
func (p *Account) MagicPetSaveTalentHandler(req *reqMsgMagicPetSaveTalent, resp *rspMsgMagicPetSaveTalent) uint32 {
	//先判断是否开启了灵宠
	if !account.CondCheck(gamedata.Mod_MagicPet, p.Account) {
		return errCode.CommonConditionFalse
	}
	hero := p.Profile.GetHero()
	if req.HeroID < 0 || int(req.HeroID) >= len(hero.HeroLevel) {
		logs.Warn("CommonInvalidParam Hero_ID")
		return errCode.CommonInvalidParam
	}
	resp.HeroID = req.HeroID

	pets := &hero.HeroMagicPets[req.HeroID]

	pet := &pets.GetPets()[0]

	//埋点数据
	beforeSaveTalents := make([]logiclog.TalentInterface, len(pet.GetTalents()))
	endSaveTalents := make([]logiclog.TalentInterface, len(pet.GetTalents()))
	for i, _ := range pet.GetTalents() {
		beforeSaveTalents[i] = pet.GetTalents()[i]
		pet.GetTalents()[i] = pet.GetCasualTalents()[i]
		endSaveTalents[i] = pet.GetTalents()[i]

	}
	//埋点数据
	saveCompreTalent := pet.CompreTalent

	if pet.CasualCompreTalent >= 0 {
		pet.CompreTalent = pet.CasualCompreTalent
	}

	//如果已经保存过，又触发了保存协议，此时不再对pet.CompreTalent赋值
	resp.CompreTalent = int64(pet.CompreTalent)
	resp.CasualCompreTalent = -1
	pet.CasualCompreTalent = -1
	resp.Talents = make([][]byte, 0)

	for _, data := range pet.GetTalents() {
		resp.Talents = append(resp.Talents, encode(MagicPetAptitude{int64(data.Type), int64(data.Value)}))
	}
	resp.CasualTalents = make([][]byte, 0)
	for _, data := range pet.GetCasualTalents() {
		resp.CasualTalents = append(resp.CasualTalents, encode(MagicPetAptitude{int64(data.Type), int64(data.Value)}))
	}
	p.Profile.GetData().SetNeedCheckMaxGS()
	resp.OnChangeMagicPetInfo()

	//埋点
	logiclog.LogHeroMagicPetTalentSaved(p.AccountID.String(),
		int(req.HeroID),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId, 0,
		int(saveCompreTalent),
		int(pet.CompreTalent),
		beforeSaveTalents,
		endSaveTalents,
		p.Profile.GetData().GetCurrGS(p.Account),
		int(p.Profile.GetVip().V),
		int(pet.SpecialChangeCountTimes),
		int(pet.NormalChangeCountTimes),
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")

	return 0
}
