package gate

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"time"

	"strconv"

	"vcs.taiyouxi.net/platform/planx/client"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/secure"
	ver "vcs.taiyouxi.net/platform/planx/version"
)

type HandShake struct {
	LoginToken  string
	ShardID     int
	GateIP      string
	SubID       uint
	HandShakeID uint
	RandomKey   uint
}

const MAX_LoginToken_Length = 512

// handShake
//TODO 所有异常都应该有监控并搜集玩家IP，防止恶意攻击服务器IP快速被发现 1095
func (g *GateServer) handShake(con net.Conn, agent *client.PacketConnAgent) (success bool, ln LoginNotify, gzipLimit uint64, sessionId int64) {
	success = false
	clientID := con.RemoteAddr().String()
	IpAddr, _, _ := net.SplitHostPort(clientID)

	con.SetReadDeadline(time.Now().Add(time.Second * 15))
	cnc := &util.ConnNopCloser{con}

	// 防止恶意客户端发送超过512字符串的命令
	limit_reader := io.LimitReader(cnc, MAX_LoginToken_Length)
	reader := bufio.NewReader(limit_reader)
	tr := textproto.NewReader(reader)
	writer := bufio.NewWriter(cnc)

	// 建立链接后的第一个消息是客户端发过来的loginToken，否则断开链接
	logs.Trace("<Gate> client [%s], handshake before ReadLine \n", clientID)
	line, err := tr.ReadLine()
	if err != nil {
		//需要实现恶意检测ip计数报告, 判断超出长度情况
		logs.Warn("<Gate> Alarm: Gate Server try to make a handshake failed with client [%s] ,error [%s]", IpAddr, err.Error())
		send_fail(writer)
		return
	}
	logs.Trace("<Gate> client [%s], handshake got loginToken:%s\n", clientID, line)
	loginToken := line
	logs.Trace("<Gate> client %s, queryUserLoginInfo loginToken:%s line:%s\n", clientID, loginToken, line)
	uInfo, sessionId := g.info.queryUserLoginInfo(loginToken, agent)
	if uInfo == nil {
		logs.Warn("<Gate> Alarm: Gate Server try to make a handshake with client [%s], failed with [%s]", IpAddr, "queryUserLoginInfo")
		send_fail(writer)
		return
	}
	ln = *uInfo
	uid := uInfo.UserId
	//gid := uInfo.GameId
	//sid := uInfo.ShardId
	accountID := ln.String() //makeAccountID(gid, sid, uid)
	accountNumberID := ln.NumberID

	gziplimit, err := tr.ReadLine()
	if err != nil {
		//需要实现恶意检测ip计数报告, 判断超出长度情况
		logs.Warn("<Gate> Alarm: Gate Server try to make a handshake failed with client [%s] ,error [%s]", IpAddr, err.Error())
		send_fail(writer)
		return
	}
	gzipLimit, err2 := strconv.ParseUint(gziplimit, 10, 64)
	if err2 != nil {
		logs.Warn("<Gate> Alarm: Gate Server try to make a handshake gzipLimit failed with client [%s] ,error [%s]", IpAddr, err2.Error())
		send_fail(writer)
		return
	}
	channelId, err := tr.ReadLine()
	if err != nil {
		//需要实现恶意检测ip计数报告, 判断超出长度情况
		logs.Warn("<Gate> Alarm: Gate Server try to make a handshake failed with client [%s] ,error [%s]", IpAddr, err.Error())
		send_fail(writer)
		return
	}

	tr.ReadLine() //reserved line 1
	tr.ReadLine() //reserved line 2
	tr.ReadLine() //reserved line 3

	//XXX: 这里的修改需要配合UnitySDK修改和botx的handshake模拟的修改 by YZH
	if uid == db.InvalidUserID {
		logs.Debug("<Gate> Alarm: Gate Server can't find loginToken for current user! Connection(%s) will be closed. login token: %s", IpAddr, loginToken)
		logs.Warn("<Gate> Alarm: Gate Server can't find loginToken for current user! Connection(%s) will be closed.", IpAddr)
		send_fail(writer)
		return
	} else {
		fmt.Fprintf(writer, "ok\r\n")
		//should pass 0:0:1001
		fmt.Fprintf(writer, "%s\r\n", secure.Encode64ForNet([]byte(accountID)))
		fmt.Fprintf(writer, "%d\r\n", time.Now().Unix()) //服务器时间
		fmt.Fprintf(writer, "%s\r\n", ver.GetVersion())
		fmt.Fprintf(writer, "%s\r\n", secure.Encode64ForNet([]byte(strconv.Itoa(accountNumberID))))
		writer.Flush()
		success = true
		logs.Debug("<Gate> Gate Server make a good handshake with client. client:%s user_id:%s, login token:%s, channel %s",
			IpAddr, uid.String(), loginToken, channelId)
		return
	}

	return
}

func send_fail(w *bufio.Writer) {
	fmt.Fprintf(w, "fail\r\n")
	w.Flush()
}
