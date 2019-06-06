package bdclog

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"bufio"

	"time"

	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	txlumberjack "vcs.taiyouxi.net/platform/planx/util/lumberjack.v2"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

const (
	gameId             = "134"
	logFileTemplet     = "userinfo.log"
	csvProfileFile     = "userinfo.csv"
	csvGuildFile       = "guildinfo.csv"
	logPayTemplet      = "charge.log"
	writer_buff_size   = 32768
	timeLocal          = "Asia/Shanghai"
	csv_profile_header = "accountId,name,渠道,手机号," +
		"注册时间,最后登陆日," +
		"累计登陆日数,VIP等级," +
		"主将," +
		"时装," +
		"战队等级,战队总战力,最强主将," +
		"最远关卡进度,最远精英关卡进度,最远地狱关卡进度," +
		"硬通数,软通数,精铁数,Boss代币数,PvP代币数,神将代币,装备代币,公会代币," +
		"神将宝石," +
		"神将," +
		"称号," +
		"武魂," +
		"战队装备," +
		"公会,功勋值," +
		"爬塔最高层," +
		"累计充值(元)," +
		"羁绊," +
		"升星符," +
		"新手引导, 引导Event\r\n"
	csv_guild_header = "工会ID,工会名称,工会等级,工会经验,创建时间," +
		"成员人数,工会总战力," +
		"当天活跃人数,当天活跃成员id," +
		"当天工会boss参与次数,当天工会boss参与人数,当天工会boss参与成员id," +
		"科技树," +
		"成员信息\r\n"
	timeLayout = "2006-01-02 15:04:05"

	commonDailyStartTime = "5:00" // TODO ！！！注意：由于不能获得gamedata，所以活动刷新时间只能写死；每次使用时注意确认是否变了！！！
)

func NewBiScanRedis(gid uint, shardId []string, ts string, gidInfo []game.GidInfo) *BiScanRedis {
	gidmap := make(map[int]string, len(gidInfo))
	for _, g := range gidInfo {
		gidmap[g.Gid] = g.Channel
	}
	var t_now int64
	if ts != "" {
		loc, _ := time.LoadLocation(timeLocal)
		t, err := time.ParseInLocation("2006-1-2", ts, loc)
		if err == nil {
			t_now = t.Unix()
			logs.Warn("use param time %s %d", ts, t_now)
		} else {
			logs.Error("use param time err %v", err)
		}

	}
	return &BiScanRedis{
		gid:       gid,
		shardId:   shardId,
		timestamp: t_now,
		gidInfo:   gidmap,
		openOnce:  &sync.Once{},
	}
}

type BiScanRedis struct {
	gid                uint
	shardId            []string
	timestamp          int64
	gidInfo            map[int]string
	logger             *txlumberjack.Logger
	writer             *bufio.Writer
	csv_profile_logger *txlumberjack.Logger
	csv_profile_writer *bufio.Writer
	csv_guild_logger   *txlumberjack.Logger
	csv_guild_writer   *bufio.Writer
	paylogger          *txlumberjack.Logger
	paywriter          *bufio.Writer
	loc                *time.Location

	openOnce     *sync.Once
	biScanCloner uint64
}

func (cs *BiScanRedis) Clone() (storehelper.IStore, error) {
	// logs.Error("BiScanRedis is not ok with Clone.")
	if atomic.LoadUint64(&cs.biScanCloner) == 0 {
		//如果Store还没有Open过，则必须Open一次
		if err := cs.Open(); err != nil {
			return nil, err
		}
	} else {
		atomic.AddUint64(&cs.biScanCloner, 1)
	}
	return cs, nil
}

func (cs *BiScanRedis) Put(key string, val []byte, rh storehelper.ReadHandler) error {
	s := strings.SplitN(key, ":", 2)
	if len(s) <= 0 {
		return fmt.Errorf("NewCustomStoreRedis key %s unvalid", key)
	}
	switch s[0] {
	case "profile":
		if err := cs.bi_profile(key, s, val, rh); err != nil {
			return err
		}
	case "guild":
		if err := cs.bi_guild(key, s, val, rh); err != nil {
			return err
		}
	default:
		return fmt.Errorf("NewCustomStoreRedis key %s unvalid", s[0])
	}

	return nil
}

func (cs *BiScanRedis) StoreKey(key string) string {
	return key
}

func (cs *BiScanRedis) Open() error {
	cs.openOnce.Do(func() {
		cs.loc, _ = time.LoadLocation(timeLocal)

		strChan := cs.gidInfo[int(cs.gid)]
		sid := fmt.Sprintf("%s%04s%06s", strChan, gameId, cs.shardId[0])
		logTemplet := gameId + "_" + sid + "_" + logFileTemplet
		cs.logger = &txlumberjack.Logger{
			FileTempletName: logTemplet,
			MaxSize:         10000, // 10g
			//XXX: 10g防止被日志切分的一个小技巧
		}
		cs.writer = bufio.NewWriterSize(cs.logger, writer_buff_size)

		payTemplet := gameId + "_" + sid + "_" + logPayTemplet
		cs.paylogger = &txlumberjack.Logger{
			FileTempletName: payTemplet,
			MaxSize:         10000, // 10g
			//XXX: 10g防止被日志切分的一个小技巧
		}
		if cs.timestamp > 0 {
			cs.paylogger.GetUTCSec = func() int64 {
				return time.Unix(cs.timestamp, 0).Unix()
			}
		}
		cs.paywriter = bufio.NewWriterSize(cs.paylogger, writer_buff_size)

		csvProfileTemplet := gameId + "_" + sid + "_" + csvProfileFile
		cs.csv_profile_logger = &txlumberjack.Logger{
			FileTempletName: csvProfileTemplet,
			MaxSize:         10000, // 10g
			Header:          csv_profile_header,
		}
		cs.csv_profile_writer = bufio.NewWriterSize(cs.csv_profile_logger, writer_buff_size)

		csvGuildTemplet := gameId + "_" + sid + "_" + csvGuildFile
		cs.csv_guild_logger = &txlumberjack.Logger{
			FileTempletName: csvGuildTemplet,
			MaxSize:         10000, // 10g
			Header:          csv_guild_header,
		}
		cs.csv_guild_writer = bufio.NewWriterSize(cs.csv_guild_logger, writer_buff_size)

		atomic.AddUint64(&cs.biScanCloner, 1)
	})
	return nil
}
func (cs *BiScanRedis) Get(key string) ([]byte, error) {
	return nil, fmt.Errorf("not support")
}
func (cs *BiScanRedis) Del(key string) error {
	return fmt.Errorf("not support")
}
func (cs *BiScanRedis) RedisKey(key_in_store string) (string, bool) {
	return "", false
}
func (cs *BiScanRedis) Close() error {
	if atomic.CompareAndSwapUint64(&cs.biScanCloner, 1, 0) {
		if err := cs.writer.Flush(); err != nil {
			return err
		}
		if err := cs.logger.Close(); err != nil {
			return err
		}
		if err := cs.csv_profile_writer.Flush(); err != nil {
			return err
		}
		if err := cs.csv_profile_logger.Close(); err != nil {
			return err
		}
		if err := cs.csv_guild_writer.Flush(); err != nil {
			return err
		}
		if err := cs.csv_guild_logger.Close(); err != nil {
			return err
		}
		if err := cs.paywriter.Flush(); err != nil {
			return err
		}
		if err := cs.paylogger.Close(); err != nil {
			return err
		}
	} else {
		atomic.AddUint64(&cs.biScanCloner, ^uint64(0))
	}
	return nil
}
