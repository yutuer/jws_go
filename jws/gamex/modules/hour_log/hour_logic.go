package hour_log

import (
	"fmt"
	"strings"
	"time"

	"math"
	"strconv"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/modules"
	metricsModules "vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (hl *HourLog) onLogin(cmd *hourCmd) {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("HourLog onLogin GetDBConn Err nil")
		return
	}
	cb := redis.NewCmdBuffer()

	// device
	nm := TableBIHourDevice(hl.sid, cmd.Channel)
	if err := cb.Send("PFADD", nm, cmd.Device); err != nil {
		logs.Error("HourLog onLogin PFADD %s %s", nm, err.Error())
	}

	// 记录渠道
	nm = TableBIHourChannels(hl.sid)
	if err := cb.Send("SADD", nm, cmd.Channel); err != nil {
		logs.Error("HourLog onLogin SADD %s %s", nm, err.Error())
	}

	if _, err := metricsModules.DoCmdBufferWrapper(db_counter_key, conn, cb, true); err != nil {
		logs.Error("HourLog onLogin DoCmdBuffer error %s", err.Error())
		return
	}
}

func (hl *HourLog) onRegister(cmd *hourCmd) {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("HourLog onLogin GetDBConn Err nil")
		return
	}
	cb := redis.NewCmdBuffer()

	// registerCount
	nm := TableBIHourRegister(hl.sid)
	if err := cb.Send("HINCRBY", nm, cmd.Channel, 1); err != nil {
		logs.Error("HourLog onLogin HINCRBY %s %s", nm, err.Error())
	}

	// 记录渠道
	nm = TableBIHourChannels(hl.sid)
	if err := cb.Send("SADD", nm, cmd.Channel); err != nil {
		logs.Error("HourLog onLogin SADD %s %s", nm, err.Error())
	}

	if _, err := metricsModules.DoCmdBufferWrapper(db_counter_key, conn, cb, true); err != nil {
		logs.Error("HourLog onLogin DoCmdBuffer error %s", err.Error())
		return
	}
}

func (hl *HourLog) onPay(cmd *hourCmd) {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("HourLog onPay GetDBConn Err nil")
		return
	}
	cb := redis.NewCmdBuffer()

	// charge acid
	nm := TableBIHourChargeAcid(hl.sid, cmd.Channel)
	if err := cb.Send("PFADD", nm, cmd.Acid); err != nil {
		logs.Error("HourLog onPay PFADD %s %s", nm, err.Error())
	}

	// charge sum
	nm = TableBIHourChargeSum(hl.sid)
	if err := cb.Send("HINCRBY", nm, cmd.Channel, cmd.Money*10000); err != nil {
		logs.Error("HourLog onPay HINCRBY %s %s", nm, err.Error())
	}

	// 记录渠道
	nm = TableBIHourChannels(hl.sid)
	if err := cb.Send("SADD", nm, cmd.Channel); err != nil {
		logs.Error("HourLog onPay SADD %s %s", nm, err.Error())
	}

	if _, err := metricsModules.DoCmdBufferWrapper(db_counter_key, conn, cb, true); err != nil {
		logs.Error("HourLog onPay DoCmdBuffer error %s", err.Error())
		return
	}
}

func (hl *HourLog) onCCU() {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("HourLog onPay GetDBConn Err nil")
		return
	}
	now_t := time.Now().Unix()
	chccu := GetCCU()
	nm := TableBIHourCCU(hl.sid)
	for ch, ccu := range chccu {
		v := fmt.Sprintf("%s:%d:%d", ch, ccu, now_t)
		if _, err := conn.Do("ZADD", nm, now_t, v); err != nil {
			logs.Error("HourLog onCCU HINCRBY %s %s", nm, err.Error())
		}
	}
}

func (hl *HourLog) onActive(cmd *hourCmd) {
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("HourLog onActive GetDBConn Err nil")
		return
	}

	// charge acid
	nm := TableBIHourActive(hl.sid, cmd.Channel)
	if _, err := conn.Do("PFADD", nm, cmd.Acid); err != nil {
		logs.Error("HourLog onActive PFADD %s %s", nm, err.Error())
	}
}

