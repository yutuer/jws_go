package merge

import (
	"fmt"

	"strconv"

	"time"

	"bufio"
	"math"
	"os"
	"path/filepath"

	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
)

type MergeConfig struct {
	EtcdEndPoint []string `toml:"etcd_endpoint"`
	EtcdRoot     string   `toml:"etcd_root"`
	OutPutPath   string   `toml:"output_path"`

	Gid int `toml:"Gid"`

	ARedis           string `toml:"ARedis"`
	ARedisDB         int    `toml:"ARedisDB"`
	ARedisDBAuth     string `toml:"ARedisDBAuth"`
	ARedisRank       string `toml:"ARedisRank"`
	ARedisRankDB     int    `toml:"ARedisRankDB"`
	ARedisRankDBAuth string `toml:"ARedisRankDBAuth"`
	AShardId         uint   `toml:"AShardId"`

	BRedis           string `toml:"BRedis"`
	BRedisDB         int    `toml:"BRedisDB"`
	BRedisDBAuth     string `toml:"BRedisDBAuth"`
	BRedisRank       string `toml:"BRedisRank"`
	BRedisRankDB     int    `toml:"BRedisRankDB"`
	BRedisRankDBAuth string `toml:"BRedisRankDBAuth"`
	BShardId         uint   `toml:"BShardId"`

	ResRedis           string `toml:"ResRedis"`
	ResRedisDB         int    `toml:"ResRedisDB"`
	ResRedisDBAuth     string `toml:"ResRedisDBAuth"`
	ResRedisRank       string `toml:"ResRedisRank"`
	ResRedisRankDB     int    `toml:"ResRedisRankDB"`
	ResRedisRankDBAuth string `toml:"ResRedisRankDBAuth"`
	ResShardId         uint   `toml:"ResShardId"`

	DelAccountCorpLvl     int `toml:"DelAccountCorpLvl"`
	DelAccountNotLoginMin int `toml:"DelAccountNotLoginMin"`

	Shard2Order map[uint]int
}

func (c *MergeConfig) Init() bool {
	// 初始化gamex cfg
	game.Cfg.Gid = c.Gid

	// 获取服务器序号
	c.Shard2Order = make(map[uint]int, 3)
	AKey := fmt.Sprintf("%s/%d/%d/%s", c.EtcdRoot, c.Gid, c.AShardId, etcd.KeyOrder)
	sAOrder, err := etcd.Get(AKey)
	if err != nil {
		logs.Error("MergeConfig Init get etcd err %s", err.Error())
		return false
	}
	AOrder, err := strconv.Atoi(sAOrder)
	if err != nil {
		logs.Error("MergeConfig Init Atoi err %s", err.Error())
		return false
	}
	c.Shard2Order[c.AShardId] = AOrder

	BKey := fmt.Sprintf("%s/%d/%d/%s", c.EtcdRoot, c.Gid, c.BShardId, etcd.KeyOrder)
	sBOrder, err := etcd.Get(BKey)
	if err != nil {
		logs.Error("MergeConfig Init get etcd err %s", err.Error())
		return false
	}
	BOrder, err := strconv.Atoi(sBOrder)
	if err != nil {
		logs.Error("MergeConfig Init Atoi err %s", err.Error())
		return false
	}
	c.Shard2Order[c.BShardId] = BOrder

	ResKey := fmt.Sprintf("%s/%d/%d/%s", c.EtcdRoot, c.Gid, c.ResShardId, etcd.KeyOrder)
	sResOrder, err := etcd.Get(ResKey)
	if err != nil {
		logs.Error("MergeConfig Init get etcd err %s", err.Error())
		return false
	}
	ResOrder, err := strconv.Atoi(sResOrder)
	if err != nil {
		logs.Error("MergeConfig Init Atoi err %s", err.Error())
		return false
	}
	c.Shard2Order[c.ResShardId] = ResOrder
	return true
}

