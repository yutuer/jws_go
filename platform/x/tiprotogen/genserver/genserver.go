package main

import (
	"github.com/codegangsta/cli"
	"os"
	"path/filepath"
	"strings"
	"vcs.taiyouxi.net/platform/x/tiprotogen/command"
	"vcs.taiyouxi.net/platform/x/tiprotogen/def"
	"vcs.taiyouxi.net/platform/x/tiprotogen/log"
	"vcs.taiyouxi.net/platform/x/tiprotogen/util"
)

func getFilelist(path string, typ string) []string {
	res := make([]string, 0, 256)
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		paths := strings.Split(path, ".")
		if strings.ToLower(paths[len(paths)-1]) == typ {
			res = append(res, path)
		}

		return nil
	})
	if err != nil {
		util.PanicInfo("filepath.Walk() returned %v\n", err)
	}
	return res
}

func main() {
	app := cli.NewApp()

	app.Name = "genserver"
	app.Usage = "gen server proto"
	app.Author = "LBB"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "f",
			Usage: "生成服务器端的proto文件",
		},
	}

	app.Action = startAction

	app.Run(os.Args)
}

func startAction(c *cli.Context) {
	defer log.Flush()
	overwrite := c.Bool("f")
	all := getFilelist("./", "json")

	defs := make([]*dsl.ProtoDef, 0, 256)

	for _, protoPath := range all {
		log.Info("processing %s", protoPath)
		d, err := command.LoadFromFile(protoPath)
		if err != nil {
			log.Err("err %s", err.Error())
			return
		}
		destFileName := strings.ToLower(protoPath)
		destFileName = strings.Split(destFileName, ".")[0]
		info, _ := os.Stat("../" + destFileName + ".go")
		if !overwrite && info != nil {
			log.Err("This has an exsit file %s", protoPath)
			continue
		}
		if info == nil {
			for i := range d {
				defs = append(defs, &d[i])
			}
		}
		command.GenGolangCodes("../", destFileName, d)
		info, _ = os.Stat("../" + destFileName + "handler" + ".go")
		if info == nil {
			log.Info("new file", protoPath)
			command.GenGolangHandlers("../", destFileName+"handler", d)
		}
		log.Info("processing %s success", protoPath)
	}

	command.AppendGolangPathCode("../", defs)
	log.Info("processing all success")
}
