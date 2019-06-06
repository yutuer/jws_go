package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/astaxie/beego/utils"
	"github.com/codegangsta/cli"
	"github.com/gogo/protobuf/proto"

	"bufio"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/tools/ticonfig/protogen"
)

func main() {
	defer logs.Close()
	app := cli.NewApp()

	app.Name = "genserver"
	app.Usage = "gen server proto"
	app.Author = "LBB"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "d",
			Usage: "指定proto文件目录",
		},
		cli.StringFlag{
			Name:  "o",
			Usage: "输出目录",
		},
		cli.StringFlag{
			Name:  "data",
			Value: "",
			Usage: "需要读取的配置文件目录, 如果不填， 表示不生成该部分代码",
		},
	}

	app.Action = startAction

	app.Run(os.Args)
}

func startAction(c *cli.Context) {
	protoDir := c.String("d")
	outDir := c.String("o")
	dataDir := c.String("data")
	configList := parseFileList(protoDir)
	logs.Debug("%v", configList)
	genAllCode(configList, outDir)
	if dataDir != "" {
		dataAbsPath := filepath.Join(GetDataPath(), dataDir, "data")
		loadAndGenDatas("", dataAbsPath, outDir)
	}
}

func genAllCode(configList []ConfigInfo, outPath string) {
	headStr := header
	loadAllStr := genLoadAllStr(configList)
	allConfigStr := genAllConfig(configList)

	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(headStr)
	buf.WriteString(loadAllStr)
	buf.WriteString(allConfigStr)
	ioutil.WriteFile(outPath+"gamedata_genereated.go", buf.Bytes(), 0666)
}

var header string = `package gamedata

import (
	"path/filepath"

	"sort"

	"github.com/gogo/protobuf/proto"
	"vcs.taiyouxi.net/comic/gamedata/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

/**
自动生成的文件，不要修改
*/
`

type ConfigInfo struct {
	ConfigName string
	Keys       []KeyStruct
}

func parseFileList(path string) []ConfigInfo {
	res := make([]ConfigInfo, 0, 256)
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		logs.Debug("find file %s", path)
		if strings.Index(path, "ProtobufGen") >= 0 &&
			strings.Index(path, ".pb.go") >= 0 {
			nameSplits := strings.Split(path[:len(path)-6], "/")
			name := nameSplits[len(nameSplits)-1]
			logs.Info("parse file begin %s", name)
			configInfo := parseFile(path, name[12:])
			logs.Info("parse file result %s", configInfo)
			if configInfo.ConfigName != "" {
				res = append(res, configInfo)
			}
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
	return res
}

func parseFile(path string, name string) ConfigInfo {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		logs.Error("", err)
		return ConfigInfo{}
	}
	protoReader := bufio.NewReader(bytes.NewReader(file))
	reqBodyLineIndex := -1 // 如果大于0表示该行是字段行
	nameByUpper := strings.ToUpper(name)
	realName := name
	keys := make([]KeyStruct, 0)
	for {
		bytes, err := protoReader.ReadBytes('\n')
		if err != nil {
			logs.Error("", err)
			return ConfigInfo{}
		}
		lineStr := string(bytes)

		if tempName := parseName(lineStr, nameByUpper); tempName != "" {
			logs.Debug("find message line %s, %s", lineStr, tempName)
			realName = tempName
			reqBodyLineIndex = 1
			continue
		}

		if reqBodyLineIndex > 0 {
			if strings.Contains(lineStr, ",req,") {
				logs.Debug("check message %s", lineStr)
				keyType, keyValue := parseKey(lineStr)
				keys = append(keys, KeyStruct{
					KeyType:  keyType,
					KeyValue: keyValue,
				})
			}
		}

		if strings.TrimSpace(lineStr) == "}" {
			break
		}
	}
	return ConfigInfo{
		ConfigName: realName,
		Keys:       keys,
	}
}

