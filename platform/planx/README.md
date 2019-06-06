#PlanServerX

## 基本开发注意事项

* Log 请使用vcs.taiyouxi.net/planx/util/logs
* 不要直接使用os.Exit,使用util.Exit()因为它不能给程序机会去进行log.Close()操作，输出最后所有的log


## 序列化

序列化在游戏引擎中有两个部分

 - **传输序列化**是客户端和服务器之间的数据传输协议。
 - **存储序列化**数据库如何存数据以及客户端如何存数据

### 传输序列化

**传输序列化**是客户端和服务器之间的数据传输协议。本身和数据库如何存数据以及客户端应该如何组织数据，是两个不同的问题（那个应该是存储序列化的问题）。

Unity MiniMessagePack目前能处理的类型必须来自:

 - 基础类型
     + string (注意s是小写)
     + bool
     + sbyte, byte
     + short, ushort
     + int, uint
     + long, ulong
     + float, double
 - IList， IDictionary类型。
 - 否则所有不认识的类型统统调用ToString函数处理为函数

**是否当然我们自己实现IList, IDictionary的接口** MiniMessagePack序列化也是可以正常？

服务器使用完整Spec的MsgPack库，所以服务器可以解析任何符合Spec的MsgPack的二进制流。所以在客户端序列化上，客户端可以通过重载序列化接口等方式，创造复杂的数据结构给服务器。不过这并不是建议的行为，简单的传输协议（单层）本身就能够完成大多数传输序列化要求。

**MiniMessagePack的接口本身Pack（序列化）和UnPack(反序列化)是递归解析**
Pack

UnPack
new byte
List<object>
Dictionary<string, object>

序列化直支持**继承**单层属性序列化，客户端SDK可以通过重载序列化和反序列化接口实现嵌套层级序列化。
但是目前服务器部分还不能够找到有效方案解析嵌套序列化。
客户端和服务器可能需要尝试新库
 - Unity3D http://qiita.com/snaka/items/8da9f89deeef17b1923a, https://github.com/masharada/msgpack-unity，看了一下，不如MiniMessagePack好用
 - Server ：https://github.com/vmihailenco/msgpack

## 路由规则

客户端发送`Chat/RoomNumberRequest`

**TODO:**: 这里描述还不够清楚，需要找时间再测试和描述一下。

服务器定义接收方式

“/”开头的定义是全路径查找

```
mux.HandleFunc("/Chat/RoomNumberRequest",
        func(r servers.Request) *servers.Response {

            passid := r.Data["passthrough"].(string)
            logs.Info("we hahahah got %s, %s", r.Code, passid)
            r.Data["number"] = 989

            return &servers.Response{
                Code: "Chat/RoomNumberRequest",
                Data: r.Data,
            }
        })
```


默认是递归查找，一级一级路由表去查。因为已经定义了Friends, 所以，Friends/Abc的声明就不会生效

```
mux.Handle("Friends/Abc", &friends)
mux.Handle("Friends", &friends)
```

# TODO List

## bugs

* telnet edge server 6667 会crash

## https://godoc.org/github.com/dropbox/godropbox

- error 可以完善优化服务器的err处理模式
- set/lrucache 集合的实现
- http://blog.golang.org/context

## ConnPool优化

    source.SetKeepAlive(true)
    source.SetKeepAlivePeriod(time.Second * 60)
    source.SetLinger(-1)


为了提高后段服务的吞吐量，edge针对每个后端的服务都可以链接过个connection.
- godropbox: net2里面使用ConnectionPool的实现有话edge-game通信
- https://github.com/fatih/pool
- http://godoc.org/github.com/youtube/vitess/go/pools 这个Pool还可以用来处理各种资源的Pool

## 优化TODO

- []byte pool , http://godoc.org/github.com/youtube/vitess/go/pools
- "runtime/pprof"等性能工具的集成
- GC 优化工具 https://github.com/davecheney/gcvis
- 对于客户端请求出现问题后如何重试，请参考 错误重试和指数退避
- gate服务器大量链接情况下， SetReadDeadline&SetWriteDeadline的方式引入大量不好处理的Timer实例化。
应该会对性能有交大的影响。因此考虑使用[Timer的优化工作](http://blog.csdn.net/siddontang/article/details/18370541).

## 分布式TODO

- etcd [x]
- confd [x]
- goagain, grace

## 代码热更新TODO

- lua

## 数据统计TODO

go stats

## Unit Test TODO

核心模块实现利用单元测试完成开发编译自动测试等工作

## Throttled

所有服务都可能需要进行限制访问模式：

- Rate Limit 指定时间窗口内，能够有多少访问
- Interval 每个请求在多少时间内完成， `Allow 10 requests per second, or one each 100ms`
- MemStats 内存增长到一定程度时，禁止所有请求

## 错误重试和指数退避

参考文献：http://docs.aws.amazon.com/zh_cn/amazondynamodb/latest/developerguide/ErrorHandling.html

网络上的大量组件（例如 DNS 服务器、交换机、负载均衡器等）都可能在某个指定请求生命周期中的任一环节出现问题。

在联网环境中，处理这些错误回应的常规技术是在客户应用程序中实施重试。此技术可以提高应用程序的可靠性和降低开发人员的操作成本。

每个 AWS 开发工具包都自动支持 Amazon DynamoDB 实施重试逻辑。适用于 Java 的 AWS 开发工具包会自动重试请求，您可以使用 ClientConfiguration 类配置重试设置。例如，有时候网页发出的请求采用最低延迟并且不想重试，您可能会希望关闭重试逻辑。您可以使用 ClientConfiguration 类，并且为 maxErrorRetry 提供 0 这个值，从而设置为不重试。有关更多信息，请参阅在 Amazon DynamoDB 中使用 AWS 开发工具包。

如果您没有使用 AWS 开发工具包，则应当对收到服务器错误 (5xx) 的原始请求执行重试。但是，客户端错误（4xx，不是 ThrottlingException 或 ProvisionedThroughputExceededException）表示您需要对请求本身进行修改，先修正了错误然后再重试。

除了简单重试之外，我们还建议使用指数退避算法来实现更好的流程控制。指数退避的原理是对于连续错误响应，重试等待间隔越来越长。例如，第一次重试最多等待 50 毫秒，第二次重试最多等待 100 毫秒，第三次重试最多等待 200 毫秒，依此类推。但是，如果过段时间后，请求仍然失败，则出错的原因可能是因为请求大小超出预配置吞吐量，而不是请求速率的问题。您可以设置最大重试次数，在大约一分钟的时候停止重试。如果请求失败，请检查您的预配置吞吐量选项。有关更多信息，请参阅 指南：对表进行操作。

以下是显示重试逻辑的一个工作流程。该工作流程逻辑首先确定遇到的是不是服务器错误 (5xx)。如果遇到的是服务器错误，代码就会重试原始请求

```
currentRetry = 0
DO
  set retry to false

  execute Amazon DynamoDB request

  IF Exception.errorCode = ProvisionedThroughputExceededException
    set retry to true
  ELSE IF Exception.httpStatusCode = 500
    set retry to true
  ELSE IF Exception.httpStatusCode = 400
    set retry to false 
    fix client error (4xx)

  IF retry = true  
    wait for (2^currentRetry * 50) milliseconds
    currentRetry = currentRetry + 1

WHILE (retry = true AND currentRetry < MaxNumberOfRetries)  // limit retries
```
