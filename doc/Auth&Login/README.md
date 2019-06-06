# Game Auth

分为登录认证(Auth)和游戏登录(Login)两个部分

多设备同时游戏的设计：后者踢掉前者

- 只保证**相同gameid**下只能有一个设备在玩
- Auth作为所有游戏的总服务器，gid:user_id 是唯一的
- Gate服务器登录退出需要指令队列化，保证任务按顺序完成?

```sequence
title:Startup & Register Gate
# Startup 
AuthLogin->AuthLogin: Startup
GateGame->GateGame: Startup

# Register Gates
#loop every 10 seconds
GateGame->AuthLogin: Register IP, CCU
```


```sequence
title:Auth:新用户注册，匿名用户绑定注册
Client->AuthLogin: Register(deviceid, username, passwd)
AuthLogin->AuthLogin: AuthTokenNotify(Authtoken, UserID)
AuthLogin-->Client: registered with AuthToken
```

```sequence
title:Auth:普通登录过程
Client->AuthLogin: GetAuthtoken(deviceid) \nor Auth(username, passwd)
AuthLogin->AuthLogin: AuthTokenNotify(Authtoken, UserID)
AuthLogin-->Client: Authtoken
```

```sequence
title:客户端Login游戏服务器功能
Client->AuthLogin: GetShards(Gameid)
AuthLogin-->Client: shardsname, shardsids

Client->AuthLogin: Client choose which shard to play, \nGetGateIP(shardid, Authtoken)

AuthLogin->GateGame: jsonRPC gate ID \n(LoginToken, UserID, GameID, shardID)
AuthLogin-->Client: IP of Gate, Logintoken
note over Client: 检查是否玩家已经在其他Gate上进行游戏，\n如果是，\n则RPC通知相关gate要求前一个玩家退出登录。
```

```sequence
title:客户端Connect连接服务器和握手过程
Client->GateGame: Connect
GateGame-->Client: Accept

Client->GateGame: handshake(Logintoken)
GateGame-->Client: handshake ok (AccountID)
note over Client,GateGame: 游戏正常请求开始
Client->GateGame: Request
GateGame->Client: Response
```

