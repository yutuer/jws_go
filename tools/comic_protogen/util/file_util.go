package util

import (
	"bytes"
	"io/ioutil"
	"os"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GenFileUtil struct {
	filePath string
	buffer   *bytes.Buffer
}

func (gf *GenFileUtil) WriteString(str string) {
	gf.buffer.WriteString(str)
}

func (gf *GenFileUtil) Flush() {
	ioutil.WriteFile(gf.filePath, gf.buffer.Bytes(), 0666)
}

func (gf *GenFileUtil) WriteStringln(str string) {
	gf.buffer.WriteString(str)
	gf.buffer.WriteString("\n")
}

func NewGenFileUtil(filePath string) *GenFileUtil {
	gf := &GenFileUtil{}
	gf.filePath = filePath
	gf.buffer = bytes.NewBuffer([]byte{})
	return gf
}

func MkDir(dirName string) {
	if _, err := os.Stat(dirName); err != nil {
		err = os.Mkdir(dirName, 0777)
		if err != nil {
			logs.Error("mk root out dir ", err)
			return
		}
	}
}
