package data

import (
	"crypto/md5"
	"github.com/stretchr/testify/assert"
	"testing"
	"vcs.taiyouxi.net/tools/dataChecker/utils"
)

// DebugGetXlsxSourceBuff 获得一个没有重复ID的XlsxSource
func DebugGetXlsxSourceBuff() *utils.XlsxSource {
	xs := &utils.XlsxSource{}
	xs.SetSourceVal("Buff")
	xs.SetSheetVal([]string{"ITEMBUFF"}, []string{"ID"})
	xs.MD5 = md5.Sum([]byte("DebugGetXlsxDataBuff"))

	return xs
}

// DebugGetXlsxDataBuff 获得一个没有重复ID的XlsxData
func DebugGetXlsxDataBuff() *XlsxData {
	xd := NewXlsxData()
	xd.ReadSource(DebugGetXlsxSourceBuff())

	return xd
}

// DebugGetXlsxSourceItem 获得一个未指定Sheets的XlsxSource
func DebugGetXlsxSourceItem() *utils.XlsxSource {
	xs := &utils.XlsxSource{}
	xs.SetSourceVal("Item")
	xs.SetSheetVal(nil, []string{"ID"})
	xs.MD5 = md5.Sum([]byte("DebugGetXlsxDataItem"))

	return xs
}

// DebugGetXlsxDataItem 获得一个未指定Sheets的XlsxData
func DebugGetXlsxDataItem() *XlsxData {
	xd := NewXlsxData()
	xd.ReadSource(DebugGetXlsxSourceItem())

	return xd
}

// DebugGetXlsxSourceHATItem 获得另一个包含Item的XlsxSource
func DebugGetXlsxSourceHATItem() *utils.XlsxSource {
	xt := &utils.XlsxSource{}
	xt.SetSourceVal("HotActivityTime")
	xt.SetSheetVal([]string{"REDPACKET"}, []string{"ItemID"})
	xt.MD5 = md5.Sum([]byte("DebugGetXlsxDataHATItem"))

	return xt
}

// DebugGetXlsxDataHATItem 获得另一个包含Item的XlsxData
func DebugGetXlsxDataHATItem() *XlsxData {
	xdt := NewXlsxData()
	xdt.ReadSource(DebugGetXlsxSourceHATItem())

	return xdt
}

// DebugGetXlsxSourceLevelInfo 获得一个有重复ID的xlsxSource
func DebugGetXlsxSourceLevelInfo() *utils.XlsxSource {
	xs2 := &utils.XlsxSource{}
	xs2.SetSourceVal("LevelInfo")
	xs2.SetSheetVal([]string{"LEVEL_INFO"}, []string{"DropItemID=\"\""})
	xs2.MD5 = md5.Sum([]byte("DebugGetXlsxDataLevelInfo"))

	return xs2
}

// DebugGetXlsxDataLevelInfo 获得一个有重复ID的xlsxData
func DebugGetXlsxDataLevelInfo() *XlsxData {
	xd2 := NewXlsxData()
	xd2.ReadSource(DebugGetXlsxSourceLevelInfo())

	return xd2
}

// DebugGetXlsxSourceHATIDS 获得一个IDS的xlsxSource
func DebugGetXlsxSourceHATIDS() *utils.XlsxSource {
	xt := &utils.XlsxSource{}
	xt.SetSourceVal("HotActivityTime")
	xt.SetSheetVal(nil, []string{"GoodsName", "TabIDS", "DesIDS", "TitleIDS", "TabIDS2", "DesIDS2", "TitleIDS2"})
	xt.MD5 = md5.Sum([]byte("DebugGetXlsxDataHATIDS"))

	return xt
}

// DebugGetXlsxDataHATIDS 获得一个IDS的xlsxData
func DebugGetXlsxDataHATIDS() *XlsxData {
	xdt := NewXlsxData()
	xdt.ReadSource(DebugGetXlsxSourceHATIDS())

	return xdt
}

