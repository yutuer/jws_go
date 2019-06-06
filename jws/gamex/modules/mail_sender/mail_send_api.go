package mail_sender

import (
	"time"

	"fmt"

	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

const (
	year_time = 365 * day_time
	day_time  = 24 * 3600
)

func SendMail2Account(accountID string, mailTyp int64, titleIDS int,
	param []string, items map[string]uint32, reason string) {

	m := genMail2Account(accountID, mailTyp, titleIDS, param, items, reason)
	a, _ := db.ParseAccount(accountID)
	GetModule(a.ShardId).sendMailCmd(m)
}

func BatchSendMail2Account(accountID string, mailTyp int64, titleIDS int,
	param []string, items map[string]uint32, reason string, isAct bool) error {
	m := genMail2Account(accountID, mailTyp, titleIDS, param, items, reason)
	a, err := db.ParseAccount(accountID)
	if nil != err {
		return fmt.Errorf("BatchSendMail2Account Failed, %v", err)
	}
	AddMailBatchSend(a.ShardId, accountID, m.Mail, isAct, m.Typ)

	return nil
}

func genMail2Account(accountID string, mailTyp int64, titleIDS int,
	param []string, items map[string]uint32, reason string) *MailToUser {

	ids := make([]string, 0, len(items))
	counts := make([]uint32, 0, len(items))
	for k, v := range items {
		ids = append(ids, k)
		counts = append(counts, v)
	}
	m := &MailToUser{
		Typ: mailTyp,
		Uid: accountID,
	}
	m.Mail.TimeBegin = time.Now().Unix()
	m.Mail.TimeEnd = m.Mail.TimeBegin + int64(day_time*7)
	m.Mail.IdsID = uint32(titleIDS)
	m.Mail.Param = param
	m.Mail.ItemId = ids
	m.Mail.Count = counts
	m.Mail.Reason = reason

	return m
}

func SendAndroidIAPMail(accountId, info string) {
	m := &MailToUser{
		Typ: timail.Mail_Send_By_AndroidIAP,
		Uid: accountId,
	}
	m.Mail.TimeBegin = time.Now().Unix()
	m.Mail.TimeEnd = m.Mail.TimeBegin + int64(year_time)
	m.Mail.Tag = info

	a, _ := db.ParseAccount(accountId)
	GetModule(a.ShardId).sendMailCmd(m)
}

func SendPayFeedBack_VN(accountId string, money int, hcBuy, hcGive uint32) {
	m := &MailToUser{
		Typ: timail.Mail_Send_By_Pay_FeedBack,
		Uid: accountId,
	}
	m.Mail = timail.MailReward{
		TimeBegin: time.Now().Unix(),
		TimeEnd:   time.Now().Unix() + util.YearSec,
		IdsID:     uint32(IDS_MAIL_GUILD_FIRSTBACKHC_TITLE),
		Param: []string{fmt.Sprintf("%d", money),
			fmt.Sprintf("%d", hcBuy), fmt.Sprintf("%d", hcGive)},
		ItemId: []string{VI_Hc_Buy, VI_Hc_Give},
		Count:  []uint32{hcBuy, hcGive},
		Reason: "SendPayFeedBack",
	}
	a, _ := db.ParseAccount(accountId)
	GetModule(a.ShardId).sendMailCmd(m)
}

func SendPayFeedBack_HMT(accountId string, hc int) {
	m := &MailToUser{
		Typ: timail.Mail_Send_By_Pay_FeedBack,
		Uid: accountId,
	}
	m.Mail = timail.MailReward{
		TimeBegin: time.Now().Unix(),
		TimeEnd:   time.Now().Unix() + util.YearSec,
		IdsID:     uint32(IDS_MAIL_HMT_PAY_FEED_BACK_TITLE),
		Param:     []string{fmt.Sprintf("%d", hc)},
		ItemId:    []string{VI_Hc_Give},
		Count:     []uint32{uint32(hc)},
		Reason:    "SendPayFeedBack",
	}
	a, _ := db.ParseAccount(accountId)
	GetModule(a.ShardId).sendMailCmd(m)
}

func SendPayFeedBack(accountId string, money, hc1, hc2 int, isFirst bool) {
	ids := IDS_MAIL_GUILD_SCENDBACKHC_TITLE
	if isFirst {
		ids = IDS_MAIL_GUILD_FIRSTBACKHC_TITLE
	}
	m := &MailToUser{
		Typ: timail.Mail_Send_By_Pay_FeedBack,
		Uid: accountId,
	}
	m.Mail = timail.MailReward{
		TimeBegin: time.Now().Unix(),
		TimeEnd:   time.Now().Unix() + util.YearSec,
		IdsID:     uint32(ids),
		Param: []string{fmt.Sprintf("%d", money),
			fmt.Sprintf("%d", hc1),
			fmt.Sprintf("%d", hc2)},
		ItemId: []string{VI_Hc_Give},
		Count:  []uint32{uint32(hc1 + hc2)},
		Reason: "SendPayFeedBack",
	}
	a, _ := db.ParseAccount(accountId)
	GetModule(a.ShardId).sendMailCmd(m)
}

func SendGangKick(accountId, actMem, guildName string) {
	m := &MailToUser{
		Typ: timail.Mail_Send_By_Single_Player,
		Uid: accountId,
	}
	m.Mail.TimeBegin = time.Now().Unix()
	m.Mail.TimeEnd = m.Mail.TimeBegin + int64(day_time*7)
	m.Mail.IdsID = IDS_MAIL_GUILD_KICKEDOUT_TITLE
	m.Mail.Param = []string{actMem, guildName}

	a, _ := db.ParseAccount(accountId)
	GetModule(a.ShardId).sendMailCmd(m)
}

func SendGangDismiss(accountId, chiefMem, guildName string) {
	m := &MailToUser{
		Typ: timail.Mail_Send_By_Single_Player,
		Uid: accountId,
	}
	m.Mail.TimeBegin = time.Now().Unix()
	m.Mail.TimeEnd = m.Mail.TimeBegin + int64(day_time*7)
	m.Mail.IdsID = IDS_MAIL_GUILD_DISBANDED_TITLE
	m.Mail.Param = []string{chiefMem, guildName}

	a, _ := db.ParseAccount(accountId)
	GetModule(a.ShardId).sendMailCmd(m)
}

func SendGangApplyRefuse(accountId, guildName string) {
	m := &MailToUser{
		Typ: timail.Mail_Send_By_Single_Player,
		Uid: accountId,
	}
	m.Mail.TimeBegin = time.Now().Unix()
	m.Mail.TimeEnd = m.Mail.TimeBegin + int64(day_time*7)
	m.Mail.IdsID = IDS_MAIL_GUILD_DECLINE_TITLE
	m.Mail.Param = []string{guildName}

	a, _ := db.ParseAccount(accountId)
	GetModule(a.ShardId).sendMailCmd(m)
}

func SendGangPosChg(accountId string, oldPos, newPos int) {
	m := &MailToUser{
		Typ: timail.Mail_Send_By_Single_Player,
		Uid: accountId,
	}
	m.Mail.TimeBegin = time.Now().Unix()
	m.Mail.TimeEnd = m.Mail.TimeBegin + int64(day_time*7)
	m.Mail.IdsID = IDS_MAIL_GUILD_CHANGELEVEL_TITLE
	m.Mail.Param = []string{fmt.Sprintf("%d", oldPos), fmt.Sprintf("%d", newPos)}

	a, _ := db.ParseAccount(accountId)
	GetModule(a.ShardId).sendMailCmd(m)
}

func SendGeneralQuestReward(accountId string, qid string, hash int, outTime int64, item_id []string, count []uint32) {
	m := &MailToUser{
		Typ: timail.Mail_Send_By_Single_Player,
		Uid: accountId,
	}
	m.Mail.TimeBegin = time.Now().Unix()
	m.Mail.TimeEnd = m.Mail.TimeBegin + int64(day_time*7)
	m.Mail.IdsID = IDS_MAIL_GENERAL_QUESTREWARD_TITLE
	m.Mail.Param = []string{fmt.Sprintf("%s,%d,%d", qid, hash, outTime)}
	m.Mail.ItemId = item_id
	m.Mail.Count = count
	m.Mail.Reason = "GeneralQuestMail"

	logs.Trace("SendGeneralQuestReward %s %v", accountId, m)

	a, _ := db.ParseAccount(accountId)
	GetModule(a.ShardId).sendMailCmd(m)
}

func SendFashionTimeOut(accountId, itemId string) {
	m := &MailToUser{
		Typ: timail.Mail_Send_By_Single_Player,
		Uid: accountId,
	}
	m.Mail.TimeBegin = time.Now().Unix()
	m.Mail.TimeEnd = m.Mail.TimeBegin + int64(day_time*7)
	m.Mail.IdsID = IDS_MAIL_FASHION_TIMEOUT_TITLE
	m.Mail.Param = []string{itemId}

	a, _ := db.ParseAccount(accountId)
	GetModule(a.ShardId).sendMailCmd(m)
}

func SendGuildInventory(accountId, guildName, loot string, item []string, count []uint32) {
	m := &MailToUser{
		Typ: timail.Mail_Send_By_Guild_Inventory,
		Uid: accountId,
	}
	m.Mail.TimeBegin = time.Now().Unix()
	m.Mail.TimeEnd = m.Mail.TimeBegin + int64(day_time*7)
	m.Mail.IdsID = IDS_MAIL_GUILD_GVEBOSSBAG_TITLE
	m.Mail.Param = []string{guildName, loot}
	m.Mail.ItemId = item
	m.Mail.Count = count
	m.Mail.Reason = "GuildInventory"

	a, _ := db.ParseAccount(accountId)
	GetModule(a.ShardId).sendMailCmd(m)
}

func BatchSend7DayGuildReward(shard uint, acid string, time_now int64, rank int,
	ritems []string, rcounts []uint32, addon int64) {
	mail := timail.MailReward{
		TimeBegin: time_now,
		TimeEnd:   time_now + util.WeekSec,
		IdsID:     IDS_MAIL_ACTIVITY_7DAYGUILD_TITLE,
		Param:     []string{fmt.Sprintf("%d", rank)},
		ItemId:    ritems,
		Count:     rcounts,
		Reason:    "SendSeverOpenGuildRankReward",
		Idx:       timail.MkMailIdByTime(time_now, timail.Mail_Send_By_Rank_ServerOpn_Guild, addon),
	}
	AddMailBatchSend(shard, acid, mail, false, timail.Mail_Send_By_Rank_ServerOpn_Guild)
}

func BatchSend7DayGuildLeaderReward(shard uint, leader string, time_now int64, rank int,
	ritems []string, rcounts []uint32, addon int64) {
	mail := timail.MailReward{
		TimeBegin: time_now,
		TimeEnd:   time_now + util.WeekSec,
		IdsID:     IDS_MAIL_ACTIVITY_7DAYGUILDLEAD_TITLE,
		Param:     []string{fmt.Sprintf("%d", rank)},
		ItemId:    ritems,
		Count:     rcounts,
		Reason:    "SendSeverOpenGuildRankReward",
		Idx:       timail.MkMailIdByTime(time_now, timail.Mail_Send_By_Rank_ServerOpn_Guild_leader, addon),
	}
	AddMailBatchSend(shard, leader, mail, false, timail.Mail_Send_By_Rank_ServerOpn_Guild_leader)
}

func BatchSend7DayPlayerRankReward(shardId uint, acid string, rank int,
	items []string, counts []uint32) {
	time_now := game.GetNowTimeByOpenServer(shardId)
	mail := timail.MailReward{
		TimeBegin: time_now,
		TimeEnd:   time_now + util.WeekSec,
		Param:     []string{fmt.Sprintf("%d", rank)},
		ItemId:    items,
		Count:     counts,
		IdsID:     IDS_MAIL_ACTIVITY_7DAYRANK_TITLE,
		Reason:    "SeverOpenRankReward",
		Idx:       timail.MkMailIdByTime(time_now, timail.Mail_Send_By_Rank_ServerOpn_Player_Rank, 0),
	}

	AddMailBatchSend(shardId, acid, mail, false, timail.Mail_Send_By_Rank_ServerOpn_Player_Rank)
}

func BatchSend7DayPlayerGsReward(shardId uint, acid string,
	items []string, counts []uint32) {
	time_now := game.GetNowTimeByOpenServer(shardId)
	mail := timail.MailReward{
		TimeBegin: time_now,
		TimeEnd:   time_now + util.WeekSec,
		Param:     []string{""},
		ItemId:    items,
		Count:     counts,
		IdsID:     IDS_MAIL_ACTIVITY_7DAYAWARD_TITLE,
		Reason:    "SeverOpenRankReward",
		Idx:       timail.MkMailIdByTime(time_now, timail.Mail_Send_By_Rank_ServerOpn_Player_Gs, 0),
	}
	AddMailBatchSend(shardId, acid, mail, false, timail.Mail_Send_By_Rank_ServerOpn_Player_Gs)
}

func BatchSendTeamPvpReward(shardId uint, acid string, rank int,
	items []string, counts []uint32) {
	time_now := game.GetNowTimeByOpenServer(shardId)
	mail := timail.MailReward{
		TimeBegin: time_now,
		TimeEnd:   time_now + util.WeekSec,
		Param:     []string{fmt.Sprintf("%d", rank)},
		ItemId:    items,
		Count:     counts,
		IdsID:     IDS_MAIL_TEAMPVP_REWARD,
		Reason:    "TeamPvpReward",
		Idx:       timail.MkMailIdByTime(time_now, timail.Mail_Send_By_TeamPvp, 0),
	}
	AddMailBatchSend(shardId, acid, mail, false, timail.Mail_Send_By_TeamPvp)
}

func SendMarketActivityMail(accountId string, titleIDS int, param []string, item_count map[string]uint32) {
	item_id := make([]string, 0, len(item_count))
	count := make([]uint32, 0, len(item_count))
	for k, v := range item_count {
		item_id = append(item_id, k)
		count = append(count, v)
	}
	m := &MailToUser{
		Typ: timail.Mail_Send_By_Market_Activity,
		Uid: accountId,
	}
	m.Mail.TimeBegin = time.Now().Unix()
	m.Mail.TimeEnd = m.Mail.TimeBegin + int64(day_time*7)
	m.Mail.IdsID = uint32(titleIDS)
	m.Mail.Param = param
	m.Mail.ItemId = item_id
	m.Mail.Count = count
	m.Mail.Reason = "MarketActivity"

	a, _ := db.ParseAccount(accountId)
	GetModule(a.ShardId).sendMailCmd(m)
}

func SendHeroGachaRaceChestMail(accountID string, items map[string]uint32) {
	ids := make([]string, 0, len(items))
	counts := make([]uint32, 0, len(items))
	for k, v := range items {
		ids = append(ids, k)
		counts = append(counts, v)
	}
	m := &MailToUser{
		Typ: timail.Mail_Send_By_Hero_GachaRace,
		Uid: accountID,
	}
	m.Mail.TimeBegin = time.Now().Unix()
	m.Mail.TimeEnd = m.Mail.TimeBegin + int64(day_time*7)
	m.Mail.IdsID = IDS_MAIL_ACTIVITY_HGRBOX_TITLE
	m.Mail.Param = []string{""}
	m.Mail.ItemId = ids
	m.Mail.Count = counts
	m.Mail.Reason = "HeroGachaRaceChest"

	a, _ := db.ParseAccount(accountID)
	GetModule(a.ShardId).sendMailCmd(m)
}

func BatchSendHeroGachaRaceRankMail(shardId uint, acid string, score, rank int,
	items []string, counts []uint32) {

	time_now := game.GetNowTimeByOpenServer(shardId)
	mail := timail.MailReward{
		TimeBegin: time_now,
		TimeEnd:   time_now + util.WeekSec,
		Param:     []string{fmt.Sprintf("%d", score), fmt.Sprintf("%d", rank)},
		ItemId:    items,
		Count:     counts,
		IdsID:     IDS_MAIL_ACTIVITY_HGRRANK_TITLE,
		Reason:    "HeroGachaRaceRank",
		Idx:       timail.MkMailIdByTime(time_now, timail.Mail_Send_By_Hero_GachaRace, 0),
	}
	AddMailBatchSend(shardId, acid, mail, true, timail.Mail_Send_By_Hero_GachaRace)
}

func BatchSendChangeChiefMail(shardId uint, acid string, params []string, addon int64) {
	now := time.Now().Unix()
	mail := timail.MailReward{
		TimeBegin: now,
		TimeEnd:   now + util.WeekSec,
		Param:     params,
		IdsID:     IDS_MAIL_GUILD_AUTO_CHANGE_CHIEF,
		Reason:    "GuildAutoChangeChief",
		Idx:       timail.MkMailIdByTime(now, timail.Mail_Send_By_Auto_Change_Chief, addon),
	}
	AddMailBatchSend(shardId, acid, mail, false, timail.Mail_Send_By_Auto_Change_Chief)
}

func BatchSendGuildBossDeath(shardId uint, acid string, bossName string, ids []string, counts []uint32, addon int64) {
	now := time.Now().Unix()
	mail := timail.MailReward{
		TimeBegin: now,
		TimeEnd:   now + util.WeekSec,
		Param:     []string{bossName},
		IdsID:     IDS_MAIL_ON_GUILD_BOSS_DIED_TITLE,
		Reason:    "OnGuildBossDied",
		Idx:       timail.MkMailIdByTime(now, timail.Mail_send_By_GuildBoss_Death, addon),
		ItemId:    ids,
		Count:     counts,
	}
	AddMailBatchSend(shardId, acid, mail, false, timail.Mail_send_By_GuildBoss_Death)
}
