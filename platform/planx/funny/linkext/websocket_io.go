package linkext

import (
	"bytes"
	"io"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

//Read 通过转换websocket的读取接口适配原有接口
//XXX: 因为golang.org/x/websocket库520af5de654dc4dd4f0f65aa40e66dbbd9043df1 [520af5d]
// websocket.go 194行的代码应该去掉，导致io.Reader语意出现问题
// 因此使用有效接口进行适配换算
func (ws *websocketConn) Read(p []byte) (n int, err error) {
	if ws.bufReader == nil {
		var b []byte
		err = ws.codec.Receive(ws.Conn, &b)
		if err == nil {
			bytesBuf := bytes.NewBuffer(b)
			ws.bufReader = io.LimitReader(bytesBuf, int64(len(b)))
		} else {
			logs.Error("websocket Message Receive problem! %s", err.Error())
		}
	}

	if ws.bufReader != nil {
		n, err = ws.bufReader.Read(p)
		if err == io.EOF {
			ws.bufReader = nil
		}
		return
	}
	return
}

func (ws *websocketConn) Write(p []byte) (n int, err error) {
	err = ws.codec.Send(ws.Conn, p)
	if err == nil {
		n = len(p)
	} else {
		logs.Error("websocket Message Send problem! %s", err.Error())
	}
	return
}
