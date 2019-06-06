package main

import (
	"encoding/csv"
	"os"
	"strconv"

	"time"

	"math/rand"

	"sync"

	"fmt"

	"bytes"

	"github.com/codegangsta/cli"
	"github.com/siddontang/go/sync2"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/account/gs"
	"vcs.taiyouxi.net/jws/gamex/models/account/update/data_update"
	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func main() {
	app := cli.NewApp()

	app.Version = "0.0.0"
	app.Name = "refresh_gs_by_offline"
	app.Usage = "calculate gs from db and save back"
	app.Author = "libingbing"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "gid",
			Value: "",
			Usage: "大区",
		},
		cli.IntFlag{
			Name:  "c",
			Value: 1,
			Usage: "同一个DB并行执行的服",
		},
		cli.BoolFlag{
			Name:  "save",
			Usage: "",
		},
		cli.StringFlag{
			Name:  "config",
			Value: "server.csv",
			Usage: "指定配置表",
		},
		cli.StringFlag{
			Name:  "out",
			Value: "",
			Usage: "输出文件指定后缀",
		},
	}

	app.Action = func(context *cli.Context) {
		logxml := config.NewConfigPath("log.xml")
		logs.LoadLogConfig(logxml)

		var err error
		outSurfix := context.String("out")
		acidGsPath := fmt.Sprintf("acid_gs_%s.txt", outSurfix)
		acid2GsFile, err = os.OpenFile(acidGsPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			logs.Error("open file err, ", err)
			return
		}
		gamedata.LoadGameData("")
		serverConfigPath := context.String("config")
		serverCfgMaps := loadServerCfg(serverConfigPath)
		FailedServers = make([]string, 0)
		FailedPlayers = make([]string, 0)

		paraRunCountPerDb := context.Int("c")
		waitter := &util.WaitGroupWrapper{}
		gsImpl := &RefreshGsImpl{}
		needSave := context.Bool("save")
		for _, cfgs := range serverCfgMaps {
			parallelRefreshGs(waitter, cfgs, paraRunCountPerDb, gsImpl, needSave)
		}
		waitter.Wait()
		reportString := generateReport()
		reportPath := fmt.Sprintf("report_%s.txt", outSurfix)
		reportFile, err := os.OpenFile(reportPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			logs.Error("open report file err, ", err)
			logs.Error("report : %s", reportString)
		} else {
			reportFile.WriteString(reportString)
		}
		acid2GsFile.Close()
	}

	app.Run(os.Args)
}

var acid2GsFile *os.File

func generateReport() string {
	bufW := bytes.NewBuffer([]byte{})
	bufW.WriteString(fmt.Sprintf("success server count, %d", successServers.Get()))
	bufW.WriteString("\n")

	serverLock.RLock()
	if len(FailedServers) > 0 {
		bufW.WriteString(fmt.Sprintf("Failed Servers:\n"))
	}
	for _, str := range FailedServers {
		bufW.WriteString(str + "\n")
	}
	serverLock.RUnlock()
	bufW.WriteString("\n")

	playerLock.RLock()
	if len(FailedPlayers) > 0 {
		bufW.WriteString(fmt.Sprintf("Failed players: \n"))
		for _, str := range FailedPlayers {
			bufW.WriteString(str + "\n")
		}
	}
	playerLock.RUnlock()

	if len(FailedPlayers) == 0 && len(FailedServers) == 0 {
		bufW.WriteString("perfect")
	}
	return bufW.String()
}

var serverLock sync.RWMutex
var FailedServers []string

var playerLock sync.RWMutex
var FailedPlayers []string

var successServers sync2.AtomicInt32

func addFailedServer(gid, serverId string) {
	serverLock.Lock()
	defer serverLock.Unlock()
	FailedServers = append(FailedServers, fmt.Sprintf("%s:%s", gid, serverId))
}

func addFailedPlayer(acid string) {
	playerLock.Lock()
	defer playerLock.Unlock()
	FailedPlayers = append(FailedPlayers, acid)
}

func increaseSuccessServer() {
	successServers.Add(1)
}

type RefreshGsInterface interface {
	RefreshGs(redisCfg ServerRedisCfg, needSave bool)
}

type RefreshGsImpl struct {
}

func (r *RefreshGsImpl) RefreshGs(redisCfg ServerRedisCfg, needSave bool) {
	refreshGs(redisCfg, needSave)
}