func parseName(line, str string) string {
	if strings.Contains(line, "type") && strings.Contains(line, "struct") {
		lineSplists := strings.Split(strings.TrimSpace(line), " ")
		validSeg := make([]string, 0)
		for _, seg := range lineSplists {
			if seg != "" && seg != " " {
				validSeg = append(validSeg, seg)
			}
		}
		if len(validSeg) < 3 {
			logs.Info("valid seg %v", validSeg)
			return ""
		} else {
			if strings.ToUpper(validSeg[1]) == strings.ToUpper(str) {
				return validSeg[1]
			}
		}
	}
	return ""
}

type KeyStruct struct {
	KeyType  string
	KeyValue string
}

func parseKey(line string) (string, string) {
	lineSplists := strings.Split(strings.TrimSpace(line), " ")
	validSeg := make([]string, 0)
	for _, seg := range lineSplists {
		if seg != "" && seg != " " {
			validSeg = append(validSeg, seg)
		}
	}
	return validSeg[1][1:], validSeg[0]
}

func genLoadAllStr(configList []ConfigInfo) string {
	loadStr := ""
	for _, config := range configList {
		lowerName := strings.ToLower(config.ConfigName)
		loadStr += fmt.Sprintf(loadOne, lowerName, config.ConfigName)
	}
	return fmt.Sprintf(loadAll, "%s", loadStr)
}

var loadAll string = `func loadAllDatas(rootPath, dataAbsPath string) {
	load := func(dfilepath string, loadfunc func(string)) {
		loadfunc(filepath.Join(rootPath, dataAbsPath, dfilepath))
		logs.Info("LoadGameData %s success", dfilepath)
	}

%s
}

`

var loadOne string = `	load("%s.data", load%sData)
`

func genAllConfig(configList []ConfigInfo) string {
	retStr := ""
	for _, config := range configList {
		retStr += genSpecificConfigCode(config.ConfigName, config.Keys)
		retStr += "\n"
	}
	return retStr
}

// 生成指定的配置文件
func genSpecificConfigCode(nameByUpper string, keys []KeyStruct) string {
	nameByLower := strings.ToLower(nameByUpper)
	varDefineStr := genDefine(nameByUpper, nameByLower, keys)
	loadfuncStr := genLoadFunc(nameByUpper, nameByLower, keys)
	getAllStr := genGetAll(nameByUpper, nameByLower)
	getByIdStr := genGetById(nameByUpper, nameByLower, keys)
	retStr := varDefineStr
	retStr += "\n"
	retStr += "\n"
	retStr += loadfuncStr
	retStr += "\n"
	retStr += "\n"
	retStr += getAllStr
	retStr += "\n"
	retStr += "\n"
	retStr += getByIdStr
	retStr += "\n"
	retStr += "\n"
	return retStr
}

var varIntDefine string = `var %ss []*ProtobufGen.%s`
var varStringDefine string = `var %ss []*ProtobufGen.%s
var %smap map[string]*ProtobufGen.%s`
var varMultiKeyDefine string = `var %ss []*ProtobufGen.%s`

func genDefine(nameByUpper, nameByLower string, keys []KeyStruct) string {
	if len(keys) == 1 {
		if keys[0].KeyType == "string" {
			return fmt.Sprintf(varStringDefine, nameByLower, nameByUpper, nameByLower, nameByUpper)
		} else {
			return fmt.Sprintf(varIntDefine, nameByLower, nameByUpper)
		}
	} else {
		return fmt.Sprintf(varMultiKeyDefine, nameByLower, nameByUpper)
	}
}

