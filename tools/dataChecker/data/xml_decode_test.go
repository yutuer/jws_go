package data

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDumpIDSXml(t *testing.T) {
	r := DumpIDSXml()

	assert.NotNil(t, r)
	//t.Logf(":", r.Worksheet[0].Row[0].Data)
}
