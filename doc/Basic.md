## 心跳

无论是客户端还是服务器，只发送PING信息，并不期待返回PONG。心跳信息只是单向发送。

服务器实现：
每30s发送一个PING信息，并携带一个uint64 的UTC
Timestamp，客户端可以根据这个时间戳更新自己的时间。

客户端实现：
每30s发送一个PING信息，不携带任何数据。客户端单向发送的原因是NAT的回收idle
connection机制，防止玩家长链接被路由器当成长时间限制链接导致链接被断开。

目前阶段，没有响应PING必须发送PONG的需求，而且Gate也不会给Game服务器传递PING。


## Log 模块

使用util/logs模块进行日志文件的输出。该模块使用了seelog作为底层模块。

下面我拿了大部分seelog的有效模块过来，并针对每一个特性进行使用上的价值说明

1. Xml configuring to be able to change logger parameters without recompilation
1. Changing configurations on the fly without app restart
1. Possibility to set different log configurations for different project files and functions
1. Simultaneous log output to multiple streams
1. Different output writers
 - Console writer
 - File writer
 - Buffered writer (Chunk writer)
 - Rolling log writer (Logging with rotation)
 - SMTP writer
 - Others... (See Wiki)
1. Log message wrappers (JSON, XML, etc.)
1. Global variables and functions for easy usage in standalone apps
1. Functions for flexible usage in libraries

说明：
- 1,2两条意味着即便你在生产环境（production）中遇到问题，你可以动态的调整服务器输出的log级别。甚至是调整输出到额外格式的额外文件中。比如一般情况，线上只输出Error,Critical级别的错误信息。当服务器出现一些问题，你可以尝试调整日志输出级别为Info，使得服务器输出更多的日志信息，这样就能够更加容易判断线上服务器出问题的原因，同时又能保证尽量少的输出日志内容，减少日志服务器的日常压力（存储和性能）
- 3,4,5,6：实现了多路log输出和输出到多种终端。方便对接任何集中式日志系统，使用任何数据结构。
- 7,8：开发者考虑了使用者可能需要在自己的库中进行二次封装，所以相关系统做过响应的调整，方便二次封装后，log系统输出函数的名称能够准确。


### Log系统初始化

`InitSentry(DSN string)`
应该在main函数中尽可能早的调用。如果被反复调用，后调用会覆盖前面的配置。

`LoadLogConfig(cfg string)`
读取log系统的xml配置。配置一旦成功加载后。系统接收HUP信号量，然后会重新刷新配置，心配置立即生效。

### Log分级

https://github.com/cihub/seelog/wiki/Log-levels

- Trace
  通常是最细粒度的调试代码，比如在你的for循环中输出每一个值的状态。
- Debug 用于输出调试系统行为的信息，帮助定位问题，用于开发过程中。
- Info
  是生产环境（production）中会输出的Log，通常用来观察系统工作状态的，比如是不是启动了，启动参数大概什么样子等等。
- Warn
  通常是小错误，可控的异常失败。通常系统仍然能够正常运转。一切仍然正常运转。
- Error 这种错误通常影响了程序的正常工作，但是还不至于导致系统退出。
- Critical
  生产级别的最严重错误，通常导致了程序的退出。这个级别的信息不被缓存，防止程序挂掉的时候丢失信息。


我们这里再定义一下Critical，游戏开发中导致玩家链接断开的所有未知错误，都认为是Critical。所以所有的panic行为都应该记录为Critial。

Info的输出量不应该影响服务器性能。

无用的Debug信息可以不注释掉，但是Trace应该适当的考虑注释去掉。以免影响服务器性能。

Log使用性能建议参考：https://github.com/cihub/seelog/wiki/Performance

### Panic

`logs.PanicCatcher()`
函数应该在任何需要捕捉异常情况的函数中作为defer调用。
这个函数会把这些异常使用Critical级别输出到Log系统。同时如果我们配置了Sentry服务，该系统负责输出堆栈信息到Sentry系统。

`PanicCatcherWithAccountID`
函数需要传入玩家accountid的参数，这样在Sentry或者Log输出系统中，就能够知道是那个玩家的运行出现了异常。能够方便利用相关玩家存档进行重现低概率问题。

### Sentry

- http://getsentry.com/
- https://github.com/getsentry/sentry
- golang sdk:"github.com/getsentry/raven-go"

Sentry
系统是异常监控系统，尽管可以自行搭建，但是使用官方的云服务速度目前仍然是很不错的。因此无论是开发环境还是production环境，目前都直接购买这个服务。

在目前的系统中：Error, Critical,
PanicCacher都会尝试输出信息到Sentry服务器。
