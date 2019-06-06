package main

import (
	"time"
	"vcs.taiyouxi.net/platform/planx/util/RW_xlsx"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gainer_beholder/tools"
)

const (
	xlsxName  string = "KoSuccessOrders.xlsx"
	sheetName string = "orders"
)

func main() {
	defer logs.Close()
	xlsx_data, err := RW_xlsx.GetDataFromXlsx(xlsxName, sheetName)
	if err != nil {
		logs.Error("[cyt]get data from xlsx error:%v", err)
		return
	}
	index_pay_description, err := RW_xlsx.GetIndexFromField(xlsx_data, "pay_description")
	if err != nil {
		logs.Error("[cyt]get field index from data error:%v", err)
		return
	}

	//获取待比较的玩家信息
	tools.GetStayForJudges(xlsx_data, index_pay_description)
	//将待比较的玩家信息与表格中信息比对,比对后出现问题的单号记录在tools.GainnerOrder中
	tools.AffirmAndRecord()
	logs.Trace("%v条记录异常", len(tools.GainnerOrder))
	//将出现问题的单号对应数据存储在gainner_data中
	var gainner_data [][]string = make([][]string, 1)
	gainner_data[0] = xlsx_data[0]
	for _, v := range tools.GainnerOrder {
		gainner_data = append(gainner_data, xlsx_data[v])
	}
	//将gainner_data生成一个新的excel存储起来。
	RW_xlsx.WriteToXlsxFile(gainner_data, "gainer_"+
		time.Now().Format("02-01-2006_15:04:05_PM")+".xlsx", "gainner_orders")
}
