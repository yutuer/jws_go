package merge

import (
	"regexp"

	"strings"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/redispool"
	"vcs.taiyouxi.net/platform/x/redis_storage/onland"
)

/*
	将需要的表现dump到最终db中
	并剔除需要删除角色表
*/
func Prepare() error {
	if err := prepareA(); err != nil {
		return err
	}
	if err := prepareB(); err != nil {
		return err
	}
	if err := prepareRankA(); err != nil {
		return err
	}
	logs.Info("total del account count %d", len(delAcids))
	return nil
}

func prepareA() error {
	ADB := GetADBConn()
	defer ADB.Close()
	ResDB := GetResDBConn()
	defer ResDB.Close()

	return _prepare(ADB, ResDB, prepare_A_table)
}

func prepareB() error {
	BDB := GetBDBConn()
	defer BDB.Close()
	ResDB := GetResDBConn()
	defer ResDB.Close()

	return _prepare(BDB, ResDB, prepare_B_table)
}

func prepareRankA() error {
	ARankDB := GetARankDBConn()
	defer ARankDB.Close()
	ResRankDB := GetResRankDBConn()
	defer ResRankDB.Close()

	return _prepare(ARankDB, ResRankDB, prepare_table_rank)
}

func _prepare(DB, ResDB redispool.RedisPoolConn, tableRegx string) error {
	// move table
	regx, err := regexp.Compile(tableRegx)
	if err != nil {
		logs.Error("Prepare regexp err %s", err.Error())
		return err
	}

	hander := func(keys []string) error {
		for _, k := range keys {
			if regx.Find([]byte(k)) != nil && !strings.Contains(k, prepare_except) {
				err, needDel := checkRoleNeedDel(DB, k)
				if err != nil {
					logs.Error("Prepare checkRoleNeedDel %s err %s", k, err.Error())
					continue
				}
				if needDel {
					logs.Debug("del account %s", k)
					continue
				}
				s, err := redis.String(DB.Do("DUMP", k))
				if err != nil {
					logs.Error("Prepare dump %s err %s", k, err.Error())
					continue
				}
				if _, err := ResDB.Do("RESTORE", k, 0, s); err != nil {
					logs.Error("Prepare RESTORE %s err %s", k, err.Error())
					continue
				}
			}
		}
		return nil
	}

	scanner := onland.NewScanner(DB.RawConn(), hander)
	if err := scanner.Start(); err != nil {
		logs.Error("Prepare scanner start err %v", err)
		return err
	}
	for !scanner.IsScanOver() {
		err := scanner.Next()
		if err != nil {
			logs.Error("Prepare scanner next err %v", err)
			return err
		}
	}
	return nil
}
