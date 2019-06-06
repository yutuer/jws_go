package gate

import (
	"github.com/ugorji/go/codec"
	"vcs.taiyouxi.net/platform/planx/client"
)

func sendKickNotify(sendpkt func(*client.Packet) bool,
	reason string, // 被踢原因
	after int, // 多少秒后服务器主动断开
	nologin int, // 多少秒内不允许再登录
) {
	var out []byte //XXX: If it comes from []byte buffer pool, it would be cool
	var kick []byte

	enck := codec.NewEncoderBytes(&kick, &mh)
	enck.Encode(struct {
		Msg     string `codec:"msg"`
		After   int    `codec:"after"`
		NoLogin int    `codec:"no_login"`
	}{
		reason,
		after,
		nologin,
	})

	enc := codec.NewEncoderBytes(&out, &mh)
	enc.Encode("NOTIFY/KICK")
	enc.Encode(kick)

	pkt := client.NewPacket(out, client.PacketIDContent)
	sendpkt(pkt)
}

type InfoNotifyToClient struct {
	GagTime int64 `codec:"g"`
}

func sendInfoNotify(sendpkt func(*client.Packet) bool, info2client InfoNotifyToClient) {
	var out []byte //XXX: If it comes from []byte buffer pool, it would be cool
	var info []byte

	enck := codec.NewEncoderBytes(&info, &mh)
	enck.Encode(info2client)

	enc := codec.NewEncoderBytes(&out, &mh)
	enc.Encode("NOTIFY/INFO")
	enc.Encode(info)

	pkt := client.NewPacket(out, client.PacketIDContent)
	sendpkt(pkt)
}
