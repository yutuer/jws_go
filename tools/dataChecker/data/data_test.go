package data

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cfg.LoadConfig()

	os.Exit(m.Run())
}

func TestInit(t *testing.T) {
	assert.NotNil(t, cfg.Dir.DataProjectDir)
}

func TestNewXlsxData(t *testing.T) {
	xd := NewXlsxData()

	assert.NotNil(t, xd)
}

func TestNewIDSData(t *testing.T) {
	idsData := NewIDSData()
	assert.NotNil(t, idsData)
	assert.Equal(t, idsData.Local, "zh-Hans")
}

func TestNewFileData(t *testing.T) {
	fd := NewFileData()
	assert.NotNil(t, fd)
}
