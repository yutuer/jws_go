package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"time"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

var outputFile *os.File
var redisHost string = "127.0.0.1:6379"

func main() {
	app := cli.NewApp()

	app.Version = "0.0.0"
	app.Name = "getInventoryFrom240"
	app.Usage = "getInventoryFrom240"
	app.Author = "LBB"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "gid",
			Value: "",
			Usage: "大区",
		},
		cli.StringFlag{
			Name:  "redis",
			Value: "",
			Usage: "redis host",
		},
	}

	app.Action = func(context *cli.Context) {
		var err error
		fileName := fmt.Sprintf("lost_inventory_%s.txt", context.String("gid"))
		outputFile, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}
		redisHost = context.String("redis")
		for i := 0; i < 32; i++ {
			fmt.Println("\nstart scan db ========", i)
			err := getInfoFromDB(i)
			if err != nil {
				fmt.Println("select db error", i, err)
			}

			fmt.Println("done <<<<<<<<<")
		}
	}

	app.Run(os.Args)
}

func getInfoFromDB(dbIndex int) error {
	conn, err := redis.Dial("tcp", redisHost,
		redis.DialConnectTimeout(1*time.Second),
		redis.DialReadTimeout(5*time.Second),
		redis.DialWriteTimeout(5*time.Second),
		redis.DialPassword(""),
		redis.DialDatabase(dbIndex))
	if err != nil {
		return err
	}

	guildKeys, err := scanGuildKey(conn)
	if err != nil {
		return err
	}
	fmt.Println(guildKeys)
	if len(guildKeys) == 0 {
		fmt.Println("find no guild info")
		return nil
	}

	getGuildLostGood(conn, guildKeys)
	return nil
}

func scanGuildKey(conn redis.Conn) ([]string, error) {
	index := 0
	guildKeys := make([]string, 0)
	for {
		reply, err := conn.Do("SCAN", index, "MATCH", "guild:Info:*")
		if err != nil {
			fmt.Println(err)
			break
		}
		reply1, err := redis.Values(reply, nil)
		if len(reply1) != 2 {
			fmt.Println("bad result ", reply1, err)
			break
		}
		index, err = redis.Int(reply1[0], nil)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println("next index ", index)
		tempKeys, err := redis.Strings(reply1[1], nil)
		if err != nil {
			fmt.Println(err)
			break
		}
		guildKeys = append(guildKeys, tempKeys...)
		if index == 0 {
			break
		}
	}

	return guildKeys, nil
}

type GuildLostInventory struct {
	GuildUid      string
	LostInventory string
}

func getGuildLostGood(conn redis.Conn, guildKeys []string) {
	for _, guildKey := range guildKeys {
		inverntory, err := redis.String(conn.Do("HGET", guildKey, "Inventory"))
		if err != nil {
			fmt.Println(err)
		} else {
			lost := GuildLostInventory{
				GuildUid:      guildKey,
				LostInventory: inverntory,
			}
			bytes, err := json.Marshal(lost)
			if err != nil {
				fmt.Println(err)
			} else {
				outputFile.Write(bytes)
				outputFile.WriteString("\n")
			}
		}
	}
}
