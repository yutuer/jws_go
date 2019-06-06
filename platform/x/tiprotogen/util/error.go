package util

import (
	"errors"
	"fmt"
	"vcs.taiyouxi.net/platform/x/tiprotogen/log"
)

func PanicInfo(format string, params ...interface{}) {
	errorInfo := fmt.Sprintf(format, params...)
	log.Err("panic by %s", errorInfo)
	log.Flush()
	panic(errors.New(errorInfo))
}