func (c *MergeConfig) checkDelAccountNotLoginTime(ts int64) bool {
	now_t := time.Now().Unix()
	return now_t-ts > int64(c.DelAccountNotLoginMin*60)
}

var (
	Cfg         MergeConfig
	aPool       redispool.IPool
	aRankPool   redispool.IPool
	bPool       redispool.IPool
	bRankPool   redispool.IPool
	resPool     redispool.IPool
	resRankPool redispool.IPool
)

func RedisInit() {
	aPool = redispool.NewSimpleRedisPool("gamex.redis.mergeA",
		Cfg.ARedis, Cfg.ARedisDB, Cfg.ARedisDBAuth, false, 10, true)

	aRankPool = redispool.NewSimpleRedisPool("gamex.redis.mergeARank",
		Cfg.ARedisRank, Cfg.ARedisRankDB, Cfg.ARedisRankDBAuth, false, 10, true)

	bPool = redispool.NewSimpleRedisPool("gamex.redis.mergeB",
		Cfg.BRedis, Cfg.BRedisDB, Cfg.BRedisDBAuth, false, 10, true)

	bRankPool = redispool.NewSimpleRedisPool("gamex.redis.mergeBRank",
		Cfg.BRedisRank, Cfg.BRedisRankDB, Cfg.BRedisRankDBAuth, false, 10, true)

	resPool = redispool.NewSimpleRedisPool("gamex.redis.mergeRes",
		Cfg.ResRedis, Cfg.ResRedisDB, Cfg.ResRedisDBAuth, false, 10, true)

	resRankPool = redispool.NewSimpleRedisPool("gamex.redis.mergeResRank",
		Cfg.ResRedisRank, Cfg.ResRedisRankDB, Cfg.ResRedisRankDBAuth, false, 10, true)
}

func GetADBConn() redispool.RedisPoolConn {
	return aPool.GetDBConn()
}

func GetARankDBConn() redispool.RedisPoolConn {
	return aRankPool.GetDBConn()
}

func GetBDBConn() redispool.RedisPoolConn {
	return bPool.GetDBConn()
}

func GetBRankDBConn() redispool.RedisPoolConn {
	return bRankPool.GetDBConn()
}

func GetResDBConn() redispool.RedisPoolConn {
	return resPool.GetDBConn()
}

func GetResRankDBConn() redispool.RedisPoolConn {
	return resRankPool.GetDBConn()
}

func OutPutRes() error {
	if len(delAcids) <= 0 && len(guild_del) <= 0 {
		return nil
	}
	fn := filepath.Join(Cfg.OutPutPath,
		fmt.Sprintf("gamex_merge_%d_%d.csv", Cfg.AShardId, Cfg.BShardId))
	file, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		logs.Error("outPut OpenFile err %s", err.Error())
		return err
	}
	defer file.Close()

	bufWriter := bufio.NewWriterSize(file, 10240)
	bufWriter.WriteString("accountid,guild\r\n")
	l := int(math.Max(float64(len(delAcids)), float64(len(guild_del))))
	buf := make([]string, l)

	sDelAcids := make([]string, 0, len(delAcids))
	sDelGuilds := make([]string, 0, len(guild_del))
	for k, _ := range delAcids {
		sDelAcids = append(sDelAcids, k)
	}
	for k, _ := range guild_del {
		sDelGuilds = append(sDelGuilds, k)
	}
	for i := 0; i < l; i++ {
		if i < len(sDelAcids) && i >= len(sDelGuilds) {
			buf[i] = fmt.Sprintf("%s\r\n", sDelAcids[i])
		} else if i < len(sDelGuilds) && i >= len(sDelAcids) {
			buf[i] = fmt.Sprintf("%s\r\n", sDelGuilds[i])
		} else if i < len(sDelGuilds) && i < len(sDelAcids) {
			buf[i] = fmt.Sprintf("%s,%s\r\n", sDelAcids[i], sDelGuilds[i])
		}
	}
	for _, s := range buf {
		bufWriter.WriteString(s)
	}
	bufWriter.Flush()
	return nil
}
