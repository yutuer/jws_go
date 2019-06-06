# IAP Gamex 概述 

## 简介

角色：CP游戏服务器，logiclog，DynamoMail

主要介绍支付消息到达CP游戏服务器后的处理

## 订单流水号
玩家收到订单后，并生成流水号，流水号就是此时的纳秒数

## DynamoMail表
只有Android会用到，主要用来解耦CP回调服务器和CP游戏服务器

CP回调服务器收到支付消息后，会在DynamoMail表中加一封邮件

在由玩家去接受此邮件，领取里面的钻石

因此Mail中的支付邮件有两种情况:

1、支付回调成功但还没有被领取的邮件

2、已被领取的邮件

## DynamoPay表
### AndroidPayDynamo表

记录内容:

1、记录CP回调服务器的回调信息

2、记录已领取时生成的流水号以及的状态信息

大致包括如下:

订单号，账号id，账号名，角色名，流水号，商品id，商品名，平台类型，渠道id，成交金额，充值状态，获得购买hc数量，获得赠送hc数量，修改到账状态，记录时间

用途:

1、用来对CP回调服务器的回调信息进行去重

2、保存信息供gmtools查询

### IOSPayDynamo表

记录内容:

1、记录已领取时生成的流水号以及的状态信息

大致包括如下:

订单号，账号id，账号名，角色名，流水号，商品id，商品名，平台类型，消费金钱数，获得购买hc数量，获得赠送hc数量，修改到账状态，记录时间

用途:

1、用来对AppStore的返回信息进行去重

2、保存信息供gmtools查询


## Logiclog
将支付成功的订单信息记录成logiclog，方便BIlog查询

## 支付gmtools查询

### 根据accountid查询
* 查看DynamoMail表的状态，查询android支付的未/已到账信息

* 查看DynamoPay表的状态，包括android和ios的到账信息，和gateway的android的回调信息

### 根据accountid和时间段查询
* 根据logiclog查询，android和ios的到账信息

* 根据logiclog查询，某玩家的硬通获得/消耗情况查询

