package gamedata

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DebugLoadLocalGamedata 根据GOPATH路径载入本地gamedata数据
func DebugLoadLocalGamedata() {
	var gamexPath string
	goPathes := GetSplitGoPath()

	for _, goPath := range goPathes {
		gamexPath = filepath.Join(goPath, "src/vcs.taiyouxi.net/jws/gamex")
		if stat, err := os.Stat(gamexPath); !os.IsNotExist(err) && stat.IsDir() {
			os.Chdir(gamexPath)
			LoadGameData("")
		}
	}
}

// GetSplitGoPath 读取环境GOPATH并转换为slice
func GetSplitGoPath() []string {
	goPathes := make([]string, 16)
	goPath := os.Getenv("GOPATH")

	if runtime.GOOS == "windows" && strings.Contains(goPath, ";") {
		goPathes = strings.Split(goPath, ";")
	} else if runtime.GOOS != "windows" && strings.Contains(goPath, ":") {
		goPathes = strings.Split(goPath, ":")
	} else {
		goPathes = []string{goPath}
	}

	return goPathes
}
