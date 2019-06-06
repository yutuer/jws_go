package main

import (
	"os"

	"fmt"

	"io/ioutil"

	"encoding/json"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "csrob_db_build"
	app.Usage = "tools for build csrob db config json"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "verbose of building",
		},
		cli.StringFlag{
			Name:  "a",
			Usage: "auth of redis",
			Value: "",
		},
		cli.StringFlag{
			Name:  "o",
			Usage: "filename of output",
			Value: "output.json",
		},
	}
	app.Action = buildAction

	app.Run(os.Args)
}

func buildAction(c *cli.Context) {
	isVerbose := c.GlobalBool("verbose")
	authStr := c.GlobalString("a")

	args := c.Args()
	if 0 == len(args) {
		fmt.Printf("need filename\n")
		return
	}

	filename := args[0]
	bs, err := ioutil.ReadFile(filename)
	if nil != err {
		fmt.Printf("ReadFile %s failed, %v", filename, err)
		return
	}

	if isVerbose {
		fmt.Printf("Read file: %+v", string(bs))
	}

	configs := []BuildConfig{}
	if err := json.Unmarshal(bs, &configs); nil != err {
		fmt.Printf("Parse file failed, %v", err)
		return
	}

	output := map[string]OutputConfig{}

	isDefault := true
	for _, cfg := range configs {
		dbIndex := 0
		serverIndex := 0
		for gid := cfg.GroupIDMin; gid <= cfg.GroupIDMax; gid++ {
			if dbIndex >= len(cfg.DBs) {
				serverIndex++
				dbIndex = dbIndex - len(cfg.DBs)
				if serverIndex >= len(cfg.Servers) {
					serverIndex = serverIndex - len(cfg.Servers)
				}
			}

			server := cfg.Servers[serverIndex]
			db := cfg.DBs[dbIndex]

			out := OutputConfig{
				AddrPort: fmt.Sprintf("%s:6379", server),
				Auth:     authStr,
				DB:       db,
			}

			if isDefault {
				output["default"] = out
				isDefault = false
			}

			output[fmt.Sprintf("%d", gid)] = out

			dbIndex++
		}
	}

	outBs, err := json.Marshal(output)
	if nil != err {
		fmt.Printf("Marshal output failed, %v", err)
		return
	}

	outfile := c.GlobalString("o")
	if err := ioutil.WriteFile(outfile, outBs, os.FileMode(0666)); nil != err {
		fmt.Printf("WriteFile %s failed, %v", outfile, err)
		return
	}

	return
}

//BuildConfig ..
type BuildConfig struct {
	Servers    []string
	DBs        []uint32
	GroupIDMin uint32
	GroupIDMax uint32
}

//OutputConfig ..
type OutputConfig struct {
	AddrPort string
	Auth     string
	DB       uint32
}
