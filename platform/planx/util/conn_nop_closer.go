package util

import "net"

type ConnNopCloser struct {
	C net.Conn
}

func (c *ConnNopCloser) Read(p []byte) (n int, err error) {
	n, err = c.C.Read(p)
	return
}

func (c *ConnNopCloser) Write(p []byte) (n int, err error) {
	n, err = c.C.Write(p)
	return
}

func (c *ConnNopCloser) Close() error {
	//logs.Critical("Gate ConnNoClose for net/textproto, should not close by textproto")
	return nil
}
