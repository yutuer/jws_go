package fenghuoproto

import (
	"github.com/google/flatbuffers/go"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/msgprocessor"
	. "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto"
	. "vcs.taiyouxi.net/jws/multiplayer/multiplay_server/server/proto/fenghuomsg"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func GenErrorRspPacket(req msgprocessor.IPacket, code int) []byte {
	builder := GetNewBuilder()
	PacketStart(builder)
	PacketAddTyp(builder, int32(msgprocessor.MsgTypRpc))
	PacketAddNumber(builder, req.Number())
	PacketAddCode(builder, int32(code))
	pkt := PacketEnd(builder)
	builder.Finish(pkt)
	return builder.FinishedBytes()
}

func GenPacketRsp(builder *flatbuffers.Builder, req msgprocessor.IPacket, dataTyp byte, datas flatbuffers.UOffsetT) []byte {
	return GenPacketRspBasic(builder, req.Typ(), req.Number(), req.Code(), dataTyp, datas)
}

func GenPacketRspBasic(builder *flatbuffers.Builder, typ int32, number int64, code int32, dataTyp byte, datas flatbuffers.UOffsetT) []byte {
	PacketStart(builder)
	PacketAddTyp(builder, (typ))
	PacketAddNumber(builder, (number))
	PacketAddDataType(builder, dataTyp)
	PacketAddData(builder, datas)
	PacketAddCode(builder, (code))
	builder.Finish(PacketEnd(builder))
	return builder.Bytes[builder.Head():]
}

func GenEquipData(builder *flatbuffers.Builder, eq helper.BagItemToClient) flatbuffers.UOffsetT {
	TableID := builder.CreateString(eq.TableID)
	ItemID := builder.CreateString(eq.ItemID)
	Data := builder.CreateString(eq.Data)

	EquipInfoStart(builder)
	EquipInfoAddId(builder, eq.ID)
	EquipInfoAddTableid(builder, TableID)
	EquipInfoAddItemid(builder, ItemID)
	EquipInfoAddCount(builder, eq.Count)
	EquipInfoAddData(builder, Data)
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
	// TODO 新的战斗属性 命中和闪避
	return AttrEnd(builder)
}

func GenAccountInfoData(builder *flatbuffers.Builder, idx int, Data *helper.Avatar2ClientByJson) flatbuffers.UOffsetT {
	avatar := Data
	eqs := avatar.GetEquips()
	afs := avatar.GetAllFashions()

	AcID := builder.CreateString(avatar.GetAcId())
	Name := builder.CreateString(avatar.Name)
	Guuid := builder.CreateString(avatar.GuildUUID)
	Gname := builder.CreateString(avatar.GuildName)
	Post := builder.CreateString(avatar.GuildPost)
	TitleOn := builder.CreateString(avatar.TitleOn)

	HeroStarVector := GenUInt32Array(builder, AccountInfoStartHeroStarVector, avatar.HeroStarLv)
	HeroLvlVector := GenUInt32Array(builder, AccountInfoStartHeroStarVector, avatar.HeroLv)
	ArousalsVector := GenUInt32Array(builder, AccountInfoStartArousalsVector, avatar.Arousals)
	SkillsVector := GenUInt32Array(builder, AccountInfoStartSkillsVector, avatar.AvatarSkills)
	SkillpsVector := GenUInt32Array(builder, AccountInfoStartSkillpsVector, avatar.SkillPractices)
	AvatarlockedsVector := GenIntArray(builder, AccountInfoStartAvatarlockedsVector, avatar.AvatarLockeds)

	logs.Trace("EquipsUPffsetTs %v", eqs)
	EquipsUPffsetTs := make([]flatbuffers.UOffsetT, 0, len(eqs))
	for i := 0; i < len(eqs); i++ {
		EquipsUPffsetTs = append(EquipsUPffsetTs, GenEquipData(builder, eqs[i]))
	}
	EquipsVector := GenUOffsetTArray(builder, AccountInfoStartEquipsVector, EquipsUPffsetTs)

	EquipUpgradeVector := GenUInt32Array(builder, AccountInfoStartEquipUpgradeVector, avatar.EquipUpgrade)
	EquipStarVector := GenUInt32Array(builder, AccountInfoStartEquipStarVector, avatar.EquipStar)
	AvatarEquipsVector := GenUInt32Array(builder, AccountInfoStartAvatarEquipsVector, avatar.AvatarEquips)

	AllFashionsUPffsetTs := make([]flatbuffers.UOffsetT, 0, len(eqs))
	for i := 0; i < len(afs); i++ {
		AllFashionsUPffsetTs = append(AllFashionsUPffsetTs, GenFashionItemData(builder, afs[i]))
	}
	AllFashionsVector := GenUOffsetTArray(builder, AccountInfoStartAllFashionsVector, AllFashionsUPffsetTs)

	GeneralsVector := GenStringArray(builder, AccountInfoStartGeneralsVector, avatar.Generals)
	GenstarVector := GenUInt32Array(builder, AccountInfoStartGenstarVector, avatar.GeneralStars)
	GenrelsVector := GenStringArray(builder, AccountInfoStartGenrelsVector, avatar.GeneralRels)
	GenrellvVector := GenUInt32Array(builder, AccountInfoStartGenrellvVector, avatar.GeneralRelLevels)
	AvatarJadeVector := GenStringArray(builder, AccountInfoStartAvatarJadeVector, avatar.EquipJade)
	DestGeneralJadeVector := GenStringArray(builder, AccountInfoStartDestGeneralJadeVector, avatar.DestGeneralJade)
	DgssVector := GenIntArray(builder, AccountInfoStartDgssVector, avatar.CurrDestinyGeneralSkill)
	TitlesVector := GenStringArray(builder, AccountInfoStartTitlesVector, avatar.Title)

	AttrUOffsetT := GenAttr(builder, avatar.Attr)
	PskillidVector := GenStringArray(builder, AccountInfoStartPskillidVector, avatar.PassiveSkillId)
	CskillidVector := GenStringArray(builder, AccountInfoStartCskillidVector, avatar.CounterSkillId)
	TskillidVector := GenStringArray(builder, AccountInfoStartTskillidVector, avatar.TriggerSkillId)

	AccountInfoStart(builder)
	AccountInfoAddIdx(builder, int32(idx))
	AccountInfoAddAccountId(builder, AcID)
	AccountInfoAddAvatarId(builder, int32(avatar.AvatarId))
	AccountInfoAddCorpLv(builder, avatar.CorpLv)
	AccountInfoAddCorpXp(builder, avatar.CorpXP)
	AccountInfoAddArousals(builder, ArousalsVector)
	AccountInfoAddSkills(builder, SkillsVector)
	AccountInfoAddSkillps(builder, SkillpsVector)
	AccountInfoAddHeroStar(builder, HeroStarVector)
	AccountInfoAddHeroLv(builder, HeroLvlVector)
	AccountInfoAddName(builder, Name)
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
	AccountInfoAddGuuid(builder, Guuid)
	AccountInfoAddGname(builder, Gname)
	AccountInfoAddGpos(builder, int32(avatar.GuildPos))
	AccountInfoAddPost(builder, Post)
	AccountInfoAddPostt(builder, (avatar.GuildPostTime))
	AccountInfoAddTitle(builder, TitleOn)
	AccountInfoAddTitles(builder, TitlesVector)
	AccountInfoAddHeroAttr(builder, AttrUOffsetT)
	AccountInfoAddPskillid(builder, PskillidVector)
	AccountInfoAddCskillid(builder, CskillidVector)
	AccountInfoAddTskillid(builder, TskillidVector)
	AccountInfoAddHeroSwing(builder, int32(avatar.HeroSwing))

	return AccountInfoEnd(builder)
}

func ForwardHPNotifyToClient(notify *HPNotify) []byte {
	builder := GetNewBuilder()
	l := notify.EnemiesHpDLength()
	HPNotifyStartEnemiesHpDVector(builder, l)
	for i := l - 1; i >= 0; i-- {
		builder.PrependInt32(notify.EnemiesHpD(i))
	}
	ehp := builder.EndVector(l)

	HPNotifyStart(builder)
	HPNotifyAddMyidx(builder, notify.Myidx())
	HPNotifyAddMyHpD(builder, notify.MyHpD())
	HPNotifyAddEnemiesHpD(builder, ehp)
	hpn := HPNotifyEnd(builder)
	return GenPacketRspBasic(builder, msgprocessor.MsgTypNotify,
		0, 0, DatasHPNotify, hpn)

}

type flatbufVecStartFunc func(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT

func GenIntArray(builder *flatbuffers.Builder, fc flatbufVecStartFunc, arr []int) flatbuffers.UOffsetT {
	fc(builder, len(arr))
	for i := len(arr) - 1; i >= 0; i-- {
		builder.PrependInt32(int32(arr[i]))
	}
	ret := builder.EndVector(len(arr))
	return ret
}

func GenUInt32Array(builder *flatbuffers.Builder, fc flatbufVecStartFunc, arr []uint32) flatbuffers.UOffsetT {
	fc(builder, len(arr))
	for i := len(arr) - 1; i >= 0; i-- {
		builder.PrependUint32(arr[i])
	}
	ret := builder.EndVector(len(arr))
	return ret
}

func GenStringArray(builder *flatbuffers.Builder, fc flatbufVecStartFunc, arr []string) flatbuffers.UOffsetT {
	index := make([]flatbuffers.UOffsetT, len(arr))
	for i := len(arr) - 1; i >= 0; i-- {
		index[i] = builder.CreateString(arr[i])
	}

	fc(builder, len(arr))
	for i := len(arr) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(index[i])
	}
	ret := builder.EndVector(len(arr))
	return ret
}

func GenUOffsetTArray(builder *flatbuffers.Builder, fc flatbufVecStartFunc, arr []flatbuffers.UOffsetT) flatbuffers.UOffsetT {
	fc(builder, len(arr))
	for i := len(arr) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(arr[i])
	}
	ret := builder.EndVector(len(arr))
	return ret
}
