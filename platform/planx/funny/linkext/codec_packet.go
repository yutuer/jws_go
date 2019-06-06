package linkext

/*
老版本Link库中，可以使用interface给自己的数据包提供结构体到二进制的相互装换
并保证封包的方式，数据最大在2个字节内。类似link.String


  length 16bits
+----------------------------+
|byte|byte|    data          |
+----------------------------+

本codec提供了slab公共库提供的gc外管理的内存池，提高效率，减少不必要的gc
*/

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"sync"

	"vcs.taiyouxi.net/platform/planx/funny/link"

	"github.com/funny/slab"
)

var ErrPacketTooLarge = errors.New("funny/link: too large packet")

type PacketUnmarshaler interface {
	UnmarshalPacket([]byte) error
}

type PacketMarshaler interface {
	BinarySize() int
	MarshalPacket([]byte) error
}

func Packet(bufioSize int, bufferPool *slab.SyncPool) link.CodecType {
	return &packetCodecType{
		bufferPool: bufferPool,
		bufioSize:  bufioSize,
	}
}

type packetCodecType struct {
	bufioSize  int
	readerPool sync.Pool
	bufferPool *slab.SyncPool
}

func (codecType *packetCodecType) NewEncoder(w io.Writer) link.Encoder {
	return &packetEncoder{
		writer:     w,
		bufferPool: codecType.bufferPool,
	}
}

func (codecType *packetCodecType) NewDecoder(r io.Reader) link.Decoder {
	reader, ok := codecType.readerPool.Get().(*bufio.Reader)
	if ok {
		reader.Reset(r)
	} else {
		reader = bufio.NewReaderSize(r, codecType.bufioSize)
	}
	return &packetDecoder{
		reader:     reader,
		readerPool: &codecType.readerPool,
		bufferPool: codecType.bufferPool,
	}
}

const packetHeadSize = 2

type packetEncoder struct {
	writer     io.Writer
	bufferPool *slab.SyncPool
}

type packetDecoder struct {
	head       [packetHeadSize]byte
	reader     *bufio.Reader
	readerPool *sync.Pool
	bufferPool *slab.SyncPool
}

func (encoder *packetEncoder) Encode(msg interface{}) (err error) {
	rsp := msg.(PacketMarshaler)

	n := rsp.BinarySize()
	if n > math.MaxUint16 {
		panic(ErrPacketTooLarge)
	}

	b := encoder.bufferPool.Alloc(n + packetHeadSize)
	defer encoder.bufferPool.Free(b)

	binary.LittleEndian.PutUint16(b, uint16(n))
	err = rsp.MarshalPacket(b[packetHeadSize:])
	if err != nil {
		return
	}
	_, err = encoder.writer.Write(b)
	return
}

func (decoder *packetDecoder) Decode(msg interface{}) (err error) {
	req := msg.(PacketUnmarshaler)

	head := decoder.head[:]
	if _, err = io.ReadFull(decoder.reader, head); err != nil {
		return
	}
	n := int(binary.LittleEndian.Uint16(head))

	if decoder.reader.Buffered() >= n {
		var b []byte
		b, err = decoder.reader.Peek(n)
		if err != nil {
			return
		}
		err = req.UnmarshalPacket(b)
		if err != nil {
			return
		}
		_, err = decoder.reader.Discard(n)
		return
	}

	b := decoder.bufferPool.Alloc(n)
	defer decoder.bufferPool.Free(b)

	if _, err = io.ReadFull(decoder.reader, b); err != nil {
		decoder.bufferPool.Free(b)
		return
	}
	return req.UnmarshalPacket(b)
}

func (decoder *packetDecoder) Dispose() {
	decoder.reader.Reset(nil)
	decoder.readerPool.Put(decoder.reader)
}
