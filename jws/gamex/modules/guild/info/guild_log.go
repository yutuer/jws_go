package guild_info

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

const (
	IDS_GUILD_LOG_1  = iota // {0}建立了军团
	IDS_GUILD_LOG_2         // {0}批准{1}加入军团
	IDS_GUILD_LOG_3         // {0}退出了军团
	IDS_GUILD_LOG_4         // {0}将{1}逐出军团
	IDS_GUILD_LOG_5         // {0}将{1}职位调整为{2}
	IDS_GUILD_LOG_6         //
	IDS_GUILD_LOG_7         // {0}成为军团长
	IDS_GUILD_LOG_8         // {0}更新了军团公告
	IDS_GUILD_LOG_9         // 军团等级提升到{0}
	IDS_GUILD_LOG_10        // {0}加入了军团
	IDS_GUILD_LOG_11        // {1}修改了军团名称。
)

type GuildLog struct {
	TimeStamp int64    `json:"ts"`
	CfgId     int      `json:"id"`
	Param     []string `json:"p"`
}

type GuildLogs struct {
	Logs []GuildLog `json:"logs"`
}

func (gLogs *GuildLogs) AddLog(logId int, param []string) {
	llimit := int(gamedata.GetCommonCfg().GetGuildLogLimit())
	if gLogs.Logs == nil {
		gLogs.Logs = make([]GuildLog, 0, llimit)
	}
	if len(gLogs.Logs) >= llimit {
		gLogs.Logs = append(gLogs.Logs[:0], gLogs.Logs[1:]...)
	}
	gLogs.Logs = append(gLogs.Logs, GuildLog{
		TimeStamp: time.Now().Unix(),
		CfgId:     logId,
		Param:     param,
	})
}
