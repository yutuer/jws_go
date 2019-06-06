# 同步战斗协议

[TOC]

-------------------------------------------

## 数据

### 1. EquipInfo 玩家装备数据
| 名称      | 类型     | 含义                   |
| :------ | ------ | -------------------- |
| id      | uint   | 当前战斗状态(等待开始\已开始\已结束) |
| tableid | string | 当前状态开始时间             |
| itemid  | string | 当前状态结束时间             |
| count   | long   | 战斗难度                 |
| data    | string | 战斗场景信息               |

### 2. FashionItemInfo 玩家时装数据
| 名称      | 类型     | 含义                   |
| ------- | ------ | -------------------- |
| id      | int    | 当前战斗状态(等待开始\已开始\已结束) |
| tableid | string | 当前状态开始时间             |
| ot      | long   | 当前状态结束时间             |

### 3. AccountInfo 玩家数据(结构上和客户端获取到的Account2Client一致)
| 名称              | 类型                | 含义                |
| --------------- | ----------------- | ----------------- |
| idx             | int               | 玩家的IDX 对应状态信息中的位置 |
| accountId       | string            | 玩家AccountID       |
| avatarId        | int               | 当前玩家角色            |
| corpLv          | uint              | 战队等级              |
| corpXp          | uint              | 战队经验              |
| arousals        | [uint]            | 玩家角色觉醒等级          |
| skills          | [uint]            | 玩家技能等级            |
| skillps         | [uint]            | 玩家修炼技能等级          |
| name            | string            | 玩家昵称              |
| vip             | uint              | 玩家vip等级           |
| avatarlockeds   | [int]             | 玩家已解锁角色           |
| gs              | int               | 玩家战队战力            |
| pvpScore        | long              | pvp积分             |
| pvpRank         | int               | pvp等级             |
| equips          | [EquipInfo]       | 玩家装备              |
| equipUpgrade    | [uint]            | 装备强化等级            |
| equipStar       | [uint]            | 装备升星等级            |
| avatarEquips    | [uint]            | 玩家当前时装            |
| allFashions     | [FashionItemInfo] | 玩家所有的时装           |
| generals        | [string]          | 玩家当前副将            |
| genstar         | [uint]            | 玩家当前副将星级          |
| genrels         | [string]          | 玩家副将羁绊            |
| genrellv        | [uint]            | 玩家副将羁绊等级          |
| avatarJade      | [string]          | 玩家角色宝石            |
| destGeneralJade | [string]          | 玩家神将宝石            |
| dg              | int               | 玩家当前最高神兽ID        |
| dglv            | int               | 玩家当前最高神兽等级        |
| dgss            | [int]             | 玩家神兽技能            |
| guuid           | string            | 玩家公会uuid          |
| gname           | string            | 玩家公会名称            |
| gpos            | int               | 玩家公会职务            |
| post            | string            | 玩家官阶              |
| postt           | long              | 玩家官阶过期时间          |

### 3. AcDataInfo Boss数据(结构上和AcDataList表一致)
略

### 4. PlayerState 玩家状态(数组中以玩家数据中的idx为索引)
| 名称    | 类型   | 含义                                      |
| ----- | ---- | --------------------------------------- |
| state | int  | 玩家状态 1-掉线 2-已退出 3-未准备 4-已准备 5-已死亡 6-战斗中 |
| hp    | int  | 玩家hp                                    |


### 5. BossState Boss状态(数组中以Boss数据中的idx为索引)
| 名称     | 类型    | 含义                         |
| ------ | ----- | -------------------------- |
| hp     | int   | boss hp                    |
| hatred | [int] | boss 仇恨值(数组中以玩家数据中的idx为索引) |

## 协议

### 1. [RPC]进入同步战斗服务器

**Req**

| 名称        | 序列化名称  | 类型     | 含义       |
| --------- | ------ | ------ | -------- |
| AccountID | acID   | string | 玩家账号ID   |
| RoomID    | roomID | string | 请求进入房间ID |
| SecretKey | secret | string | 房间密码     |

**Rsp**

| 名称          | 序列化名称      | 类型            | 含义                   |
| ----------- | ---------- | ------------- | -------------------- |
| GameState   | stat       | int           | 当前战斗状态(等待开始\已开始\已结束) |
| StartTime   | startTime  | long          | 当前状态开始时间             |
| EndTime     | endTime    | long          | 当前状态结束时间             |
| GameClass   | GameClass  | int           | 战斗难度                 |
| GameScene   | GameScene  | string        | 战斗场景信息               |
| PlayerData  | accDatas   | [AccountInfo] | 当前房间中玩家的数据           |
| AcDataList  | acDatas    | [AcDataInfo]  | 当前房间中Boss的数据         |
| PlayerState | playerStat | [PlayerState] | 当前房间中玩家的状态           |
| BossState   | bossStat   | [BossState]   | 当前房间中Boss的状态         |

