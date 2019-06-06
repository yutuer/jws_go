Dynamodb
==========================
# Dynamodb总结
Dynamodb特点：
- KV数据库:主键对内容 内容假定的很大
- 分布式:Item存储在不同的Part上，索引独立存 --> 读写索引由此要另收费
- 要钱:根据吞吐量/压力付费，不需调优和运维，只要省钱就行

## KV
key（主键 Primary Key）有两种:

- 独立key Hash Primary Key
- key + range Hash and Range Primary Key

key + range 可以根据key和range单独做查询
大多数时候独立key可以满足

## 读取模式

- 最终一致(evently)：可能不是最新的，节省钱[^queryevently].
- 绝对一致：保证最新的，double钱

[^query_event]: 查询Query只能最终一致

Dynamodb假定Item是很大的，所以最小传输以400KBytes为单位

> Each of your DynamoDB items can now occupy up to 400 KB. The size of a given item includes the attribute name (in UTF-8) and the attribute value. 


## 查询与索引

作为一个KV数据库，依然要一些比较复杂的查询，
因为是NoSQL，所以要查询就得加索引，

索引其实就是第二个key，但是是独立存的，所以任何对于索引的读写都要单算钱，另一方面，读出索引之后，只是知道了item的位置，要读item里的内容还要再有一个读操作，再花一份钱

索引可以加Projection，其实就是把Item里的一部分直接加到索引中，这样读索引直接可以读出这部分内容，但是一旦这个被Projection的Item发生修改，会多写一次

由于Dynamodb的索引实现，索引更新是比较麻烦的，需要先删除旧索引，再加入新索引。

- Global Secondary Indexes
- Local Secondary Indexes

Dynamodb存储Item的时候，对于key+range的键，相同key的Item是存储在一个Part上的，
对于索引，Hash算法是一样，所以如果索引的key是Item的主键当中的key的话，Dynamodb就会确保查询的Part，否则Dynamodb不知道最后要去那些Part去找最终的Item，这就是上面两种索引的区别。

对于Local Secondary Indexes 只需要查两次，因为后一次取Item时Dynamodb是知道去哪个Part的

对于Global Secondary Indexes 就不知道有多少次了，因为后一次的Item不定在哪个Part上，所以计费时按结果Item的个数算，总不能有三个Item，却查了四个Part

## 最大化吞吐量和维护

