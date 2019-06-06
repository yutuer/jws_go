package main

import (
	"fmt"
	"log"
	"os"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"

	"io/ioutil"

	"encoding/json"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "csrob_rank_fix"
	app.Usage = "fix csrob rank data for new version in csrob db"

	app.Commands = []cli.Command{
		cli.Command{
			Name:   "fix",
			Usage:  "fix csrob db data",
			Action: fixAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "batch",
					Usage: "batch of rank data",
					Value: "0",
				},
				cli.StringFlag{
					Name:  "f",
					Usage: "filename of db info",
					Value: "nil.json",
				},
			},
		},
		cli.Command{
			Name:   "clear",
			Usage:  "clear batch csrob db data",
			Action: clearAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "batch",
					Usage: "batch of rank data",
					Value: "0",
				},
				cli.StringFlag{
					Name:  "f",
					Usage: "filename of db info",
					Value: "nil.json",
				},
			},
		},
	}

	app.Run(os.Args)
}

func fixAction(c *cli.Context) {
	filename := c.String("f")
	batch := c.String("batch")

	fileBytes, err := ioutil.ReadFile(filename)
	if nil != err {
		fmt.Println("ReadFile Err:", err)
		return
	}

	dbList := map[string]DBConfig{}
	if err := json.Unmarshal(fileBytes, &dbList); nil != err {
		fmt.Println("Unmarshal Err:", err)
		return
	}

	for group, db := range dbList {
		if "default" == group {
			continue
		}
		func() {
			conn, err := redis.Dial("tcp", db.AddrPort, redis.DialDatabase(int(db.DB)), redis.DialPassword(db.Auth))
			if nil != err {
				log.Printf("!!WARN, connect db failed, group [%s], db [%+v], err: %v", group, db, err)
				return
			}
			defer conn.Close()

			//先转存robrank
			oldTableRank := fmt.Sprintf("csrob:%s:common:robrank", group)
			newTableRank := fmt.Sprintf("csrob:%s:common:robrank:%s", group, batch)
			values, err := redis.Values(conn.Do("ZREVRANGE", oldTableRank, 0, -1, "withscores"))
			if nil != err && redis.ErrNil != err {
				log.Printf("!!WARN, redis do ZREVRANGE failed, group [%s], db [%+v], err: %v", group, db, err)
				return
			}
			if redis.ErrNil == err {
				log.Printf("Ignore group [%s] in robrank by no rank data", group)
				return
			}
			guildList := []string{}
			count := 0
			for i := 0; i+1 < len(values); i += 2 {
				guild, err := redis.String(values[i], nil)
				if nil != err {
					log.Printf("!!WARN, redis parse guild ID failed, group [%s], err: %v", group, err)
					continue
				}
				scores, err := redis.Float64(values[i+1], nil)
				if nil != err {
					log.Printf("!!WARN, redis parse scores failed, group [%s], err: %v", group, err)
					continue
				}
				if _, err := conn.Do("ZADD", newTableRank, scores, guild); nil != err {
					log.Printf("!!WARN, redis do ZADD failed, group [%s] guild [%s] score [%f], err: %v", group, guild, scores, err)
					continue
				}
				guildList = append(guildList, guild)
				count++
			}
			log.Printf("Done group [%s] robrank [%s] to [%s], count [%d] to [%d]", group, oldTableRank, newTableRank, len(values)/2, count)

			//转存robtimes
			newTableTimes := fmt.Sprintf("csrob:%s:common:robtimes:%s", group, batch)
			count = 0
			for _, guild := range guildList {
				oldTableTimes := fmt.Sprintf("csrob:%s:guild:%s:info", group, guild)
				rob, err := redis.Int(conn.Do("HGET", oldTableTimes, "robs"))
				if nil != err && redis.ErrNil != err {
					log.Printf("!!WARN, redis HGET robs failed, group [%s] guild [%s], err: %v", group, guild, err)
					continue
				}
				if redis.ErrNil == err {
					log.Printf("Ignore group [%s] guild [%s] in robtimes by no rob", group, guild)
					return
				}
				last, err := redis.Int64(conn.Do("HGET", oldTableTimes, "last"))
				if nil != err {
					log.Printf("!!WARN, redis HGET last failed, group [%s] guild [%s], err: %v", group, guild, err)
					continue
				}
				newFieldRobs := fmt.Sprintf("robs:%s", guild)
				if _, err := conn.Do("HSET", newTableTimes, newFieldRobs, rob); nil != err {
					log.Printf("!!WARN, redis do HSET robs failed, group [%s] guild [%s] robs [%d], err: %v", group, guild, rob, err)
					continue
				}
				newFieldLast := fmt.Sprintf("last:%s", guild)
				if _, err := conn.Do("HSET", newTableTimes, newFieldLast, last); nil != err {
					log.Printf("!!WARN, redis do HSET robs failed, group [%s] guild [%s] robs [%d], err: %v", group, guild, rob, err)
					continue
				}
				count++
			}
			log.Printf("Done group [%s] robtimes [%s], count [%d] to [%d]", group, newTableTimes, len(guildList), count)
		}()
	}

	return
}

func clearAction(c *cli.Context) {
	filename := c.String("f")
	batch := c.String("batch")

	fileBytes, err := ioutil.ReadFile(filename)
	if nil != err {
		fmt.Println("ReadFile Err:", err)
		return
	}

	dbList := map[string]DBConfig{}
	if err := json.Unmarshal(fileBytes, &dbList); nil != err {
		fmt.Println("Unmarshal Err:", err)
		return
	}

	for group, db := range dbList {
		if "default" == group {
			continue
		}
		func() {
			conn, err := redis.Dial("tcp", db.AddrPort, redis.DialDatabase(int(db.DB)), redis.DialPassword(db.Auth))
			if nil != err {
				log.Printf("!!WARN, connect db failed, group [%s], db [%+v], err: %v", group, db, err)
				return
			}
			defer conn.Close()

			newTableRank := fmt.Sprintf("csrob:%s:common:robrank:%s", group, batch)
			newTableTimes := fmt.Sprintf("csrob:%s:common:robtimes:%s", group, batch)
			if _, err := conn.Do("DEL", newTableRank, newTableTimes); nil != err && redis.ErrNil != err {
				log.Printf("!!WARN, redis do DEL failed, err: %v", err)
			}
			log.Printf("Clear group [%s]", group)
		}()
	}
}

//DBConfig ..
type DBConfig struct {
	AddrPort string
	Auth     string
	DB       uint32
}
