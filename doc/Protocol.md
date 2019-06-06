服务器-客户端协议文档
========================================
注意：更新时别忘了更新 http://wiki.taiyouxi.net/w/3k_engineer/client/net_protocol/
<!-- toc -->

## 游戏玩家账号逻辑协议 Account级别


### 1. 获取玩家基本信息

通过此接口客户端可以获取一些玩家基本信息

#### 1.1 消息定义

**Req** RequestGetPlayerAttr /PlayerAttr/GetPlayerAttributesRequest

|     名称     |    序列化名称     |   类型   |        含义        |
|--------------|-------------------|----------|--------------------|
| ProfileID    | profileid         | string   | 玩家账号ID         |
| AttrNameList | attributesKeyList | []string | 请求数据名称的列表 |

**Rsp** ResponseGetPlayerAtt /PlayerAttr/GetPlayerAttributesResponse

|  名称  | 序列化名称 |          类型          |             含义            |
|--------|------------|------------------------|-----------------------------|
| Result | result     | map[string]interface{} | 请求的数据的map，以名称索引 |

**错误码**

- 0 成功
- 1 失败 获取信息失败

**msg**

- "ok" 成功
- "no" 失败 获取信息失败

#### 1.2 客户端接口

```
void GetAttrRequest(List<object> attrList)
```

attrList 请求信息名称的列表

#### 1.3 客户端命令
暂无

### 2. 设置玩家基本信息
通过此接口客户端可以获取一些玩家基本信息

#### 2.1 消息定义

**Req** RequestSetPlayerAtt /PlayerAttr/SetPlayerAttributesRequest

|    名称  | 序列化名称 |          类型          |             含义            |
|---------|------------|------------------------|--------------------------|
| ProfileID  | profileid | string                 | 玩家账号ID               |
| Attributes | attributes| map[string]interface{} | 修改的数据的map，以名称索引 |

**Rsp** Resp /PlayerAttr/SetPlayerAttributesResponse

**错误码**

- 0 成功
- 101 失败 修改信息失败

**msg**

- "ok" 成功
- "no" 失败 修改信息失败

#### 2.2 客户端接口

```
void SetAttrRequest()
```
#### 2.3 客户端命令
暂无

### 3. 申请进入关卡

在进入关卡之前需要发送这个请求获取掉落数据

#### 3.1 消息定义
**Req** RequestPrepareLootForLevelEnemy /PlayerAttr/GetLvlEnmyLootRequest

| 名称            | 序列化名称          | 类型         | 含义   |
| -------------- | ------------------ | ----------- | ---------------------- |
| LevelID        | levelid            | string      | 关卡ID |

**Rsp** ResponsePrepareLootForLevelEnemy /PlayerAttr/GetLvlEnmyLootResponse

| 名称            | 序列化名称  | 类型                    | 含义  |
| -------------- | ---------- | ---------------------- | ----------- |
| Result         | result     | map[string][]loot | 返回敌兵loot信息:代表[(掉落ID1,内容)， (掉落ID2,内容)]] |
| ResultN        | resultN     | map[string]int | 返回敌兵最大数量 代表[(掉落ID1,敌兵最大数量)...] |

说明：Result中value为[]loot, 其中loot结构如下

```
type loot struct {
    ID    uint16 `codec:"id"`     掉落id
    Data  string `codec:"data"`   掉落物品数据（现在仅为ItemId）
    Count uint16 `codec:"count"`  掉落数量
}
```

**错误码**

- 0 成功
- 201 失败:关卡ID不存在
- 202 失败:关卡掉落信息存储错误

**msg**

- "ok" 成功
- "no" 失败 关卡ID不存在

#### 3.2 客户端接口

```
void GetPlayLevelEnmyLoot (Connection.OnMessageCallback callback, 
                           string levelid)
```

levelid为关卡id

#### 3.3 客户端命令
按钮 PlayLevelDeclareLoot

当命令框中文本不为空时，申请Id为命令框文本的关卡,
否则申请“k1_1”

### 4. 申请关卡结算

关卡结算之后利用其向服务器段申请结算

#### 4.1 消息定义

