#tiprotogen使用文档
=====================
###概述
tiprotogen是一个自动生成客户端服务器网络接口API定义代码的小工具,
tiprotogen定义了如下的以json为格式的协议描述文件:

```json
{
  "name"    : "TestProto",
  "title"   : "测试生成协议",
  "path"    : "Attr",
  "comment" : "用来测试生成代码的协议",
  "req": {
    "params": [
        ["AccountID",  "string",   "膜拜的玩家ID", "acid"   ],
        ["ReqID2s",    "string[]", "请求2ID",     "reqid2" ]
    ]
  },
  "rsp": {
    "base": "WithRewards",
    "params": [
        ["ResAccountID", "string",   "膜拜的玩家ID"            ],
        ["ReqID2s",      "long[]",   "请求2ID",      "reqid2" ]
        ["ReqObj,        "NetData.Room", "Room信息"]
    ]
  }
}
```

通过这个文件,工具可以生成客户端C# API实现代码和服务器端golangAPI定义代码

对于客户端C#, 上述文件会生成如下代码:

```C#
using System.Text;
using PlanXNet;
using System.Collections.Generic;
using System.IO;

//
// TestProto
// 测试生成协议
// 用来测试生成代码的协议
// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgTestProto 测试生成协议请求消息定义
public class reqMsgTestProto : AbstractReqMsg
{

    public string acid { set; get; } // 膜拜的玩家ID  传输用
    public string[] reqid2 { set; get; } // 请求2ID  传输用
    public override string ToString()
    {
        return string.Format("reqMsgTestProto - [passthrough:]{0}",
                              passthrough);
    }
}

// rspMsgTestProto 测试生成协议回复消息定义
public class rspMsgTestProto : AbstractRspWithRewardMsg
{
    // 传输用的数据定义, 不要直接使用 
    public string resaccountid { set; get; } // 膜拜的玩家ID  传输用
    public List<object> reqid2 { set; get; } // 请求2ID  传输用
    public byte[] reqobj {set; get; } //

    public class msgData
    {
        // 使用这里的定义获取回包数据
        public string ResAccountID; // 膜拜的玩家ID
        public long[] ReqID2s; // 请求2ID
        public NetData.Room ReqObj; //
    }

    public msgData Data()
    {
        var res = new msgData();

        res.ResAccountID = resaccountid;
        res.ReqID2s = MkArrayFromObjList<long>(reqid2);
        res.ReqObj = new NetData.Room()
        res.ReqObj.Deserialize(new NetData.Deserializer(reqobj));

        return res;
    }

    public override string ToString()
    {
        return string.Format("rspMsgTestProto - [passthrough:]{0} [msg:]{1} [code:]{2}",
                              passthrough,
                              msg,
                              code);
    }
}
public partial class AttrClient : Client
{
    private bool isTestProtoPathHasReg = false;

    // TestProto : 测试生成协议, 用来测试生成代码的协议
    public RpcHandler TestProto(Connection.OnMessageCallback callback, string acid, string[] reqid2)
    {
        // 当没注册path时注册path
        if (!isTestProtoPathHasReg) {
            RegMsgPath("Attr/TestProtoReq", typeof(reqMsgTestProto));
            RegMsgPath("Attr/TestProtoRsp", typeof(rspMsgTestProto));
            isTestProtoPathHasReg = true;
        }

        var msg = new reqMsgTestProto();

        msg.callback = callback;
        msg.acid = acid;
        msg.reqid2 = reqid2;

        return SendByRpcHandler(msg);
    }
}

```

==========================
###协议描述文件
使用工具,需要使用协议描述文件来定义协议,
协议描述文件为JSON格式,一个文件只描述一个协议.
协议描述文件分三部分: 基本信息, 请求定义, 回复定义.
**注意:所有英文名不要以下划线开头\结尾**
#### 1.基本信息
需要以下字段:

|     名称      |    描述            |   限制                         | 
|--------------|-------------------|-------------------------------|
| name         | 协议英文名称,用来作为API函数名  | 大写开头,驼峰格式,如"TestProtoName"|
| title        | 协议中文名                   | 如 "测试API协议"      |
| path         | 路径,用来区分协议种类,如无特殊需求统一是 "Attr"  | 大写开头字符串|
| comment      | 协议注释                | 一句话描述协议是干嘛的,不要为空 |

如:

```json
{
  "name"    : "TestProto",
  "title"   : "测试生成协议",
  "path"    : "Attr",
  "comment" : "用来测试生成代码的协议",
  //其他信息略...
}
```

#### 2.请求
请求统一在req字段下, 目前只有一项params, 是一个数组, 用来表示协议中的参数.

如:

```json
  "req": {
    "params": [
        ["AccountID", "string", "膜拜的玩家ID", "acid"  ],
        ["ReqID2s",   "string[]", "请求2ID",      "reqid2" ]
    ]
  },
```

