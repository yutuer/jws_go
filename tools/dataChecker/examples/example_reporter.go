package examples

import "vcs.taiyouxi.net/tools/dataChecker/utils"

func ReportSample() {
	reporter := utils.NewReporter()

	// Record 依次记录如下内容：Checklist行号，来源文件名，检查类型，错误类型，错误LOGs
	reporter.Record(11, "Item",
		utils.IS_IDS_EXIST,
		utils.DATA_UNEXPECTED,
		[]string{
			"第25行有错误1",
			"第33行有错误2",
			"第2555行错误n",
		})

	// 输出到 dataChecker/log 目录
	reporter.Report()
}
