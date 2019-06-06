package utils

import (
	"crypto/md5"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/tealeg/xlsx"
)

type SourceInfo struct {
	SourceName string   // 来源文件名
	MD5        [16]byte // 用文件名+表名+列名作为Hash源，保证唯一性
}

type SheetInfo struct {
	SheetNames []string // 表名
	ColNames   []string // 列名
}

type FileInfo struct {
	Path    string // 从Asset目录开始的路径
	ExtName string // 扩展名
}

type NumRange struct {
	Left          float64
	ContainsLeft  bool
	Right         float64
	ContainsRight bool
}

type XlsxSource struct {
	SourceInfo
	SheetInfo
	NumRange
}

type FileSource struct {
	SourceInfo
	SheetInfo
	FileInfo
}

func (x *SourceInfo) SetSourceVal(name string) {
	x.SourceName = name
}

func (x *SheetInfo) SetSheetVal(sheetNames []string, colNames []string) {
	x.SheetNames = sheetNames
	x.ColNames = colNames
}

func (x *FileInfo) SetFileVal(path string, extName string) {
	x.Path = path
	x.ExtName = extName
}

func (x *SourceInfo) SetMD5(md5 [16]byte) {
	x.MD5 = md5
}

type CheckList struct {
	Index     int // 对应的行号
	CheckType int // 检查类型

	IDSSource  string      // 源：IDS
	DataSource *XlsxSource // 源：Xlsx
	DataTarget *XlsxSource // 目标：Xlsx
	FileSource *FileSource // 源：文件
}

// 添加为来源文件
func (c *CheckList) SetFileSource(fileSrc *FileSource) {
	c.FileSource = fileSrc
}

// 添加为目标文件
func (c *CheckList) SetDataTarget(dataTar *XlsxSource) {
	c.DataTarget = dataTar
}

func getCheckListFile(fName string) *xlsx.File {
	// 本地
	cl := SearchInConf(fName)

	// teamcity
	if cfg.Dir.RunOnTeamCity == true {
		cl = filepath.Join(cfg.Dir.DataProjectDir, fName)
	}

	f, err := xlsx.OpenFile(cl)
	if err != nil {
		panic(err.Error())
	}
	return f
}

// rowIdx 定位错误， row 为内容
func sendChecklistSource(f *xlsx.File, rowIdx chan<- int, row chan<- *xlsx.Row) {
	// 第一张表说明
	sheet := f.Sheets[1]

	// 正式数据始于R5
	for i := cfg.Sheet.DataStartRow; i < sheet.MaxRow; i++ {
		row <- sheet.Rows[i]
		rowIdx <- i
	}

	close(row)
}

func GenChecklist(fname string, cl chan<- *CheckList) {
	f := getCheckListFile(fname)

	rowIdx := make(chan int)
	row := make(chan *xlsx.Row)
	doneSend := make(chan bool)

	go sendChecklistSource(f, rowIdx, row)
	go sendChecklist(rowIdx, row, cl, doneSend)

	<-doneSend
}

func sendChecklist(rowIdx <-chan int, row <-chan *xlsx.Row, clOut chan<- *CheckList, done chan<- bool) {
	for {
		r, ok := <-row
		if ok {
			// 如果没有赋值，跳过此行
			if r.Cells[COL_CHECK_TYPE].Value == "" {
				<-rowIdx
				continue
			}

			checkType, err := strconv.Atoi(r.Cells[COL_CHECK_TYPE].Value)
			if err != nil {
				panic(err)
			}

			cl := &CheckList{Index: <-rowIdx, CheckType: checkType}

			// 检查文件存在性
			if checkType == IS_FILE_EXIST {
				cl.FileSource = row2FileSource(r)
				clOut <- cl
				continue
			}

			// 如果需要目标文件
			if r.Cells[COL_TAR_FILE].Value != "" {
				cl.DataTarget = row2XlsxTarget(r)
			}

			// 如果需要检查范围
			if checkType == IS_IN_RANGE {
				content := r.Cells[COL_RANGE].Value

				if strings.Count(content, ",") != 1 {
					panic("Invalid cell value!")
				}

				n := len(content)
				if content[0] == '[' {
					cl.DataTarget.NumRange.ContainsLeft = true
				}
				if content[n-1] == ']' {
					cl.DataTarget.NumRange.ContainsRight = true
				}

				values := strings.Split(content[1:n-1], ",")

				cl.DataTarget.NumRange.Left, err = strconv.ParseFloat(strings.TrimSpace(values[0]), 64)
				if err != nil {
					panic(err)
				}

				cl.DataTarget.NumRange.Right, err = strconv.ParseFloat(strings.TrimSpace(values[1]), 64)
				if err != nil {
					panic(err)
				}
			}

			// 来源为IDS
			if r.Cells[COL_IDS].Value != "" {
				cl.IDSSource = r.Cells[COL_IDS].Value
			}

			// 需要有来源
			if r.Cells[COL_SRC_FILE].Value != "" {
				cl.DataSource = row2XlsxSource(r)
			}

			clOut <- cl
		} else {
			close(clOut)
			done <- true
			return
		}
	}
}

// 单元格转Slice，排序保证Hash唯一性
func Cell2Slice(str string) []string {
	s := strings.Split(str, ",")
	for i, c := range s {
		s[i] = strings.TrimSpace(c)
	}
	sort.Strings(s)
	return s
}

func row2XlsxSource(row *xlsx.Row) *XlsxSource {
	xs := &XlsxSource{}

	file := row.Cells[COL_SRC_FILE].Value
	sheets := Cell2Slice(row.Cells[COL_SRC_SHEETS].Value)
	cols := Cell2Slice(row.Cells[COL_SRC_COLS].Value)

	xs.SetSourceVal(file)
	xs.SetSheetVal(sheets, cols)

	xs.SetMD5(md5.Sum([]byte(file + strings.Join(sheets, "") + strings.Join(cols, ""))))

	return xs
}

func row2FileSource(row *xlsx.Row) *FileSource {
	fs := &FileSource{}

	fs.SetSourceVal(row.Cells[COL_TAR_FILE].Value)
	fs.SetSheetVal(Cell2Slice(row.Cells[COL_TAR_SHEETS].Value), Cell2Slice(row.Cells[COL_TAR_COLS].Value))
	fs.SetFileVal(row.Cells[COL_PATH].Value, row.Cells[COL_EXT].Value)

	fs.SetMD5(md5.Sum([]byte(row.Cells[COL_TAR_FILE].Value + row.Cells[COL_TAR_SHEETS].Value + row.Cells[COL_TAR_COLS].Value)))

	return fs
}

func row2XlsxTarget(row *xlsx.Row) *XlsxSource {
	xs := &XlsxSource{}

	xs.SetSourceVal(row.Cells[COL_TAR_FILE].Value)
	xs.SetSheetVal(Cell2Slice(row.Cells[COL_TAR_SHEETS].Value), Cell2Slice(row.Cells[COL_TAR_COLS].Value))

	xs.SetMD5(md5.Sum([]byte(row.Cells[COL_TAR_FILE].Value + row.Cells[COL_TAR_SHEETS].Value + row.Cells[COL_TAR_COLS].Value)))

	return xs
}