**Req** RequestDeclareLootForLvlEnmy /PlayerAttr/DeclareLvlEnmyLootRequest

|    名称   | 序列化名称 |   类型   |       含义       |
|-----------|------------|----------|------------------|
| LootIDs   | lootids    | []uint16 | 确认掉落的掉落id |
| AvatarIDs | avatar_id  | []int    | 参战武将id       |
| IsSuccess | is_success | bool     | 是否关卡成功     |

**Rsp** ResponseDeclareLootForLvlEnmy /PlayerAttr/DeclareLvlEnmyLootResponse

|     名称     | 序列化名称  |      类型     |               含义               |
|--------------|-------------|---------------|----------------------------------|
| StageRewards | rewards     | []StageReward | 掉落奖励                         |
| ScType       | sc_t        | int           | 增加的Sc的类型                   |
| ScValue      | sc_v        | int64         | 增加的Sc的值                     |
| AvatarXp     | avatar_xp   | uint32        | 增加的角色Xp（每个武将都会增加） |
| CorpXpAdd    | corp_xp_add | uint32        | 增加的战队等级经验               |

说明：StageReward结构如下

```
type StageReward struct {
    Item_id string `codec:"id"`     掉落的ItemId
    Count   uint16 `codec:"count"`  掉落的数量
}
```

IsSuccess 为false时，即认为关卡失败，不会发奖励

**错误码**

- 0 成功
- 201 失败:没有申请过开启关卡
- 102 警告:奖励物品发送失败
- 103 警告:申请的掉落ID无效

**msg**

- "ok" 成功
- "no" 失败

#### 4.2 客户端接口

```
void DeclarePlayLevelEnmyLoot (Connection.OnMessageCallback callback, 
                                        List<ushort> lootids,
                                        bool is_success,
                                        List<int> avatar_ids)
```

lootids为确认掉落的掉落id

is_success 是否关卡成功

avatar_ids 参战武将id

#### 4.3 客户端命令
按钮 PlayLevelDeclareLoot,
会默认所有的东西都掉落

### 5. 改变当前装备 

请求改变某一个角色的装备

#### 5.1 消息定义

**Req** RequestChangeEquip /PlayerAttr/ChangeEquipRequest

|名称            | 序列化名称          | 类型         | 含义   |
|-------------- | ------------------ | ----------- | ---------------------- |
|AvatarID       | avatar_id          | int         | 改变装备的武将的ID |
|Equips         | avatar_equip       | []uint32    | 改变之后的装备 |

**Rsp** ResponseChangeEquip /PlayerAttr/ChangeEquipResponse

**错误码**

- 0 成功
- 201 失败:装备不存在
- 202 失败:装备状态错误：同一件装备不能装备两次
- 203 失败:装备位置与类型不一致
- 204 失败:装备类型信息无效
- 105 警告:装备不能被脱下

**msg**

- "ok" 成功

#### 5.2 客户端接口

```
void ChangeEquip (Connection.OnMessageCallback callback, 
                  int avatar, List<uint> equips)
```

avatar改变装备的武将的ID,
equips改变之后的装备

#### 5.3 客户端命令

命令 Equip:[AvatarId],[Slot],[WeaponId]

会将AvatarId武将的Slot位置装备WeaponId，其他位置统统卸掉

### 7. 获取玩家体力信息
获取玩家体力信息

#### 7.1 消息定义
**Req** RequestGetEnergy /PlayerAttr/GetEnergyRequestRequest

**Rsp** ResponseGetEnergy /PlayerAttr/GetEnergyRequestResponse

| 名称          | 序列化名称  | 类型                    | 含义   |
| ------------ | ------------ | -------------- | ----------- |
| Value        | value        | int64  | 当前体力值 |
| RefershTime  | refersh_time | int64  | 服务器刷新体力的时刻 unix时间戳 |
| LastTime     | last_time    | int64  | 上次刷新体力时未结算入体力的时间间隔 |

LastTime即为下图的s, RefershTime为下图的now

```
                   + s +
     ----+---------+---+---+--------
        last     this now  next
     ----+---------+---+---+--------
         +    add  +   + r +

                   + one   +
```


