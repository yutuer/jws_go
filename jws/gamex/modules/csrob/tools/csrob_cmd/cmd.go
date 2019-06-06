package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

func main() {
	app := cli.NewApp()

	app.Name = "csrob_cmd"
	app.Usage = "tools for operating csrob db"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "s",
			Usage: "IP of redis server",
			Value: "127.0.0.1",
		},
		cli.IntFlag{
			Name:  "p",
			Usage: "Port of redis server",
			Value: 6379,
		},
		cli.IntFlag{
			Name:  "d",
			Usage: "DB of CSRob in redis server",
			Value: -1,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "clear",
			Usage:  "clear csrob db data",
			Action: clearAction,
		},
	}

	app.Run(os.Args)
}

func clearAction(c *cli.Context) {
	ip := c.GlobalString("s")
	port := c.GlobalInt("p")
	db := c.GlobalInt("d")

	if -1 == db {
		fmt.Println("Invalid DB :", db)
		return
	}

	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", ip, port), redis.DialDatabase(db))
	if nil != err {
		fmt.Println("Err:", err)
		return
	}

	keys, err := redis.Strings(conn.Do("Keys", "csrob*"))
	if nil != err {
		fmt.Println("Err:", err)
		return
	}

	for _, key := range keys {
		if _, err := conn.Do("DEL", key); nil != err {
			fmt.Println("Err:", err)
			continue
		}
	}

	return
}
