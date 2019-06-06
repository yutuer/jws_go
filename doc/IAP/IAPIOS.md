# AppStore IOS 支付系统对接

## 简介

1、角色：游戏客户端，CP游戏服务器，AppStore服务器

## 系统开发需求

```sequence
title: IOS 支付流程图
participant CP游戏客户端 as CPClient
participant CP游戏服务器 as CPServer
participant Dynamodb
participant AppStore

CPClient->AppStore: 玩家触发支付
AppStore-->CPClient: receipt_data

CPClient->CPServer: data
CPServer->Dynamodb: 查询订单号
Dynamodb-->CPServer: 返回结果，是否订单号重复

Note over CPServer,AppStore: 如果订单不重复逻辑

CPServer->AppStore: verify
AppStore-->CPServer: verify_res
CPServer->CPServer: 加钻石，生成账单流水号，记录账单

Note over CPServer,AppStore: 如果订单不重复逻辑

CPServer-->CPClient: 通知充值结果
```

IOS支付不需要考虑玩家不在线和CP服务器不在线的情况，只要AppStore没有收到CP游戏服务器的verify，下次CP游戏客户端上线时还会收到AppStore的receipt_data

