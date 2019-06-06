package client

var clientHeader = `using System.Text;
using PlanXNet;
using System.Collections.Generic;
using System.IO;
`

var clientReqResp = `public class %sReqMessage : AbstractReqMsg
{
	public protogen.%sReq req;
	public override void Serialize(Stream stream)
	{
		ProtoBuf.Serializer.Serialize(stream, req);
	}
	public override string ToString ()
	{
		return string.Format ("%sReqMessage - [passthrough:]{0}",
			req.Req.PassthroughID);
	}
}

public class %sRespMessage : AbstractRspMsg
{
	public protogen.%sResp resp;
	public override void Deserialize(Stream stream)
	{
		resp = ProtoBuf.Serializer.Deserialize<protogen.%sResp>(stream);
		passthrough = resp.Resp.PassthroughID;
		msg = resp.Resp.MsgOK;
		code = resp.Resp.Code;
		change = resp.Resp.PlayerChange;
		rewards = resp.Resp.Rewards;
	}
	public override string ToString ()
	{
		return string.Format ("%sRespMessage - [passthrough:]{0} ",
			resp.Resp.PassthroughID);
	}
}
`

var clientHanlder = `public partial class AttrClient : Client
{
	private void RegRpcHandler%sPathHasReg()
	{
		RegMsgPath("Attr/%sReq", typeof(%sReqMessage));
		RegMsgPath("Attr/%sRsp", typeof(%sRespMessage));
	}

	public RpcHandler %sRequest (Connection.OnMessageCallback callback%s)
	{
		%sReqMessage msg = new %sReqMessage ();
		msg.callback = callback;
		msg.req = new protogen.%sReq ();
		msg.req.Req = new protogen.Req ();
		%s

		initReq (msg, msg.req.Req);

		return SendByRpcHandler (msg);
	}
}`

var clientPushCode string = `using System.Text;
using PlanXNet;
using System.Collections.Generic;
using System.IO;

public class %sPushMessage : AbstractRspMsg
{
	public protogen.%sPush %sPush;
	public override void Deserialize(Stream stream)
	{
		%sPush = ProtoBuf.Serializer.Deserialize<protogen.%sPush>(stream);
	}
	public override string ToString ()
	{
		return string.Format ("%sPush - [passthrough:]{0} ", %sPush.ToString());
	}
}

public partial class AttrClient : Client
{

	public Connection.OnMessageCallback On%sPush;

	private void RegPushHandler%sPush()
	{
		RegPushPath("Push/%sPush", typeof(%sPushMessage), On%sPushCallback);
	}

	public void On%sPushCallback(object message)
	{
		if (On%sPush != null) {
			On%sPush (message);
		} else {
			Output.Log ("not listened push On%sPush");
		}
	}
}`
