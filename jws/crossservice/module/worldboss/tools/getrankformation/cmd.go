package main

import (
	"os"

	"fmt"

	"log"

	"encoding/json"

	"bytes"

	"encoding/gob"

	"io/ioutil"

	"strings"

	"strconv"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/jws/crossservice/module/worldboss"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		cli.Command{
			Name:   "getRankFormation",
			Usage:  "get rank formation data to file",
			Action: getRankFormation,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "i",
					Usage: "ip of redis db",
					Value: "127.0.0.1",
				},
				cli.IntFlag{
					Name:  "p",
					Usage: "port of redis db",
					Value: 6379,
				},
				cli.IntFlag{
					Name:  "n",
					Usage: "db of redis db",
					Value: 8,
				},
				cli.StringFlag{
					Name:  "a",
					Usage: "auth of redis db",
					Value: "",
				},
				cli.IntSliceFlag{
					Name:  "g",
					Usage: "group id list",
				},
				cli.StringSliceFlag{
					Name:  "d",
					Usage: "days to get",
				},
				cli.StringFlag{
					Name:  "o",
					Usage: "output file name",
					Value: "db_data.data",
				},
			},
		},
		cli.Command{
			Name:   "setRankFormation",
			Usage:  "set rank formation data to redis",
			Action: setRankFormation,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "i",
					Usage: "ip of redis db",
					Value: "127.0.0.1",
				},
				cli.IntFlag{
					Name:  "p",
					Usage: "port of redis db",
					Value: 6379,
				},
				cli.IntFlag{
					Name:  "n",
					Usage: "db of redis db",
					Value: 4,
				},
				cli.StringFlag{
					Name:  "s",
					Usage: "source file name",
					Value: "db_data.data",
				},
			},
		},
		cli.Command{
			Name:   "csv",
			Usage:  "make data to csv format",
			Action: makeCsv,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "i",
					Usage: "ip of redis db",
					Value: "127.0.0.1",
				},
				cli.IntFlag{
					Name:  "p",
					Usage: "port of redis db",
					Value: 6379,
				},
				cli.IntFlag{
					Name:  "n",
					Usage: "db of redis db",
					Value: 4,
				},
				cli.StringFlag{
					Name:  "a",
					Usage: "auth of redis db",
					Value: "",
				},
				cli.IntSliceFlag{
					Name:  "g",
					Usage: "group id list",
				},
				cli.StringSliceFlag{
					Name:  "d",
					Usage: "days to get",
				},
				cli.StringFlag{
					Name:  "o",
					Usage: "output file name prefix",
					Value: "csv",
				},
			},
		},
	}

	app.Run(os.Args)
}