func (hl *HourLog) logCheck(t time.Time, forceOutput bool) {
	nt_tl := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 59, 59, 0, logiclog.GetBILogTL())
	hour_end_t := nt_tl.Unix()
	if hl.LastLogHourTime <= 0 {
		hl.LastLogHourTime = hour_end_t
	} else if hour_end_t > hl.LastLogHourTime || forceOutput {
		conn := modules.GetDBConn()
		defer conn.Close()
		if conn.IsNil() {
			logs.Error("HourLog onLogin GetDBConn Err nil")
			return
		}

		ch_log_data := make(map[string]*logData, 10)
		nm := TableBIHourChannels(hl.sid)
		chs, err := redis.Strings(conn.Do("SMEMBERS", nm))
		if err != nil {
			logs.Warn("HourLog logCheck SMEMBERS %s %s", nm, err.Error())
		}
		for _, ch := range chs {
			nm := TableBIHourRegister(hl.sid)
			registerC, err := redis.Int(conn.Do("HGET", nm, ch))
			if err != nil {
				logs.Warn("HourLog logCheck HGET %s %s", nm, err.Error())
			}
			nm = TableBIHourDevice(hl.sid, ch)
			deviceC, err := redis.Int(conn.Do("PFCOUNT", nm))
			if err != nil {
				logs.Warn("HourLog logCheck PFCOUNT %s %s", nm, err.Error())
			}
			nm = TableBIHourActive(hl.sid, ch)
			activeC, err := redis.Int(conn.Do("PFCOUNT", nm))
			if err != nil {
				logs.Warn("HourLog logCheck PFCOUNT %s %s", nm, err.Error())
			}
			nm = TableBIHourChargeAcid(hl.sid, ch)
			chargeAcidC, err := redis.Int(conn.Do("PFCOUNT", nm))
			if err != nil {
				logs.Warn("HourLog logCheck PFCOUNT %s %s", nm, err.Error())
			}
			nm = TableBIHourChargeSum(hl.sid)
			chargeSumC, err := redis.Int(conn.Do("HGET", nm, ch))
			if err != nil {
				logs.Warn("HourLog logCheck HGET %s %s", nm, err.Error())
			}
			// 记录
			data, ok := ch_log_data[ch]
			if !ok {
				data = &logData{}
				ch_log_data[ch] = data
			}
			data = ch_log_data[ch]
			data.register = registerC
			data.device = deviceC
			data.active = activeC
			data.chargeSum = chargeSumC
			data.chargeAcid = chargeAcidC
		}

		// ccu pcu
		nm = TableBIHourCCU(hl.sid)
		logs.Debug("HourLog logCheck ccu log time %d %d",
			hl.LastLogHourTime-util.HourSec, hl.LastLogHourTime)
		ccu_ss, err := redis.Strings(conn.Do("ZRANGEBYSCORE", nm,
			hl.LastLogHourTime-util.HourSec, hl.LastLogHourTime))
		if err != nil {
			logs.Warn("HourLog logCheck ZRANGEBYSCORE %s %s", nm, err.Error())
		}

		ccumap := make(map[string]*ccuData, 10)
		for _, ccu_ch := range ccu_ss {
			_ss := strings.Split(ccu_ch, ":")
			ch := _ss[0]
			s_ccu := _ss[1]
			data, ok := ccumap[ch]
			if !ok {
				data = &ccuData{make([]int, 0, 60)}
				ccumap[ch] = data
			}
			data = ccumap[ch]
			ccu, err := strconv.Atoi(s_ccu)
			if err != nil {
				logs.Error("HourLog logCheck Atoi ccu %s", err.Error())
			}
			data.ccus = append(data.ccus, ccu)
		}

		for ch, _ccus := range ccumap {
			var acu, pcu, ccu_sum int
			for _, ccu := range _ccus.ccus {
				if ccu > pcu {
					pcu = ccu
				}
				ccu_sum += ccu
			}
			acu = int(math.Ceil(float64(ccu_sum) / float64(len(_ccus.ccus))))

			data, ok := ch_log_data[ch]
			if !ok {
				data = &logData{}
				ch_log_data[ch] = data
			}
			data = ch_log_data[ch]
			data.acu = acu
			data.pcu = pcu
		}

		if hour_end_t > hl.LastLogHourTime {
			hl.LastLogHourTime += util.HourSec
		}
		hl.writeLog(hl.LastLogHourTime, ch_log_data)
	}
}

func (hl *HourLog) clearCheck(t time.Time) {
	daybegin := logiclog.BIDailyBeginUnix(t.Unix())
	if daybegin > logiclog.BIDailyBeginUnix(hl.CurDayTimeBegin) {
		conn := modules.GetDBConn()
		defer conn.Close()
		if conn.IsNil() {
			logs.Error("HourLog clearCheck GetDBConn Err nil")
			return
		}

		nm := TableBIHourChannels(hl.sid)
		chs, err := redis.Strings(conn.Do("SMEMBERS", nm))
		if err != nil {
			logs.Warn("HourLog clearCheck SMEMBERS %s %s", nm, err.Error())
		}

		// 删除表
		cb := redis.NewCmdBuffer()

		nm = TableBIHourRegister(hl.sid)
		if err := cb.Send("DEL", nm); err != nil {
			logs.Warn("HourLog clearCheck DEL %s %s", nm, err.Error())
		}

		for _, ch := range chs {
			nm = TableBIHourDevice(hl.sid, ch)
			if err := cb.Send("DEL", nm); err != nil {
				logs.Warn("HourLog clearCheck DEL %s %s", nm, err.Error())
			}
			nm = TableBIHourChargeAcid(hl.sid, ch)
			if err := cb.Send("DEL", nm); err != nil {
				logs.Warn("HourLog clearCheck DEL %s %s", nm, err.Error())
			}
			nm = TableBIHourActive(hl.sid, ch)
			if err := cb.Send("DEL", nm); err != nil {
				logs.Warn("HourLog clearCheck DEL %s %s", nm, err.Error())
			}
		}
		nm = TableBIHourChargeSum(hl.sid)
		if err := cb.Send("DEL", nm); err != nil {
			logs.Warn("HourLog clearCheck DEL %s %s", nm, err.Error())
		}
		nm = TableBIHourChannels(hl.sid)
		if err := cb.Send("DEL", nm); err != nil {
			logs.Warn("HourLog clearCheck DEL %s %s", nm, err.Error())
		}
		nm = TableBIHourCCU(hl.sid)
		if err := cb.Send("DEL", nm); err != nil {
			logs.Warn("HourLog clearCheck DEL %s %s", nm, err.Error())
		}

		if _, err := metricsModules.DoCmdBufferWrapper(db_counter_key, conn, cb, true); err != nil {
			logs.Error("HourLog clearCheck DoCmdBuffer error %s", err.Error())
		}

		hl.CurDayTimeBegin = daybegin
	}
}

type logData struct {
	register   int
	device     int
	active     int
	chargeSum  int
	chargeAcid int
	acu        int
	pcu        int
}

type ccuData struct {
	ccus []int
}