**错误码**

- 0 成功
- 401 失败 获取信息失败

**msg**

- "ok" 成功

#### 7.2 客户端接口

```
void GetEnergy (Connection.OnMessageCallback callback)
```

#### 7.3 客户端命令
暂无

### 9. 卖道具 

#### 9.1 消息定义

**Req** RequestBagSell /PlayerAttr/SellRequest

| 名称            | 序列化名称    | 类型       | 含义   |
| -------------- | ------------ | --------- | ---------------------- |
| SellList       | sells        | []uint32  | 卖出物品的bagid列表 |
| CountList      | counts       | []uint16  | 卖出物品的数量列表，索引与上面一一对应 |

**Rsp** ResponseBagSell /PlayerAttr/SellResponse

|名称          | 序列化名称  | 类型                    | 含义   |
|------------ | --------- | ------- | ----------- |
|NewSc        | int64     | new_sc  | 新增金钱（类型为0的软通）的数量 |


**错误码**

- 0 成功
- 201 失败:要买的物品不是道具
- 202 失败:服务器没找到道具对应的数据（ItemID不正确）
- 203 失败:玩家没有足够的道具

**msg**

- "ok" 成功

#### 9.2 客户端接口

```
 void Sell(Connection.OnMessageCallback callback, uint[] ids, ushort[] counts)
```

ids 卖出物品的bagid列表

counts 卖出物品的数量列表，索引与上面一一对应

例如 ids [1,2,3] counts [11,22,33]，

表示卖出道具id为1的道具11个，卖出道具id为2的道具22个，卖出道具id为3的道具33个

#### 9.3 客户端命令

命令 Sell:[itemid1],[count1],[itemid2],[count2],...[itemidN],[countN],...

如果卖出道具id为1的道具11个，卖出道具id为2的道具22个，卖出道具id为3的道具33个, 命令为

```
Sell:1,11,2,22,3,33
```
### 10. 升级精炼 

升级精炼装备位置

#### 10.1 消息定义

**Req** RequestEquipOp /PlayerAttr/EquipOpReq

| 名称            | 序列化名称          | 类型         | 含义   |
| -------------- | ------------------ | ----------- | ---------------------- |
| Op        | op   | int         | 操作类型 3为升级 4为精炼 |
| Avatar_Id | aid  | int         | 武将id |
| Slot      | slot | int         | 位置 |
| P1        | p1   | int         | 升的等级（升了几级就是几） |

**Rsp** ResponseEquipOp /PlayerAttr/EquipOpRsp

说明：装备升级和精炼目前是针对位置的

**错误码**

升级

- 0 成功
- 201 失败:达到上限，不能超过角色等级
- 202 失败:达到上限，不能超过装备最大等级
- 203 失败:没有足够的物品
- 204 失败:升级精炼位置错误，没有装备
- 205 失败:装备信息缺失

精炼

- 0 成功
- 201 失败:达到上限，不能超过角色等级
- 202 失败:达到上限，不能超过装备最大等级
- 203 失败:没有足够的物品
- 204 失败:升级精炼位置错误，没有装备
- 205 失败:装备信息缺失

**msg**

- "ok" 成功

#### 10.2 客户端接口
升级

```
RpcHandler EquipUpgrade (Connection.OnMessageCallback callback, 
                                    int avatar_id, int slot, int lv_add)
```

- avatar_id 武将id
- slot 位置
- lv_add 升的等级（升了几级就是几）

精炼

```
RpcHandler EquipEvolution (Connection.OnMessageCallback callback, 
                                      int avatar_id, int slot, int lv_add)
```
- avatar_id 武将id
- slot 位置
- lv_add 升的等级（升了几级就是几）

#### 10.3 客户端命令

升级命令 Upgrade:[avatar_id],[slot],[lv_up]

精炼命令 Evolution:[avatar_id],[slot],[lv_up]

### 13. 装备熔炼 
熔炼几件装备

#### 13.1 消息定义
**Req** RequestBagEquipResolve /PlayerBag/EquipResolveReq

