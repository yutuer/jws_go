package imp

import (
	"fmt"

	"github.com/siddontang/go/timingwheel"

	"time"

	"bufio"
	"os"

	"path/filepath"

	"os/signal"
	"syscall"

	elastic "gopkg.in/olivere/elastic.v5"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	elasticClient            *elastic.Client
	begin_st                 int64
	data_time_begin          string
	data_time_begin_graphite string
	data_time_end            string
	next_data_time           int64
	log_file_name            string
	channel2Sid2Data         map[string]map[string]*data
)

type data struct {
	registerCount   int
	deviceCount     int
	activeCount     int
	acu             int
	pcu             int
	chargeSum       int
	chargeAcidCount int
}

func (d data) String() string {
	return fmt.Sprintf("reg %d dev %d act %d pcu %d acu %d cgs %d cgacid %d",
		d.registerCount, d.deviceCount, d.activeCount, d.pcu, d.acu, d.chargeSum, d.chargeAcidCount)
}

func StartImp() error {
	//client, err := elastic.NewClient(elastic.SetURL("http://es-search.es50-prod.yingxiong.net:9200"))
	client, err := elastic.NewClient(elastic.SetURL(Cfg.ElasticUrl))
	if err != nil {
		logs.Error("elastic.NewClient err %s", err.Error())
		return err
	}
	elasticClient = client
	updateTime("")

	stopch := make(chan os.Signal, 1)
	signal.Notify(stopch, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	debugch := make(chan os.Signal, 1)
	signal.Notify(debugch, syscall.SIGHUP)

	timing := timingwheel.NewTimingWheel(time.Second, 10)
	for {
		select {
		case <-timing.After(time.Second):
			if time.Now().Unix() >= next_data_time {
				_imp()
				// 下次时间
				updateTime("")
			}
		case <-debugch:
			logs.Info("debug imp...")

			next_data_time = 0
			updateTime(TimeParam)

			_imp()

			next_data_time = 0
			updateTime("")
		case <-stopch:
			logs.Info("stoped!!")
			return nil
		}
	}
	return nil
}

func _imp() {
	channel2Sid2Data = make(map[string]map[string]*data, 10)
	// 收集数据
	collectChargeAccountCount()
	collectChargeSum()
	collectLoginAndDevice()
	collectActivePlayer()
	collectAPCU()
	writeLog()
}

func updateTime(t string) {
	var h_utc int
	var nt_tl time.Time
	if next_data_time <= 0 { // 首次
		_t := time.Now()
		if t != "" {
			ttmp, err := time.ParseInLocation("2006-01-02 15", t, TimeLocal)
			if err != nil {
				logs.Error("param time err %v %v", t, err)
			} else {
				_t = ttmp
				logs.Debug("param time %v", _t)
			}
		}
		h_utc = _t.UTC().Hour()
		t := _t.In(TimeLocal)

		nt_tl = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 59, 59, 0, TimeLocal)
		next_data_time = nt_tl.Unix()
		begin_st = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, TimeLocal).Unix()
	} else {
		next_data_time += util.HourSec
		begin_st += util.HourSec
		_t := time.Unix(next_data_time, 0)
		nt_tl = _t.In(TimeLocal)
		h_utc = _t.UTC().Hour()
	}

	data_time_end = fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d.999",
		nt_tl.Year(), nt_tl.Month(), nt_tl.Day(), nt_tl.Hour(), nt_tl.Minute(), nt_tl.Second())
	data_time_begin = fmt.Sprintf("%d-%02d-%02dT00:00:00.000",
		nt_tl.Year(), nt_tl.Month(), nt_tl.Day())
	data_time_begin_graphite = fmt.Sprintf("%02d:00_%d%02d%02d",
		h_utc, nt_tl.Year(), nt_tl.Month(), nt_tl.Day())

	nt_next := nt_tl.Add(time.Hour).In(TimeLocal)
	log_file_name = fmt.Sprintf("%s%d-%02d-%02d_%02d.log", Cfg.LogPrefix,
		nt_next.Year(), nt_next.Month(), nt_next.Day(), nt_next.Hour())

	logs.Info("nexttime %d %d %s %s %s %s", begin_st, next_data_time,
		data_time_begin, data_time_begin_graphite, data_time_end, log_file_name)
}

func writeLog() error {
	file, err := os.OpenFile(filepath.Join(Cfg.LogOutput, log_file_name), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		logs.Error("writeLog OpenFile err %s", err.Error())
		return err
	}

	bufWriter := bufio.NewWriterSize(file, 1024)

	for channel, sidData := range channel2Sid2Data {
		for sid, data := range sidData {
			gid := _sid2gid(sid)
			if gid < 0 {
				logs.Error("writeLog _sid2gid gid is nil, sid=%s", sid)
				continue
			}
			strChan := game.Gid2Channel[gid]
			gidSid := fmt.Sprintf("%s%04s%06s", strChan, gameId, sid)
			line := fmt.Sprintf("%s$$%s$$%d$$%d$$%d$$%d$$%d$$%d$$%d$$%d$$%d\n",
				gidSid, channel, data.chargeSum, data.chargeAcidCount,
				data.activeCount, data.activeCount-data.registerCount,
				data.registerCount, data.registerCount, data.deviceCount,
				data.pcu, data.acu)
			bufWriter.Write([]byte(line))
		}
	}
	bufWriter.Flush()
	file.Close()
	return nil
}

type graphite_res struct {
	Target     string          `json:"target"`
	Datapoints [][]interface{} `json:"datapoints"`
}
