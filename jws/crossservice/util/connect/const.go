package connect

import "time"

//Default definition
const (
	DefaultServerRequestQueueLength = 4096

	DefaultRequestInTimeout = 500 * time.Millisecond

	DefaultClientGetConnTimeout = 100 * time.Millisecond
)
