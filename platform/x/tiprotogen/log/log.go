package log

import (
	"fmt"
	"runtime/debug"

	"github.com/timesking/seelog"
)

// A Simple Wapper to seelog

// Init init a globa logger
func Init(path string) error {
	logger, err := seelog.LoggerFromConfigAsFile(path)
	if err != nil {
		return err
	}
	seelog.ReplaceLogger(logger)
	return nil
}

// Flush flush curr log data, need call before exit
func Flush() {
	seelog.Flush()
}

// Trace Log
func Trace(format string, params ...interface{}) {
	seelog.Tracef(format, params...)
}

// Debug Log
func Debug(format string, params ...interface{}) {
	seelog.Debugf(format, params...)
}

// Info Log
func Info(format string, params ...interface{}) {
	seelog.Infof(format, params...)
}

// Error Log
func Err(format string, params ...interface{}) {
	seelog.Errorf(format, params...)
}

// Critical Log
func Critical(format string, params ...interface{}) {
	seelog.Critical(
		fmt.Sprintf(format, params...),
		",",
		debug.Stack())
}

// PanicHandle defer when need log panic err data
func PanicHandle() {
	if err := recover(); err != nil {
		seelog.Critical(
			"Panic Err by ",
			err,
			" , stack: ",
			debug.Stack())
		seelog.Flush()
		panic(err)
	}
}
