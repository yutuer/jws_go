package linkext

import (
	"io"
	"time"

	"golang.org/x/net/websocket"

	"vcs.taiyouxi.net/platform/planx/funny/link"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func WebsocketType() link.CodecType {
	return websocketCodecType{}
}

func WebsocketWithWriteDeadline() link.CodecType {
	return websocketCodecType{
		writeDeadLine: time.Second * 10,
	}
}

type websocketCodecType struct {
	writeDeadLine time.Duration
	// ReadDeadLine  time.Time
}

func (codecType websocketCodecType) NewEncoder(w io.Writer) link.Encoder {
	if ws, ok := w.(*websocketConn); ok {
		return &websocketEnDecoder{
			WSConn:        ws,
			writeDeadLine: codecType.writeDeadLine,
		}
	}
	panic("codec_websocket only work with golang.org/x/net/websocket")
	return nil
}

func (codecType websocketCodecType) NewDecoder(r io.Reader) link.Decoder {
	if ws, ok := r.(*websocketConn); ok {
		return &websocketEnDecoder{
			WSConn:        ws,
			writeDeadLine: codecType.writeDeadLine,
		}
	}
	panic("codec_websocket only work with golang.org/x/net/websocket")
	return nil
}

//可以直接使用golang.org/x/net/websocket库的websocket.Conn，而不用使用这层封装。
//这里依然使用了这个封装，是因为websocket.go文件中的GinWebSocketListen需要使用websocketConn这个类型
//因此，如果使用websocket.Conn请自己封装新的GinWebSocketListen
type websocketEnDecoder struct {
	WSConn        *websocketConn
	writeDeadLine time.Duration
}

func (encoder *websocketEnDecoder) Encode(msg interface{}) error {
	if encoder.writeDeadLine > time.Second {
		encoder.WSConn.SetWriteDeadline(time.Now().Add(encoder.writeDeadLine))
	}
	return websocket.Message.Send(encoder.WSConn.Conn, msg)
}

func (decoder *websocketEnDecoder) Decode(msg interface{}) error {
	err := websocket.Message.Receive(decoder.WSConn.Conn, msg)
	if err != nil {
		if err != io.EOF {
			logs.Error("websocketEnDecoder Decode err %v", err.Error())
		}
		return err
	}

	switch data := msg.(type) {
	case *string:
		s := *data
		if len(s) == 0 {
			*data = ""
			return io.EOF
		}
		return nil
	case *[]byte:
		sli := *data
		if sli == nil || len(sli) == 0 {
			*data = nil
			return io.EOF
		}
		return nil
	}
	return nil
}
