package linkext

/*
移植了老版本link库中link.String(link.U16BE)

  length 16bits
+----------------------------+
|byte|byte|    data          |
+----------------------------+

发送和解析的数据包格式
*/

import (
	"encoding/binary"
	"io"

	"vcs.taiyouxi.net/platform/planx/funny/link"
)

func String() link.CodecType {
	return stringCodecType{}
}

type stringCodecType struct {
}

func (codecType stringCodecType) NewEncoder(w io.Writer) link.Encoder {
	return stringEncoder{
		Writer: w,
	}
}

func (codecType stringCodecType) NewDecoder(r io.Reader) link.Decoder {
	return stringDecoder{
		Reader: r,
	}
}

type stringEncoder struct {
	Writer io.Writer
}

func (encoder stringEncoder) Encode(msg interface{}) error {
	b := []byte(msg.(string))
	lenb := len(b)
	if err := binary.Write(encoder.Writer, binary.BigEndian, lenb); err != nil {
		return err
	}
	return binary.Write(encoder.Writer, binary.BigEndian, b)
}

type stringDecoder struct {
	Reader io.Reader
}

func (decoder stringDecoder) Decode(msg interface{}) error {
	var lenb int
	binary.Read(decoder.Reader, binary.BigEndian, &lenb)
	b := make([]byte, lenb)
	err := binary.Read(decoder.Reader, binary.BigEndian, b)
	*(msg.(*string)) = string(b)
	return err
}
