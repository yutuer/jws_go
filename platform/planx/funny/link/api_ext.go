// +build old

package link

import "github.com/gin-gonic/gin"

type Codec interface {
	Decode(msg interface{}) error
	Encode(msg interface{}) error
}

type MonoCounter interface {
	Inc()
}

var (
	full_buff_close_session_counter MonoCounter
	sum_rec_counter                 MonoCounter
	sum_send_counter                MonoCounter
	sum_want_send_counter           MonoCounter
	sum_conn_counter                MonoCounter
)

// const (
//     // 发送队列长度
//     SendQueueLen = 1024
// )

func RegisterCounter(fullBuffCloseSessionCounter, sumRecCounter, sumSendCounter, sumWantSendCounter, sumConnCounter MonoCounter) {
	full_buff_close_session_counter = fullBuffCloseSessionCounter
	sum_rec_counter = sumRecCounter
	sum_send_counter = sumSendCounter
	sum_want_send_counter = sumWantSendCounter
	sum_conn_counter = sumConnCounter
}

func WebSocketServe(address string, codecType CodecType, g *gin.Engine) (*ServerWebsocket, error) {
	return NewServerWebsocket(address, codecType, g), nil
}

func WebSocketServe2(url string, codecType CodecType, g *gin.Engine) (*ServerWebsocket2, error) {
	return NewServerWebsocket2(url, codecType, g), nil
}
