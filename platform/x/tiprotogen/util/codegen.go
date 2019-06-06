package util

import (
	"bytes"
	"fmt"
)

type CodeGenData struct {
	b *bytes.Buffer
}

func (c *CodeGenData) WriteLine(format string, params ...interface{}) {
	data := fmt.Sprintf(format, params...)
	c.b.WriteString(data)
	c.b.WriteString("\n") // TODO 区别换行符
}

func NewCodeGenData() *CodeGenData {
	res := new(CodeGenData)
	res.b = bytes.NewBuffer([]byte{})
	return res
}

func (c *CodeGenData) Bytes() []byte {
	return c.b.Bytes()
}
