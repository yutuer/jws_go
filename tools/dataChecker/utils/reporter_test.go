package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReporter(t *testing.T) {
	r := NewReporter()

	assert.NotEmpty(t, r.ReportPath)
	assert.NotEmpty(t, len(r.GitStatus))
	assert.True(t, len(r.ErrorCode2String) > 0)
	assert.True(t, len(r.CheckType2String) > 0)
}

func TestGetGitStatus(t *testing.T) {
	r := NewReporter()

	for _, gitInfo := range r.GitStatus {
		assert.NotEmpty(t, gitInfo)
	}
}

func TestReporter_ReadGitInfo(t *testing.T) {
	r := NewReporter()
	srvGit := filepath.Join(GetVCSRootPath(), ".git/FETCH_HEAD")
	r.ReadGitInfo(srvGit)
	assert.NotEmpty(t, r.GitStatus[0])
}

func TestReporter_Record(t *testing.T) {
	r := NewReporter()
	r.Record(255, "TestFile", 13, 0, []string{"Error1", "Error2", "Error3"})

	assert.Equal(t, 1, len(r.Unexceptions))
	assert.Equal(t, 255, r.Unexceptions[0].Index)
	assert.Equal(t, "TestFile", r.Unexceptions[0].ExtraInfo)
	assert.Equal(t, 13, r.Unexceptions[0].CheckType)
	assert.Equal(t, 0, r.Unexceptions[0].ErrorType)
	assert.Equal(t, 3, len(r.Unexceptions[0].Unexpections))
}

func TestReporter_Report(t *testing.T) {
	r := NewReporter()
	r.GetGitStatus()
	r.Record(255, "TestFile", 1, 0, []string{"Error1", "Error2", "Error3"})
	r.Record(288, "", 0, 2, nil)

	r.Report()

	s, err := os.Stat(r.lastlog)
	assert.Nil(t, err)
	assert.True(t, s.Size() > 0)

	os.Remove(r.lastlog)

	// 验证错误数为0时，不再生成报告
	r = NewReporter()
	r.Report()

	s, err = os.Stat(r.lastlog)
	assert.Nil(t, s)
}