|名称            | 序列化名称          | 类型         | 含义   |
|-------------- | ------------------ | ----------- | ---------------------- |
|Equips         | equips             | []uint32    | 要熔炼的装备的列表 |

**Rsp** ResponseBagEquipResolve /PlayerBag/EquipResolveRsp

**错误码**

- 0 成功
- 201 失败:要买的物品不是装备
- 202 失败:服务器没找到装备对应的数据（ItemID不正确）
- 203 失败:服务器没找到装备对应的熔炼数据（ItemID不正确）
- 204 失败:玩家没有这件装备
- 205 失败:玩家装备已锁定(当前已装备中)
- 406 错误:给予物品时失败

**msg**

- "ok" 成功

#### 13.2 客户端接口
```
RpcHandler EquipResolve(Connection.OnMessageCallback callback, uint[] ids)
```
ids 要熔炼的装备的列表
注意当失败时返回包中不会全量更新

#### 13.3 客户端命令

命令 EquipResolve:[EquipId1],[EquipId2],[EquipId3],...

会将EquipId1,...,EquipIdN都熔炼

### 15. 获取玩家信息 

通过这个接口获取一些需要刷新的数据

#### 15.1 消息定义

**Req** RequestGetInfo /PlayerAttr/GetInfoRequest
|      名称      |  序列化名称  | 类型 |              含义              |
|----------------|--------------|------|--------------------------------|
| NeedBag        | need_bag     | bool | 背包数据需要更新               |
| NeedSc         | need_sc      | bool | 软通数据需要更新               |
| NeedAvatarExps | need_avatars | bool | 角色经验需要更新               |
| NeedCorp       | need_corp    | bool | 战队信息（主要是等级）需要更新 |
| NeedEnergy     | need_energ   | bool | 体力信息需要更新               |
| NeedStageAll   | need_stage   | bool | 关卡信息需要全量更新           |


**Rsp** ResponseGetInfo /PlayerAttr/GetInfoResponse

**说明**

走通用的同步逻辑

**错误码**

- 0 成功

**msg**

- "ok" 成功

#### 15.2 客户端接口

获取部分信息

```
RpcHandler GetInfo (Connection.OnMessageCallback callback, 
                             bool need_bag, 
                             bool need_sc, 
                             bool need_avatar_exps, 
                             bool need_corp, 
                             bool need_energy, 
                             bool need_stage_all )
```

- need_bag 背包数据需要更新
- need_sc 软通数据需要更新
- need_avatar_exps 角色经验需要更新
- need_corp 战队信息（主要是等级）需要更新
- need_energy 体力信息需要更新
- need_stage_all 关卡信息需要全量更新


获取全部信息更新

```
RpcHandler GetAllInfo (Connection.OnMessageCallback callback)
```

#### 15.3 客户端命令
按钮 GetInfo

### 16. 材料合成

材料合成，扣除一些材料，增加合成结果，
一个包中包含多次合成动作，按照参数数组顺序依次合成

#### 16.1 消息定义

**Req** RequestCompose PlayerBag/ComposeRequest

|    名称   | 序列化名称 |  类型 |   含义   |
|-----------|------------|-------|----------|
| FormulaID | fid        | []int | 合成配方id的数组 |

**Rsp** ResponseCompose PlayerBag/ComposeResponse

说明：要注意合成的顺序，会按照数组中的id的顺序依次合成

**错误码**

- 0 成功
- 201 失败:配方ID不存在
- 202 失败:配方消耗错误，玩家原料不全
- 203 失败:添加合成结果错误

**msg**

- "ok" 成功

#### 16.2 客户端接口

```
RpcHandler Compose (Connection.OnMessageCallback callback, int[] fid)
```

- fid 合成配方id的数组


#### 16.3 客户端命令

按钮compose, 会按command发起一次合成

命令Compose:[fid1],[fid2],[fid3],...,[fidN],...

例如按配方1合成：
```
Compose:1
```

先按配方1合成，再按2合成
```
Compose:1,2
```

### 17. 商城购买

购买某一商城中某一位置上的物品，两次刷新之间，一个物品只能购买一次