func getRankFormation(c *cli.Context) {
	paramIP := c.String("i")
	paramPort := c.Int("p")
	paramDB := c.Int("n")
	paramAuth := c.String("a")
	paramGroupList := c.IntSlice("g")
	paramDayList := c.StringSlice("d")
	paramOutput := c.String("o")

	conn, err := redis.Dial("tcp4", fmt.Sprintf("%s:%d", paramIP, paramPort), redis.DialDatabase(paramDB), redis.DialPassword(paramAuth))
	if nil != err {
		log.Printf("redis.Dial failed, %v", err)
		return
	}
	defer conn.Close()

	allDay := make(map[string]map[int]map[string]*worldboss.FormationRankElem)
	for _, dayStr := range paramDayList {
		allDay[dayStr] = make(map[int]map[string]*worldboss.FormationRankElem)

		groupList := paramGroupList
		if 0 == len(groupList) {
			list := []int{}
			scanIndex := int(0)
			for {
				res, err := redis.Values(conn.Do("SCAN", scanIndex, "MATCH", fmt.Sprintf("worldboss:*:formationrank:%s", dayStr), "COUNT", 10000))
				if nil != err {
					log.Printf("SCAN failed, %v", err)
					return
				}
				ni, err := redis.Int(res[0], nil)
				if nil != err {
					log.Printf("SCAN Int failed, %v", err)
					return
				}
				keys, err := redis.Strings(res[1], nil)
				if nil != err {
					log.Printf("SCAN Strings failed, %v", err)
					return
				}

				for _, key := range keys {
					ret := strings.Split(key, ":")
					groupID, err := strconv.Atoi(ret[1])
					if nil != err {
						log.Printf("Atoi failed, %v", err)
						return
					}
					list = append(list, groupID)
				}

				if 0 == ni {
					break
				}
			}
			groupList = list
		}

		for _, group := range groupList {
			key := fmt.Sprintf("worldboss:%d:formationrank:%s", group, dayStr)
			res, err := redis.StringMap(conn.Do("HGETALL", key))
			if nil != err && redis.ErrNil != err {
				log.Printf("HGETALL failed, %v", err)
				return
			}
			if redis.ErrNil == err {
				continue
			}

			allDay[dayStr][group] = make(map[string]*worldboss.FormationRankElem)
			for acid, info := range res {
				elem := &worldboss.FormationRankElem{}
				if err := json.Unmarshal([]byte(info), elem); nil != err {
					continue
				}
				allDay[dayStr][group][acid] = elem
			}
		}
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(allDay); nil != err {
		log.Printf("gob.Encode failed, %v", err)
		return
	}
	if err := ioutil.WriteFile(paramOutput, buf.Bytes(), os.ModePerm); nil != err {
		log.Printf("ioutil.WriteFile failed, %v", err)
		return
	}
}

func setRankFormation(c *cli.Context) {
	paramIP := c.String("i")
	paramPort := c.Int("p")
	paramDB := c.Int("n")
	paramSource := c.String("s")

	bs, err := ioutil.ReadFile(paramSource)
	if nil != err {
		log.Printf("ioutil.ReadFile failed, %v", err)
		return
	}

	buf := bytes.NewBuffer(bs)

	allDay := make(map[string]map[int]map[string]*worldboss.FormationRankElem)
	if err := gob.NewDecoder(buf).Decode(&allDay); nil != err {
		log.Printf("gob.Decode failed, %v", err)
		return
	}

	conn, err := redis.Dial("tcp4", fmt.Sprintf("%s:%d", paramIP, paramPort), redis.DialDatabase(paramDB))
	if nil != err {
		log.Printf("redis.Dial failed, %v", err)
		return
	}
	defer conn.Close()

	for dayStr, dayInfo := range allDay {
		for group, groupInfo := range dayInfo {
			key := fmt.Sprintf("worldboss:%d:formationrank:%s", group, dayStr)
			for acid, elem := range groupInfo {
				bs, err := json.Marshal(elem)
				if nil != err {
					log.Printf("Marshal failed, %v", err)
					return
				}
				if _, err := conn.Do("HSET", key, acid, string(bs)); nil != err {
					log.Printf("HSET failed, %v", err)
					return
				}
			}
		}
	}
}

func makeCsv(c *cli.Context) {
	paramIP := c.String("i")
	paramPort := c.Int("p")
	paramDB := c.Int("n")
	paramAuth := c.String("a")
	paramGroupList := c.IntSlice("g")
	paramDayList := c.StringSlice("d")
	paramOutput := c.String("o")

	conn, err := redis.Dial("tcp4", fmt.Sprintf("%s:%d", paramIP, paramPort), redis.DialDatabase(paramDB), redis.DialPassword(paramAuth))
	if nil != err {
		log.Printf("redis.Dial failed, %v", err)
		return
	}
	defer conn.Close()

	allDay := make(map[string]map[int]map[string]*worldboss.FormationRankElem)
	for _, dayStr := range paramDayList {
		allDay[dayStr] = make(map[int]map[string]*worldboss.FormationRankElem)

		groupList := paramGroupList
		if 0 == len(groupList) {
			list := []int{}
			scanIndex := int(0)
			for {
				res, err := redis.Values(conn.Do("SCAN", scanIndex, "MATCH", fmt.Sprintf("worldboss:*:formationrank:%s", dayStr), "COUNT", 10000))
				if nil != err {
					log.Printf("SCAN failed, %v", err)
					return
				}
				ni, err := redis.Int(res[0], nil)
				if nil != err {
					log.Printf("SCAN Int failed, %v", err)
					return
				}
				keys, err := redis.Strings(res[1], nil)
				if nil != err {
					log.Printf("SCAN Strings failed, %v", err)
					return
				}

				for _, key := range keys {
					ret := strings.Split(key, ":")
					groupID, err := strconv.Atoi(ret[1])
					if nil != err {
						log.Printf("Atoi failed, %v", err)
						return
					}
					list = append(list, groupID)
				}

				if 0 == ni {
					break
				}
			}
			groupList = list
		}

		for _, group := range groupList {
			key := fmt.Sprintf("worldboss:%d:formationrank:%s", group, dayStr)
			res, err := redis.StringMap(conn.Do("HGETALL", key))
			if nil != err && redis.ErrNil != err {
				log.Printf("HGETALL failed, %v", err)
				return
			}
			if redis.ErrNil == err {
				continue
			}

			allDay[dayStr][group] = make(map[string]*worldboss.FormationRankElem)
			for acid, info := range res {
				elem := &worldboss.FormationRankElem{}
				if err := json.Unmarshal([]byte(info), elem); nil != err {
					continue
				}
				allDay[dayStr][group][acid] = elem
			}
		}
	}

	for dayStr, dayInfo := range allDay {
		for group, groupInfo := range dayInfo {
			filename := fmt.Sprintf("%s_%s_%d.csv", paramOutput, dayStr, group)
			fileContext := ""
			for _, elem := range groupInfo {
				gs := int64(0)
				strHeros := ""
				for _, h := range elem.Team {
					gs += h.BaseGs + h.ExtraGs
					if "" != strHeros {
						strHeros += ","
					}
					strHeros += fmt.Sprintf("%d", h.Idx)
				}
				str := fmt.Sprintf("%s,%d,%d,%d,%s\n", elem.Acid, elem.Damage, elem.BuffLevel, gs, strHeros)
				fileContext += str
			}

			if err := ioutil.WriteFile(filename, []byte(fileContext), os.ModePerm); nil != err {
				if nil != err {
					log.Printf("ioutil.WriteFile failed, %v", err)
					return
				}
			}
		}
	}
}
