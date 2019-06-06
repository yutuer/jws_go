package storehelper

import (
	"errors"
	"fmt"
	"hash/fnv"

	"github.com/jackc/pgx"
	"sync"
	"vcs.taiyouxi.net/platform/planx/util/account_json"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const MAX_SUB_TABLE = 64 // 分表个数 不要改, 一旦修改会导致之前存的key找不到

var key2sql = []string{
	"profile", "bag", "store", "pguild", "tmp", "simpleinfo", "general", "anticheat",
}

var PutSQLs map[string][]string
var GetSQLs map[string][]string
var DelSQLs map[string][]string
var ListSQLHeads map[string][]string

func mkSQLs(pool *pgx.ConnPool, nameHead string, sqlStr string) []string {
	sqls := [MAX_SUB_TABLE]string{}
	for i := 0; i < MAX_SUB_TABLE; i++ {
		sqls[i] = fmt.Sprintf(sqlStr, nameHead, i)
		pool.Prepare("", sqls[i])
	}
	return sqls[:]
}

func mkAllSQLs(pool *pgx.ConnPool) {
	PutSQLs = make(map[string][]string, len(key2sql))
	GetSQLs = make(map[string][]string, len(key2sql))
	DelSQLs = make(map[string][]string, len(key2sql))
	ListSQLHeads = make(map[string][]string, len(key2sql))
	for _, k := range key2sql {
		PutSQLs[k] = mkSQLs(pool, k, "insert into %s_%d (uid,datas) values($1,$2) on conflict (uid) do update set datas=$2")
		GetSQLs[k] = mkSQLs(pool, k, "select datas from %s_%d where uid = $1")
		DelSQLs[k] = mkSQLs(pool, k, "delete from %s_%d where uid = $1")
		ListSQLHeads[k] = mkSQLs(pool, k, "select uid from %s_%d order by uid ")
	}
}

func CreateAllTable(pool *pgx.ConnPool) {
	sql := `
	create table %s_%d (
		uid	varchar(64)	PRIMARY KEY,
		t	timestamptz	DEFAULT current_timestamp,
		datas	jsonb )
	`
	for _, k := range key2sql {
		for i := 0; i < MAX_SUB_TABLE; i++ {
			_, err := pool.Exec(fmt.Sprintf(sql, k, i))
			if err != nil {
				logs.Error("create %s %d err by %s", k, i, err.Error())
				return
			}
		}
	}
}

func getTableIDxByKey(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() % MAX_SUB_TABLE)
}

func GetDataFromPQ(pool *pgx.ConnPool, key, id string) ([]byte, error) {
	idx := getTableIDxByKey(id)
	s, ok := GetSQLs[key]
	if !ok {
		logs.Error("NoSQLKey %s, data no get", key)
		return []byte{}, nil
	}
	if len(s) <= idx {
		return []byte{}, errors.New("NoSQLIdx")
	}

	sql := s[idx]

	var datas []byte
	res := pool.QueryRow(sql, id)
	if res == nil {
		return []byte{}, errors.New("nullRes")
	}
	err := res.Scan(&datas)
	if err != nil {
		return []byte{}, err
	}
	return datas, nil
}

func SetDataToPQ(pool *pgx.ConnPool, key, id string, data []byte) error {
	idx := getTableIDxByKey(id)
	s, ok := PutSQLs[key]
	if !ok {
		logs.Error("NoSQLKey %s, data no set", key)
		return nil
	}
	if len(s) <= idx {
		return errors.New("NoSQLIdx")
	}

	j, err := accountJson.MkPrueJson(string(data))

	if err != nil {
		return err
	}

	ds, err := j.Encode()
	if err != nil {
		return err
	}

	_, err = pool.Exec(s[idx], id, ds)
	return err
}

func DelDataFromPQ(pool *pgx.ConnPool, key, id string) error {
	idx := getTableIDxByKey(id)
	s, ok := DelSQLs[key]
	if !ok {
		logs.Error("NoSQLKey %s, data no del", key)
		return nil
	}
	if len(s) <= idx {
		return errors.New("NoSQLIdx")
	}

	_, err := pool.Exec(s[idx], id)
	return err
}

func GetAllKeys(pool *pgx.ConnPool, keysChann func(key string)) error {
	wg := sync.WaitGroup{}
	wg.Add(1)
	for _, k := range key2sql {
		for i := 0; i < MAX_SUB_TABLE; i++ {
			key := ListSQLHeads[k][i]
			wg.Add(1)
			go func(keyHead, kh string) {
				defer wg.Done()
				err := getAllKeysByOneTable(pool, keyHead, kh+":", keysChann)
				if err != nil {
					logs.Critical("getAllKeysByOneTable Err By %s", err.Error())
				}
			}(key, k)
		}
	}
	wg.Done()
	wg.Wait()
	return nil
}

func getAllKeysByOneTable(pool *pgx.ConnPool, headSQL, profileName string, keysChann func(key string)) error {
	// select uid from profile_32 order by uid |limit 100 offset 0
	limitSQL := "limit 100 offset %d" // limit := 100
	nowOffset := 0
	for {
		sql := headSQL + fmt.Sprintf(limitSQL, nowOffset)
		nowOffset += 100
		rows, err := pool.Query(sql)
		if err != nil {
			logs.Error("rows nil %s", err.Error())
			return err
		}
		if rows == nil {
			logs.Error("rows nil")
			return errors.New("rows nil")
		}

		i := 0
		for rows.Next() {
			var uid string
			err := rows.Scan(&uid)
			if err != nil {
				return err
			}
			i++
			keysChann(profileName + uid)
		}
		if i == 0 {
			return nil
		}
	}
	return nil
}
