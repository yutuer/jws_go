package gen2csharp

/*
using System.Text;
using PlanXNet;
using System.Collections.Generic;
using System.IO;

//
// TestProto
// XXX
// XXX
// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

*/

var headerTemplate string = `using System.Text;
using PlanXNet;
using System.Collections.Generic;
using System.IO;

//
// %s
// %s
// %s
// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

`

/*
public class reqMsgTestProto : AbstractReqMsg
{
    public string accid { set; get; } // AccountID info

    public override string ToString()
    {
        return string.Format("reqMsgTestProto - [passthrough:]{0} {1}",
                              passthrough,
                              accid);
    }
}
*/

var reqMsgBegin string = `// %s %s请求消息定义
public class %s : %s
{
`
var reqMsgEnd string = `    public override string ToString()
    {
        return string.Format("%s - [passthrough:]{0}",
                              passthrough);
    }
}
`

/*
// reqMsgTestProto title: rsp msg
type rspMsgTestProto struct {
	SyncRespWithRewards
	RespID int `codec:"respid"` // RespID info
}
*/

var rspMsgBegin string = `// %s %s回复消息定义
public class %s : %s
{
    // 传输用的数据定义, 不要直接使用 `
var rspMsgEnd string = `    public override string ToString()
    {
        return string.Format("%s - [passthrough:]{0} [msg:]{1} [code:]{2}",
                              passthrough,
                              msg,
                              code);
    }
}`

var objectMsgBegin string = `// %s %s回复消息定义
public class %s
{
    // %s传输用的数据定义, 不要直接使用 `

var objectMsgEnd string = `    public override string ToString()
    {
    	return "";
    }
}`

/*
public partial class AttrClient : Client
{
    // TestProto info
    public RpcHandler TestProto(Connection.OnMessageCallback callback, string accid)
    {
        var msg = new reqMsgTestProto();
        msg.callback = callback;
        msg.accid = accid;
        return SendByRpcHandler(msg);
    }
}
*/

var funcBegin string = `public partial class AttrClient : Client
{
	private void RegRpcHandler%sPathHasReg()
	{
		RegMsgPath("%s/%sReq", typeof(%s));
		RegMsgPath("%s/%sRsp", typeof(%s));
	}

    // %s : %s, %s
    public RpcHandler %s(Connection.OnMessageCallback callback%s)
    {
        var msg = new %s();

        msg.callback = callback;`
var funcEnd string = `        return SendByRpcHandler(msg);
    }
}
`
var funcBeginWithCheat string = `public partial class AttrClient : Client
{
	private void RegRpcHandler%sPathHasReg()
	{
		RegMsgPath("%s/%sReq", typeof(%s));
		RegMsgPath("%s/%sRsp", typeof(%s));
	}

    // %s : %s, %s
    public RpcHandler %s(Connection.OnMessageCallback callback, string hackjson%s)
    {
        var msg = new %s();
        msg.hackjson = hackjson;
        msg.callback = callback;`

//    public string accid { set; get; } // AccountID info
var paramTemplate string = "    public %s %s { set; get; } // %s  传输用"

//    msg.accid = accid;
var paramInit string = "        msg.%s = %s;"

var msgDataDefBegin string = `
    public class msgData
    {
        // 使用这里的定义获取回包数据`
var msgDataDefParam string = `        public %s %s; // %s`
var msgDataDefParamWithoutTab string = `    public %s %s; // %s`

var msgDataDefEnd string = `    }
`

var msgDataGetBegin string = `    public %s Data()
    {
        var res = new %s();
`
var msgDataGetArrayParam string = `        res.%s = %s<%s>(%s);`
var msgNetDataGetArrayParam string = `        Net%s[] temp%s = %s<Net%s>(%s);
		res.%s = new %s[temp%s.Length];
		for (int i = 0 ; i < temp%s.Length; i++ )
		{
			res.%s[i] = temp%s[i].Data();
		}
`

/*
   res.Room = new NetData.Room();
   res.Room.Deserialize(new NetData.Deserializer(_p3_));
*/
var msgDataGetObjParam string = `        Net%s tmp = MkObjectFromNet<Net%s>(%s);
        res.%s = tmp.Data();`
var msgDataGetParam string = `        res.%s = %s;`
var msgDataGetEnd string = `
        return res;
    }
`
