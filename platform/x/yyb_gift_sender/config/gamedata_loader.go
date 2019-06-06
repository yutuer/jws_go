package config

import (
	"fmt"
	"github.com/astaxie/beego/utils"
	"io"
	"os"
	"path/filepath"
)

func loadBin(cfgname string) ([]byte, error) {
	errgen := func(err error, extra string) error {
		return fmt.Errorf("gamex.models.gamedata loadbin Error, %s, %s", extra, err.Error())
	}

	//	path := GetDataPath()
	//	appConfigPath := filepath.Join(path, cfgname)

	file, err := os.Open(cfgname)
	if err != nil {
		return nil, errgen(err, "open")
	}

	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return nil, errgen(err, "stat")
	}

	buffer := make([]byte, fi.Size())
	_, err = io.ReadFull(file, buffer) //read all content
	if err != nil {
		return nil, errgen(err, "readfull")
	}

	return buffer, nil
}

func GetDataPath() string {
	workPath, _ := os.Getwd()
	workPath, _ = filepath.Abs(workPath)
	// initialize default configurations
	AppPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	appConfigPath := filepath.Join(AppPath, "conf")
	if workPath != AppPath {
		if utils.FileExists(appConfigPath) {
			os.Chdir(AppPath)
		} else {
			appConfigPath = filepath.Join(workPath, "conf")
		}
	}
	return appConfigPath
}
