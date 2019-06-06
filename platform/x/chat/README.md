# chattown城镇广播服务器

主要承担城镇多人显示和跑马灯功能

配套测试机器人请使用 titools/chatroomrobot/client

## 对外公开端口

在conf/chat.toml中net_address配置，是客户端websocket连接用的端口，完成城镇广播功能

## 未对外公开端口

为对外公开的限制，由在conf/chat.toml中LimitConfig和limit.InternalOnlyLimit()方法实现

在conf/chat.toml中golang.org/x/net/context配置

/auth - gamex发起，获得accountid对应的认证信息

/broadcast - gamex发起，对指定shard进行消息广播

/filterroom/:shard - 外部发起，如一天执行一次，清理人数过少的房间，但保证至少省一个有空位的房间, 如 127.0.0.1:10011/filterroom/1:16

## 协议

新改版使用FlatBuffers实现协议，协议定义在chatserver/proto/schema中，协议生成在chatserver/proto/gen中

其中协议ProtoWrap是包装协议，由一个string和byte数组组成，string是自己定义的协议id，byte数组是具体协议的字节流，需要再转换成具体协议

客户端使用同样机制处理协议

## 逻辑结构

shard->city->room->player