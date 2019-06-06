package data

import (
	"os"
	"path/filepath"
	"strconv"

	"vcs.taiyouxi.net/tools/dataChecker/utils"
)

func RunAllChecklist(checklistFlag *string) {
	checklists := make(chan *utils.CheckList)
	done := make(chan bool)

	go func() {
		for {
			cl, ok := <-checklists
			if ok {
				RunChecklist(cl)
			} else {
				done <- true
				return
			}
		}
	}()

	// 发送Checklist
	utils.GenChecklist(*checklistFlag, checklists)

	<-done

	// 生成报告
	Finish()
}

// RunChecklist执行该条用例，并记录异常
func RunChecklist(cl *utils.CheckList) {
	var (
		idsData  *IDSData
		source   *XlsxData
		target   *XlsxData
		fileData *FileData
		ok       bool
	)

	if cl.IDSSource != "" {
		idsData, ok = TryLoadIDS(cl.IDSSource)
		if !ok {
			idsData.SetLocal(cl.IDSSource)
			idsData.LoadLocalizationData()
		}
	}

	if cl.DataSource != nil {
		source, ok = TryLoadXData(cl.DataSource.MD5)
		if !ok {
			source.ReadSource(cl.DataSource)
			warns, ok := source.CheckInvalidSeparators()
			if !ok {
				reporter.Record(cl.Index+1, cl.DataSource.SourceName, 0, 3, warns)
			}
		}
	}

	if cl.DataTarget != nil {
		target, ok = TryLoadXData(cl.DataTarget.MD5)
		if !ok {
			target.ReadSource(cl.DataTarget)
			warns, ok := target.CheckInvalidSeparators()
			if !ok {
				reporter.Record(cl.Index+1, cl.DataTarget.SourceName, 0, 3, warns)
			}
		}
	}

	if cl.FileSource != nil {
		fileData, ok = TryLoadFData(cl.FileSource.MD5)
		if !ok {
			fileData.ReadSource(cl.FileSource)
		}
	}

	switch cl.CheckType {
	case utils.IS_ID_EXIST:
		if source != nil && target != nil {
			if r, ok := VerifyContain(source, target); !ok {
				reporter.Record(cl.Index+1, cl.DataTarget.SourceName, cl.CheckType, 0, r)
			}
		} else {
			reporter.Record(cl.Index+1, "", cl.CheckType, 1, nil)
		}
	case utils.IS_DUP:
		if idsData != nil {
			r := idsData.Duplicates
			if len(r) != 0 {
				reporter.Record(cl.Index+1, cl.IDSSource, cl.CheckType, 0, r)
			}
		} else if source != nil {
			if r, ok := VerifyDuplicate(source); !ok {
				reporter.Record(cl.Index+1, cl.DataSource.SourceName, cl.CheckType, 0, r)
			}
		} else {
			reporter.Record(cl.Index+1, "", cl.CheckType, 1, nil)
		}
	case utils.IS_IN_RANGE:
		if target != nil && target.NumRange.Left <= target.NumRange.Right {
			if r, ok := VerifyInRange(target); !ok {
				reporter.Record(cl.Index+1, cl.DataTarget.SourceName, cl.CheckType, 0, r)
			}
		} else {
			reporter.Record(cl.Index+1, "", cl.CheckType, 1, nil)
		}
	case utils.IS_FILE_EXIST:
		if fileData != nil {
			if r, ok := VerifyFileExist(fileData); !ok {
				reporter.Record(cl.Index+1, cl.FileSource.SourceName, cl.CheckType, 0, r)
			}
		} else {
			reporter.Record(cl.Index+1, "", cl.CheckType, 1, nil)
		}
	case utils.IS_EMPTY:
		return
	case utils.IS_TIME_CORRECT:
		return
	case utils.IS_IDS_EXIST:
		if idsData != nil && target != nil {
			if r, ok := VerifyIDS(idsData, target); !ok {
				reporter.Record(cl.Index+1, cl.DataTarget.SourceName, cl.CheckType, 0, r)
			}
		} else {
			reporter.Record(cl.Index+1, "", cl.CheckType, 1, nil)
		}
	case utils.IS_REPEAT_CORRECT:
		if target != nil {
			if r, ok := target.CheckRepeat(); !ok {
				reporter.Record(cl.Index+1, cl.DataTarget.SourceName, cl.CheckType, 0, r)
			}
		} else {
			reporter.Record(cl.Index+1, "", cl.CheckType, 1, nil)
		}
	default:
		return
		reporter.Record(cl.Index+1, "", 0, 2, nil)
	}
}

