package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Dir   Dir   `toml:"Dir"`
	Sheet Sheet `toml:"Sheet"`
	Local Local `toml:"Local"`
	// eg. Characters/Items/Pottery/box.prefab => key:[prefab] value:Characters\Items\Pottery\box
	AllClientFiles map[string][]string
}

type Dir struct {
	RunOnTeamCity       bool   // 是否在TC上运行
	ClientProjectDir    string // 客户端工程目录
	DataProjectDir      string // 数据工程目录
	LocalProjectDir     string // 多语言资源目录
	GamedataDir         string // data文件所在目录
	TCCheckOutDir       string // Teamcity使用的Checkout目录
	TCClientArtiRelPath string // TC Client传递的Artifacts相对路径
	TCDataArtiRelPath   string // TC Data传递的Artifacts相对路径
	TCClientFileList    string // Client 传递的文件名列表txt文件
}

type Sheet struct {
	SheetStartNum int // Sheet开始序号；0号表经常用于说明，一般从1号开始
	FieldType     int // 列类型所在行数
	DataTypeRow   int // 数据类型行数
	ColNameIdx    int // 列标题所在行数
	DataStartRow  int // 正式数据开始行数
}

type Local struct {
	Locals       []string // 多语言文件名
	DefaultLocal string   // 首选多语言
}

// SearchInConf 在当前目录和上级目录搜索指定文件名，如果存在则返回文件的完整路径，不存在则返回空值
func SearchInConf(fileName string) string {
	vcsRoot := GetVCSRootPath()
	fn := filepath.Join(vcsRoot, "tools/dataChecker/conf", fileName)

	if _, err := os.Stat(fn); !os.IsNotExist(err) {
		return fn
	}
	return ""
}

// LoadConfig 读取config.toml文件并将设置转换成对应go结构。如果RunOnTeamCity为true，则将三个目录指向teamcity workdir的构建目录
func (c *Config) LoadConfig() {
	if cf := SearchInConf("config.toml"); cf == "" {
		panic("Cannot find any config file!")
	} else {
		if md, e := toml.DecodeFile(cf, c); e != nil {
			fmt.Printf("Undecoded keys: %q\n", md.Undecoded())
		}
	}

	// 以下是teamcity配置
	if c.Dir.RunOnTeamCity == true {
		c.Dir.TCCheckOutDir = os.Getenv("TCCheckOutDir")
		if c.Dir.TCCheckOutDir == "" {
			panic("Unable to Get TCCheckOutDir")
		}
		c.Dir.ClientProjectDir = filepath.Join(c.Dir.TCCheckOutDir, cfg.Dir.TCClientArtiRelPath)
		c.Dir.DataProjectDir = filepath.Join(c.Dir.TCCheckOutDir, cfg.Dir.TCDataArtiRelPath)
		c.Dir.LocalProjectDir = filepath.Join(c.Dir.TCCheckOutDir, cfg.Dir.TCClientArtiRelPath)

		// 读取客户端扫描过来的文件名
		c.LoadAllClientFileName()
	}
}

// DebugLoadLocalConfig 强制将cfg设置为非teamcity模式
func (c *Config) DebugLoadLocalConfig() {
	if cf := SearchInConf("config.toml"); cf == "" {
		panic("Cannot find any config file!")
	} else {
		if md, e := toml.DecodeFile(cf, c); e != nil {
			fmt.Printf("Undecoded keys: %q\n", md.Undecoded())
		}
	}

	c.Dir.RunOnTeamCity = false
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
	return filepath.Join(GetVCSRootPath(), cfg.Dir.GamedataDir, filename+".data")
}

// DebugSetGamedataDir 是为了单元测试Mock预留的接口
func DebugSetGamedataDir(path string) {
	cfg.Dir.GamedataDir = path
}

// LoadAllClientFileName从Client传递的Artifacts文件里，获取之前生成的所有文件列表，并以map形式存储到AllClientFiles
func (c *Config) LoadAllClientFileName() {
	fName := filepath.Join(cfg.Dir.ClientProjectDir, cfg.Dir.TCClientFileList)
	f, err := os.Open(fName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	c.AllClientFiles = make(map[string][]string, 64)
	c.AllClientFiles[""] = make([]string, 0)

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()

		// 扩展名塞到对应的slice里, 包括nil
		ext := filepath.Ext(line)
		if _, ok := c.AllClientFiles[ext]; ok {
			c.AllClientFiles[ext] = append(c.AllClientFiles[ext], line)
		} else {
			nameSlice := make([]string, 0)
			nameSlice = append(nameSlice, line)
			c.AllClientFiles[ext] = nameSlice
		}
	}
}

// DebugLoadGamedataAll Load gamedata
func DebugLoadGamedataAll() {
	rootPath := GetVCSRootPath()
	gamexPath := filepath.Join(rootPath, "jws/gamex")
	os.Chdir(gamexPath)
	gamedata.LoadGameData("")
}
