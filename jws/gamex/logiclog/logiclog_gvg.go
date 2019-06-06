package logiclog

import (
	"fmt"
	"vcs.taiyouxi.net/platform/planx/util/logiclog"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type LogicInfo_GvgInfo struct {
	PlayerId        string
	PlayerCityId    int
	PlayerGs        int
	PlayerCropLvl   uint32
	PlayerAvatarIds []int
	EnemyId         string
	EnemyCityId     int
	EnemyGs         int
	EnemyCropLvl    uint32
	EnemyAvatarIds  []int
	IsWin           int32
	WinCount        int
}

type LogicInfo_GvgGuildInfo struct {
	GuildId          string
	GuildName        string
	GuildMemberCount int
	GuildPlayerCount int
	GuildHoldCity    int
	GuildSource      int64
}

type LogicInfo_GvgGuildplayInfo struct {
	PlayerCount int
	HoldCityId  int
}

type LogicInfo_GvgGuildCityScoreInfo struct {
	Infos []LogicInfo_GvgGuildCityScoreItem
}

type LogicInfo_GvgGuildCityScoreItem struct {
	CityID    int
	GuildInfo []LogicInfo_GvgGuildScoreItem
}

type LogicInfo_GvgGuildScoreItem struct {
	GuildID   string
	GuildName string
	Score     int
}

func LogGvGFinishFight(accountId string, avatar int, corpLvl uint32, channel string,
	playerCityId int, playerGs int, playerAvatarIds []int, wincount int, iswin int32,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[GvGFinishFight][%s]  playerCityId %d info %v ", accountId, playerCityId, info)

	r := LogicInfo_GvgInfo{
		PlayerId:        accountId,
		PlayerCityId:    playerCityId,
		PlayerGs:        playerGs,
		PlayerCropLvl:   corpLvl,
		PlayerAvatarIds: playerAvatarIds,
		IsWin:           iswin,
		WinCount:        wincount,
	}
	TypeInfo := LogicTag_GvGFinishFight
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogGvGstartFight(accountId string, avatar int, corpLvl uint32, channel string,
	playerCityId int, playerGs int, playerAvatarIds []int, enemyId string, enemyCityid int, enemyGs int,
	enemycroplvl uint32, enemyAvatarids []int,
	fgs GetLastSetCurLogType, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[GvGFinishFight][%s]  playerCityId %d info %v ", accountId, playerCityId, info)

	r := LogicInfo_GvgInfo{
		PlayerId:        accountId,
		PlayerCityId:    playerCityId,
		PlayerGs:        playerGs,
		PlayerCropLvl:   corpLvl,
		PlayerAvatarIds: playerAvatarIds,
		EnemyId:         enemyId,
		EnemyCityId:     enemyCityid,
		EnemyGs:         enemyGs,
		EnemyCropLvl:    enemycroplvl,
		EnemyAvatarIds:  enemyAvatarids,
	}
	TypeInfo := LogicTag_GvGstartFight
	logiclog.Error(accountId, avatar, corpLvl, channel, TypeInfo, r, fgs(TypeInfo), format, params...)
}

func LogGvgGuildInfo(guildId string, guildName string, guildMemberCount int,
	guildPlayerCount int, guildHoldCity int, guildSource int64, format string, params ...interface{}) {

	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[GvGGuildFinish][%s]  guildId %d guildName %s  %s", guildId, guildName, info)

	r := LogicInfo_GvgGuildInfo{
		GuildId:          guildId,
		GuildName:        guildName,
		GuildMemberCount: guildMemberCount,
		GuildPlayerCount: guildPlayerCount,
		GuildHoldCity:    guildHoldCity,
		GuildSource:      guildSource,
	}
	TypeInfo := LogicTag_GvGGuildFinish
	logiclog.ErrorForGuild("", "", guildId, TypeInfo, r, format, params...)
}

func LogGvgGuildScoreInfo(logicInfo *LogicInfo_GvgGuildCityScoreInfo, format string, params ...interface{}) {
	format = BITag + format
	info := fmt.Sprintf(format, params...)
	logs.Trace("[GvGGuildInterruptByGMTool]  %s", info)
	r := *logicInfo
	TypeInfo := LogicTag_GvGGuildScoreGM
	logiclog.ErrorForGuild("", "", "", TypeInfo, r, format, params...)
}
