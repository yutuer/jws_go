package connect

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
	"time"
)

//Conn ..
type Conn struct {
	oc  *origConn
	err error

	isReleased  bool
	releaseFunc func(*origConn, bool)
}

func newConn(oc *origConn) *Conn {
	conn := &Conn{
		oc:  oc,
		err: nil,

		isReleased:  false,
		releaseFunc: nil,
	}
	return conn
}

func newConnInErr(err error) *Conn {
	return &Conn{
		oc:  nil,
		err: err,
	}
}

//Invalid ..
func (c *Conn) Invalid() bool {
	if nil == c.oc {
		return true
	}
	if c.isReleased {
		return true
	}

	return false
}

//Err ..
func (c *Conn) Err() error {
	return c.err
}

//LocalAddr ..
func (c *Conn) LocalAddr() *net.TCPAddr {
	return c.oc.tc.LocalAddr().(*net.TCPAddr)
}

//RemoteAddr ..
func (c *Conn) RemoteAddr() *net.TCPAddr {
	return c.oc.tc.RemoteAddr().(*net.TCPAddr)
}

//Send ..
func (c *Conn) Send(req *Message) error {
	if c.Invalid() {
		return ErrInvalid
	}
	pac := &packet{data: req.Payload[:req.Length], length: req.Length}
	err := c.oc.writePacket(pac)
	c.err = err
	if nil != err {
		if io.EOF == err {
			return ErrClosed
		}
		if e, eok := err.(*net.OpError); eok {
			if se, sok := e.Err.(*os.SyscallError); sok && se.Err == syscall.EPIPE {
				return ErrClosed
			}
		}
	}
	return err
}

//Recv ..
func (c *Conn) Recv() (rsp *Message, err error) {
	if c.Invalid() {
		return nil, ErrInvalid
	}
	c.oc.tc.SetReadDeadline(time.Time{})
	pac, err := c.oc.readPacket()
	if nil != err {
		c.err = err
		return nil, err
	}
	msg := NewMessage(pac.data, pac.length)
	return msg, nil
}

//RecvWithTimeout ..
func (c *Conn) RecvWithTimeout(t time.Time) (rsp *Message, err error) {
	if c.Invalid() {
		return nil, ErrInvalid
	}
	c.oc.tc.SetReadDeadline(t)
	pac, err := c.oc.readPacket()
	if nil != err {
		c.err = err
		return nil, err
	}
	msg := NewMessage(pac.data, pac.length)
	return msg, nil
}

//Release ..
func (c *Conn) Release() bool {
	if c.isReleased {
		return false
	}
	c.isReleased = true
	if nil != c.releaseFunc && nil != c.oc {
		c.releaseFunc(c.oc, nil != c.err)
	}
	return true
}

func (c *Conn) setReleaseFunc(rf func(*origConn, bool)) {
	c.releaseFunc = rf
}

type origConn struct {
	tc *net.TCPConn
}

func newOrigConn(tc *net.TCPConn) *origConn {
	oc := &origConn{
		tc: tc,
	}
	return oc
}

func (oc *origConn) readPacket() (*packet, error) {
	od, ol, first, max, err := oc.readSubPacket()
	if nil != err {
		return nil, err
	}

	length := int(ol)
	for seq := first + 1; seq < max; seq++ {
		subData, subLen, subSeq, subMax, err := oc.readSubPacket()
		if nil != err {
			return nil, err
		}
		if max != subMax {
			return nil, fmt.Errorf("read subPacket un-match max value, %d != %d (seq:%d, subSeq:%d, Data:%+v), current-od: %+v", max, subMax, seq, subSeq, subData, od)
		}
		if seq != subSeq {
			return nil, fmt.Errorf("read subPacket unordered, %d != %d (max:%d, subMax:%d, Data:%+v), current-od: %+v", seq, subSeq, max, subMax, subData, od)
		}
		od = append(od, subData...)
		length += int(subLen)
	}

	pac := &packet{data: od, length: length}
	return pac, nil
}

func (oc *origConn) readSubPacket() ([]byte, uint16, int8, int8, error) {
	ol := uint16(0)
	if err := binary.Read(oc.tc, binary.BigEndian, &ol); nil != err {
		return nil, 0, 0, 0, err
	}
	seq := int8(0)
	if err := binary.Read(oc.tc, binary.BigEndian, &seq); nil != err {
		return nil, 0, 0, 0, err
	}
	max := int8(0)
	if err := binary.Read(oc.tc, binary.BigEndian, &max); nil != err {
		return nil, 0, 0, 0, err
	}
	od := make([]byte, ol)
	if err := binary.Read(oc.tc, binary.BigEndian, od); nil != err {
		return nil, 0, 0, 0, err
	}
	return od, ol, seq, max, nil
}

const (
	maxSubPacketLen = 0xFFF0
)

func (oc *origConn) writePacket(pac *packet) error {
	max := int8(pac.length / maxSubPacketLen)
	if 0 != pac.length%maxSubPacketLen {
		max++
	}
	for seq := int8(0); seq < max; seq++ {
		start := int(seq) * maxSubPacketLen
		end := int(seq+1) * maxSubPacketLen
		if end > pac.length {
			end = pac.length
		}
		if err := oc.writeSubPacket(pac.data[start:end], uint16(end-start), seq, max); nil != err {
			return err
		}
	}
	return nil
}

func (oc *origConn) writeSubPacket(data []byte, length uint16, seq int8, max int8) error {
	ol := uint16(length)
	if err := binary.Write(oc.tc, binary.BigEndian, ol); nil != err {
		return err
	}
	oSeq := int8(seq)
	if err := binary.Write(oc.tc, binary.BigEndian, oSeq); nil != err {
		return err
	}
	oMax := int8(max)
	if err := binary.Write(oc.tc, binary.BigEndian, oMax); nil != err {
		return err
	}
	od := data
	if err := binary.Write(oc.tc, binary.BigEndian, od); nil != err {
		return err
	}
	return nil
}

func (oc *origConn) close() {
	oc.tc.Close()
}

//-------debug

//SendBad ..
func (c *Conn) SendBad(req *Message) error {
	if c.Invalid() {
		return ErrInvalid
	}
	pac := &packet{data: req.Payload[:req.Length], length: req.Length}
	err := c.oc.writeBadPacket(pac)
	c.err = err
	if nil != err {
		if io.EOF == err {
			return ErrClosed
		}
		if e, eok := err.(*net.OpError); eok {
			if se, sok := e.Err.(*os.SyscallError); sok && se.Err == syscall.EPIPE {
				return ErrClosed
			}
		}
	}
	return err
}

func (oc *origConn) writeBadPacket(pac *packet) error {
	max := int8(pac.length / maxSubPacketLen)
	if 0 != pac.length%maxSubPacketLen {
		max++
	}
	for seq := int8(0); seq < max; seq++ {
		start := int(seq) * maxSubPacketLen
		end := int(seq+1) * maxSubPacketLen
		if end > pac.length {
			end = pac.length
		}
		if 0 == seq {
			if err := oc.writeSubPacket(pac.data[start:end], uint16(end-start), seq, max); nil != err {
				return err
			}
		} else {
			if err := oc.writeSubPacket(pac.data[start:end], uint16(end-start), seq, 0); nil != err {
				return err
			}
		}
	}
	return nil
}
