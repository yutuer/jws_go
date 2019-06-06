package chat

/*
import (
	"reflect"

	"vcs.taiyouxi.net/platform/planx/servers"

	"github.com/ugorji/go/codec"

	"vcs.taiyouxi.net/platform/planx/client"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type reqest struct {
	pkt *client.Packet
	t   *client.Transport
}

type Server struct {
	*servers.LimitConnServer

	writers   map[*client.Transport]bool
	router    chan *client.Packet
	routerreq chan reqest
}

func NewServer(listento string, maxconn uint) *Server {
	chatServer := &Server{
		writers:   make(map[*client.Transport]bool, maxconn),
		router:    make(chan *client.Packet),
		routerreq: make(chan reqest),
	}
	chatServer.LimitConnServer = servers.NewLimitConnServer(listento, maxconn)
	return chatServer
}

func (c *Server) incoming(t *client.Transport) {
	r := t.GetPacketChan()
	for {
		select {
		case <-c.Quit:
			logs.Info("router quit")
			return
		case pkt, ok := <-r:
			if ok {
				switch pkt.GetContentType() {
				case client.PacketIDContent:
					c.router <- pkt
				case client.PacketIDReqResp:
					c.routerreq <- reqest{pkt: pkt, t: t}
				}

			} else {
				logs.Info("client transport closed already")
				c.writers[t] = true
				return
			}

		}
	}
}

func (c *Server) broadcast() {
	for {
		select {
		case <-c.Quit:
			logs.Info("broadcast, server quit")
			return
		case pkt := <-c.router:
			//What's the performance it is?
			for t, del := range c.writers {
				if del {
					logs.Info("delete one writer")
					//a little delay but should work
					delete(c.writers, t)
				} else {
					//why it is go? becasue I hope, if anyone blocked, shoudl not affect others.
					go t.Send(pkt)
				}
			}
		case conn := <-c.WaitingConn:
			t := client.NewTransport(conn, c.QuitingConn)
			c.writers[t] = false
			go c.incoming(t)
		}
	}
}

func (c *Server) request() {
	for {
		select {
		case <-c.Quit:
			logs.Info("reqest, server quit")
			return
		case pkt := <-c.routerreq:
			var mh codec.MsgpackHandle
			var (
				action string
				v      map[string]interface{}
			)
			//XXX: this line is important, otherwise it would be map[interface{}]interface{}
			mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
			mh.RawToString = true

			dec := codec.NewDecoderBytes(pkt.pkt.GetBytes(), &mh)
			dec.Decode(&action)
			dec.Decode(&v)

			//TODO: sessionid is needed?
			response := servers.DefaultServeMux.Serve(servers.Request{action, "nosessionid", v})
			transport := pkt.t

			if response != nil && response.Data != nil {
				var mh codec.MsgpackHandle
				mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
				mh.RawToString = true

				var out []byte //XXX: If it comes from []byte buffer pool, it would be cool
				enc := codec.NewEncoderBytes(&out, &mh)
				enc.Encode(response.Code) //TODO: 返回值应该使用不同的类型，所以需要考虑如何传递过来
				enc.Encode(response.Data)

				pkt := client.NewPacket(out, client.PacketIDReqResp)
				transport.Send(pkt)
			}
		}
	}
}

func (c *Server) Start() {

	servers.HandleFunc("ChatRoomNumberRequest", func(r servers.Request) *servers.Response {
		passid := r.Data["passthrough"].(string)
		logs.Info("we hahahah got %s, %s", r.Code, passid)
		r.Data["number"] = 999
		return &servers.Response{
			Code: "ChatRoomNumberRequest",
			Data: r.Data,
		}
	})

	var waitGroup util.WaitGroupWrapper
	waitGroup.Wrap(func() { c.broadcast() })
	waitGroup.Wrap(func() { c.request() })
	waitGroup.Wrap(func() { c.LimitConnServer.Start() })
	waitGroup.Wait()
}
*/