其中参数是一个字符串的数组, 各项定义如下:

|     Index    |    描述            |   限制                         | 
|--------------|-------------------|-------------------------------|
| 0 | 参数正式英文名称  | 大写开头,驼峰格式,如"AccountID"|
| 1 | 参数类型, 只支持long,string,long[],string[]四种| 只能是四种之一 |
| 2 | 参数注释 | 描述参数是干啥的,不要为空 |
| 3 | (可选)参数英文名称简写 | 用来在传输协议中表明参数,需要全小写 |

#### 3.回复
回复在rsp字段下,有两项:base和params, 
base项用来标识是否需要返回奖励增量数据, 
params定义回复参数,结构和req中一致.

如

```json
  "rsp": {
    "base": "WithRewards",
    "params": [
        ["ResAccountID", "string",   "膜拜的玩家ID"],
        ["ReqID2s",      "long[]", "请求2ID",      "reqid2" ]
        ["ReqObj,        "NetData.Room", "Room信息"]
    ]
  }
```

其中参数是一个字符串的数组, 各项定义如下:

|     Index    |    描述            |   限制                         | 
|--------------|-------------------|-------------------------------|
| 0 | 参数正式英文名称  | 大写开头,驼峰格式,如"AccountID"|
| 1 | 参数类型, 只支持long,string,long[],string[]四种, 参数类型支持继承于NetObj复杂的数据结构| 只能是四种之一, 其他视作NetObj的基类 |
| 2 | 参数注释 | 描述参数是干啥的,不要为空 |
| 3 | (可选)参数英文名称简写 | 用来在传输协议中表明参数,需要全小写 |

对于我们的回复包中, 很多时候一个协议的处理过程中会导致资源产出, 那么客户端需要显示产出的资源,
如领取任务协议,客户端需要显示玩家都具体获得了什么道具,
为了方便显示这些信息,我们有一个统一的数据结构:

```C#
        public List<object> rids { set; get; }
        public List<object> rcs  { set; get; }
        public List<object> rds  { set; get; }

        public RewardsFromServer rewards;
```

这里面 rids rcs rds 是用来传输用的, 客户端通过rewards来获取到底得到了什么.

为了减少重复代码, 回包有一个基类:

```C#
    public class AbstractRspWithRewardMsg : AbstractRspMsg {
        public List<object> rids { set; get; }
        public List<object> rcs  { set; get; }
        public List<object> rds  { set; get; }

        public RewardsFromServer rewards;

        public override void Deserialize(Stream stream)
        {
            base.Deserialize(stream);
            var rewardIDs    = MkArrayFromObjList<string>(rids);
            var rewardCounts = MkArrayFromObjList<long>(rcs);
            var rewardDatas  = MkArrayFromObjList<string>(rds);

            rewards = new RewardsFromServer(rewardIDs.Length);
            rewards.loadFromNet(rewardIDs, rewardCounts, rewardDatas);

            Output.Logf("rewards : {0}", rewards.ToString());
        }
    }
```
如果是有奖励信息的回包,其回复消息继承自 AbstractRspWithRewardMsg, 而不是 AbstractRspMsg

base项用来表示回包消息是否是附带奖励信息的, 如果是**WithRewards** 则会附带奖励信息, 如果为空,或者没有base项, 就是通常的协议.

很多时候需要给客户端返回比较复杂的数据结构, 返回包支持返回实现NetObj接口的复杂数据信息, 
目前生成代码时会将所有除了四种基本类型之外的类型视为实现NetObj接口的类型.

===============================
###C#生成代码使用方法

生成的C#代码和之前手写的定义相似, 分三部分: 请求消息定义, 回复消息定义和API函数定义.
客户端需要使用的是API函数和回复消息

**1.API函数**
工具会生成参数列表(按照json中的顺序),直接调用即可.
**2.回复消息使用**
对于带奖励的回包,可以取rewards项来获得奖励内容.
对于其他回包信息,会生成如下代码

```C#
    public class msgData
    {
        // 使用这里的定义获取回包数据
        public string ResAccountID; // 膜拜的玩家ID
        public long[] ReqID2s; // 请求2ID
    }

    public msgData Data()
    {
        var res = new msgData();

        res.ResAccountID = resaccountid;
        res.ReqID2s = MkArrayFromObjList<long>(reqid2);

        return res;
    }
```

调用Data()函数获取回包对应的msgData结构,从其中获取信息.

==============================
###客户端工具使用说明

因为客户端同学只需要生成客户端代码,所以专门构建了一个供客户端使用的命令行工具 genclient, 可以在公司共享文件夹的tiprotogen目录找对应的版本
genclient不需要配置及参数, 直接运行即可, 这个程序根据自己目录下(包括子目录)所有的*.json\*.JSON 协议文件在./gen/目录中生成C#代码.

