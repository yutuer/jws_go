package check

import "fmt"

func panicf(format string, v ...interface{}) {
	panic(fmt.Errorf(format, v...))
}