// VerifyDuplicate返回重复的ID, 以及结果(pass = true, fail = false)
func VerifyDuplicate(src *XlsxData) ([]string, bool) {
	wrongIDs := []string{}

	for srcID, count := range src.Data {
		if count > 1 {
			wrongIDs = append(wrongIDs, srcID)
		}
	}

	return wrongIDs, len(wrongIDs) == 0
}

//  VerifyContain返回所有不在来源中的ID, 以及结果(pass = true, fail = false)
func VerifyContain(src *XlsxData, tar *XlsxData) ([]string, bool) {
	wrongIDs := []string{}

	for tarID := range tar.Data {
		if _, ok := src.Data[tarID]; !ok {
			wrongIDs = append(wrongIDs, tarID)
		}
	}

	return wrongIDs, len(wrongIDs) == 0
}

// VerifyInRange 返回所有范围超出预设区间的ID
func VerifyInRange(tar *XlsxData) ([]string, bool) {
	wrongIDs := []string{}

	for data := range tar.Data {
		value, err := strconv.ParseFloat(data, 64)
		if err != nil {
			wrongIDs = append(wrongIDs, data)
		}

		isMatch := false
		left, right := tar.NumRange.Left, tar.NumRange.Right

		// 判断Value是否落在区间
		if value > left && value < right {
			isMatch = true
		} else if tar.NumRange.ContainsLeft && value == left {
			isMatch = true
		} else if tar.NumRange.ContainsRight && value == right {
			isMatch = true
		}

		if !isMatch {
			wrongIDs = append(wrongIDs, data)
		}
	}

	return wrongIDs, len(wrongIDs) == 0
}

// VerifyIDS返回所有不存的IDS
func VerifyIDS(id *IDSData, target *XlsxData) ([]string, bool) {
	wrongIDSs := []string{}

	for tarID := range target.Data {
		if _, ok := id.Data[tarID]; !ok {
			wrongIDSs = append(wrongIDSs, tarID)
		}
	}

	return wrongIDSs, len(wrongIDSs) == 0
}

// VerifyFileExist返回所有不存在的文件
func VerifyFileExist(fd *FileData) ([]string, bool) {
	wrongFiles := []string{}

	// teamcity 模式下
	if cfg.Dir.RunOnTeamCity == true {
		for fName := range fd.Data {
			fileName := filepath.Join(fd.Path, fName+fd.ExtName)
			// 因为xlsx里，有时会填带扩展名的值，所以这里只能先拼起来再拆开处理
			isExist := false
			ext := filepath.Ext(fileName)

			if nameSlice, ok := cfg.AllClientFiles[ext]; ok {
				for _, name := range nameSlice {
					if name == fileName {
						isExist = true
						continue
					}
				}
			}

			if !isExist {
				wrongFiles = append(wrongFiles, fName)
			}
		}
	} else {
		// 本地模式
		for fName := range fd.Data {
			//  D:/Projects/the-last-of-shuang/trunk,  Assets/UI/bundle/Comic/, Name, .jpg
			fullFileName := filepath.Join(cfg.Dir.ClientProjectDir, fd.Path, fName+fd.ExtName)
			// fmt.Println("Name: ", fullFileName)
			if _, err := os.Stat(fullFileName); os.IsNotExist(err) {
				wrongFiles = append(wrongFiles, fName)
			}
		}
	}

	return wrongFiles, len(wrongFiles) == 0
}
