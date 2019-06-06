
数据落地


为了能够批量处理（保存，迁移，冷热数据交换）需要有一个和玩家所有Key的大集合。

SADD avatarkeys:{avatarid} keyname

原生方法备份(Redis Spec format)

* DUMP/RESTORE
* PTTL用户获取KEY的时效

或者转意成Json，方法也有两个参考

* redigo + json https://github.com/nesv/Scarlet/blob/master/read.go
* DUMP->json https://github.com/cupcake/rdb 需要额外的实现，但是估计性能会较好, Redis负担（读写次数）应该相对低


性能监控

http://www.nkrode.com/article/real-time-dashboard-for-redis

内存分析

redis-rdb-tools


所有存储分成两个部分
1. AddFlat 时间结构体-> Redis
   HashMap的转换。

需要弄清楚，如果结构体的变量类型不是string, []byte,
int, int64, float64, bool, nil, 时， Redigo将会调用
```
		default:
			var buf bytes.Buffer
			fmt.Fprint(&buf, arg)
			err = c.writeBytes(buf.Bytes())
```
https://github.com/garyburd/redigo/wiki/FAQ#does-redigo-provide-a-way-to-serialize-structs-to-redis
根据这里的描述，有关序列化和反序列化的问题，不能使用默认的这个FPrint来完成~

TODO 

- http://godoc.org/github.com/youtube/vitess/go/pools Does the Redigo pool support blocking get and other advanced features in Vitess pools? 需要了解这个数据库pool更改后的差别

