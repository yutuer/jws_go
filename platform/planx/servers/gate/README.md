# 网关服务器设计


## 基本概要

能够维护客户端和服务器之间的链接

需要维护一个SessionID,
每个客户端长链接链接到gate服务器后，gate服务器需要给一个唯一标识（UUID）。

客户端发送流程：

Client --(pkt)--> Gate Server --(sessionID, pkt)-->Game server

客户端接收过程：

Game server --(sessionID, pkt)--> GateServer --(pkt)--> Client 

所以从流程上来看，Gate Server的一些作用和好处

* 实现移动客户端安全握手和登录服务器CCU汇报等机制，使得Game Server逻辑更加简单
* 中间封包，用于维护Game Server和Client之间的消息关系
  * 消息分流：Gate可以让不同的Game Server运行特定的服务，例如（Rank Server只接收Ranking相关的服务）。好处是提高了负载的能力便于规模化
  * 均衡负载扩展：即便是相同的服务，可以有很多个Game
    Server运算，Gate进行RoundRobin类似的负载均衡，使得指定服务能够更好的水平扩展。
* 能够在后端服务器重启的过程中维护玩家链接不掉线，实现热升级
  * 滚动升级
  * 后端服务发现
* 能够有机会在后端Game
  Server出现问题时帮助玩家重发到能够维持服务的其他服务器上，提高用户体验和实现HA

问题：对于开发阶段构架复杂度略高，一般规模的游戏服务器可能不需要Gate的复杂度。
因此如果需要Game Server能够仅需Gate Server的All in One模式。单个Game
server则需要维护所有功能。

为了实现这个功能需要游戏服务器能够实现如下功能，或者说以下模块需要模块化

 - Login服务器的配合（CCU汇报，login）
 - 客户端握手安全机制

## Gate Server的角色

统一Gate集群(需要客户端声明自己gid, sid)

成为所有游戏总服务调度，无论任何分服，每个分服有多少Game
Server，所有玩家都是链接到统一的Gate server集群。

 * 好处：在很多云服务器上可以节省IP资源和主机租用资源
 * 问题：集群稳定性影响所有游戏


每个分服都有自己的Gate集群(传入固定gid, sid)

 * 好处：运维逻辑简单，各个集群之间是隔离的。不容易出现连锁问题。
 * 问题：运维工作量大。