### 2. [Notify]主动离开战斗服务器(状态算退出)

**Notify**

| 名称        | 序列化名称  | 类型     | 含义       |
| --------- | ------ | ------ | -------- |
| AccountID | acID   | string | 玩家账号ID   |
| RoomID    | roomID | string | 请求进入房间ID |


### 3. [Notify]准备开始战斗

**Notify**

| 名称        | 序列化名称  | 类型     | 含义       |
| --------- | ------ | ------ | -------- |
| AccountID | acID   | string | 玩家账号ID   |
| RoomID    | roomID | string | 请求进入房间ID |


### 4. [Notify]伤害\损失HP通知

**Notify**

| 名称           | 序列化名称     | 类型     | 含义             |
| ------------ | --------- | ------ | -------------- |
| AccountID    | acID      | string | 玩家账号ID         |
| RoomID       | roomID    | string | 请求进入房间ID       |
| PlayerHpDeta | playerHpD | int    | 玩家自身血量变化       |
| BossHpDeta   | bossHpD   | [int]  | 玩家对Boss造成的血量变化 |

### 5. [Push]进入玩家信息通知 <不需要了>

**Push**

| 名称          | 序列化名称      | 类型            | 含义         |
| ----------- | ---------- | ------------- | ---------- |
| PlayerData  | accDatas   | [AccountInfo] | 当前房间中玩家的数据 |
| PlayerState | playerStat | [PlayerState] | 当前房间中玩家的状态 |

### 6. [Push]当前战斗状态
**Push**

| 名称          | 序列化名称      | 类型            | 含义                   |
| ----------- | ---------- | ------------- | -------------------- |
| GameState   | stat       | int           | 当前战斗状态(等待开始\已开始\已结束) |
| StartTime   | startTime  | long          | 当前状态开始时间             |
| EndTime     | endTime    | long          | 当前状态结束时间             |
| PlayerState | playerStat | [PlayerState] | 当前房间中玩家的状态           |
| BossState   | bossStat   | [BossState]   | 当前房间中Boss的状态         |

### 7. [RPC]获取当前战斗状态

**Req**

| 名称        | 序列化名称  | 类型     | 含义       |
| --------- | ------ | ------ | -------- |
| AccountID | acID   | string | 玩家账号ID   |
| RoomID    | roomID | string | 请求进入房间ID |

**Rsp**

| 名称        | 序列化名称     | 类型     | 含义                   |
| --------- | --------- | ------ | -------------------- |
| GameState | stat      | int    | 当前战斗状态(等待开始\已开始\已结束) |
| StartTime | startTime | long   | 当前状态开始时间             |
| EndTime   | endTime   | long   | 当前状态结束时间             |
| GameClass | GameClass | int    | 战斗难度                 |
| GameScene | GameScene | string | 战斗场景信息               |

### 8. [RPC]获取战斗数据

**Req**

| 名称        | 序列化名称  | 类型     | 含义       |
| --------- | ------ | ------ | -------- |
| AccountID | acID   | string | 玩家账号ID   |
| RoomID    | roomID | string | 请求进入房间ID |

**Rsp**

| 名称          | 序列化名称      | 类型            | 含义           |
| ----------- | ---------- | ------------- | ------------ |
| PlayerData  | accDatas   | [AccountInfo] | 当前房间中玩家的数据   |
| AcDataList  | acDatas    | [AcDataInfo]  | 当前房间中Boss的数据 |
| PlayerState | playerStat | [PlayerState] | 当前房间中玩家的状态   |
| BossState   | bossStat   | [BossState]   | 当前房间中Boss的状态 |

### 9. [RPC]获取奖励

**Req**

| 名称        | 序列化名称  | 类型     | 含义       |
| --------- | ------ | ------ | -------- |
| AccountID | acID   | string | 玩家账号ID   |
| RoomID    | roomID | string | 请求进入房间ID |

**Rsp**
| 名称       | 序列化名称    | 类型       | 含义                |
| -------- | -------- | -------- | ----------------- |
| IsDouble | IsDouble | int      | 是否双倍(0-否 1-是)     |
| IsUseHc  | IsUseHc  | int      | 是否使用HC双倍(0-否 1-是) |
| Rewards  | Rewards  | [string] | 奖励ID列表            |
| Counts   | Counts   | [uint]   | 奖励数量列表(已做双倍四倍处理)  |
