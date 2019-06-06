package data

import (
	"github.com/tealeg/xlsx"
	"path/filepath"
	"time"
	"vcs.taiyouxi.net/tools/dataChecker/utils"
)

func TransHotActivities(language string) string {
	// 获取IDS
	idata := NewIDSData()
	idata.SetLocal(language)
	idata.LoadLocalizationData()

	hact := GetDataFile("HotActivityTime")

	transItemID2ItemIDS(hact, idata)

	savedFileName := filepath.Join(utils.GetVCSRootPath(), "/tools/dataChecker/log",
		"HotActivitiesTime"+time.Now().Format("-20060102-150405")+".xlsx")

	hact.Save(savedFileName)

	return savedFileName
}

// getItemNameMap 返回 map[ID]IDS
func getItemNameMap() map[string]string {
	item := GetDataFile("Item")
	itemNameMap := make(map[string]string, 32768)
	colID, colIDS := -1, -1

	for i := cfg.Sheet.SheetStartNum; i < len(item.Sheets); i++ {
		sheet := item.Sheets[i]
		// 先定位ID和IDS的列号
		for col := 0; col < sheet.MaxCol; col++ {
			cell := sheet.Cell(cfg.Sheet.ColNameIdx, col).Value
			if cell == "ID" {
				colID = col
			} else if cell == "NameIDS" {
				colIDS = col
			}
		}
		// 如果都存在，搞之
		if colID != -1 && colIDS != -1 {
			for row := cfg.Sheet.DataStartRow; row < sheet.MaxRow; row++ {
				itemNameMap[sheet.Cell(row, colID).Value] = sheet.Cell(row, colIDS).Value
			}
		}
	}

	return itemNameMap
}

func transItemID2ItemIDS(f *xlsx.File, ids *IDSData) {
	// 获取ItemID -> ItemIDS的映射Map
	itemMap := getItemNameMap()

	// 设定HotActivitiesTime里需要替换的Item列名，哦，只有一个，算了……
	// hactItemColNames := []string {"ItemID"}

	for i := cfg.Sheet.SheetStartNum; i < len(f.Sheets); i++ {
		sheet := f.Sheets[i]
		for col := 0; col < sheet.MaxCol; col++ {
			colTitle := sheet.Cell(cfg.Sheet.ColNameIdx, col).Value
			if colTitle == "ItemID" || colTitle == "FinalRewardID" || colTitle == "ShowItemID" {
				for row := cfg.Sheet.DataStartRow; row < sheet.MaxRow; row++ {
					cell := sheet.Cell(row, col)
					if name, ok := ids.Data[itemMap[cell.Value]]; ok {
						// 替换掉，没有就留着
						cell.SetString(name)
					}
				}
			}
		}
	}
}