// DebugGetFileSourceHAT 获得一个FileSource
func DebugGetFileSourceHAT() *utils.FileSource {
	fs := &utils.FileSource{}
	fs.SetSourceVal("HotActivityTime")
	fs.SetSheetVal(nil, []string{"Icon=\"\""})
	//fs.SetFileVal("", "")
	fs.MD5 = md5.Sum([]byte("DebugGetFileDataHAT"))

	return fs
}

// DebugGetFileDataHAT 获得一个FileData
func DebugGetFileDataHAT() *FileData {
	fd := NewFileData()
	fd.ReadSource(DebugGetFileSourceHAT())

	return fd
}

func TestVerifyDuplicate(t *testing.T) {
	xd := DebugGetXlsxDataBuff()

	r1, ok := VerifyDuplicate(xd)
	assert.Empty(t, r1)
	assert.True(t, ok)

	xd2 := DebugGetXlsxDataLevelInfo()

	r2, ok := VerifyDuplicate(xd2)
	assert.NotEmpty(t, r2)
	assert.False(t, ok)
}

func TestVerifyContain(t *testing.T) {
	// 源
	xd := DebugGetXlsxDataItem()

	// 目标
	xdt := DebugGetXlsxDataHATItem()

	// 复制品
	xd_dup := NewXlsxData()
	*xd_dup = *xd

	// 开跑
	r1, ok := VerifyContain(xd, xdt)
	//t.Logf(":", r1)
	assert.Empty(t, r1)
	assert.True(t, ok)

	r2, ok := VerifyContain(xd, xd_dup)
	assert.Empty(t, r2)
	assert.True(t, ok)
}

func TestVerifyIDS(t *testing.T) {
	// 源
	d := NewIDSData()
	d.LoadLocalizationData()

	// 目标
	xdt := DebugGetXlsxDataHATIDS()

	// 开跑
	r, ok := VerifyIDS(d, xdt)
	assert.Empty(t, r)
	assert.True(t, ok)
}

func TestVerifyFileExist(t *testing.T) {
	// 源
	fd := DebugGetFileDataHAT()

	r, ok := VerifyFileExist(fd)
	assert.NotEmpty(t, r)
	assert.False(t, ok)
}

// TestRunChecklist0 验证IS_ID_EXIST时执行用例的各种情况
func TestRunChecklist0(t *testing.T) {
	cl := &utils.CheckList{}
	rIdx := 0

	// 验证IS_ID_EXIST, 没有source和target
	cl.CheckType = utils.IS_ID_EXIST
	RunChecklist(cl)
	assert.Equal(t, 1, reporter.Unexceptions[rIdx].ErrorType)
	rIdx++

	// 只有source
	cl.DataSource = DebugGetXlsxSourceItem()
	RunChecklist(cl)
	assert.Equal(t, 1, reporter.Unexceptions[rIdx].ErrorType)
	rIdx++

	// 只有target
	cl.DataSource = nil
	cl.DataTarget = DebugGetXlsxSourceHATItem()
	RunChecklist(cl)
	assert.Equal(t, 1, reporter.Unexceptions[rIdx].ErrorType)
	preErrorCount := len(reporter.Unexceptions)
	rIdx++

	// 都有，不出错，reporter不增加
	cl.DataSource = DebugGetXlsxSourceItem()
	RunChecklist(cl)
	assert.Equal(t, preErrorCount, len(reporter.Unexceptions))

	// 手动改出一个错误使其报错
	d := usedData.XData[cl.DataTarget.MD5]
	d.Data["NotExistKey"] = 1
	RunChecklist(cl)
	// 其他已经在reporter_test中验证过了
	assert.Equal(t, 0, reporter.Unexceptions[rIdx].ErrorType)
}

// TestRunChecklist1 验证IS_DUP的运行情况
func TestRunChecklist1(t *testing.T) {}
