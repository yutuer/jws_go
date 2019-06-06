#公会设计
---------

##1.词汇表

-	公会: Guild
-	公会成员: Guild Member
-	公会职位: Guild Position
-	公会会长: President
-	公会副会长: Vice President
-	申请: apply
-	公会ID: GID
-	玩家ID: AID
-	公会活跃度: Guild Activity Value
-	公会日志: Guild Log

##2.设计

###2.1 成员管理

**成员数据库**

- HSET	guild:{SID}:{GID}:members {AID} -> {玩家基本信息 json}
- HSET	guild:{SID}:{GID}:position {AID} -> {positionId}

**成员操作**

**开除公会成员**

需要同时读position, 写members, 写position, 应该用Lua脚本实现, 最后要发送邮件给玩家通知

**变更成员职位**

需要同时读position, 读members, 写position, 应该用Lua脚本实现, 最后要发送邮件给玩家通知

**更换公会会长**

需要同时读position, 读members, 写position, 应该用Lua脚本实现, 最后要发送邮件给玩家通知

**玩家退会**

需要同时读写members, 写members, 写position, 应该用Lua脚本实现, 玩家退会后在自己的存档中标识一个退回时间,来实现8小时不能进公会功能. 最后要发送邮件给玩家通知

###2.2 进入公会申请

**申请数据库**

-	guild:{SID}:{GID}:apply {AID} -> {玩家基本信息 json} Hash为了去重，多人操作
-	guild:{SID}:{AID}:apply {GID} -> {公会基本信息 json} 申请时：只有玩家自己操作，进公会成功：清除空
-	guild:{SID}:{GID}:applytime {AID} -> {玩家申请时间 }
-	guild:{SID}:{AID}:applytime {GID} -> {玩家申请时间 }
-   guild:{SID}:{AID}:guild {GID}

玩家和公会都需要读取各自的发送和收到的申请请求, 所以建立两个表, 分别用公会ID和玩家ID索引, 两者都设置为24小时过期

玩家提交申请时需要判断已提交申请的个数, 这个在玩家端坐判断即可, 对于公会申请数需要判断公会总数.

整个过程需要用lua脚本来实现

**批准进入公会**

需要删除两个表中的申请, 同时改member表, 最后要发送邮件给玩家通知

整个过程需要用lua脚本来实现

###2.3 公会基本数据

**公会活跃度**

-	guild:{SID}:{GID}:info activity -> 公会活跃度
-	guild:{SID}:{GID}:members_act {AID} -> 当日贡献活跃度数据(包括时间, 用来清零)

玩家消耗体力 直接加就可以, guild:{SID}:{GID}:members_act中的项只有玩家自己改, guild:{SID}:{GID}:members_act是需要每天清零的, 通过members_act里面的时间来判断是否需要清零.

###2.4 公会Boss 没有策划文档

###2.5 公会日志

guild:{SID}:{GID}:log [{日志信息 json}, {日志信息 json}, ...]
LIST
使用列表来存储日志信息, 如果超过最大上限就删除旧的
