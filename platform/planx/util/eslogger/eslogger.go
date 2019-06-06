package eslogger

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"

	"github.com/timesking/seelog"
	"vcs.taiyouxi.net/platform/planx/servers/db"
)

type ESLogger struct {
	logger seelog.LoggerInterface
	lock   sync.Mutex
}

type ESLoggerInfo struct {
	Level     string      `json:level`
	TimeOut   int64       `json:"logtime"`
	TimeKey   string      `json:"time"` //time_key in td-agent must string
	TimeES    string      `json:"@timestamp"`
	TimeUTC8  string      `json:"utc8"`
	Type      string      `json:"type_name"` //type_name for https://github.com/uken/fluent-plugin-elasticsearch
	Gid       uint        `json:"gid,omitempty"`
	Sid       uint        `json:"sid,omitempty"`
	UserId    string      `json:"userid,omitempty"`
	AccountID string      `json:"accountid,omitempty"`
	Avatar    string      `json:"avatar,omitempty"`
	CorpLvl   uint32      `json:"corplvl,omitempty"`
	Channel   string      `json:"channel,omitempty"`
	GuildID   string      `json:"guild,omitempty"`
	Info      interface{} `json:"info,omitempty"`
	Last      string      `json:"last,omitempty"`
	Extra     string      `json:"extra"`
}

/*
线上部署的时候应该考虑使用size rolling防止文件尺寸不均匀
https://github.com/cihub/seelog/wiki/Receiver-reference#rolling-file-writer-or-rotation-file-writer
<!--conf/logiclog.xml-->
<seelog>
    <outputs formatid="common">
        <rollingfile type="date" filename="logics" datepattern="02.01.2006.log" maxrolls="7" />
	</outputs>
	<formats>
		<format id="common"  format="%Msg%n"/>
	</formats>
</seelog>
*/
func (l *ESLogger) ReturnLoadLogger() func(n string, cmd config.LoadCmd) {
	return func(n string, cmd config.LoadCmd) {
		l.LoadLogger(n, cmd)
	}
}
func (l *ESLogger) LoadLogger(lcfgname string, cmd config.LoadCmd) {
	l.lock.Lock()
	defer l.lock.Unlock()

	switch cmd {
	case config.Load, config.Reload:
		logger, err := seelog.LoggerFromParamConfigAsFile(lcfgname, nil)
		if err != nil {
			logs.Warn("logs.LoadLogger load error, %s", err.Error())
			return
		}
		if l.logger == nil {
			l.logger = logger
			logs.Info("ESLogger loadlogger success %v", lcfgname)
		} else {
			l.logger.Flush()
			l.logger.Close()
			l.logger = logger
			logs.Info("ESLogger loadlogger close old an success %v", lcfgname)
		}

	case config.Unload:
		if l.logger != nil {
			l.logger.Flush()
			l.logger.Close()
		}
		l.logger = nil
		logs.Info("ESLogger loadlogger unload %v", lcfgname)
	}
}

func (l *ESLogger) StopLogger() {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.logger != nil {
		l.logger.Flush()
		l.logger.Close()
	}
	logs.Info("ESLogger StopLogger ")
}

func Error(logger *ESLogger, accountID string, avatar int, corpLvl uint32, channel string,
	guildID string, typeInfo string, info interface{},
	last string, format string, v ...interface{}) {
	if logger.logger == nil {
		//		logs.Error("logger error but nil")
		return
	}
	errorf := func(format string, params ...interface{}) {
		logger.logger.Errorf(format, params...)
	}
	send(errorf, "Error", accountID, avatar, corpLvl, channel, guildID, typeInfo, info, last, format, v...)
}

func Info(logger *ESLogger, accountID string, avatar int, corpLvl uint32, channel string,
	typeInfo string, info interface{}, format string, v ...interface{}) {
	if logger.logger == nil {
		//		logs.Error("logger info but nil")
		return
	}
	send(logger.logger.Infof, "Info", accountID, avatar, corpLvl, channel, "", typeInfo, info, "", format, v...)
}

func Debug(logger *ESLogger, accountID string, avatar int, corpLvl uint32, channel string,
	typeInfo string, info interface{},
	last string, format string, v ...interface{}) {
	if logger.logger == nil {
		//		logs.Error("logger info but nil")
		return
	}
	send(logger.logger.Debugf, "Debug", accountID, avatar, corpLvl, channel, "", typeInfo, info, last, format, v...)
}

func Trace(logger *ESLogger, accountID, typeInfo string, info interface{}, format string, v ...interface{}) {
	if logger.logger == nil {
		//		logs.Error("logger info but nil")
		return
	}
	send(logger.logger.Tracef, "Trace", accountID, -1, 0, "", "", typeInfo, info, "", format, v...)
}

// Send 记录玩家重要行为日志，用于客服和玩家行为分析
func send(logger func(format string, params ...interface{}),
	level string,
	accountID string, avatar int, corpLvl uint32, channel, guildID string, typeInfo string, info interface{}, last string,
	format string, v ...interface{}) {

	nt := time.Now()
	t := nt.UnixNano()
	st := fmt.Sprintf("%d", t)
	mst := nt.Format("2006-01-02T15:04:05.000Z07:00")
	utc := nt.UTC()
	utc8 := utc.Add(8 * time.Hour)
	utc8st := utc8.Format("2006-01-02 15:04:05")
	logic := ESLoggerInfo{
		Level:     level,
		TimeOut:   t,
		TimeKey:   st,
		TimeES:    mst,
		TimeUTC8:  utc8st,
		Type:      typeInfo,
		Info:      info,
		Last:      last,
		AccountID: accountID,
		GuildID:   guildID,
		Extra:     fmt.Sprintf(format, v...),
	}
	if guildID == "" {
		if accountID != "" {
			ac, err := db.ParseAccount(accountID)
			if err == nil {
				logic.Sid = ac.ShardId
				logic.Gid = ac.GameId
				logic.UserId = fmt.Sprintf("%s", ac.UserId)
			}
			logic.Avatar = fmt.Sprintf("%d", avatar)
			logic.CorpLvl = corpLvl
			logic.Channel = channel
		}
	}
	data, err := json.Marshal(logic)
	if err != nil {
		logs.Error("Logger Error for %s: %s %v", typeInfo, err.Error(), logic)
	}
	//TODO: 需要锁吗？
	logger("%s", data)
}
