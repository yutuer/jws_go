package client

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"

	"github.com/ugorji/go/codec"
	// _ "vcs.taiyouxi.net/platform/planx/util/logs"
)

type PacketID int32

const (
	PacketIDContent PacketID = iota
	PacketIDReqResp
	PacketIDPingPong
	PacketIDGateSession
	PacketIDGatePkt
)

type Packet struct {
	len     int32
	id      PacketID
	rawData []byte
	// crc     uint32
}

type SessionPacket struct {
	SessionID  string
	AccountID  string //0:0:1001
	PacketData *Packet
}

var BytePing []byte

//var PacketPing *Packet

func init() {
	BytePing = []byte("PING")
}

func (p *Packet) MarshalBinary() ([]byte, error) {
	var mh codec.MsgpackHandle
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	mh.RawToString = true
	var out []byte //XXX: If it comes from []byte buffer pool, it would be cool
	enc := codec.NewEncoderBytes(&out, &mh)
	err := enc.Encode(p.len)
	if err != nil {
		return nil, err
	}
	err = enc.Encode(p.id)
	if err != nil {
		return nil, err
	}
	err = enc.Encode(p.rawData)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (p *Packet) UnmarshalBinary(data []byte) error {
	var mh codec.MsgpackHandle
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	mh.RawToString = true
	dec := codec.NewDecoderBytes(data, &mh)
	err := dec.Decode(&p.len)
	if err != nil {
		return err
	}
	err = dec.Decode(&p.id)
	if err != nil {
		return err
	}
	err = dec.Decode(&p.rawData)
	if err != nil {
		return err
	}
	return nil
}

func (p *Packet) GetContentType() PacketID {
	return p.id
}

func (p *Packet) IsContent() bool {
	return p.id == PacketIDContent
}

func (p *Packet) IsReqResp() bool {
	return p.id == PacketIDReqResp
}

func (p *Packet) String() string {
	return string(p.rawData)
}

func (p *Packet) GetBytes() []byte {
	return p.rawData
}

func (p *Packet) GetId() PacketID {
	return p.id
}

func NewPingPacket() *Packet {
	var buf bytes.Buffer
	buf.Write(BytePing)
	fmt.Fprintf(&buf, "%d", time.Now().Unix())
	logs.Trace("Ping To Client %s", string(buf.Bytes()))
	return NewPacket(buf.Bytes(), PacketIDPingPong)
}

func NewPacket(data []byte, id PacketID) *Packet {
	pkt := &Packet{
		len:     int32(len(data)),
		rawData: data,
		id:      id,
		// crc:     uint32(0),
	}
	return pkt
}

const max_Msg_Size = 1024 * 1024

var msg_size_err = errors.New("msg_size_err")
var msg_read_err = errors.New("msg_read_err")

//Instead of binary.LittleEndian.Uint32, better performance without reflect
func ReadLittleEndianUint32(r io.Reader) (uint32, error) {
	var buf32 [4]byte
	n, err := io.ReadFull(r, buf32[:])
	if err != nil {
		if n != 0 {
			logs.Warn("ReadLittleEndianUint32 got error with ReadFull: size(%d), %s", n, err.Error())
		}
		return 0, err
	}
	if n != 4 {
		return 0, fmt.Errorf("ReadPacket header failed, len is not 4")
	}

	ubufint32 := binary.LittleEndian.Uint32(buf32[:])
	return ubufint32, nil
}

func ReadPacket(r io.Reader, setReadDeadline func(int)) (*Packet, error) {
	var msgSize int32
	var umsgSize uint32

	var msgID PacketID
	var err error

	//TODO: by YZH 这个第一个Deadline是用来处理消息消息主循环的等待，这个方法不是很理想
	//如果恰好读完这个头，然后引起了系一个数据的Read出错timeout则是有问题的。
	//因此在此处和下面分别设置不同的数据Deadline
	//理应在这个消息头到达后，应该立即有后续数据跟进，因此如果后续数据未能在指定时间出现
	//则认为timeout也应该断线。
	setReadDeadline(15)
	// message size
	if umsgSize, err = ReadLittleEndianUint32(r); err != nil {
		return nil, err
	} else {
		msgSize = int32(umsgSize)
	}
	//protect bigger than 1MB message, Have to protect msgSize to avoid memory exhaust problem
	if msgSize < 0 || msgSize >= max_Msg_Size {
		return nil, msg_size_err
	}

	if err != nil {
		e, ok := err.(net.Error)
		if ok && e.Temporary() {
			if e.Timeout() {
				return nil, err
			}
		}
	}
	// message binary data
	setReadDeadline(5)
	readSize := msgSize + 4
	buf := make([]byte, readSize)
	if n, err := io.ReadFull(r, buf); err != nil {
		if n != 0 {
			logs.Warn("ReadPacket got error with ReadFull: size(%d), %s", n, err.Error())
		}

		e, ok := err.(net.Error)
		if ok && e.Temporary() {
			if e.Timeout() {
				return nil, fmt.Errorf("ReadPacket error: the  msgbody part come too late!! you are killed.")
			}
		}
		return nil, err
	} else {
		if int32(n) != readSize {
			return nil, msg_read_err
		}
	}
	upktid := binary.LittleEndian.Uint32(buf[:4])
	msgID = PacketID(upktid)

	// logs.Info("ReadPacket id %x", buf)
	// logs.Info("got packet size %d, content:%s", msgSize, string(buf))
	return &Packet{msgSize, msgID, buf[4:]}, nil
}

func SendBytes(w io.Writer, data []byte, id PacketID) (int, error) {
	return SendPacket(w, NewPacket(data, id))
}

func SendPacket(w io.Writer, pkt *Packet) (int, error) {
	framelen := 4 + 4
	beBuf := make([]byte, framelen)
	size := pkt.len
	id := pkt.id

	binary.LittleEndian.PutUint32(beBuf, uint32(size))
	binary.LittleEndian.PutUint32(beBuf[4:], uint32(id))

	n, err := w.Write(beBuf)

	if err != nil {
		return n, err
	}

	//rawData might be longer than its content.
	n, err = w.Write(pkt.rawData[:size])
	return n + framelen, err
}
