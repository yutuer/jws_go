package utils

import (
	"io"
	"os"
)

var (
	cfg Config
)

func init() {
	cfg.LoadConfig()
	DebugLoadGamedataAll()
}

// 定义错误码
const (
	NONE = iota
	DATA_UNEXPECTED
	DATA_SOURCE_MISSING
	DATA_CHECK_TYPE_UNKNOWN
	DATA_CONTAINS_INVALID_SEPARATOR
	HOT_ACTIVITY_TIME_INVALID
	HOT_ACTIVITY_TIME_RANGE_INVALID
	HOT_ACTIVITY_SEVER_ID_INVALID
	ERROR_CODE_TYPE_COUNT
)

// 和Checklist检查类型一一对应
const (
	IS_ID_EXIST = iota
	IS_DUP
	IS_IN_RANGE
	IS_FILE_EXIST
	IS_EMPTY
	IS_TIME_CORRECT
	IS_IDS_EXIST
	IS_REPEAT_CORRECT
	IS_SERVER_GROUP_OVERLAP
	IS_TIME_RANGE_CORRECT
	LOOT
	GACHA
	TEAMBOSS
	CHECK_TYPE_COUNT
)

// 和Checklist表每一列序号一一对应
const (
	COL_TAR_FILE = iota
	COL_TAR_SHEETS
	COL_TAR_COLS
	COL_CHECK_TYPE
	COL_IDS
	COL_SRC_FILE
	COL_SRC_SHEETS
	COL_SRC_COLS
	COL_PATH
	COL_EXT
	COL_RANGE
	MAX_COL
)

// loadBin2Buff 将.data文件读取为 []byte
func LoadBin2Buff(binFilename string) ([]byte, error) {
	file, err := os.Open(binFilename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, fi.Size())
	_, err = io.ReadFull(file, buffer)
	if err != nil {
		return buffer, err
	}

	return buffer, nil
}

// WriteBuff2Bin 将buff写入文件
func WriteBuff2Bin(binFilename string, buff []byte) (err error) {
	f, err := os.Create(binFilename)
	if err != nil {
		return
	}

	defer f.Close()

	_, err = f.Write(buff)
	if err != nil {
		return
	}

	err = f.Sync()

	return
}
