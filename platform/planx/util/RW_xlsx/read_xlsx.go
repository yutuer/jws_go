package RW_xlsx

import (
	"errors"
	"fmt"
	"vcs.taiyouxi.net/Godeps/_workspace/src/github.com/tealeg/xlsx"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//传入xlsx文件名及页签名称和字段名称，获取该页签上指定字段的值
//返回两个参数，分别为xlsx文件sheet页签field字段中的数据[]string和错误信息
func GetFieldDataFromXlsx(XLSX_LOC, sheet, field string) ([]string, error) {
	//读文件
	file, err := xlsx.OpenFile(XLSX_LOC)
	if err != nil {
		logs.Error("[cyt]open file: %v,error:%v", XLSX_LOC, err)
		return []string{}, err
	}
	//获取标签页中的内容
	sheet_data := file.Sheet[sheet]
	//获取field的索引index
	maxCol := sheet_data.MaxCol
	index := -1
	for i := 0; i < maxCol; i++ {
		colname, err := sheet_data.Rows[0].Cells[i].String()
		if err != nil {
			logs.Error("[cyt]find field %v error:%v", field, err)
			return []string{}, err
		}
		if colname == field {
			index = i
		}
	}
	if index == -1 {
		logs.Error("[cyt]cannot find field %v in xlsx file:%v,sheet:%v", field, XLSX_LOC, sheet)
		return []string{}, fmt.Errorf("no such field:%v", field)
	}
	maxRow := sheet_data.MaxRow
	results := make([]string, maxRow)
	//读取信息存入results中
	for i := 0; i < maxRow; i++ {
		row, err := sheet_data.Rows[i].Cells[index].String()
		if err != nil {
			logs.Error("[cyt]read rownum:%v cells:%v error:%v", i, index, err)
			return []string{}, err
		}
		results[i] = row
	}
	return results, nil
}

//传入xlsx文件名及页签名称，获取页面信息并返回
//返回两个参数，分别为xlsx文件sheet页签中的数据[][]string,
//如果有报错，返回error，其余返回值为零值
func GetDataFromXlsx(XLSX_LOC, sheet string) ([][]string, error) {
	//读文件
	file, err := xlsx.OpenFile(XLSX_LOC)
	if err != nil {
		logs.Error("[cyt]open file: %v,error:%v", XLSX_LOC, err)
		return [][]string{}, err
	}
	//获取标签页中的内容
	sheet_data := file.Sheet[sheet]
	maxRow := sheet_data.MaxRow
	maxCol := sheet_data.MaxCol
	results := make([][]string, maxRow)
	//读取信息存入results中
	for i := 0; i < maxRow; i++ {
		row := make([]string, maxCol)
		for j := 0; j < maxCol; j++ {
			row[j], err = sheet_data.Rows[i].Cells[j].String()
			//第i、j个数据有误，i、j从0开始计数
			if err != nil {
				logs.Error("[cyt]read file :%v, sheet :%v, "+
					"rownum :%v, colnum :%v, error :%v", XLSX_LOC, sheet, i, j, err)
				return [][]string{}, err
			}
		}
		results[i] = row
	}
	return results, nil
}

//根据传入的字段名查询该字段在信息中的索引
//找到则返回该索引且error为nil
//未查找到则返回-1,且error不为nil
func GetIndexFromField(data [][]string, field string) (int, error) {
	for i := 0; i < len(data[0]); i++ {
		if field == data[0][i] {
			return i, nil
		}
	}
	return -1, errors.New("cannot found field in data[0]")
}
