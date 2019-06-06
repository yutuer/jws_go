#用户登录状态的维护和多重登录

## 用户登录状态的维护

Gate自己维护一个KV Cache(loginCache in loginInfo struct) 用来接收来自login服务器的LoginToken和和即将登录的用户的信息。
Cache中的信息代表了**可能**会来Gate服务器进行游戏的玩家。如果玩家长时间没有来到当前Gate，相关的K会再60s后被清除。
所以客户端应该保证，login服务器获取到Gate IP地址后，60s内登录到相关Gate,
否则应该重新进行login api调用。


用户成功登录Gate后(即成功完成handshake)，玩家发送的LoginToken会从这个Cache里面清除，
然后同时发送UpdateLoginStatus的http api(`/notifylogin`)请求给login服务器,
login服务器会利用redis维护一个K(ls:gid:uid)， 保存`struct loginStatus`，
里面包括玩家当前真正登录的Gate服务器的详细信息，以及LOGIN的状态值。

如果玩家断线，即Gate Server上和玩家之间的TCP链接断开，Gate服务器一定会发送一个http api(`/NotifyUserLogout`)给login服务器。
login服务器会设置redis上维护的K(ls:gid:uid)的状态值为LOGOFF。

这个状态值K通常设置24小时的TTL，以便于在意外情况下导致redis上的数据状态和玩家在线情况的不一致(login服务器当机，等等)，
redis服务器能清理删除这些错误的状态。玩家的再次登录触发检查（在玩家GatGate时），自己之前的登录状态用的信息是否是有问题的。
如果有问题则会`DeleteLoginStatus`，删除相关的K。




## 多重登录

利用有效的Authtoken去尝试GetGate的客户端，会先检查当前loginStatus的K(ls:gid:uid)，如果存在**可能**已在线的登录，
则应该通知已登录客户端下线，然后尝试再给当前客户端查找需要登录的Gate IP

## 
