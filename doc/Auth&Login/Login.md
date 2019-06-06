## 游戏登录(Login) HTTP


名词解释

 - gid 渠道上的划分，定义为game id/product id，例如iOS 的值为1， googleplay的值为2
 - sid shard id，分服时的分服数量的唯一数字
 - shardname 保存在客户端的随机字符串，服务器记录了这个字符串和sid之间的关系。(安全)
 - handshakeid 客户端第一次链接Gate/Game时由客户端生成的数值，每次断线重新握手+1，老数值不再使用。
 - randomkey 用于握手，游戏服务器在加密流的开头，先回应这个 randomkey ，如果用不匹配的密钥，会被游戏服务器检查出来断开连接
 - subid 是随机生成，且不重复的，当单个登陆允许多重登陆，subid 有实质意义，而在频繁登陆时，处理一些边界情况时也能发挥作用



参考：
 - http://blog.codingnow.com/2014/07/skynet_short_connection.html

登录系统数据库,都是热数据。登录系可以是每一个Shard一个，也可以是所有游戏共用一个。
登录前玩家应该已经选择了要去哪一个分服（shard）去玩游戏。

下面是有关Login, Gate/Game, Client之间通信的协议描述

1. Client 拿着authtoken到login服务器请求login(authtoken, gid,
   shardname). [^snid]
1. Login服务器验证authtoken有效并得知user_id,
   然后根据shardname换算成shardid(sid)。根据sid能够找到有效的Gate服务器的IP列表。返回给客户端
1. Login服务器找到分配给Client的Gate
  1. Login服务器检查当前auth帐号在当前game id范围内，是否在其他Gate服务器上登录,如果有登录，则RPC强制已登录帐号下线。
  1. 分配一个合适的Gate地址给玩家
  1. 告知客户端应该去哪个Gate, 返回GateIP, 登录信息(subid, loginToken),
1. 将登录信息(IPAddress, loginToken)推送到客户端
1. 将(loginToken, user_id)推送到game servers上。
1. 客户端链接握手和断线链接的过程。（客户端和服务器建立链接后第一条信息发送文本"loginToken\n"，否则断开链接）
1. 维持用户的在线状态，至于当用户已经在登陆状态时，维护一个user_id所有登录状态(loginToken)的表。可以踢出已经登录链接
1. 可以接受在线登录状态查询，通过 loginToken/user_id 可以查到该所有在线状态
1. 如果玩家主动离线、Gate/Game Server应该向登陆服务器发送 RPC 请求，注销在线登录状态


[^snid]: 请参考[Shard](Auth&Login/Login_Shard.md)



