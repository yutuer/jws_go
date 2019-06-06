package logs

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"
	//"time"

	"vcs.taiyouxi.net/platform/planx/util/signalhandler"

	raven "github.com/getsentry/raven-go"
	"github.com/timesking/seelog"
	//"github.com/howeyc/fsnotify"
)

const defaultStackTraceDepth = 1

var _logger seelog.LoggerInterface
var _stacktraceDepth int
var lastCfg string
var _sentry *raven.Client

const PANICHEADER = "SENTRY_SENT:"

type SentryReceiver struct {
	//client *raven.Client
}

func (ar *SentryReceiver) ReceiveMessage(message string, level seelog.LogLevel, context seelog.LogContextInterface) error {
	if ar == nil {
		return nil
	}
	//if ar.client == nil {
	//return nil
	//}
	if level <= seelog.WarnLvl {
		return nil
	}
	if strings.Contains(message, PANICHEADER) {
		return nil
	}
	var msg string
	if context.IsValid() {
		msg = fmt.Sprintf("%s, func: %s, file: %s, line %d",
			message,
			context.Func(),
			context.FullPath(),
			context.Line())
	} else {
		msg = message
	}

	packet := raven.NewPacket(msg, &raven.Message{msg, nil})
	packet.Logger = "seelog"
	switch level {
	case seelog.ErrorLvl:
		packet.Level = raven.ERROR
	case seelog.CriticalLvl:
		packet.Level = raven.FATAL
	default:
		packet.Level = raven.ERROR
	}
	//packet.ServerName =
	//packet.Release =
	//ar.client.Capture(packet, nil)
	if _sentry != nil {
		_sentry.Capture(packet, nil)
	}

	return nil
}

func (ar *SentryReceiver) AfterParse(initArgs seelog.CustomReceiverInitArgs) error {
	return nil
}

func (ar *SentryReceiver) Flush() {

}

func (ar *SentryReceiver) Close() error {
	//if ar != nil && ar.client != nil {
	//ar.client.Close()
	//}
	return nil
}

func init() {
	seelog.RegisterReceiver("sentry", &SentryReceiver{
	//client: _sentry,
	})
	defaultConfig := `
	<seelog>
		<outputs>
			<filter levels="warn">
				<console formatid="coloredblue"/>
			</filter>
			<filter levels="error,critical">
				<console formatid="coloredred"/>
			</filter>
			<console formatid="common"/>
		</outputs>
		<formats>
			<format id="coloredblue"  format="[%Date %Time] %EscM(34)[%LEV] [%Func] %Msg%EscM(39)%n%EscM(0)"/>
			<format id="coloredred"  format="[%Date %Time] %EscM(31)[%LEV] [%Func] %Msg%EscM(39)%n%EscM(0)"/>
			<format id="common"  format="[%Date %Time] [%LEV] [%Func] %Msg%n"/>
		</formats>
	</seelog>
	`

	logger, err := seelog.LoggerFromConfigAsBytes([]byte(defaultConfig))
	if err != nil {
		panic(err)
		return
	}
	seelog.ReplaceLogger(logger)
	_logger = seelog.Current
	SetStackTraceDepth(defaultStackTraceDepth)

	signalhandler.SignalReloadFunc(func() {
		LoadLogConfig(lastCfg)
		fmt.Printf("Got A SIGUSR2 Signal! Now Reloading Conf....\n")
	})
}

type SentryCfg struct {
	SentryConfig Config
}
type Config struct {
	DSN string
}

func InitSentry(DSN string) error {
	InitSentryTags(DSN, nil)
	return nil
}

func InitSentryTags(DSN string, tags map[string]string) error {
	s, err := raven.NewClient(DSN, tags)
	if err != nil {
		Warn("Init Sentry failed with error %s", err.Error())
		_sentry = nil
		return err
	}
	_sentry = s
	return nil
}

func LoadLogConfig(cfg string) {
	logger, err := seelog.LoggerFromParamConfigAsFile(cfg, nil)
	if err != nil {
		fmt.Printf("load logger failed in location %s. \nlogger starts with default console log system\n", cfg)
		fmt.Println(err.Error())
		return
	}
	//go func(clogger seelog.LoggerInterface) {
	//defer func() {
	//recover()
	//}()
	//time.Sleep(time.Second * 2)
	//fmt.Println("***try to close it again", clogger)
	//clogger.Close()
	//}(_logger)

	if err := seelog.ReplaceLogger(logger); err != nil {
		fmt.Printf("replace logger with new logger failed. %s\n", cfg)
		return
	}
	lastCfg = cfg
	_logger = seelog.Current
	SetStackTraceDepth(_stacktraceDepth)
}

