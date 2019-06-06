package data

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIDSData_SetLocal(t *testing.T) {
	d := NewIDSData()
	assert.NotNil(t, d.Local)
	assert.Error(t, errors.New("No local found!\n"), d.SetLocal("us"))

	d.SetLocal("en")
	assert.Equal(t, "en", d.Local)
}

func TestIDSData_LoadLocalizationData(t *testing.T) {
	d := NewIDSData()
	d.LoadLocalizationData()
	d.SetLocal("zh-HMT")

	assert.True(t, len(d.Data) > 0)
	assert.NotNil(t, d.MD5)
	//t.Logf(":", len(d.Data))
	//t.Logf("md5: %x", d.MD5)

	_, ok := usedData.IData[d.MD5]
	assert.True(t, ok)
}

func TestTryLoadIDS(t *testing.T) {
	// 错误
	_, ok := TryLoadIDS("Chinese")
	assert.False(t, ok)

	// 未加载过
	_, ok = TryLoadIDS("en")
	assert.False(t, ok)

	d := NewIDSData()
	d.SetLocal("en")
	d.LoadLocalizationData()

	// 已加载
	e, ok := TryLoadIDS("en")
	assert.True(t, ok)
	assert.EqualValues(t, d, e)

	_, ok = TryLoadIDS("zh-HMT")
	assert.False(t, ok)
}
