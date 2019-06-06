### 利用Hash Set 而不是 独立数据库列散KV模式

HashSet是另外一种实现方式，目前没有使用是因为不确定数据量很大的情况下的性能表现会是怎样的，所以目前没有使用。
此外Auth数据库应该是需要分离程序里数据库的。
 
用于查询当前设备ID对应的真正user_id
 ``HSET devices {device_id} {user_id} `` 


### 数据库设计注意事项 TODO

实际上此数据库的冷数据会越来越多，可以考虑使用SSDB这种LevelDB的数据库，或者Mysql+Redis模式。DynamoDB数据库比较适合。