#### 17.1 消息定义

**Req** BuyInStoreRequestMessage PlayerAttr/BuyInStoreRequest

|   名称  | 序列化名称 | 类型 |         含义         |
|---------|------------|------|----------------------|
| StoreId | s          | int  | 购买的商店的数组索引 |
| BlankId | b          | int  | 购买的位置的数组索引 |

**Rsp** BuyInStoreResponseMessage PlayerAttr/BuyInStoreResponse

**错误码**

- 0 成功
- 201 失败:购买失败

**msg**

- "ok" 成功

#### 17.2 客户端接口

```
RpcHandler BuyInStore (Connection.OnMessageCallback callback, 
                       int store_id, int blank_id)
```

- store_id 购买的商店的数组索引
- blank_id 购买的位置的数组索引


#### 17.3 客户端命令

按钮BuyInStore, command中以 store_id,blank_id 格式填写参数


### 18. 商城刷新

使用硬通刷新商城物品

#### 18.1 消息定义

**Req** RefreshStoreRequestMessage PlayerAttr/RefreshStoreRequest

|   名称  | 序列化名称 | 类型 |         含义         |
|---------|------------|------|----------------------|
| StoreId | s          | int  | 购买的商店的数组索引 |

**Rsp** RefreshStoreResponseMessage PlayerAttr/RefreshStoreResponse

**错误码**

- 0 成功
- 201 失败:购买失败

**msg**

- "ok" 成功

#### 18.2 客户端接口

```
RpcHandler RefreshStore( Connection.OnMessageCallback callback, 
                         int store_id )
```

- store_id 购买的商店的数组索引


#### 18.3 客户端命令

按钮RefreshStore, command中以 store_id 格式填写参数




## Debug协议

### 201. 背包Debug操作
通过此接口客户端可以获取一些玩家基本信息

#### 201.1 消息定义
**Req** RequestBagDebugOp /DEBUG/BagOpRequest

|     名称    | 序列化名称 |   类型   |          含义         |
|-------------|------------|----------|-----------------------|
| AddList     | addlist    | []string | 增加的物品ItemId列表  |
| RemoveList  | removelist | []uint32 | 要去掉的物品bagid列表 |
| IsRemoveAll | remove_all | bool   |   是否要清空背包        |

**Rsp** ResponseBagDebugOp /DEBUG/BagOpResponse

| 名称          | 序列化名称  | 类型                    | 含义   |
| -------- | --------- | -------- -- | ----------- |
| Total    | total     | int         | 物品总数 |
| Added    | added     | []uint32    | 增加了的物品bagid |
| Data     | data      | []string    | 请求中的AddList原样返回 |

说明：清空背包的同时会清空所有当前装备

**错误码**

- 0 成功
- 401 失败

**msg**

- "ok" 成功

#### 201.2 客户端接口

```
void BagDebugRequest(Connection.OnMessageCallback callback, 
                              string[] add, 
                              long[] remove, 
                              bool is_remove_all = false)
```

- add 增加的物品ItemId列表
- remove 要去掉的物品bagid列表
- is_remove_all 是否要清空背包

#### 201.3 客户端命令
按钮 DebugBag
增加以下三个道具

- "WP_ZF_0_0"
- "WP_GY_0_0"
- "WP_GY_0_1"

如果命令行中输入removeall时，按按钮DebugBag将会清除所有的背包



### 202. 软通Debug操作
软通相关的Debug接口，可以增加花去软通

#### 202.1 消息定义
**Req** RequestSCDebugOp /Debug/SCOpRequest

| 名称        | 序列化名称  | 类型     | 含义   |
| ---------- | --------- | -------- | ---------------------- |
| OpType     | op        | int      | 操作类型 1 增加 2 减少 |
| SCType     | typ       | int      | 操作软通类型 |
| Value      | value     | int64    | 操作软通值 |

**Rsp** ResponseSCDebugOp /Debug/SCOpResponse

| 名称          | 序列化名称  | 类型           | 含义   |
| ------------ | --------- | -------------- | ----------- |
| SC           | sc        | []int64   | 当前软通全量（用于更新客户端） |

