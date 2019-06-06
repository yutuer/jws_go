# Chat Server

---
## Stories

实现一个基本的聊天服务器。

### 0.0.1
服务器目前只实现一个基本功能 -- 连接进来的客户端可以聊天

 * 传输协议：文本
 * 多Room：无
 * 限定服务器容纳人数
 * 目的：这个服务器的进化开始提取client通用部分，然后实现edge server

### 0.0.2
 * 封装协议: 二进制传输
 * 提取通用edge server, 提取通用服务器listen机制

### 0.0.3
 * 登录验证

### 1.0.0
 * 传输协议：二进制 + 压缩 + 加密
 * 多Room：支持
 * 优化广播：广播过程定时按计划压缩推送到所有玩家 


---
## DONE
* 上线限定可以考虑用 buffer channel 来实现： go 1.3 http://golang.org/doc/effective_go.html#channels

---

## TODO LIST

* chat机器人，进行压力测试
* 从config来处理配置
* 高效率聊天室


## Notes

### 如何高效率的进行聊天室消息的广播？

一个消息进来，然后服务器立即进行转发，是不高效率的，是CPU/IO不友好的.

“立即”模式，如果发消息的人很多，立即会出发频繁的CPU轮询，导致频繁的IO。比较理想的情况下是所有人都被均匀的分在CPU时间片上并且打包压缩了**一个阶段内**所有需要发送的消息组。当然，这个模式应当会更加消耗CPU。

如何分片？
服务器Accept一个链接，这个链接就由一个timer来管理。同时，这个链接组应该还有一条自己独立的消息缓冲.
时间到来，所有消息缓冲打包成一个大消息，批量发送到所有客户端。

接下来的问题是，一个服务器分多少个组，每个组管理多少个链接合适呢？

也许是PubSub模式
* 本地pubsub https://github.com/tuxychandru/pubsub 不能够扩展，只能单服使用
* 支持redsi https://github.com/Rafflecopter/golang-messageq 可以由多个Chat edge组成一个大的集群。

