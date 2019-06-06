package lumberjack

import (
	"log"

	"vcs.taiyouxi.net/platform/planx/util/lumberjack.v2"
)

// To use lumberjack with the standard library's log package, just pass it into
// the SetOutput function when your application starts.
func Example() {
	log.SetOutput(&lumberjack.Logger{
		FileTempletName: "/var/log/myapp/foo.log",
		MaxSize:         500, // megabytes
		MaxBackups:      3,
		MaxAge:          28, // days
	})
}
