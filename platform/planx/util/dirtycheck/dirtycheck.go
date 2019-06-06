package dirtycheck

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type DirtyChecker interface {
	Check(v interface{}) (hash string, dirty bool)
}

func NewDirtyChecker() DirtyChecker {
	return new(GobDirtyCheck)
}

type GobDirtyCheck struct {
	md5 string
}

func (dc *GobDirtyCheck) Check(v interface{}) (hash string, dirty bool) {
	hash = ""
	dirty = true
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		logs.Warn("GobDirtyCheck Encoded Gob Failed by %s, %v", err.Error(), v)
		return "", true
	}
	s := md5.Sum(buf.Bytes())
	hash = string(s[:])
	dirty = (dc.md5 != hash)
	dc.md5 = hash
	return
}
