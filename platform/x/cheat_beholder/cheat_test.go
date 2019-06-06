package main

import (
	"fmt"
	"github.com/astaxie/beego/utils"
	"os"
	"path/filepath"
	"testing"
)

func getMyDataPath() string {
	workPath, _ := os.Getwd()
	fmt.Printf("1os.Getwd():%v\n", workPath)
	workPath, _ = filepath.Abs(workPath)
	fmt.Printf("2filepath.Abs(workPath):%v\n", workPath)
	// initialize default configurations
	AppPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	fmt.Printf("3filepath.Abs(filepath.Dir(os.Args[0])):%v\n", AppPath)

	appConfigPath := filepath.Join(AppPath, "conf")
	fmt.Printf("4filepath.Join(AppPath, `conf`):%v\n", appConfigPath)

	if workPath != AppPath {
		if utils.FileExists(appConfigPath) {
			os.Chdir(AppPath)
		} else {
			appConfigPath = filepath.Join(workPath, "conf")
		}
	}
	fmt.Printf("5result:%v", appConfigPath)
	return appConfigPath
}

func TestGetMyDataPath(t *testing.T) {
	fmt.Println(getMyDataPath())
}