func refreshGs(redisCfg ServerRedisCfg, needSave bool) {
	rankInfo, rankConn, rankTable := loadRankInfo(redisCfg)
	if rankInfo == nil || rankConn == nil {
		addFailedServer(redisCfg.Gid, redisCfg.Sid)
		return
	}

	logs.Info("server %s, get rank count %d", redisCfg.Sid, len(rankInfo))

	driver.SetupRedisByCap(redisCfg.Url, redisCfg.Db, redisCfg.Auth, 1)

	for i := range rankInfo {
		acid := rankInfo[i]
		acc, err := db.ParseAccount(acid)
		if err != nil {
			logs.Error("parse account", err)
			addFailedPlayer(acid)
			continue
		}
		p, err := account.LoadFullAccount(acc, false)
		if err != nil {
			logs.Error("load account", err)
			addFailedPlayer(acid)
			continue
		}
		err = data_update.Update(p.Profile.Ver, true, p)
		if err != nil {
			logs.Error("update account", err)
			addFailedPlayer(acid)
			continue
		}
		p.Profile.OnAfterLogin() // 一些初始化的工作
		_, _, _, _, _, gs, topHero := gs.GetCurrAttr(
			account.NewAccountGsCalculateAdapter(p))

		score := int64(gs*rank.RankByCorpDelayPowBase + int(rand.Int31n(10000)))
		logs.Info("acid->gs, %s, %d", acid, score)
		top3GsStr := ""
		for i, gs := range topHero {
			if i >= 3 {
				break
			}
			top3GsStr = fmt.Sprintf("%s, %d, %d", top3GsStr, gs.HeroId, gs.Gs)
		}
		acid2GsFile.WriteString(fmt.Sprintf("%s, %d, %s\n", acid, score, top3GsStr))
		if needSave {
			rankConn.Do("ZADD", rankTable, score, acid)
		}

	}

	increaseSuccessServer()
}

func loadRankInfo(redisCfg ServerRedisCfg) ([]string, redis.Conn, string) {
	// 连接rankDb
	rankCon, err := redis.Dial("tcp", redisCfg.RankUrl,
		redis.DialConnectTimeout(1*time.Second),
		redis.DialReadTimeout(5*time.Second),
		redis.DialWriteTimeout(5*time.Second),
		redis.DialPassword(redisCfg.RankAuth),
		redis.DialDatabase(redisCfg.RankDb))
	if err != nil {
		logs.Error("connect rank redis err ", err)
		return nil, nil, ""
	}
	rankTable := fmt.Sprintf("%s:%s:RankCorpGs", redisCfg.Gid, redisCfg.Sid)
	logs.Info("rankTable %s", rankTable)
	rankInfo, err := redis.Strings(rankCon.Do("ZRANGE", rankTable, 0, -1, "WITHSCORES"))
	if err != nil {
		logs.Error("zrange err", err)
		return nil, nil, ""
	}
	acids := make([]string, 0)
	for i := 0; i < len(rankInfo); i += 2 {
		acids = append(acids, rankInfo[i])
	}
	return acids, rankCon, rankTable
}

type ServerRedisCfg struct {
	Gid      string
	Sid      string
	Url      string
	Db       int
	Auth     string
	RankUrl  string
	RankDb   int
	RankAuth string
}

const (
	CSV_GID = iota
	CSV_SID
	CSV_URL
	CSV_DB
	CSV_AUTH
	CSV_RANK_URL
	CSV_RANK_DB
	CSV_RANK_AUTH
)

// 返回 map<db, serverCfg>
func loadServerCfg(configPath string) map[string][]ServerRedisCfg {
	serverFile, err := os.Open(configPath)
	if err != nil {
		panic(err)
	}
	csvReader := csv.NewReader(serverFile)
	csvAll, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}
	//fmt.Println(csvAll)
	cfgByDb := make(map[string][]ServerRedisCfg)
	for _, str := range csvAll[1:] {
		db, err := strconv.Atoi(str[CSV_DB])
		if err != nil {
			panic(err)
		}
		rankDb, err := strconv.Atoi(str[CSV_RANK_DB])
		cfg := ServerRedisCfg{
			Gid:      str[CSV_GID],
			Sid:      str[CSV_SID],
			Url:      str[CSV_URL],
			Db:       db,
			Auth:     str[CSV_AUTH],
			RankUrl:  str[CSV_RANK_URL],
			RankDb:   rankDb,
			RankAuth: str[CSV_RANK_AUTH],
		}
		dbKey := fmt.Sprintf("%s:%s", cfg.Url, cfg.Db)
		if cfgList, ok := cfgByDb[dbKey]; ok {
			cfgByDb[dbKey] = append(cfgList, cfg)
		} else {
			cfgByDb[dbKey] = append([]ServerRedisCfg{}, cfg)
		}
	}
	return cfgByDb
}

func parallelRefreshGs(waitter *util.WaitGroupWrapper, cfgs []ServerRedisCfg, paraCount int, impl RefreshGsInterface, needSave bool) {
	for i := 0; i < len(cfgs); i += paraCount {
		ii := i
		waitter.Wrap(func() {
			for j := ii; j < len(cfgs) && j < ii+paraCount; j++ {
				impl.RefreshGs(cfgs[j], needSave)
			}
		})
	}
}
