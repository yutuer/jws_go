package main

import (
	"flag"

	"vcs.taiyouxi.net/tools/dataChecker/data"
	"vcs.taiyouxi.net/tools/dataChecker/gacha"
	"vcs.taiyouxi.net/tools/dataChecker/hot_activities"
	"vcs.taiyouxi.net/tools/dataChecker/loot"
)

func main() {
	modeFlag := flag.String("mode", "checklist",
		`The running mode of the tool: checklist/loc/loot/gacha/hotactivity
		checklist: 表格检查
		loc: 本地化检查
		hotactivity: 执行活动数据检查
		loot：生成所有关卡掉落结果
		gacha：生成所有Gacha掉落结果
		`,
	)
	languageFlage := flag.String("lang", "zh-Hans", "The language for use, eg. zh-Hans, ko")
	checklistFlag := flag.String("checklist", "Checklist.xlsx", "The filename of the checklist. e.g. Checklist.xlsx ")
	flag.Parse()

	switch *modeFlag {
	case "checklist":
		data.RunAllChecklist(checklistFlag)
	case "loc":
		data.TransHotActivities(*languageFlage)
	case "loot":
		loot.GetAllLevelLootInfo()
		loot.GetTeambossResult()
	case "gacha":
		gacha.GetAllNormalGachaLoot()
		gacha.GetGachaLootDetails(12) // 典韦限时神将
	case "hotactivity":
		hot_activities.CheckHotActivityData()
		hot_activities.Report()
	default:
		panic("Run mode error!!!")
	}
}
