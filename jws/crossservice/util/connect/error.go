package connect

import (
	"fmt"
)

//error pre-definition
var (
	ErrInvalid = fmt.Errorf("ErrInvalid")
	ErrClosed  = fmt.Errorf("ErrClosed")
	ErrTimeout = fmt.Errorf("ErrTimeout")
)
