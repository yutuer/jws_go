#广播系统架构
-------------------

广播系统当中,对于客户端有 

- 进入房间
- 离开房间
- 换房间
- 说

三种操作

```sequence
title:Register
Client->City: Register 
City->Room: Hash Room
Room->Client: Broadcast Register
```

```sequence
title:Say
Client->Room: Say Msg
Room->Client: Broadcast Msg
```

```sequence
title:UnRegister
Client->Room: UnRegister
Room->Client: Broadcast UnRegister
```

```sequence
title:Dynamic Add Room
Room->City: Room Mem Num > Max_Mem
City->Room: Add New Room and \nKick(Rejoin) Some Players
```

```sequence
title:Dynamic Del Room
Room->City: Room Mem Num < Min_Mem && \nRoom Num > Init_Room_Num
City->Room: Kick(Rejoin) All Players \nIn Room and Del Room
```

## 1. 广播时阻塞问题

每个Player对应一个客户端,起一个单独的gocrountine进行收发包,

每一个Player有一个SendQueue channel, 在Player的gocrountine中会依次取其中的Message发送给客户端.

如果这个SendQueue上的Message满了的话, 视为无法向其发包, 就踢掉这个Player.

这个设计为了解决房间向所有Player广播时可能出现的阻塞问题, 使得整个广播过程中不会因为一个客户端收包慢而卡住


## 2. City和Room设计

根据需求：一个city下多个room，每个room玩家数量有上限，room内player行为互相广播；尽量保证player能回到同一room

服务器收到客户端的注册协议时，city分配room给这个player，之后这个player变进入这个room，并通知room内其他player

服务器收到客户端的说话、移动、换装等协议时，将广播给room内所有player

服务器收到客户端的反注册协议或链接断开时，将直接将player从room中删除，并通知room中其他player

目前服务器采用github.com/serialx/hashring，根据player的uuid进行一致性hash，尽量保证玩家总能回到同一个room；在room没有增减的情况下，player肯定会进入同一room


## 3. 根据City人数动态增减Room数量

当city中room中player数量变化时进行检测：

当room中player人数 > room_mem_max，创建一个新room，并将原room部分player踢掉（客户端会自动重连），服务器会保证player优先进入新创建的空room，在新room人数达到一定人数时，便和其他room一起进行hash

当room中player人数 < room_mem_min，踢掉此room中所有player(客户端会自动重连），并将次room删除


## 4. TODO

目前服务器程序是一个进程，优化方向为city和room是独立进程的分布式结构

city只有一个，和client之间用http通信，用来路由client到正确的room所在的机器上，client和room之间为websocket连接

city上记录着各个room机器上room的人数状态，可以采用room定时上报给city的方式

city根据room的人数，决定创建或删除room

room的机器可以动态增减，增减是主动通知city进行注册/反注册

city根据room的负载，决定在那台room机器上创建新room



 



