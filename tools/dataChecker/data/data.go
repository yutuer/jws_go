package data

import (
	"vcs.taiyouxi.net/tools/dataChecker/utils"

	"github.com/tealeg/xlsx"
)

var (
	cfg      utils.Config
	reporter *utils.Reporter
	usedData = &UsedData{
		XData: make(map[[16]byte]*XlsxData, 1024),
		IData: make(map[[16]byte]*IDSData, 16),
		FData: make(map[[16]byte]*FileData, 1024),
	}
)

func init() {
	// 主要是为了获取Project路径
	cfg.LoadConfig()
	reporter = utils.NewReporter()
}

type UsedData struct {
	XData map[[16]byte]*XlsxData
	IData map[[16]byte]*IDSData
	FData map[[16]byte]*FileData
}

type XlsxData struct {
	Data     map[string]int    // val用于查重
	Repeats  map[string]Repeat // [SheetName] Repeat
	NumRange utils.NumRange
	File     *xlsx.File
	DataType string
	MD5      [16]byte
}

type Repeat struct {
	err          error
	HasOptStruct bool           // 是否存在optional_struct
	StartCol     int            // repeat所在的列号
	DefVals      map[int]string // [列号]默认值
	MaxRNum      int            // Repeat最大次数设定值
	RUnitWidth   int            // Repeat单元宽度
	ErrLines     map[int]string // [行号] "数值为%s, 实际为%d"
}

type IDSData struct {
	MD5        [16]byte
	Data       map[string]string
	Duplicates []string
	Local      string // 语言缩写，譬如"en", "zh-Hans"
}

type FileData struct {
	XlsxData
	Path    string
	ExtName string
}

func NewXlsxData() *XlsxData {
	// 目前ID最多的是Item, 1926. 2048应该是比较安全的
	x := &XlsxData{
		Data:    make(map[string]int, 2048),
		Repeats: make(map[string]Repeat),
	}

	return x
}

func NewIDSData() *IDSData {
	// 目前IDS有10363行
	d := &IDSData{
		Local: cfg.Local.DefaultLocal,
		Data:  make(map[string]string, 16384),
	}

	return d
}

func NewFileData() *FileData {
	// 目前IDS有10363行
	d := &FileData{}
	d.Data = make(map[string]int, 2048)

	return d
}

func NewRepeat() *Repeat {
	r := &Repeat{
		DefVals:  make(map[int]string, 32),
		ErrLines: make(map[int]string, 32),
	}

	return r
}

func Finish() {
	reporter.Report()
}
