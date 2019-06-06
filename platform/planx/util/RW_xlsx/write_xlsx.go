package RW_xlsx

import (
	"fmt"
	"vcs.taiyouxi.net/Godeps/_workspace/src/github.com/tealeg/xlsx"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//向文件XLSX_Name的sheet_name页签写入数据data
func WriteToXlsxFile(data [][]string, XLSX_Name, sheet_name string) {
	file := xlsx.NewFile()
	sheet, err := file.AddSheet(sheet_name)
	if err != nil {
		logs.Error("[cyt]AddSheet :%v fail ,error data :%v")
		return
	}
	for i := range data {
		row := sheet.AddRow()
		for j := range data[i] {
			cell := row.AddCell()
			cell.Value = data[i][j]
		}
	}
	err = file.Save(XLSX_Name)
	if err != nil {
		fmt.Printf(err.Error())
	}
}
