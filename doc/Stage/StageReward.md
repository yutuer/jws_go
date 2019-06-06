#关卡结算
----------
##1. 奖励结算
副本奖励分Limit奖励和Rand奖励，数据分别在LootData.xlsx的STAGELIMITREWAD表和STAGERANDREWARD中
###1.1 Limit奖励
Limit奖励 包括

- ItemGroupID="" 物品组ID  
- LootNum=1      (数量控制）掉落次数
- LootSpace	    (数量控制）随机区间 
- Offset=0        (数量控制）区间偏移量
- MItemGroupID="" 补偿物品组ID       

首先会计算一个随机的区间值N 根据LootSpace和Offset随机值N
这里表示每N次掉落中有LootNum次掉落ItemGroupID对应的物品，
其他则掉落MItemGroupID对应的物品

###1.2 Rand奖励
Rand奖励比较简单，
每当副本结算时，
对于副本id对应的所有的ItemGroupId，进行发奖


##2. 奖励算法实现
随机算法参考TAOCP 3.4.2节中的算法S

算法S是一次性生成结果，而我们需要分次随机，所以存储了算法S中的中间变量，
在算法S中t表示已经随机过的次数，m表示已经摇中的次数，N即为随机区间，n即为掉落次数，
每次检验时生成一个0-1的随机值U，
检验 (N-t)U >= n-m 如成立则没有摇中（注意），t自增1，
如不成立则摇中 t自增1，m自增1，
如果m >= n 则算法完成。

这里为了便于实现计

- Reward_count = n - m
- Space_num = N - t

储存这两个值便于分次随机

##3. 玩家奖励中间数据
由于玩家结算副本是分次的，这期间会上下线，
所以上面算法中的Reward_count和Space_num需要存入数据库，
每个玩家每个副本会存入多组奖励数据。

玩家副本对应的数据存在Proflie的stage_imp中，通过Stage变量进行序列化和反序列化，
存数据库时将stage_imp转为json格式存入Stage中，随Profile存入Redis，
取数据库时将Stage上得json串反序列化到stage_imp中，外部通过GetStage()接口进行读写

##4. 关卡结算限制

关卡有如下限制

- 体力是否足够
- 战队等级是否满足
- 角色等级是否满足
- 是否满足角色专属

申请关卡掉落（进入关卡）和申请关卡结算时都需要验证条件。

在服务器端有两个接口：

关卡前检查

```
func (p *Account) IsStageCanPlay(id string, avatars []int, is_sweep bool) uint32 
```

关卡后检查，增加体力的消耗

```
func (p *Account) CostStagePay(id string, avatars []int, is_sweep bool) uint32 
```

如果检查不通过，服务器端会驳回请求