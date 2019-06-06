package data

import (
	"bufio"
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func (idata *IDSData) SetLocal(local string) error {
	for _, l := range cfg.Local.Locals {
		if l == local {
			idata.Local = l
		}
	}
	err := errors.New("No local found!\n")

	return err
}

// TryLoadIDS尝试从usedData中读取数据
func TryLoadIDS(language string) (*IDSData, bool) {
	var hash [16]byte

	for _, l := range cfg.Local.Locals {
		if l == language {
			hash = md5.Sum([]byte(filepath.Join(cfg.Dir.LocalProjectDir, language+".txt")))
			if hashData, ok := usedData.IData[hash]; ok {
				return hashData, true
			}
		}
	}

	return NewIDSData(), false
}

// LoadLocalizationData读取IDS来源并进行查重，去掉等号以及两边的空格，其他保留
func (idata *IDSData) LoadLocalizationData() {
	fullFileName := filepath.Join(cfg.Dir.LocalProjectDir, idata.Local+".txt")
	f, err := os.Open(fullFileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	idata.MD5 = md5.Sum([]byte(fullFileName))

	dupChecker := make(map[string]int, 16384)

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()

		for i, c := range line {
			if c == rune('=') {
				ids := line[:i-1]
				dupChecker[ids] += 1
				// 如果等号后是空的
				if i == len(line)-1 {
					// 去掉等号前面的空格
					idata.Data[ids] = ""
				} else {
					// 去掉等号以及前后的空格
					idata.Data[ids] = line[i+2:]
				}
			}
		}
	}

	if err := s.Err(); err != nil {
		fmt.Println("reading standard input:", err)
	}

	for ids, count := range dupChecker {
		if count > 1 {
			idata.Duplicates = append(idata.Duplicates, ids)
		}
	}
	// 新建数据放入usedData
	usedData.IData[idata.MD5] = idata
}