本段落部分主要来自[DynamoDB Table使用指南](http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GuidelinesForTables.html#GuidelinesForTables.Partitions)

DynamoDB会如何把数据列散到更多的数据库上？

'''
numPartitions.total = MAX ( numPartitions.tableSize | numPartitions.throughput)
numPartitions.tableSize = tableSizeInBytes / 10 GB
numPartitions.throughput = ( readCapacityUnits / 3000 ) + ( writeCapacityUnits / 1000 )
'''

- 数据库尺寸的增长， 每10GB会被独立到一个分区（Partition）上
- 因为单个Partition上的Throughput会有上限， read上限3000 unit[^capacityunity], write上限1000 unit, 因此，当Thoughput增长时会因此分区的增长。

[^capacityunity]: capacity units are based on a data item size of 4 KB per read or 1 KB per write
而数据应该在哪些键值上，是由hash决定的。另外分区只会增长不会下降。

所以如何更好的利用Throughput就是一个优化过程。也可以认为是如何最大化利用Throughput是需要精心设计的。这个优化过程我们需要平衡Read/WriteCapacityUnit和Throughput。

优化方法: 最大列散化hash分布。[0:0:1001] < [0:0:1001].profile, [0:0:1001].store ... < [0:0:1001].profile.1, [0:0:1001].profile.2, [0:0:1001].store.10gatcha

> This does not mean that you must access all of the hash keys to achieve your throughput level; nor does it mean that the percentage of accessed hash keys needs to be high. However, do be aware that when your workload accesses more distinct hash keys, those requests will be spread out across the partitioned space in a manner that better utilizes your allocated throughput level. In general, you will utilize your throughput more efficiently as the ratio of hash keys accessed to total hash keys in a table grows.

### Cache Popular Items

ElasticCache -> DynamoDB(Last Week Table, Last Month Table, ColdTable) -> S3(Cold, 可选)
有些数据模型是，老数据基本不会被访问到。可以考虑使用更多的Table
Last Week Table, Last Month Table, ColdTable， 其中ColdTable的Throughput的设定可以很低。

### Use Burst Capacity Sparingly

> DynamoDB currently reserves up 5 minutes (300 seconds) of unused read and write capacity.  During an occasional burst of read or write activity, this reserved throughput can be consumed very quickly — even faster than the per-second provisioned throughput capacity that you've defined for your table. 

> However, do not design your application so that it depends on burst capacity being available at all times: DynamoDB can and does use burst capacity for background maintenance and other tasks without prior notice.

虽然我们可以使用5分钟内积攒的未使用Throughput，但是应用程序应该只用这个burst应对突发事件，而不是常规的访问。

### Consider Workload Uniformity When Adjusting Provisioned Throughput

> If you reduce the amount of provisioned throughput for your table, DynamoDB will not decrease the number of partitions . Suppose that you created a table with a much larger amount of provisioned throughput than your application actually needed, and then decreased the provisioned throughput later. In this scenario, the provisioned throughput per partition would be less than it would have been if you had initially created the table with less throughput.

这段主要说明了，分区是不会减少的。如果之前因为要在短时间内完成写操作，而选择了大的Write Throughput，当你减少这个阈值时，分区数量不会减少。这将会导致，每个分区上的阈值会非常小！！！


通常一个引用初期的数据量比较小，随着时间的流逝，数据量变大后，Dynamo会自动Scale out你的数据分布到更多的分区上。这是你每个分区的Throughput的值自然会下降。

这时如果某些数据的访问仍然是局部热点，那么他们的Throughput可能就会出现不够的问题。

> If it isn't possible for you to generate a large amount of test data, you can create a table that has very high provisioned throughput settings. This will create a table with many partitions; you can then use UpdateTable to reduce the settings, but keep the same ratio of storage to throughput that you determined for running the application at scale. You now have a table that has the throughput-per-partition ratio that you expect after it grows to scale. Test your application against this table using a realistic workload.

测试的方式是：1生成大量的数据，如果这不好做就可以这样操作，创建一个高Throughput的Table.然后突然降低这个配额。因为分区的数量不会减少，会导致每个分区可用的Thoughput减少到一定程度。

所以，针对时间序列数据，随着时间的流逝，数据库的性能会下降到一个无法容忍的地步。所以如果你能够从实时数据TAble中移除就得时间数据，把它们保存到别的地方，这将能够大大的提高每个分区的吞估量

## 键值数据结构模式下，分析玩家存档和如何利用DynamoDB之间的问题思考
TODO

## 文档模式下，分析玩家存档和如何利用DynamoDB之间的问题思考

文档模式下，V是Json数据存档

### 如果一个玩家的存档是400KB， 玩家用一个Key存储在DynamoDB中

读取一个玩家需要100 Read Capacity Units, 写一个玩家存档需要400 Write Capacity Units.

一个Key会被存储到一个固定的Partition上，而单个Paritition的读写上限是:Read 3000, Write 1000.

每个Partition上可以同时读的玩家数量？3000/100=30玩家读取/每秒, 1000/400=2.5玩家写入/每秒

如果一个玩家的存档是80KB， 玩家用一个Key存储在DynamoDB中:

```
80KB/4KB = 20 ReadUnits
80KB/1KB = 80 WriteUnits
3000/20=150玩家读取/每秒, 1000/80=12.5玩家写入/每秒
```

### DynamoDB要求设计者能够更合理设计运行模式和数据模式（Workload Uniformity）

我们无法保证哪些分区上都是活跃用户，哪些分区上是不活跃用户的情况。所以我们应该至少使用两个DynamoDB来保证较好的用户体验。

- 活跃Table: 高Throughput，但是整体尺寸只是活跃用户的数据。以充分利用Thoughput为最高优先级。玩家数据用hash详细列散
- 不活跃Table: 低Throughput，因为玩家多所以数据尺寸大。以数据存储为最高优先级，能够使在Throughput整体很低的情况下，每个分区有足够的Read/Write能力。

如果活跃Table用Redis替代， 不活跃Table用S3替代呢？
S3的吞吐量是如何设计的？稳定性如何。Dynamo对比S3的优势就是保证响应时间。
而S3的优势是操作简单尺寸大。

## 操作的消耗总结

###Global Secondary Indexes

- 索引要单独加钱 

####读

- 两方面的消耗
- 读取索引 = 匹配的item索引总量
- 读取Item = item * item个数

####写

- 所有索引相关的修改都要单独写
- 索引键的创建要有一次写
- 索引键的修改要有两次写，一次删，一次加
- 索引键的删要有一次写
- 如果修改影响到projected attributes 即便没有影响索引，也会触发额外的写


###Local Secondary Indexes

- 索引要单独加钱 

####读

- 仅仅查询索引或者仅仅需要已经projected的内容的话，不需额外消耗
- 如果查询的是没有projected的内容，只是增加数据量，

####写

- 所有索引相关的修改都要单独写
- 索引键的创建要有一次写
- 索引键的修改要有两次写，一次删，一次加
- 索引键的删要有一次写
- 如果修改影响到projected attributes 即便没有影响索引，也会触发额外的写

## 过载处理

Dynamodb是预设置吞吐量的，但总会有过载的时候，这就需要梳理错误处理。

Dynamodb返回的错误分三类：

- Amazon的错误
- 过载
- 我们的错误

我们的错误要我们自己解决，重发不能解决问题，
Amazon的错误，需要（也只能）重发，并且这时应该告警。
过载的话，分两种情况，一种是偶尔的大操作导致满负荷，另一种是原来设定的吞吐量上限已经不够了。
对于登陆这种需求，应该没有大操作，所有操作都差不多，遇到过载情况等待无效时应该增加上限。

问题：如何保证过载期间所有数据读写不丢失？

分两种情况:

- 可重复的操作 确保成功之后重试即可，可以有一个写队列管理
- 不可重复的操作 --> TODO 目前似乎没有这样的需求

## Auth/Login的需求

特点：

- 对空间不敏感 完全可以一次把一个账号下的数据全取出来
- 读写不均衡 常登陆，但注册是有限的，并且可以预估

## Auth By Dynamodb实现

当前的Auth系统中共有五个表：

|     名称      |索引类型| 索引 | 索引格式        |   说明             |
|--------------|------|------|----------------|-------------------|
| Device       | HASH | Id   | {device_id}    |匿名登录数据         |
| Name         | HASH | Name | un:{usersname} |用户名密码登录数据    |
| UserInfo     | HASH | UId  | uid:{user_id}  |认证系统数据         |
| DeviceTotal  | HASH | Id   | -              |device_total       |
| AuthToken    | HASH | Token| {authtoken}    |AuthToken          |

这些表机构与Redis表一致

为了实现生成user_id，数据库中单独记录了一个device_total值，利用DynamoDB的Atomic Counter功能，
每次将device_total加一， 之后用其生成user_id。这个过程和redis逻辑上是一致的。

## Scan

这个太费按理来说我们不应该有这样的需求
