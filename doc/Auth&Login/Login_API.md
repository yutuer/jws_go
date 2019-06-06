Client 接口(login/getgate):

1. Client传输{authtoken}， {gid}  {shardname} 到Login服务器
1. Login服务器验证authtoken有效性
1. Login服务器根据Shard上的Gate用户数量找出Gate IP。
1. Login服务器生成logintoken
1. 返回给Gate/Game Server用户信息
    1. Call Gate RPC 传递如下信息
	    - {user_id}@{shard}@{gate}@subid, subid是login服务器随机生成的
    1. 返回给客户端 Gate IP, loginToken
        - loginToken，在后续的Gate操作中使用loginToken进行消息有效性的校验

内部API
1. AuthTokenNotify(/api/authtoken) Auth服务

1. Userlogin(/api/notifylogin): Auth 服务调用， 推送{authtoken} {user_id}过来
1. Userlogout(/api/notifylogout): Gate发现玩家掉线后调用 或者 玩家主动注销。
1. GateRegister(/api/gateregister): Gate服务调用， 推送当前Gate上的玩家数量. ？？如何校验Gate是否还活着？？
	* ZADD gates:{gid}:{sid} counter IP
		* 记录当前登录上来的玩家的数量
		* Gate通过RPC向login服务汇报，数据记录在内存中或者记录在Redis里面
		* ZRANGE gates:{gid}:{sid} 0 0 来获取当前的Gate IP
1. 在线登录状态查询(/api/onlinestatusquery)



