package redis

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type CmdBuffer interface {
	Err() error
	GetCmdNumber() int
	Send(commandName string, args ...interface{}) error
	String() string
	Bytes() []byte
}

// CmdBuffer is not goroutine safe
type cmdBuffer struct {

	// Shared
	mu      *sync.Mutex
	pending int
	err     error

	bw *bytes.Buffer

	// Scratch space for formatting argument length.
	// '*' or '$', length, "\r\n"
	lenScratch [32]byte

	// Scratch space for formatting integers and floats.
	numScratch [40]byte
}

// NewCmdBuffer returns a new Redigo connection for the given net connection.
func NewCmdBuffer() CmdBuffer {
	return &cmdBuffer{
		mu: &sync.Mutex{},
		bw: bytes.NewBuffer(nil),
	}
}

func (c *cmdBuffer) fatal(err error) error {
	c.mu.Lock()
	if c.err == nil {
		c.err = err
		c.bw.Reset()
	}
	c.mu.Unlock()
	return err
}

func (c *cmdBuffer) GetCmdNumber() int {
	c.mu.Lock()
	pending := c.pending
	c.mu.Unlock()
	return pending
}

func (c *cmdBuffer) Err() error {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()
	return err
}

func (c *cmdBuffer) writeLen(prefix byte, n int) error {
	c.lenScratch[len(c.lenScratch)-1] = '\n'
	c.lenScratch[len(c.lenScratch)-2] = '\r'
	i := len(c.lenScratch) - 3
	for {
		c.lenScratch[i] = byte('0' + n%10)
		i--
		n = n / 10
		if n == 0 {
			break
		}
	}
	c.lenScratch[i] = prefix
	_, err := c.bw.Write(c.lenScratch[i:])
	return err
}

func (c *cmdBuffer) writeString(s string) error {
	c.writeLen('$', len(s))
	c.bw.WriteString(s)
	_, err := c.bw.WriteString("\r\n")
	return err
}

func (c *cmdBuffer) writeBytes(p []byte) error {
	c.writeLen('$', len(p))
	c.bw.Write(p)
	_, err := c.bw.WriteString("\r\n")
	return err
}

func (c *cmdBuffer) writeInt64(n int64) error {
	return c.writeBytes(strconv.AppendInt(c.numScratch[:0], n, 10))
}

func (c *cmdBuffer) writeFloat64(n float64) error {
	return c.writeBytes(strconv.AppendFloat(c.numScratch[:0], n, 'g', -1, 64))
}

func (c *cmdBuffer) writeCommand(cmd string, args []interface{}) (err error) {
	c.writeLen('*', 1+len(args))
	err = c.writeString(cmd)
	for _, arg := range args {
		if err != nil {
			break
		}
		switch arg := arg.(type) {
		case string:
			err = c.writeString(arg)
		case []byte:
			err = c.writeBytes(arg)
		case int:
			err = c.writeInt64(int64(arg))
		case int64:
			err = c.writeInt64(arg)
		case float64:
			err = c.writeFloat64(arg)
		case bool:
			if arg {
				err = c.writeString("1")
			} else {
				err = c.writeString("0")
			}
		case nil:
			err = c.writeString("")
		default:
			var buf bytes.Buffer
			fmt.Fprint(&buf, arg)
			err = c.writeBytes(buf.Bytes())
		}
	}
	return err
}

func (c *cmdBuffer) Send(cmd string, args ...interface{}) error {
	c.mu.Lock()
	c.pending++
	c.mu.Unlock()
	if err := c.writeCommand(cmd, args); err != nil {
		return c.fatal(err)
	}
	return nil
}

func (c *cmdBuffer) String() string {
	return c.bw.String()
}

func (c *cmdBuffer) Bytes() []byte {
	return c.bw.Bytes()
}

//###for conn

func (c *conn) DoCmdBuffer(cb CmdBuffer, transaction bool) (interface{}, error) {
	c.mu.Lock()
	pending := cb.GetCmdNumber()
	c.mu.Unlock()

	if pending == 0 {
		return nil, nil
	}

	if c.writeTimeout != 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}

	if transaction {
		c.writeCommand("MULTI", nil)
		pending++
	}

	if _, err := c.bw.Write(cb.Bytes()); err != nil {
		return nil, c.fatal(err)
	}
	var cmd string
	if transaction {
		cmd = "EXEC"
		c.writeCommand(cmd, nil)
	}

	if err := c.bw.Flush(); err != nil {
		return nil, c.fatal(err)
	}

	if c.readTimeout != 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	}

	if cmd == "" {
		reply := make([]interface{}, pending)
		for i := range reply {
			r, e := c.readReply()
			if e != nil {
				return nil, c.fatal(e)
			}
			reply[i] = r
		}
		return reply, nil
	}

	var err error
	var reply interface{}
	for i := 0; i <= pending; i++ {
		var e error
		if reply, e = c.readReply(); e != nil {
			return nil, c.fatal(e)
		}
		if e, ok := reply.(Error); ok && err == nil {
			err = e
		}
	}
	return reply, err
}

//for pool.go

func (pc *pooledConnection) DoCmdBuffer(cb CmdBuffer, t bool) (interface{}, error) {
	return pc.c.DoCmdBuffer(cb, t)
}

func (ec errorConnection) DoCmdBuffer(cb CmdBuffer, t bool) (interface{}, error) { return nil, ec.err }
