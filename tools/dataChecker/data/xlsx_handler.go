package data

import (
	"fmt"
	"github.com/tealeg/xlsx"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"vcs.taiyouxi.net/tools/dataChecker/utils"
)

func GetDataFile(fName string) *xlsx.File {
	fullName := filepath.Join(cfg.Dir.DataProjectDir, fName+".xlsx")

	f, err := xlsx.OpenFile(fullName)
	if err != nil {
		panic(err.Error())
	}
	return f
}

/*
// 多语言文件是xml在工程目录，真悲伤
func GetIDSFile() *xlsx.File {
	fullName := filepath.Join(cfg.Dir.ClientProjectDir, "Data/Localization/TextRes/Languages.xml")

	f, err := xlsx.OpenFile(fullName)
	if err != nil {
		panic(err.Error())
	}

	return f
}
*/

// dumpRow 读取表格内容，如果遇到半角逗号自动切割
func (xd *XlsxData) dumpRow(sheet *xlsx.Sheet, colName string) {
	for col := 0; col < sheet.MaxCol; col++ {
		if sheet.Cell(cfg.Sheet.ColNameIdx, col).Value == colName {
			// 从cfg.Sheet.DataStartCol配置的开始行读数据
			for row := cfg.Sheet.DataStartRow; row < sheet.MaxRow; row++ {
				if sheet.Cell(row, col).Value != "" {
					content := sheet.Cell(row, col).Value
					// 处理表格内的逗号分隔符
					if strings.Contains(content, ",") {
						contents := strings.Split(content, ",")
						for _, c := range contents {
							xd.Data[c] += 1
						}
					} else {
						xd.Data[content] += 1
					}
				}
			}
		}
	}
}

// CheckStruct 根据表头初始化数据
func (repeat *Repeat) CheckStruct(sheet *xlsx.Sheet) {
	for col := 0; col < sheet.MaxCol; col++ {
		if sheet.Cell(cfg.Sheet.FieldType, col).Value == "repeated" {
			repeat.StartCol = col

			// 将dataType行的内容转为int
			maxRNum, err := strconv.Atoi(sheet.Cell(cfg.Sheet.DataTypeRow, col).Value)
			if err == nil {
				repeat.MaxRNum = maxRNum
			} else {
				repeat.err = err
				return
			}

			// 如果存在optional_struct
			if sheet.Cell(cfg.Sheet.FieldType, col+1).Value == "optional_struct" {
				repeat.HasOptStruct = true

				rUnitWidth, err := strconv.Atoi(sheet.Cell(cfg.Sheet.DataTypeRow, col+1).Value)
				if err == nil {
					repeat.RUnitWidth = rUnitWidth
				} else {
					repeat.err = err
					return
				}
			}

			// 储存Title中有默认值的列
			for col2 := repeat.StartCol; col2 < sheet.MaxCol; col2++ {
				val := sheet.Cell(cfg.Sheet.ColNameIdx, col2).Value
				if strings.Contains(val, "=") {
					parts := strings.Split(val, "=")
					if parts[1] != "\"\"" {
						repeat.DefVals[col2] = parts[0]
					}
				}
			}

			// 每张Sheet只有一个repeat, 找到了就跑
			break
		}
	}

	/*
	if repeat.MaxRNum == 0 {
		repeat.err = fmt.Errorf("No repeat columns in %s!", sheet.Name)
	}
	*/

}

// CheckData 核对repeat列数字是否为整数，是否与内容数量一致
func (repeat *Repeat) CheckData(sheet *xlsx.Sheet) {
	for row := cfg.Sheet.DataStartRow; row < sheet.MaxRow; row++ {
		rValue := sheet.Cell(row, repeat.StartCol).Value
		ExpectedRVal, err := strconv.Atoi(rValue)

		// 非整数直接报错
		if err != nil {
			if rValue == "" {
				ExpectedRVal = 0
			} else {
				repeat.ErrLines[row] = "非整数"
				continue
			}
		}

		// 超出范围也直接报错
		if ExpectedRVal > repeat.MaxRNum {
			repeat.ErrLines[row] = fmt.Sprintf("%d超出最大值%d", ExpectedRVal, repeat.MaxRNum)
			continue
		}

		// 如果没有optional_struct, 从数据下一列开始，间隔1
		col := repeat.StartCol + 1
		step := 1
		if repeat.RUnitWidth != 0 {
			step += repeat.RUnitWidth
		}
		count := 0

		// 防止越界
		maxCol := col + repeat.MaxRNum*step
		if maxCol > sheet.MaxCol {
			maxCol = sheet.MaxCol
		}

		if !repeat.HasOptStruct {
			for ; col < maxCol; col++ {
				_, hasDefVal := repeat.DefVals[col]
				if sheet.Cell(row, col).Value != "" || hasDefVal {
					count++
				}
			}
		} else {
			// 如果有，需要把自身占用的col算进去
			for ; col < maxCol; col += step {
				// 遍历Unit
				for lessStep := 1; lessStep < step; lessStep++ {
					curCol := col + lessStep
					//_, hasDefVal := repeat.DefVals[curCol]

					// 内容不为""，或有default值时，认为这个格子非空 fixed: 单元格为空时不计算default
					//if sheet.Cell(row, curCol).Value != "" || hasDefVal {
					if sheet.Cell(row, curCol).Value != "" {
						count++
						break
					}
				}
			}
		}

		if count != ExpectedRVal {
			repeat.ErrLines[row] = fmt.Sprintf("Repeat数值%d, 实际个数%d", ExpectedRVal, count)
		}
	}
}

