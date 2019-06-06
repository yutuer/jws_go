package examples

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

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

// GetVCSRootPath 返回vcs.taiyouxi.net目录的全路径，如果无法找到，返回""
func GetVCSRootPath() string {
	rootPath := ""

	// 从当前目录往根目录搜
	curPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// 如果当前路径不包括vcs，不会进入for loop;
	for ; strings.Contains(curPath, "vcs.taiyouxi.net"); curPath = filepath.Dir(curPath) {
		rootPath = curPath
	}

	// 如果本地搜不到，则搜索每个GoPath
	if rootPath == "" {
		goPathes := GetSplitGoPath()
		for _, goPath := range goPathes {
			vcsPath := filepath.Join(goPath, "/src/vcs.taiyouxi.net")
			if stat, err := os.Stat(vcsPath); !os.IsNotExist(err) && stat.IsDir() {
				rootPath = vcsPath
			}
		}
	}

	return rootPath
}

// GetDataFilePath 返回读取data文件的路径，可以Mock
func GetDataFileFullPath(filename string) string {

	return filepath.Join(GetVCSRootPath(), "jws/gamex/conf/data", filename+".data")
}
