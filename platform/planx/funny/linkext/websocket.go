package linkext

import (
	"io"
	"net"
	"sync"
	"sync/atomic"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/mailgun/oxy/utils"
	"github.com/mailgun/timetools"
	"golang.org/x/net/websocket"
	// "vcs.taiyouxi.net/platform/planx/util/logs"
	"errors"
	"strings"
	"time"
)

var closeErr = errors.New("Listener has been closed.")

type websocketConn struct {
	*websocket.Conn
	codec     websocket.Codec
	closeFlag sync.WaitGroup
	closeOnce sync.Once

	bufReader io.Reader
}

func newWebSocketConn(c *websocket.Conn) *websocketConn {
	wsc := &websocketConn{
		Conn:  c,
		codec: websocket.Codec{marshal, unmarshal},
		// codec: websocket.Message,
	}
	wsc.closeFlag.Add(1)
	return wsc
}

func marshal(v interface{}) (msg []byte, payloadType byte, err error) {
	switch data := v.(type) {
	case string:
		return []byte(data), websocket.TextFrame, nil
	case []byte:
		return data, websocket.BinaryFrame, nil
	}
	return nil, websocket.UnknownFrame, websocket.ErrNotSupported
}

func unmarshal(msg []byte, payloadType byte, v interface{}) (err error) {
	switch data := v.(type) {
	case *string:
		*data = string(msg)
		return nil
	case *[]byte:
		*data = msg
		return nil
	}
	return websocket.ErrNotSupported
}

func (ws *websocketConn) waitDone() {
	ws.closeFlag.Wait()
}

func (ws *websocketConn) Close() error {
	// logs.Info("websocket closed")
	ws.closeOnce.Do(func() {
		ws.closeFlag.Done()
	})
	return ws.Conn.Close()
}

type GinWebSocketListen struct {
	g         *gin.Engine
	wsChan    chan *websocketConn
	closeFlag int32
	listener  net.Listener
}

func NewGinWebSocketListen(g *gin.Engine, listener net.Listener, uri string) (*GinWebSocketListen, *gin.Engine) {
	if g == nil {
		g = gin.Default()
	}
	gsl := &GinWebSocketListen{
		g:        g,
		wsChan:   make(chan *websocketConn, 64),
		listener: listener,
	}

	g.GET(uri, func(c *gin.Context) {
		// logs.Debug("web2 accept: %s", c.Request.URL.RequestURI())
		handler := websocket.Handler(
			func(ws_conn *websocket.Conn) {
				if cf := atomic.LoadInt32(&gsl.closeFlag); cf == 0 {
					ws_conn.PayloadType = websocket.BinaryFrame
					// websocket.Message.Send(ws_conn, "test")
					wsc := newWebSocketConn(ws_conn)
					gsl.wsChan <- wsc
					wsc.waitDone()
				}
			})
		handler.ServeHTTP(c.Writer, c.Request)
	})

	//go func() {
	//	http.Serve(l, g)
	//}()

	return gsl, g
}

func (g *GinWebSocketListen) Addr() net.Addr {
	return g.listener.Addr()
}

func (g *GinWebSocketListen) Close() error {

	if atomic.CompareAndSwapInt32(&g.closeFlag, 0, 1) {
		//close all connections in queue
		for i := 0; i < len(g.wsChan); i++ {
			c := <-g.wsChan
			c.Close()
		}
		close(g.wsChan)
	}
	return g.listener.Close()
}

func (g *GinWebSocketListen) Accept() (net.Conn, error) {
	select {
	case wsConn, ok := <-g.wsChan:
		if ok {
			return wsConn, nil
		} else {
			return nil, closeErr
		}
	}
	return nil, nil
}

func NewGinDebugUrl(g *gin.Engine) {
	if g == nil {
		g = gin.Default()
	}
	tools := g.Group("/game/")
	tools.Use(RateLimit())
	tools.GET("/echoip", func(c *gin.Context) {
		clientIP := c.ClientIP()
		clientIPonly := strings.Split(clientIP, ":")[0]
		c.JSON(200, clientIPonly)
	})
}

var (
	rateLimit *ratelimit.TokenLimiter
)

func init() {
	rates := ratelimit.NewRateSet()
	rates.Add(time.Second, 5, 10)

	extr, err := utils.NewExtractor("client.ip")
	if err != nil {
		panic("init ratelimite NewExtractor err: " + err.Error())
	}

	l, err := ratelimit.NewForHttp(
		nil,
		extr,
		rates,
		ratelimit.Clock(&timetools.RealTime{}))
	if err != nil {
		panic("init ratelimite new limiter err: " + err.Error())
	}
	rateLimit = l
}

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := rateLimit.Consume(c.Request); err != nil {
			logs.Warn("[limit] ratelimit ip: %s", c.ClientIP())
			c.JSON(429, "Too Many Requests")
			c.Abort()
			return
		}
		c.Next()
	}
}
