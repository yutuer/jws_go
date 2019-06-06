package data

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestTransHotActivities(t *testing.T) {
	f := TransHotActivities("zh-Hans")

	s, err := os.Stat(f)
	assert.Nil(t, err)
	assert.True(t, s.Size() > 0)

	os.Remove(f)
}
