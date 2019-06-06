package linkext

/*
RawBytes这个解码器的特点是，认为拿到字节数据后不需要理解字节数据的内容。
因为特别适合，底层io接口保证每次传输进来的是完整的数据包。
因此可以配合websocket.go下的NewGinWebSocketListen来使用。
效率上不如codec_websocket.go

数据格式
 +------------------------------+
 |          data                |
 +------------------------------+
*/

import (
	"fmt"
	"io"
	"io/ioutil"

	"vcs.taiyouxi.net/platform/planx/funny/link"
)

func RawBytes() link.CodecType {
	return rawBytesCodecType{}
}

type rawBytesCodecType struct {
}

func (codecType rawBytesCodecType) NewEncoder(w io.Writer) link.Encoder {
	return rawBytesEncoder{
		Writer: w,
	}
}

func (codecType rawBytesCodecType) NewDecoder(r io.Reader) link.Decoder {
	return rawBytesDecoder{
		Reader: r,
	}
}

type rawBytesEncoder struct {
	Writer io.Writer
}

func WriteFull(w io.Writer, p []byte) error {
	n, err := w.Write(p)
	if err == nil && n != len(p) {
		err = fmt.Errorf("ErrShortWrite")
	}
	return err
}

func (encoder rawBytesEncoder) Encode(msg interface{}) error {
	b := msg.([]byte)
	return WriteFull(encoder.Writer, b)
}

type rawBytesDecoder struct {
	Reader io.Reader
}

const MinRead = 100

func (decoder rawBytesDecoder) Decode(msg interface{}) error {
	b, err := ioutil.ReadAll(decoder.Reader)
	// buf := bytes.NewBuffer(make([]byte, 0, MinRead))
	// n, err := buf.ReadFrom(decoder.Reader)
	if err != nil {
		return err
	}

	//XXX: 深度阅读io.util.ReadAll之后，当长度0的情况，读取工作就已经结束了
	if b == nil || len(b) == 0 {
		return io.EOF
	}
	// if n == 0 {
	//     return io.EOF
	// }
	// b := buf.Bytes()
	switch data := msg.(type) {
	case *[]byte:
		*data = b
		return nil
	default:
	}

	return fmt.Errorf("rawBytesEncoder need []bytes only!")
}