func genLoadFunc(nameByUpper, nameByLower string, keys []KeyStruct) string {
	if len(keys) == 1 {
		funcParmas := make([]interface{}, 4)
		funcParmas[0] = nameByUpper
		funcParmas[1] = nameByUpper
		funcParmas[2] = nameByLower
		funcParmas[3] = genLoadSort(nameByUpper, nameByLower, keys[0].KeyType, keys[0].KeyValue)
		return fmt.Sprintf(loadFunc, funcParmas...)
	} else {
		funcParmas := make([]interface{}, 4)
		funcParmas[0] = nameByUpper
		funcParmas[1] = nameByUpper
		funcParmas[2] = nameByLower
		funcParmas[3] = ""
		return fmt.Sprintf(loadFunc, funcParmas...)
	}
}

func genLoadSort(nameByUpper, nameByLower string, keyType, keyValue string) string {
	if keyType == "string" {
		funcParmas := make([]interface{}, 6)
		funcParmas[0] = nameByLower
		funcParmas[1] = nameByLower
		funcParmas[2] = nameByUpper
		funcParmas[3] = nameByLower
		funcParmas[4] = nameByLower
		funcParmas[5] = keyValue
		return fmt.Sprintf(loadMap, funcParmas...)
	} else {
		funcParmas := make([]interface{}, 6)
		funcParmas[0] = nameByLower
		funcParmas[1] = nameByLower
		funcParmas[2] = nameByLower
		funcParmas[3] = keyValue
		funcParmas[4] = nameByLower
		funcParmas[5] = keyValue
		return fmt.Sprintf(loadSort, funcParmas...)
	}
}

var loadFunc string = `func load%sData(filePath string) {
	buffer, err := loadBin(filePath)
	panicIfError(err)

	ar := new(ProtobufGen.%s_ARRAY)
	panicIfError(proto.Unmarshal(buffer, ar))

	%ss = ar.GetItems()
	%s
}`

var loadSort string = `	if len(%ss) > 100 {
		sort.Slice(%ss, func(i, j int) bool {
			return %ss[i].Get%s() < %ss[j].Get%s()
		})
	}`

var loadMap string = `	if len(%ss) > 100 {
		%smap = make(map[string]*ProtobufGen.%s, len(ar.GetItems()))
		for _, item := range %ss {
			%smap[item.Get%s()] = item
		}
	}`

func genGetAll(nameByUpper, nameByLower string) string {
	return fmt.Sprintf(getAll, nameByUpper, nameByUpper, nameByLower)
}

var getAll string = `func GetAll%sS() []*ProtobufGen.%s {
	return %ss
}`

func genGetById(nameByUpper, nameByLower string, keys []KeyStruct) string {
	if len(keys) == 1 {
		keyType := keys[0].KeyType
		keyValue := keys[0].KeyValue
		if keyType == "string" {
			funcParmas := make([]interface{}, 9)
			funcParmas[0] = nameByUpper
			funcParmas[1] = keyValue
			funcParmas[2] = nameByUpper
			funcParmas[3] = nameByLower
			funcParmas[4] = nameByLower
			funcParmas[5] = keyValue
			funcParmas[6] = nameByLower
			funcParmas[7] = keyValue
			funcParmas[8] = keyValue
			return fmt.Sprintf(getByStringId, funcParmas...)
		} else {
			funcParmas := make([]interface{}, 16)
			funcParmas[0] = nameByUpper
			funcParmas[1] = keyValue
			funcParmas[2] = keyType
			funcParmas[3] = nameByUpper
			funcParmas[4] = nameByLower
			funcParmas[5] = nameByLower
			funcParmas[6] = nameByLower
			funcParmas[7] = keyValue
			funcParmas[8] = keyValue
			funcParmas[9] = nameByLower
			funcParmas[10] = keyValue
			funcParmas[11] = keyValue
			funcParmas[12] = nameByLower
			funcParmas[13] = nameByLower
			funcParmas[14] = keyValue
			funcParmas[15] = keyValue
			return fmt.Sprintf(getByIntId, funcParmas...)
		}
	} else {
		funcParmas := make([]interface{}, 5)
		funcParmas[0] = nameByUpper
		params := ""
		for i, item := range keys {
			params += fmt.Sprintf("%s %s", item.KeyValue, item.KeyType)
			if i != len(keys)-1 {
				params += ","
			}
		}
		funcParmas[1] = params
		funcParmas[2] = nameByUpper
		funcParmas[3] = nameByLower
		conditions := ""
		for i, item := range keys {
			conditions += fmt.Sprintf(" item.Get%s() == %s", item.KeyValue, item.KeyValue)
			if i != len(keys)-1 {
				conditions += " &&"
			}
		}
		funcParmas[4] = conditions
		return fmt.Sprintf(getByMultiId, funcParmas...)
	}
}

