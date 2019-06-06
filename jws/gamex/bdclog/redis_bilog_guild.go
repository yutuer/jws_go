package bdclog

import (
	"strings"

	"encoding/json"
	"fmt"

	"time"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/activity/guild_boss"
	"vcs.taiyouxi.net/jws/gamex/modules/guild/info"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

func (cs *BiScanRedis) bi_guild(key string, skey []string, val []byte, rh storehelper.ReadHandler) error {
	s := strings.SplitN(skey[1], ":", 2)
	if s[0] != "Info" {
		return nil
	}

	a, err := db.ParseAccount(s[1])
	if err != nil {
		return err
	}
	has := false
	for _, sid := range cs.shardId {
		if fmt.Sprintf("%d", a.ShardId) == sid {
			has = true
			break
		}
	}
	if !has {
		return nil
	}

	var dat map[string]string
	if err := json.Unmarshal(val, &dat); err != nil {
		return fmt.Errorf("BiScanRedis unmarshal val %v %s", err, key)
	}

	base := &guild_info.GuildSimpleInfo{}
	if err := json.Unmarshal([]byte(dat["Base"]), base); err != nil {
		return fmt.Errorf("BiScanRedis unmarshal GuildSimpleInfo %v %s %s", err, key, dat["base"])
	}

	mems := &[guild_info.MaxGuildMember]helper.AccountSimpleInfo{}
	if err := json.Unmarshal([]byte(dat["Members"]), mems); err != nil {
		return fmt.Errorf("BiScanRedis unmarshal []AccountSimpleInfo %v %s %s", err, key, dat["members"])
	}
	log_mems := make(MemInfos, 0, base.MemNum)
	for _, m := range mems {
		if m.AccountID == "" {
			continue
		}
		log_mems = append(log_mems, MemInfo{
			Acid:   m.AccountID,
			Name:   m.Name,
			CorpLv: m.CorpLv,
			Pos:    m.GuildPosition,
		})
	}

	sciences := &[guild_info.MaxGuildScienceCount]guild_info.GuildScience{}
	if err := json.Unmarshal([]byte(dat["Sciences"]), sciences); err != nil {
		return fmt.Errorf("BiScanRedis unmarshal []GuildScience %v %s %s", err, key, dat["Sciences"])
	}

	t_now := time.Now().In(cs.loc)
	t_now = t_now.Add(-1 * time.Hour) // 如果用当前时间跑过去一天的数据，则当前时间减去1小时
	if cs.timestamp > 0 {
		t_now = time.Unix(cs.timestamp, 0).In(cs.loc)
	}
	tb := util.DailyBeginUnix(t_now.Unix())

	guildAct := &logiclog.DailyStatistics{}
	if err := json.Unmarshal([]byte(dat["ActivePlayerStatistic"]), guildAct); err != nil {
		return fmt.Errorf("BiScanRedis unmarshal ActivePlayerStatistic %v %s %s", err, key, dat["ActivePlayerStatistic"])
	}
	actC := 0
	actMem := ""
	for _, info := range guildAct.Infos {
		if info.TS == tb {
			actC = len(info.JoinMem)
			actMem = fmt.Sprintf("\"%v\"", info.JoinMem)
			break
		}
	}

	actBoss := &guild_boss.ActivityState{}
	if err := json.Unmarshal([]byte(dat["ActBoss"]), actBoss); err != nil {
		return fmt.Errorf("BiScanRedis unmarshal guild_boss %v %s %s", err, key, dat["ActBoss"])
	}
	gBossC := 0
	gBossJoinC := 0
	gBossMem := ""
	for _, info := range actBoss.Statictic.Infos {
		if info.TS == tb {
			gBossJoinC = info.JoinCount
			gBossC = len(info.JoinMem)
			gBossMem = fmt.Sprintf("\"%v\"", info.JoinMem)
			break
		}
	}

	t := time.Unix(base.CreateTS, 0).In(cs.loc)
	cts := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	csvStr := fmt.Sprintf("%s,%s,%d,%d,%s,"+ // 工会ID,工会名称,工会等级,工会经验,创建时间
		"%d,%d,"+ // 成员人数,工会总战力
		"%d,%s,"+ // 当天活跃人数,当天活跃成员id
		"%d,%d,%s,"+ //当天工会boss参与次数,当天工会boss参与人数,当天工会boss参与成员id
		"%s,"+ // 科技树
		"%s\r\n", // 成员信息
		base.GuildUUID, base.Name, base.Level, base.XpCurr, cts,
		base.MemNum, base.GuildGSSum,
		actC, actMem,
		gBossJoinC, gBossC, gBossMem,
		fmt.Sprintf("\"%v\"", sciences[1:]),
		log_mems.String())

	if _, err := cs.csv_guild_writer.WriteString(csvStr); err != nil {
		return err
	}

	logs.Info("RedisBiLog end key %s", key)
	return nil
}

type MemInfo struct {
	Acid   string
	Name   string
	CorpLv uint32
	Pos    int
}

func (ae MemInfo) String() string {
	return fmt.Sprintf("[%s,%s,%d,%d]", ae.Acid, ae.Name, ae.CorpLv, ae.Pos)
}

type MemInfos []MemInfo

func (es MemInfos) String() string {
	res := "\""
	for _, e := range es {
		res += e.String()
	}
	res += "\""
	return res
}