// dumpRepeat 根据表格内容生成Repeat
func (xd *XlsxData) dumpRepeat(sheet *xlsx.Sheet) {
	repeat := *NewRepeat()
	repeat.CheckStruct(sheet)

	if repeat.err == nil && repeat.MaxRNum >0 {
		repeat.CheckData(sheet)
	}

	xd.Repeats[sheet.Name] = repeat
}

func (xd *XlsxData) getColData(sheetName string, colNames []string) {
	/*
		如果Sheet为空，遍历所有sheet
		cfg.SheetStartNum是配置的开始表序号
		一般从1开始，因为0是说明表
		如果没有说明表，就去干掉那个策划
	*/
	checkRepeat := false
	if len(colNames) == 1 && colNames[0] == "" {
		checkRepeat = true
	}

	if sheetName == "" {
		sheets := xd.File.Sheets
		for i := cfg.Sheet.SheetStartNum; i < len(sheets); i++ {
			if checkRepeat {
				// 读取Repeat
				xd.dumpRepeat(sheets[i])
			} else {
				for _, colName := range colNames {
					xd.dumpRow(sheets[i], colName)
				}
			}
		}
	} else {
		if sheet, ok := xd.File.Sheet[sheetName]; ok {
			if checkRepeat {
				// 读取Repeat
				xd.dumpRepeat(sheet)
			} else {
				for _, colName := range colNames {
					xd.dumpRow(sheet, colName)
				}
			}
		} else {
			fmt.Println("No sheet named %s !", sheetName)
		}
	}
}

func (xd *XlsxData) ReadSource(xs *utils.XlsxSource) {
	xd.File = GetDataFile(xs.SourceName)
	xd.NumRange = xs.NumRange
	xd.MD5 = xs.MD5

	// 如果需要遍历，构造一个空的slice让for loop跑起来
	if len(xs.SheetNames) == 0 {
		xs.SheetNames = make([]string, 1)
	}

	for _, sheetName := range xs.SheetNames {
		xd.getColData(sheetName, xs.ColNames)
	}

	// 存入usedData
	usedData.XData[xd.MD5] = xd
}

// CheckInvalidSeparators 检查数据中是否有全角逗号、分号、冒号等非法分隔符
func (xd *XlsxData) CheckInvalidSeparators() ([]string, bool) {
	invalids := make([]string, 0)
	invalidSeparators := []string{"；", "，", "："}

	for d := range xd.Data {
		for _, c := range invalidSeparators {
			if strings.Contains(d, c) {
				invalids = append(invalids, d)
			}
		}
	}

	return invalids, len(invalids) == 0
}

// checkRepeat 检查表格中Repeat数量是否为整数，以及是否与后面表格一致
func (xd *XlsxData) CheckRepeat() ([]string, bool) {
	errLogs := make([]string, 0)

	for sheetName, repeat := range xd.Repeats {
		if repeat.err != nil {
			errLogs = append(errLogs, repeat.err.Error())
		}
		if len(repeat.ErrLines) != 0 {
			for lineNum, errLog := range repeat.ErrLines {
				output := fmt.Sprintf("%s %d行%s", sheetName, lineNum+1, errLog)
				errLogs = append(errLogs, output)
			}
		}
	}

	sort.Slice(errLogs, func(i, j int) bool { return errLogs[i] < errLogs[j] })

	return errLogs, len(errLogs) == 0
}

func (fd *FileData) ReadSource(fs *utils.FileSource) {
	fd.File = GetDataFile(fs.SourceName)
	fd.Path = fs.Path
	fd.ExtName = fs.ExtName
	fd.MD5 = fs.MD5

	// 如果Sheets为空，遍历所有sheet
	if len(fs.SheetNames) == 0 {
		fs.SheetNames = make([]string, 1)
	}

	for _, sheetName := range fs.SheetNames {
		fd.getColData(sheetName, fs.ColNames)
	}

	// 存入usedData
	usedData.FData[fd.MD5] = fd
}

// TryLoadXData尝试读取已经存在的数据，如果不存在，返回一个新的
func TryLoadXData(hash [16]byte) (*XlsxData, bool) {
	if _, ok := usedData.XData[hash]; ok {
		return usedData.XData[hash], true
	}

	return NewXlsxData(), false
}

// TryLoadXData尝试读取已经存在的数据，如果不存在，返回一个新的
func TryLoadFData(hash [16]byte) (*FileData, bool) {
	if _, ok := usedData.FData[hash]; ok {
		return usedData.FData[hash], true
	}

	return NewFileData(), false
}
