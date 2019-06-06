# Redis落地工具设计文档
---------------------------

## 1. 综述

Redis落地工具包括以下功能：

- 1 实时监听Redis，将变化的key对应的内容落地
- 2 落地Redis中Key对应的内容
- 3 落地Redis中所有的Key对应的内容
- 4 将S3或者DynamoDB中的某一个Key的落地信息回档到Redis
- 5 将S3或者DynamoDB中一个库中的所有Key的落地信息回档到Redis

实时监听功能是自动触发的，其他功能需要通过启动参数或者tcp命令发起。

## 2. Store封装
为了隐藏数据库API的不同，实现了一套Store接口，分别对应S3、Redis、DynamoDB、SSDB实现

```
type IStore interface {
    Open() error
    Close() error
    Put(key string, val []byte) error
    Get(key string) ([]byte, error)
    Del(key string) error
    StoreKey(key string) string
    RedisKey(key_in_store string) (string, bool)
    Clone() IStore
}
```

其中

StoreKey() \ RedisKey()用来装换存在落地库和Redis中的实际key。

Clone()接口会创建一个自身的副本，配置一样但是调用API新建了链接，
这个是为了避免IStore实现不是线程安全的问题，另外一方面可以新建链接，增加IO效率。

**注意**：对于Del,有的因为用不到而没实现。

## 3. Scan数据库
由于这几种数据库的Scan操作差异比较大，所以并没有统一接口。

对于**Redis**：

Scan是通过onland包中的redisScanner类中实现的，

```
type RedisKeysHander func(keys []string) error
```

外层传入一个RedisKeysHander回调去实现逻辑。

通过NewScanner()创建一次Scan，通过Start()开始，循环调用Next()获取数据，
Next()中会调用RedisKeysHander，通过IsScanOver()判断是否Scan完成。

对于**S3**：

Scan是通过restore包中的s3Scanner类中实现的，

```
type S3KeysHander func(keys []string) error
```

NewS3Scanner()创建一次Scan，通过Start()开始，这其中会循环调用S3KeysHander

对于**DynamoDB**：

Scan是通过StoreDynamoDB类中的Scan函数实现的，

```
type KeyScanHander func(idx int, key, data string) error
```

Scan会调用KeyScanHander，由于DynamoDB Scan时会返回数据，所以回调中直接包含数据。

DynamoDB ParallelScan：

DynamoDB支持在数据库端发起并发Scan，通过Segment，TotalSegments参数可以指定多次scan线程调用来完成一次scan，在planx/util/dynamodb/dynamodb.go的ParallelScan()函数中封装实现，
实现中开启了多个scan协程分别进行scan，结果通过channel返回给外部进行处理。


## 4. 落地


- [ ] TODO

**OnlandStores安全退出**
OnlandStores Stop()会确保所有key_queue中的任务完成，之后所有work协程退出。
通过

```
    wg                sync.WaitGroup
    wg_key_queue      sync.WaitGroup
```

实现。

wg保证协程创建销毁正确：

当Init()时回wg.Add(1)确保Start()中的主协程正常创建，这个会在主协程退出时Done。

每创建一个工作协程会wg.Add(1)，这个会在工作协程退出时Done。

wg_key_queue保证key_queue中的任务完成：

通过NewKeyDumpJob()添加任务其中wg_key_queue.Add(1)，工作协程中完成任务后会Done。

需要注意CloseJobQueue()：由于任务是其他模块发起的，所以应该由发起任务的模块关闭key_queue。



## 5. 回档

由于DynamoDB和S3的Scan功能不一致，DynamoDB Scan时会将key和数据都发送回来，
而S3只是将key发回来，需要另外取数据。
针对这里的不同，分别实现了两个Restore类来封装逻辑。

从DynamoDB回档

- [ ] TODO

从S3回档

- [ ] TODO
-
**Restore安全退出**
Restore Stop()会确保所有jobs中的任务完成，之后所有work协程退出。
由于两种实现原理一样，这里统一叙述。

通过

```
    wg           sync.WaitGroup
    wg_jobs      sync.WaitGroup
```

实现。

wg保证协程创建销毁正确：

当Init()时回wg.Add(1)确保Start()中的主协程正常创建，这个会在主协程退出时Done。

每创建一个工作协程会wg.Add(1)，这个会在工作协程退出时Done。

wg_jobs保证jobs中的任务完成：

添加任务其中wg_jobs.Add(1)，工作协程中完成任务后会Done。

Stop时会在任务完成之后关闭jobs，触发工作协程退出


## 6. TODO
- 需要优化DynamoDB的Scan，现在的实现速度上不去
- 需要统一各个模块安全退出逻辑的实现，现阶段实现虽正确，但容易出错

