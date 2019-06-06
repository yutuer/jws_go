package main

import (
	"os"
	"path/filepath"
	"strings"
	"vcs.taiyouxi.net/platform/x/tiprotogen/command"
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
	defer log.Flush()
	all := getFilelist("./", "json")

	for _, protoPath := range all {
		log.Info("processing %s", protoPath)
		d, err := command.LoadFromFile(protoPath)
		if err != nil {
			log.Err("err %s", err.Error())
			return
		}
		command.GenMultipleCSharpCode("./", d)
		log.Info("processing %s success", protoPath)
	}
}
