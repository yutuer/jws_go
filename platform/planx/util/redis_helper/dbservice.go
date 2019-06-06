package redis_helper

import (
	"fmt"
	"time"

	gm "github.com/rcrowley/go-metrics"
	"github.com/timesking/seelog"
	"vcs.taiyouxi.net/platform/planx/metrics"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

type DBService struct {
	dbLog            seelog.LoggerInterface
	metricNumberSave gm.Counter
}

var dbss map[uint]*DBService
var dbsStarted bool

func init() {
	dbss = make(map[uint]*DBService, 4)
}

//TODO 时间可能需要调整稍微大一点
var defaultConfig1 = `
	<seelog type="adaptive" mininterval="1000" maxinterval="1000000" critmsgcount="100">
		<outputs formatid="common">
			<rollingfile type="date" filename="/opt/supervisor/log/dbfail_%d" datepattern="02.01.2006.log" maxrolls="60" />
		</outputs>`
var defaultConfig2 = `
		<formats>
			<format id="common"  format="%Msg%n"/>
		</formats>
	</seelog>
	`
const account_dbsaves = "accountdbsaves"

func InitDBService(ShardId uint) {
	if _, ok := dbss[ShardId]; !ok {
		logCfg := fmt.Sprintf(defaultConfig1, ShardId) + defaultConfig2
		logger, _ := seelog.LoggerFromConfigAsBytes([]byte(logCfg))

		dbss[ShardId] = &DBService{
			metricNumberSave: metrics.NewCounter(account_dbsaves),
			dbLog:            logger,
		}
	} else {
		panic("InitDBService should only be called once.")
	}
}

func GetDBService(ShardId uint) *DBService {
	dbs, ok := dbss[game.Cfg.GetShardIdByMerge(ShardId)]
	if !ok {
		panic("GetDBService should be after InitDBService.")
	}
	return dbs
}

func (s *DBService) LogDBError(accountId string, err error, buffer []byte) {
	if s.dbLog != nil {
		t := time.Now().UnixNano()
		s.dbLog.Errorf("[%d], Account:[%s], Error:[%s], buffer:[%q]", t, accountId, err, buffer)
	}
}

func (s *DBService) MetricsCountDBSaves() {
	s.metricNumberSave.Inc(1)
}
