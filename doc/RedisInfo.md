# Redis数据键值整理
----------------------

Redis数据信息分为 

- 服务器信息 服务器相关的信息 读写量为常数
- 玩家登陆信息 玩家登陆所用的信息 写量与注册量相关 读与玩家登陆次数相关
- 玩家账号信息 玩家数据信息 读与玩家登陆次数相关, 写请参考下面详细信息

##1. 服务器信息

|           键名          |           内容          |  读  |  写  |
|-------------------------|-------------------------|------|------|
| cfg:gamenames           | 未使用                  |      |      |
| cfg:gameids             | authGamePrime信息                | 常数 | 常数 |
| gates:{sid}             | 记录当前登录上来的Gates | 常数 | 每Gate每6s |
| gateinfo:{gate_address} | Gate信息                | 常数 | 常数 |
| gate:{sid}:ip           | TTL: 5min Gate定时注册放保证服务存在   | loginCount | 可配置game.toml， 6s |

##2. 登入信息

loginCount 玩家登陆次数

regCount 玩家注册次数

|        键名       |            内容            |     读     |     写     |
|-------------------|----------------------------|------------|------------|
| cfg:client_shards | shard信息                  | loginCount | 常数       |
| cfg:shards        | 通过shardname对应的信息    | loginCount | 常数       |
| device_total      | 注册数量                   | regCount   | regCount   |
| {did}             | Device信息                 | loginCount | regCount   |
| uid:{avatarid}    | 玩家账号信息               | loginCount | regCount   |
| un:{usersname}    | 用户名对应的账号Id         | loginCount | regCount   |
| at:{uuid}         | AuthToken UUID对应的账号Id, TTL:SetAuthToken 5min | loginCount | regCount   |
| ls:{gid}:{aid}    | 玩家登陆状态信息TTL:24hours| loginCount | loginCount |

##3. 玩家角色信息

读与玩家登陆次数相关 写是 :

 - 每个玩家10次请求写一次
 - 根据请求返回flag可以强制写 TODO
 - 每玩家每30秒写一次
 - 登出时写一次


|       键名       |       内容       |
|------------------|------------------|
| profile:0:0:1001 | 玩家角色基础信息 |
| tmp:0:0:1001     | 玩家角色临时信息 |
| bag:0:0:1001     | 玩家角色背包信息 |
| general:0:0:1001 | 玩家角色副将信息 |
| store:0:0:1001   | 玩家角色商店信息 |
