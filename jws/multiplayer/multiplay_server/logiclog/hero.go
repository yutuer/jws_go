package logiclog

import (
	"encoding/json"
	"regexp"

	"github.com/timesking/seelog"

	"time"

	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/eslogger"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	txlumberjack "vcs.taiyouxi.net/platform/planx/util/lumberjack.v2"
)

const (
	writer_buff_size = 32768
)

type heroLog struct {
	logger *txlumberjack.Logger
}

func init() {
	seelog.RegisterReceiver("herologiclog_multiplayer", &heroLog{})
}

func (heroLog *heroLog) AfterParse(initArgs seelog.CustomReceiverInitArgs) error {
	filename := "hero_multiplayer.log"
	if initArgs.XmlCustomAttrs != nil && len(initArgs.XmlCustomAttrs) > 0 {
		filename = initArgs.XmlCustomAttrs["filename"]
	}

	heroLog.logger = &txlumberjack.Logger{
		FileTempletName: filename,
		MaxSize:         10000, // 10g
		BufSize:         32768, // 30k
		TimeLocal:       "Asia/Shanghai",
		GetUTCSec:       func() int64 { return time.Now().Unix() },
	}

	return nil
}

func (heroLog *heroLog) ReceiveMessage(message string, level seelog.LogLevel, context seelog.LogContextInterface) error {
	if level != seelog.InfoLvl {
		return nil
	}
	var log eslogger.ESLoggerInfo
	err := json.Unmarshal([]byte(message), &log)
	if err != nil {
		logs.Error("reading standard input json.Unmarshal err %v line %s", err, message)
		return err
	}
	ok, _ := regexp.Match(BITag+".*", []byte(log.Extra))
	if !ok {
		return nil
	}
	var resLine string
	switch log.Type {
	case LogicTag_GameOver:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_GameOver{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_GameOver)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err %s %v", log.Type, _log.Info)
			return err
		}
		resLine = fmt.Sprintf("%s$$PVETEAMBOSS_GAMEOVER$$%s$$%d$$%d$$%v$$%v$$%d", log.TimeUTC8,
			info.GameId, info.Duration, info.IsWin, info.IsHard, info.BossId, info.PlayerLvlAvg)
	}
	if len(resLine) > 0 {
		if _, err := heroLog.logger.Write([]byte(resLine + "\n")); err != nil {
			logs.Error("hero ReceiveMessage write %v", err)
			return err
		}
	}
	return nil
}

func (heroLog *heroLog) Flush() {

}

func (heroLog *heroLog) Close() error {
	return heroLog.logger.Close()
}
