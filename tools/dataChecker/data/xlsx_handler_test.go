package data

import (
	"crypto/md5"
	"github.com/stretchr/testify/assert"
	"testing"
	"vcs.taiyouxi.net/tools/dataChecker/utils"
)

func TestGetDataFile(t *testing.T) {
	f := GetDataFile("Item")
	//t.Logf(": ", len(f.Sheets))

	assert.True(t, len(f.Sheets) > 0)
}

func TestXlsxData_ReadSource(t *testing.T) {
	xs := &utils.XlsxSource{}
	xs.SetSourceVal("Item")
	xs.SetSheetVal(nil, []string{"ID"})
	xs.MD5 = md5.Sum([]byte("TestXlsxData_ReadSource"))

	xd := NewXlsxData()
	xd.ReadSource(xs)

	// 不应该有空值
	_, ok := xd.Data[""]
	assert.False(t, ok)

	// 也没有表头
	_, ok = xd.Data["string"]
	assert.False(t, ok)

	// 开始行正确
	_, ok = xd.Data["VI_BC"]
	assert.True(t, ok)

	assert.True(t, len(xd.Data) > 0)
	//t.Logf(": ", len(xd.Data))
}

func TestFileData_ReadSource(t *testing.T) {
	fs := &utils.FileSource{}
	fs.SetSourceVal("Skill")
	fs.SetSheetVal([]string{"SKILLLIST"}, []string{"Icon"})
	fs.Path = "/Test1/Test2/Test3"
	fs.ExtName = "AVI"
	fs.MD5 = md5.Sum([]byte("TestFileData_ReadSource"))

	fd := NewFileData()
	fd.ReadSource(fs)

	// 不应该有空值
	_, ok := fd.Data[""]
	assert.False(t, ok)

	// 表头也没有
	_, ok = fd.Data["string"]
	assert.False(t, ok)

	assert.Equal(t, fs.Path, fd.Path)
	assert.Equal(t, fs.ExtName, fd.ExtName)

	//t.Logf(": %x", fd)
	//t.Logf(": %x", fd.MD5)
	assert.True(t, len(fd.Data) > 0)
	//t.Logf(": ", len(fd.Data))
}

func TestTryLoadXData(t *testing.T) {
	xs := &utils.XlsxSource{}
	xs.SetSourceVal("Item")
	xs.SetSheetVal(nil, []string{"ID"})
	xs.MD5 = md5.Sum([]byte("TestTryLoadXData"))

	hashData, ok := TryLoadXData(xs.MD5)
	assert.False(t, ok)
	assert.Equal(t, hashData, NewXlsxData())

	xd := NewXlsxData()
	xd.ReadSource(xs)

	hashData, ok = TryLoadXData(xs.MD5)
	assert.True(t, ok)
	assert.Equal(t, hashData, xd)

	xs.MD5 = md5.Sum([]byte("BliBliPliPli"))

	_, ok = TryLoadXData(xs.MD5)
	assert.False(t, ok)
}

func TestTryLoadFData(t *testing.T) {
	fs := &utils.FileSource{}
	fs.SetSourceVal("Skill")
	fs.SetSheetVal([]string{"SKILLLIST"}, []string{"Icon"})
	fs.MD5 = md5.Sum([]byte("TestTryLoadFData"))

	hashData, ok := TryLoadFData(fs.MD5)
	assert.False(t, ok)
	assert.Equal(t, hashData, NewFileData())

	fd := NewFileData()
	fd.ReadSource(fs)

	hashData, ok = TryLoadFData(fs.MD5)
	assert.True(t, ok)
	assert.Equal(t, hashData, fd)
}

func TestXlsxData_CheckInvalidSeparators(t *testing.T) {
	xs := &utils.XlsxSource{}
	xs.SetSourceVal("Item")
	xs.SetSheetVal(nil, []string{"ID"})
	xs.MD5 = md5.Sum([]byte("TestXlsxData_CheckInvalidSeparators"))

	xd := NewXlsxData()
	xd.ReadSource(xs)

	d, r := xd.CheckInvalidSeparators()
	assert.Empty(t, d)
	assert.True(t, r)

	// 手动添加错误数据
	xd.Data["ExpectedError1；"] = 1
	xd.Data["ExpectedError2："] = 1
	xd.Data["ExpectedError3，"] = 1
	xd.Data["ExpectedNormal1;"] = 1
	xd.Data["ExpectedNormal2:"] = 1
	xd.Data["ExpectedNormail3,"] = 1

	d, r = xd.CheckInvalidSeparators()

	assert.Equal(t, 3, len(d))
	assert.False(t, r)
}

func TestXlsxData_CheckRepeat(t *testing.T) {
	// Let's mock
	dataPath := cfg.Dir.DataProjectDir
	mockPath := utils.GetVCSRootPath() + "/tools/dataChecker/test"
	cfg.Dir.DataProjectDir = mockPath

	t.Run("Correct", func(*testing.T) {
		xs := &utils.XlsxSource{}
		xs.SetSourceVal("TestRepeat")
		xs.SetSheetVal([]string{"NoOptCorrect", "OptCorrect", "HOTGACHASHOW"}, []string{""})
		xs.MD5 = md5.Sum([]byte("TestXlsxData_TestCheckRepeat_Wrong"))

		xd := NewXlsxData()
		xd.ReadSource(xs)

		d, r := xd.CheckRepeat()

		assert.Empty(t, d)
		assert.True(t, r)
	})

	t.Run("Wrong", func(t *testing.T) {
		xs := &utils.XlsxSource{}
		xs.SetSourceVal("TestRepeat")
		xs.SetSheetVal([]string{"NoOptWrong", "OptWrong"}, []string{""})
		xs.MD5 = md5.Sum([]byte("TestXlsxData_TestCheckRepeat_Wrong"))

		xd := NewXlsxData()
		xd.ReadSource(xs)

		d, r := xd.CheckRepeat()

		assert.Equal(t, 15, len(d))
		assert.False(t, r)
	})

	// Restore mocked
	cfg.Dir.DataProjectDir = dataPath
}
