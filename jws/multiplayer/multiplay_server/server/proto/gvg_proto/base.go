package gvg_proto

import (
	"github.com/google/flatbuffers/go"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	. "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/multiplayMsg"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func GenEquipData(builder *flatbuffers.Builder, eq helper.BagItemToClient) flatbuffers.UOffsetT {
	h := NewFlatBufferHelper(builder, 3)
	TableID := h.Pre(builder.CreateString(eq.TableID))
	ItemID := h.Pre(builder.CreateString(eq.ItemID))
	Data := h.Pre(builder.CreateString(eq.Data))

	EquipInfoStart(builder)
	EquipInfoAddId(builder, eq.ID)
	EquipInfoAddTableid(builder, h.Get(TableID))
	EquipInfoAddItemid(builder, h.Get(ItemID))
	EquipInfoAddCount(builder, eq.Count)
	EquipInfoAddData(builder, h.Get(Data))
	return EquipInfoEnd(builder)
}

func GenFashionItemData(builder *flatbuffers.Builder, f helper.FashionItem) flatbuffers.UOffsetT {
	TableID := builder.CreateString(f.TableID)
	FashionItemInfoStart(builder)
	FashionItemInfoAddId(builder, f.ID)
	FashionItemInfoAddTableid(builder, TableID)
	FashionItemInfoAddOt(builder, f.ExpireTimeStamp)
	return FashionItemInfoEnd(builder)
}

func GenAccountInfoData(builder *flatbuffers.Builder, idx int, avatar *helper.Avatar2ClientByJson) flatbuffers.UOffsetT {
	eqs := avatar.GetEquips()
	afs := avatar.GetAllFashions()

	h := NewFlatBufferHelper(builder, 32)

	AcID := h.Pre(builder.CreateString(avatar.GetAcId()))
	Name := h.Pre(builder.CreateString(avatar.Name))
	Guuid := h.Pre(builder.CreateString(avatar.GuildUUID))
	Gname := h.Pre(builder.CreateString(avatar.GuildName))
	Post := h.Pre(builder.CreateString(avatar.GuildPost))
	TitleOn := h.Pre(builder.CreateString(avatar.TitleOn))

	HeroStarVector := h.CreateUInt32Array(AccountInfoStartHeroStarVector, avatar.HeroStarLv)
	HeroLvlVector := h.CreateUInt32Array(AccountInfoStartHeroStarVector, avatar.HeroLv)
	ArousalsVector := h.CreateUInt32Array(AccountInfoStartArousalsVector, avatar.Arousals)
	SkillsVector := h.CreateUInt32Array(AccountInfoStartSkillsVector, avatar.AvatarSkills)
	SkillpsVector := h.CreateUInt32Array(AccountInfoStartSkillpsVector, avatar.SkillPractices)
	AvatarlockedsVector := h.CreateIntArray(AccountInfoStartAvatarlockedsVector, avatar.AvatarLockeds)
	logs.Trace("EquipsUPffsetTs %v", eqs)
	EquipsUPffsetTs := make([]flatbuffers.UOffsetT, 0, len(eqs))
	for i := 0; i < len(eqs); i++ {
		EquipsUPffsetTs = append(EquipsUPffsetTs, GenEquipData(builder, eqs[i]))
	}
	EquipsVector := h.CreateUOffsetTArray(AccountInfoStartEquipsVector, EquipsUPffsetTs)
	EquipUpgradeVector := h.CreateUInt32Array(AccountInfoStartEquipUpgradeVector, avatar.EquipUpgrade)
	EquipStarVector := h.CreateUInt32Array(AccountInfoStartEquipStarVector, avatar.EquipStar)
	AvatarEquipsVector := h.CreateUInt32Array(AccountInfoStartAvatarEquipsVector, avatar.AvatarEquips)
	AllFashionsUPffsetTs := make([]flatbuffers.UOffsetT, 0, len(eqs))
	for i := 0; i < len(afs); i++ {
		AllFashionsUPffsetTs = append(AllFashionsUPffsetTs, GenFashionItemData(builder, afs[i]))
	}
	AllFashionsVector := h.CreateUOffsetTArray(AccountInfoStartAllFashionsVector, AllFashionsUPffsetTs)
	GeneralsVector := h.CreateStringArray(AccountInfoStartGeneralsVector, avatar.Generals)
	GenstarVector := h.CreateUInt32Array(AccountInfoStartGenstarVector, avatar.GeneralStars)
	GenrelsVector := h.CreateStringArray(AccountInfoStartGenrelsVector, avatar.GeneralRels)
	GenrellvVector := h.CreateUInt32Array(AccountInfoStartGenrellvVector, avatar.GeneralRelLevels)
	AvatarJadeVector := h.CreateStringArray(AccountInfoStartAvatarJadeVector, avatar.EquipJade)
	DestGeneralJadeVector := h.CreateStringArray(AccountInfoStartDestGeneralJadeVector, avatar.DestGeneralJade)
	DgssVector := h.CreateIntArray(AccountInfoStartDgssVector, avatar.CurrDestinyGeneralSkill)
	TitlesVector := h.CreateStringArray(AccountInfoStartTitlesVector, avatar.Title)
	AttrUOffsetT := GenAttr(builder, avatar.Attr)
	PskillidVector := h.CreateStringArray(AccountInfoStartPskillidVector, avatar.PassiveSkillId)
	CskillidVector := h.CreateStringArray(AccountInfoStartCskillidVector, avatar.CounterSkillId)
	TskillidVector := h.CreateStringArray(AccountInfoStartTskillidVector, avatar.TriggerSkillId)

	AccountInfoStart(builder)
	AccountInfoAddIdx(builder, int32(idx))
	AccountInfoAddAccountId(builder, h.Get(AcID))
	AccountInfoAddAvatarId(builder, int32(avatar.AvatarId))
	AccountInfoAddCorpLv(builder, avatar.CorpLv)
	AccountInfoAddCorpXp(builder, avatar.CorpXP)
	AccountInfoAddArousals(builder, ArousalsVector)
	AccountInfoAddSkills(builder, SkillsVector)
	AccountInfoAddSkillps(builder, SkillpsVector)
	AccountInfoAddHeroStar(builder, HeroStarVector)
	AccountInfoAddHeroLv(builder, HeroLvlVector)
	AccountInfoAddName(builder, h.Get(Name))
	AccountInfoAddVip(builder, avatar.VipLv)
	AccountInfoAddAvatarlockeds(builder, AvatarlockedsVector)
	AccountInfoAddGs(builder, int32(avatar.Gs))
	AccountInfoAddPvpScore(builder, avatar.SimplePvpScore)
	AccountInfoAddPvpRank(builder, int32(avatar.SimplePvpRank))
	AccountInfoAddEquips(builder, EquipsVector)
	AccountInfoAddEquipUpgrade(builder, EquipUpgradeVector)
	AccountInfoAddEquipStar(builder, EquipStarVector)
	AccountInfoAddAvatarEquips(builder, AvatarEquipsVector)
	AccountInfoAddAllFashions(builder, AllFashionsVector)
	AccountInfoAddGenerals(builder, GeneralsVector)
	AccountInfoAddGenstar(builder, GenstarVector)
	AccountInfoAddGenrels(builder, GenrelsVector)
	AccountInfoAddGenrellv(builder, GenrellvVector)
	AccountInfoAddAvatarJade(builder, AvatarJadeVector)
	AccountInfoAddDestGeneralJade(builder, DestGeneralJadeVector)
	AccountInfoAddDg(builder, int32(avatar.DestinyGeneralID))
	AccountInfoAddDglv(builder, int32(avatar.DestinyGeneralLv))
	AccountInfoAddDgss(builder, DgssVector)
	AccountInfoAddGuuid(builder, h.Get(Guuid))
	AccountInfoAddGname(builder, h.Get(Gname))
	AccountInfoAddGpos(builder, int32(avatar.GuildPos))
	AccountInfoAddPost(builder, h.Get(Post))
	AccountInfoAddPostt(builder, (avatar.GuildPostTime))
	AccountInfoAddTitle(builder, h.Get(TitleOn))
	AccountInfoAddTitles(builder, TitlesVector)
	AccountInfoAddHeroAttr(builder, AttrUOffsetT)
	AccountInfoAddPskillid(builder, PskillidVector)
	AccountInfoAddCskillid(builder, CskillidVector)
	AccountInfoAddTskillid(builder, TskillidVector)
	AccountInfoAddHeroSwing(builder, int32(avatar.HeroSwing))
	AccountInfoAddMagicPetfigure(builder, uint32(avatar.MagicPetfigure))

	return AccountInfoEnd(builder)
}

func GenPlayerStateWithPos(builder *flatbuffers.Builder, stat GVGPlayerState, lead string) flatbuffers.UOffsetT {
	h := NewFlatBufferHelper(builder, 16)
	ID := h.Pre(builder.CreateString(stat.AcID))
	length := len(stat.Hp)
	PlayerStateStartAvatarIDVector(builder, length)
	key := make([]int32, 0, 3)
	for k, _ := range stat.Hp {
		builder.PrependInt32(int32(k))
		key = append(key, k)
	}
	playerAvatarVec := builder.EndVector(length)

	PlayerStateStartHpVector(builder, length)
	for _, item := range key {
		builder.PrependInt32(stat.Hp[item])
	}
	playerHpVec := builder.EndVector(length)

	PlayerStateStart(builder)
	PlayerStateAddAccountID(builder, h.Get(ID))
	PlayerStateAddState(builder, int32(stat.State))
	PlayerStateAddHp(builder, playerHpVec)
	PlayerStateAddAvatarID(builder, playerAvatarVec)
	if lead == stat.AcID {
		PlayerStateAddPos(builder, int32(1))
	} else {
		PlayerStateAddPos(builder, int32(0))
	}
	PlayerStateAddCurAvatar(builder, stat.AvatarID)
	return PlayerStateEnd(builder)
}

func GenPlayerState(builder *flatbuffers.Builder, stat GVGPlayerState) flatbuffers.UOffsetT {
	h := NewFlatBufferHelper(builder, 16)
	ID := h.Pre(builder.CreateString(stat.AcID))
	PlayerStateStart(builder)
	PlayerStateAddAccountID(builder, h.Get(ID))
	PlayerStateAddState(builder, int32(stat.State))
	return PlayerStateEnd(builder)
}

func GenStateParam(builder *flatbuffers.Builder, id string, param string) flatbuffers.UOffsetT {
	h := NewFlatBufferHelper(builder, 16)
	ID := h.Pre(builder.CreateString(id))
	Param := h.Pre(builder.CreateString(param))
	StateParamStart(builder)
	StateParamAddAccountID(builder, h.Get(ID))
	StateParamAddParam(builder, h.Get(Param))
	return StateParamEnd(builder)
}

func GenBossState(builder *flatbuffers.Builder, stat GVGBossState) flatbuffers.UOffsetT {
	length := len(stat.Hatred)
	BossStateStartHatredVector(builder, length)
	for i := length - 1; i >= 0; i-- {
		builder.PrependInt32(int32(stat.Hatred[i]))
	}
	HatredVector := builder.EndVector(length)
	BossStateStart(builder)
	BossStateAddHp(builder, int32(stat.Hp))
	BossStateAddHatred(builder, HatredVector)
	return BossStateEnd(builder)
}

func GenAttr(builder *flatbuffers.Builder, attr helper.AvatarAttr_) flatbuffers.UOffsetT {
	AttrStart(builder)
	AttrAddAtk(builder, attr.ATK)
	AttrAddDef(builder, attr.DEF)
	AttrAddHp(builder, attr.HP)
	AttrAddCritRate(builder, attr.CritRate)
	AttrAddResilienceRate(builder, attr.ResilienceRate)
	AttrAddCritValue(builder, attr.CritValue)
	AttrAddResilienceValue(builder, attr.ResilienceValue)
	AttrAddIceDamage(builder, attr.IceDamage)
	AttrAddIceBonus(builder, attr.IceBonus)
	AttrAddIceResist(builder, attr.IceResist)
	AttrAddFireDamage(builder, attr.FireDamage)
	AttrAddFireDefense(builder, attr.FireDefense)
	AttrAddFireBonus(builder, attr.FireBonus)
	AttrAddFireResist(builder, attr.FireResist)
	AttrAddLightingDamage(builder, attr.LightingDamage)
	AttrAddLightingDefense(builder, attr.LightingDefense)
	AttrAddLightingBonus(builder, attr.LightingBonus)
	AttrAddLightingResist(builder, attr.LightingResist)
	AttrAddPoisonDamage(builder, attr.PoisonDamage)
	AttrAddPoisonDefense(builder, attr.PoisonDefense)
	AttrAddPoisonBonus(builder, attr.PoisonBonus)
	AttrAddPoisonResist(builder, attr.PoisonResist)
	return AttrEnd(builder)
}