**错误码**

- 0 成功
- 401 失败

**msg**

- "ok" 成功

#### 202.2 客户端接口
```
void DebugSCOp(Connection.OnMessageCallback callback, 
                          long op, long typ, long value)
```
- op    操作类型 1 增加 2 减少
- typ   操作软通类型
- value 操作软通值

#### 202.3 客户端命令
命令 SCOp:[op],[sc_t],[sc_v]

op 1 增加 2 减少

例如 增加1000金钱（软通类型是0）

```
SCOp:1,0,1000
```

减少10精铁（软通类型是1）

```
SCOp:2,1,10
```

- 按钮 Add Money 1000 增加金钱1000
- 按钮 Use Money 300 减少金钱300

### 203. 角色经验Debug操作
通过此接口客户端可以修改角色的经验

#### 203.1 消息定义
**Req** RequestAvatarExpOp /Debug/AvatarExpOpRequest

| 名称        | 序列化名称 | 类型      | 含义   |
| ---------- | --------- | -------- | ---------------------- |
| OpType     | op        | int      | 操作类型 1 增加经验 2设置级别 |
| SCType     | typ       | int      | 操作武将id |
| Value      | value     | int64    | 操作值 |

**Rsp** ResponseAvatarExpOp /Debug/AvatarExpOpResponse

| 名称 | 序列化名称 |   类型   |              含义             |
|------|------------|----------|-------------------------------|
| Exps | exps       | []uint32 | 当前经验全量（用于更新客户端) |


**错误码**

- 0 成功
- 401 失败

**msg**

- "ok" 成功

#### 203.2 客户端接口

```
 RpcHandler DebugAvatarExpOp( Connection.OnMessageCallback callback, 
                              long op, long typ, long value )
```

#### 203.3 客户端命令
AvatarXpOp:[op],[avatar_id],[value]
例如 给武将id0的增加1000经验

```
AvatarXpOp:1,0,1000
```

设置武将id2为10级

```
AvatarXpOp:2,2,10
```

### 204. 通用Debug操作
各种杂七杂八

#### 204.1 消息定义
**Req** RequestDebugOp /Debug/DebugOpRequest

| 名称        | 序列化名称 | 类型      | 含义   |
| ---------- | --------- | -------- | ---------------------- |
| Type     | typ        | string      | 操作类型 |
| P1     | p1       | int      | 参数1 |
| P2      | p2     | int    | 参数2 |
| P3     | p3       | int      | 参数3 |
| P4      | p4     | int    | 参数4 |

**Rsp** ResponseDebugOp /Debug/DebugOpResponse

**Debug操作**

|          类型          |          说明         |             P1含义             | P2含义 |
|------------------------|-----------------------|--------------------------------|--------|
| 将自身信息存入初始存档 | SaveSelfToInitAccount | 存档Id,为1时为默认的初始化存档 | -      |
| 增加体力               | AddEnergy             | 增加的体力值                   | -      |
| 重置关卡信息           | ResetStage            | -                              | -      |
| 用初始存档重置自身      | ResetSelfToInitAccount | -                              | -      |



**错误码**

- 0 成功

**msg**

- "ok" 成功
- 其他 失败

#### 204.2 客户端接口

##### 1 将自身信息存入初始存档

```
void DebugSaveSelfToInitAccount (Connection.OnMessageCallback callback)
```

##### 2 增加150体力

```
void DebugAddEnergy (Connection.OnMessageCallback callback)
```

##### 3 重置关卡信息

```
void DebugResetStageInfo (Connection.OnMessageCallback callback)
```

##### 3 用初始存档重置自身

```
void DebugResetSelfToInitAccount (Connection.OnMessageCallback callback)
```


#### 204.3 客户端命令
DebugOp:[typ],[p1],[p2],[p3],[p4]
例如 将自身存入初始存档

```
DebugOp:SaveSelfToInitAccount,1
```

增加150体力

```
DebugOp:AddEnergy,150
```

重置关卡信息

```
DebugOp:ResetStage
```

用初始存档重置自身

```
DebugOp:ResetSelfToInitAccount
```