var getByIntId string = `func Get%sById(%s %s) *ProtobufGen.%s {
	if len(%ss) > 100 {
		index := sort.Search(len(%ss), func(i int) bool {
			return %ss[i].Get%s() >= %s
		})
		if %ss[index].Get%s() != %s {
			return nil
		} else {
			return %ss[index]
		}
	} else {
		for _, item := range %ss {
			if item.Get%s() == %s {
				return item
			}
		}
	}
	return nil
}`

var getByStringId string = `func Get%sById(%s string) *ProtobufGen.%s {
	if len(%ss) > 100 {
		return %smap[%s]
	} else {
		for _, item := range %ss {
			if item.Get%s() == %s {
				return item
			}
		}
	}
	return nil
}`

var getByMultiId string = `func Get%sById(%s) *ProtobufGen.%s {
	for _, item := range %ss {
		if %s {
			return item
		}
	}
	return nil
}`

func loadAndGenDatas(rootPath, dataAbsPath string, outDir string) {
	load := func(dfilepath string, loadfunc func(string, string)) {
		loadfunc(filepath.Join(rootPath, dataAbsPath, dfilepath), outDir)
		logs.Info("LoadGameData %s success", dfilepath)
	}

	load("pkconst.data", loadAndGenPKConstData)
}

func loadAndGenPKConstData(filePath string, outDir string) {
	buffer, err := loadBin(filePath)
	panicIfError(err)

	ar := new(ProtobufGen.PKConst_ARRAY)
	panicIfError(proto.Unmarshal(buffer, ar))

	vars := ""
	gets := ""
	for _, item := range ar.Items {
		vars += fmt.Sprintf("var %s %s", item.GetConstID(), getVarType(item))
		vars += "\n"
		gets += fmt.Sprintf(`	%s = GetPKConstById("%s").GetValue%s()`, item.GetConstID(), item.GetConstID(), getFuncType(item))
		gets += "\n"
	}
	lastCodes := fmt.Sprintf(constCode, vars, gets)
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(lastCodes)
	ioutil.WriteFile(outDir+"gamedata_const.go", buf.Bytes(), 0666)

}

func getVarType(item *ProtobufGen.PKConst) string {
	if item.ValueInt != nil {
		return "int32"
	}
	if item.ValueFloat != nil {
		return "float32"
	}
	if item.ValueString != nil {
		return "string"
	}
	return "string"
}

func getFuncType(item *ProtobufGen.PKConst) string {
	if item.ValueInt != nil {
		return "Int"
	}
	if item.ValueFloat != nil {
		return "Float"
	}
	if item.ValueString != nil {
		return "String"
	}
	return "String"
}

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

func panicIfError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func GetDataPath() string {
	workPath, _ := os.Getwd()
	workPath, _ = filepath.Abs(workPath)
	// initialize default configurations
	AppPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	logs.Debug("data path work=%s, app=%s", workPath, AppPath)
	appConfigPath := AppPath
	if workPath != AppPath {
		if utils.FileExists(appConfigPath) {
			os.Chdir(AppPath)
		} else {
			appConfigPath = workPath
		}
	}
	return appConfigPath
}

var constCode string = `package gamedata

// 自动生成, 不要修改

%s

func constPostLoad() {
%s
}
`
