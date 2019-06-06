# 服务器合服设计

## 所有表

A类表：Key带有玩家id相关的表：

* profile：包含玩家名称
* bag
* general
* pguild
* store
* simpleinfo
* anticheat
* friend
* tmp 
* msgs:SimplePvpRecord **合服不保留**
* msgs:gank **合服不保留**
* guild:playerapply **合服不保留**

B类表：Key带有公会id的表：

* guild:Info：包含玩家名称；包含公会显示id
* guild:guildapply **合服不保留**

C类表：Key带有shardId的表，如：names:0:10；guild:name:10

* guild:account2guild：玩家id->公会id，直接合并
* names：玩家名称
* guild:name：公会名称
* guild:id2uuid：公会id->uuid，用于公会查找功能
* guild:uuid：公会uuid的set，用于得到随机公会列表
* guild:idseed：公会id自增发生器
* fishreward：钓鱼全服奖池，**合并规则需要策划**
* global:levelfinish：首次通关信息，用于跑马灯
* globalcount:acid7day：全服数量，7日狂欢商店物品数量，**不应该在新服里出现和服的情况**
* teampvp：组队pvp
* festivalboss: 年兽，**合服不保留**
* moneycat：招财猫，**合服不保留**
* 限时神将: 没有影响，单独的db
* RankSimplePvpForTitle,RankTeamPvpForTitle,Rank7DayGsForTitle: **合服不保留**
* wantgenbest: **合服不保留**
* gvg:
* topAccountWorship：**合服不保留**
* destgenfirst：选时间早的

	排行榜相关

	SortedSet表可以直接合并，合并排序后，重新生成topN表

	* RankCorpGs
	* RankCorpGs:topN
	* RankCorpGsSvrOpn
	* RankCorpGsSvrOpn:topN
	* RankSimplePvp
	* RankSimplePvp:topN
	* RankGuildGS
	* RankGuildGS:topN
	* RankGuildGsSvrOpn:topN
	* RankGuildGateEnemy
	* RankGuildGateEnemyLast
	* RankGuildGateEnemy:topN
	* RankCorpTrial
	* RankCorpTrial:topN
	* RankBalance：排行榜上次刷新时间，用于启服务器检查是否需要发奖；**合并时有一个shard需要重发奖，就重发**

## 合并步骤

应该有合服组的概念，即多少shardid范围的服务器能够被合，如限制一个合服组最多能有999个服

### 步骤一：对C类表进行合并

#### names和guild:name表，并涉及到玩家和公会名称重复的处理：
1、服务器不修改玩家和公会的名称，只是将names和guild:name表合并，由前端对每个玩家姓名根据shardid加个后缀，但需要服务器提供合服标记，而且保证显示名字的地方都要有uuid

2、服务器离线将所有名字，加上shard标记

3、服务器离线，将重复的名字的账号或公会统计出来，并只将重复的名字进行修改，并做标记；
在玩家或会长上线后，给玩家一次修改名字的机会；

之后names和guild:name表的内容需要合并

#### guild:id2uuid和guild:uuid表，公会显示id处理：

1、如果有合服组，就可以修改现有公会id的生成方式，如shardId(后三位)*1000+id；之后表的数据就可以直接合并；这里默认要被合服的服务器的公会数量不会超过999个；

2、重新生成两个服务器的公会的id

#### Rank表处理

SortedSet的表（即不带“:topN”）包括基本所有玩家或公会，合服先合并SortedSet的表，之后再生成新的topN，覆盖之前两个shard的topN

### 步骤二：对A类和B类表的处理，带有shardId的表（包括Rank）有问题：

目前此类表都是解析Accountid或GuildId得到shardId，再拼出表明，如names:0:10，再访问表

若合服的话，上述过程就会出现问题
	
#### 方案一

修改所有的AccountId和GuildId，将shardid统一，此方案需要离线修改：

1、所有的表名

2、所有表里记录了id的地方

3、所有的dynamo表里记录id的地方

4、以后加表都要留意是否合服需要修改

5、不知是否还有其他地方受影响，如BIlog？

#### 方案二(采用)

1、将表的内容进行合并

2、在根据Accountid或GuildId得出shardid拼表名的时候，统一经过一个接口，此方法给出当前应该用哪个shardid

3、之后再加新C类表的访问，也得经过统一接口

### 步骤三：最后将A类和B类表放到同一个db即可

## 合服后影响
1、战斗记录清空，包括：pvp和切磋
2、公会申请清空
3、首次通关关卡记录以一服为准
4、钓鱼奖池以一服为准
5、开服狂欢商店物品数量以一服为准
6、排行榜重新排
7、开服7天排行榜没有处理，我认为合服都是在开服7天之后才会进行
8、全服邮件已一服为准
9、排行榜称号少一天的
10、我要名将最高清空
11、神兽第一用时间最早的
12、排行榜膜拜清空
13、招财猫、年兽清空
14、重名的玩家和工会，二服的加后缀
15、军团战奖励开服后，奖励一次性发放，城主占领信息清除

# 在什么时候不能合服
1、开服七日狂欢期间
2、限时神将开放时（其他运营活动呢？）
3、各种结算奖励期间
4、各种活动期间，如：兵临，工会战