func SetStackTraceDepth(depth int) {
	_stacktraceDepth = depth
	if _logger != nil {
		_logger.SetAdditionalStackDepth(depth)
	}
}

func Close() {
	if _logger != nil {
		_logger.Close()
	}
}

func Flush() {
	if _logger != nil {
		_logger.Flush()
	}
}

func Trace(format string, v ...interface{}) {
	if _logger != nil {
		_logger.Tracef(format, v...)
	}
}

func Debug(format string, v ...interface{}) {
	if _logger != nil {
		_logger.Debugf(format, v...)
	}
}

func Info(format string, v ...interface{}) {
	if _logger != nil {
		_logger.Infof(format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	if _logger != nil {
		_logger.Warnf(format, v...)
	}
}

func Error(format string, v ...interface{}) {
	if _logger != nil {
		_logger.Errorf(format, v...)
	}
}

func Critical(format string, v ...interface{}) {
	if _logger != nil {
		_logger.Criticalf(format, v...)
	}
}

//SentryLogicCritical 所有逻辑上不应该出现的错误，配置错误，或者基于某个玩家存档的特殊异常
// 都应该通过这个函数输出。
func SentryLogicCritical(accountId string, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	sentryCatcherWithAll(accountId, "", msg)
}

// PanicCatcher 在defer中调用recover()，输出处理异常情况
func PanicCatcher() {
	sentryCatcherWithAll("", "", recover())
}

func PanicCatcherWithInfo(format string, v ...interface{}) {
	if err := recover(); err != nil {
		msg := fmt.Sprintf(format, v...)
		errmsg := fmt.Sprintf("%s, err:%v", msg, err)
		sentryCatcherWithAll("", "", errmsg)
	}
}

func PanicCatcherWithAccountID(accountId string) {
	sentryCatcherWithAll(accountId, "", recover())
}
func PanicCatcherWithAll(accountId string, ipAddr string) {
	sentryCatcherWithAll(accountId, ipAddr, recover())
}
func sentryCatcherWithAll(accountId string, ipAddr string, raisedErr interface{}) {
	if raisedErr != nil {
		if _sentry != nil {
			ri := make([]raven.Interface, 0, 2)
			var user raven.User
			var bUserInfo bool
			if accountId != "" {
				bUserInfo = true
				user.ID = accountId
			}
			if ipAddr != "" {
				bUserInfo = true
				user.Email = ipAddr
			}

			if bUserInfo {
				ri = append(ri, &user)
			}

			var packet *raven.Packet
			switch rval := raisedErr.(type) {
			case nil:
				return
			case error:
				trace := raven.NewStacktrace(3, 3, nil)
				ri = append(ri, raven.NewException(rval, trace))
				packet = raven.NewPacket(rval.Error(), ri...)
			default:
				rvalStr := fmt.Sprint(rval)
				trace := raven.NewStacktrace(3, 3, nil)
				ri = append(ri, raven.NewException(errors.New(rvalStr), trace))
				packet = raven.NewPacket(rvalStr, ri...)
			}
			packet.Logger = "seelog"
			packet.Level = raven.FATAL
			_sentry.Capture(packet, nil)

		}

		//如果当前有玩家账户ID则输出账户ID
		critical_header := PANICHEADER
		if accountId != "" {
			critical_header = fmt.Sprintf("%s account:%s", PANICHEADER, accountId)
		}

		switch rval := raisedErr.(type) {
		case nil:
			return
		case error:
			trace := make([]byte, 2048)
			count := runtime.Stack(trace, true)
			tracej, err := json.Marshal(string(trace[:count]))
			if err != nil {
				Error("%s error %s, stack string is: %s", critical_header, rval.Error(), trace[:count])
			} else {
				Error("%s error %s, stack json is: %s", critical_header, rval.Error(), tracej)
			}
		default:
			trace := make([]byte, 2048)
			count := runtime.Stack(trace, true)
			rvalStr := fmt.Sprint(rval)
			tracej, err := json.Marshal(string(trace[:count]))
			if err != nil {
				Error("%s msg %s, stack string is: %s", critical_header, rvalStr, trace[:count])
			} else {
				Error("%s msg %s, stack json is: %s", critical_header, rvalStr, tracej)
			}

		}

	}
}
