# QuickSDK 安卓支付系统对接

## 简介
本系统和IAP统计服务器是两个不同的服务，数据库的设计目的也不一样。本系统只用来保证，Android的各种支付成功后，最终一致性：服务器能够**最终**保证玩家拿到支付结果。

中国安卓支付渠道比较复杂，最复杂的例如短信充值。短信充值不像互联网充值，当玩家在客户端上支付后，客户端SDK往往能够自己确认返回回掉。用户短信充值后，客户端只能通过轮询等待第三方支付回掉服务器的返回信息。有可能1分钟，也有可能10分钟。


### 角色：
- Quick客户端SDK
- Quick支付服务器
- CP回掉服务器
- CP游戏服务器, MailDynamo表

### QuickSDK支付接入

CP需要提供一个唯一指定的回掉URL给Quick服务器配置。
所有Android渠道统一一个,通过POST请求传输XML文件,具体格式请参考相关Quick服务器SDK对接文档。

  1. 游戏方利用xml的数据进行自身业务.成功后回写SUCCESS.失败回写FAILED.
  1. SDK服务器请求同步失败(返回不为SUCCESS)时. 天象SDK服务器会对以一定的机制再次补发请求. 服务器请求游戏方进行同步失败后.
   - 会在当天之内.以1分钟、4分钟、16分钟、64分钟、4小时、14小时间隔进行再次补单.
   - 其中任意一次补单成功后.后续将不再补单.若全部不成功.最多会补单6次.
  
  服务器认为同步失败的条件为.同步返回不为”SUCCESS”.可能为FAILED、timeout、域名不能解析等



## 系统开发需求

1. 实时性：尽可能的保证实时性，如果所有服务器正常，玩家在线等待时，应该尽可能的保证实时性。（短信支付例外）
1. 可用性：游戏服务器不在线或者玩家不再线的情况，应该能够保证玩家下次上线后能够收到购买的产品。
  1. 并尽可能清晰的通过客户端提醒玩家。
  1. 服务器的日志能够清晰的给出日志，方便客服查询。

第一阶段的起点: 客户端发起支付，时序图如下

```sequence
title:客户端发起支付(第一阶段)
participant CP游戏客户端 as CPClient

participant Quick客户端SDK as QuickSDK
participant Quick支付服务器 as QuickServer

CPClient->QuickSDK: 玩家触发支付
note left of CPClient: 玩家触发支付参数\n(game_order, extras_params), \ngame_order，extras_params \n游戏在调用QucikSDK发起支付时传递的，\nCP服务器会收到这两个参数。
QuickSDK->QuickSDK: 弹窗引导完成付费（钱扣掉！）
QuickSDK-->CPClient: 通知CP客户端支付完成(不一定会有)
QuickSDK->QuickServer: 通知玩家支付成功

```

当第一阶段客户端的支付流程完成后，客户端只能进入等待并轮询的方式等待阶段（第二阶段）。
在CP的回掉服务器也等待着Quick支付服务器的“支付成功”回掉。
最后经过去重等处理，把Quick的支付成功信息先写入DynamoDB数据库。


```sequence
title:支付并写入数据库：CP游戏服务器和玩家都正常在线
#participant CP游戏客户端 as CPClient
participant Quick客户端SDK as QuickSDK
participant Quick支付服务器 as QuickServer
participant CP回掉服务器 as CPPayCBServer
participant CP游戏服务器 as CPServer
participant MailDynamo

QuickSDK->QuickServer: 通知玩家支付成功
QuickServer->CPPayCBServer: channel， channel_uid， \ngame_order， order_no，\npay_time， amount， \nstatus， extras_params


CPPayCBServer->CPPayCBServer: 检查重复并保存到数据库\n(+tistatus:Paid,\n+uid(gid:uid:uuid))
CPPayCBServer->MailDynamo: 订单数据保存到Mail表

CPPayCBServer-->QuickServer: 返回SUCCESS
#在线打开着UI并定时拉取Mail，直到收到支付邮件或超时
```

客户端等待阶段（第二阶段）
这里有必要说明一下客户端在`请求邮件`后，客户端发现信息中有支付相关信息，主动发送ReceiveMail请求，游戏服务器的ReceiveMail根据邮件是支付类型触发逻辑，处理支付状态给玩家发放购买的东西。

因为Mail系统的异步同步特点，CP游戏服务器不需要在这个请求过程中处理DynamoDB的数据状态。Mail系统会根据玩家内存中的邮件状态，必要时自动同步到DynamoDB数据库。


```sequence
title:客户端定时UI中Pull拉取邮件(登录上线也会有第一时间拉取邮件)
participant CP游戏客户端 as CPClient
participant CP游戏服务器 as CPServer
participant MailDynamo
CPClient->CPServer: 请求邮件
CPServer->MailDynamo: 拉取邮件
MailDynamo-->CPServer: 返回所有邮件
CPServer-->CPClient: 返回邮件
CPClient->CPServer: ReceiveMail
CPServer-->CPServer: 加钻石，标记Mail已读，生成账单流水号，记录账单
CPServer-->CPClient: 同步钻石信息
```

根据QuickSDK文档说明status返回值可能是0或者1， 0为成功充值。1为充值失败。所以失败的重置这里是不保存到数据库的。
需要和英雄确认。

### 借助Mail完成订单充值

建个新的邮件类型用来保存订单数据，客户端并不显示此类邮件，收集就冲相应的购买钻

### 玩家在线情况

客户端在支付完成后，首次等待200ms拉取一次邮件，若没有支付邮件则，再等待500ms再次拉取，间隔时间依次是200ms，500ms，1s，总共持续10s，若期间成功拉取到支付邮件，则停止此查询过程

按照DAU是10000人，CCU是DAU的10%即1000人，支付的人是CCU的3%的话，即30人，最差情况是每人每秒2次查询，即总共60次/秒

### 玩家不在线情况

当玩家在支付完后，在没有等到支付邮件就下线了，下次上线就会拉取邮件，获取支付邮件

## 如何使用game_order 和 extras_params

- game_order：用户购买的IAP的唯一ID， 例如HC100
- extras_params: uid(gid:sid:uuid):timestamp

## DynamoDB 的数据库设计

1. 回掉有可能被调用多次，所以主键应该能够去重
1. 需要建立Global Secondary Index，accountid->order_no, 方便根据accountid查询订单[^GSI]

[^GSI]: http://docs.amazonaws.cn/amazondynamodb/latest/developerguide/GSI.html
### DynamoDB Table
#### Key

order_no, 满足Quick支付服务器多次回掉中都能快速查询当前订单的状态

#### Attributes

回掉发过来的基础属性channel， channel_uid， game_order, pay_time， amount， status， extras_params,

维护支付状态的属性:
 - tistatus: Paid 或者 Delivered
 - uid: gid:sid:uuid  （从extras_params中解析出来的）
 - receiveTimestamp: 接收到回掉的时间, 即创建本记录的时间

#### Key

- hash: uid: gid:sid:uuid

#### Attributes

- order_no
- game_order
- amount


