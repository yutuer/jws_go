# Redis落地工具使用说明
---------------------------

Redis监听落地的功能需要使用Reids 2.8以上版本， 参考文档[
Redis Keyspace Notifications](http://redis.io/topics/notifications)。

Redis需要配置：notify-keyspace-events=ghE
AWS ElastiCache 里面不能配置redis.conf和运行命令CONFIG SET，
但是能通过“Cache Parameter Groups”来配置这个参数。

## 启动参数

工具提供了很多中运行模式，也可以一次启动所有功能。

| sub command | 运行模式   | 相关功能                                                 |
|:------------|:-----------|:---------------------------------------------------------|
| monitor     | 服务模式   | 实时监听Redis，将变化的Hash key对应的内容落地            |
| allinone    | 服务模式   | 监听落地和回档功能都作为服务启动                         |
| onlandone   | 命令行模式 | 落地Redis中Key对应的内容                                 |
| onlandall   | 命令行模式 | 落地Redis中所有的Key对应的内容                           |
| restoreone  | 命令行模式 | 将S3或者DynamoDB中的某一个Key的落地信息回档到Redis       |
| restoreall  | 命令行模式 | 将S3或者DynamoDB中一个库中的所有Key的落地信息回档到Redis |

### 实时监听Redis，将变化的key对应的内容落地

   - monitor	启动监听Redis修改同步模式

启动后开启监听程序不退出。利用了Redis本身的Key Space
Notification的能力对有修改的key进行落地。**我们只监听hdel, hset两种命令带来的修改**

如果在没有任何修改的情况下，需要强制触发数据落地，有如下两种方法：

- 向TCP指令端口发送命令：服务启动后，其他程序可通过TCP端口(配置中参考`Command_Addr`配置)发送指令，达到强制要求落地用户的目的。参考下面章节**指令**
- 向redis数据库特殊pub/sub发送key名字：我们如果设置了配置`Redis_Channel="onland"`,则可以通过其他进程向这个pub/sub发送，key的名字要求立即落地单个存档。例如发送key字符串`profile:0:0:1001`.


### 落地Redis中Key对应的内容

   - onlandone	落地库中指定账号
   ```
   ./redis_storage  onlandone --user profile:0:0:1001,bag:0:0:1001
   ```
   程序会将profile:0:0:1001和bag:0:0:1001落地

   落地结束后程序退出


### 落地Redis中所有的Key对应的内容

   - onlandall	将库中所有账号落地

   回档结束后程序退出


### 将S3或者DynamoDB中的某一个Key的落地信息回档到Redis

   - restoreone	回档单个用户
   ```
   ./redis_storage  restoreone --user 2014/01/02/profile:0:0:1001,2014/01/02/bag:0:0:1001
   ```
   程序会将2014/01/02/profile:0:0:1001和2014/01/02/bag:0:0:1001回档到redis

   回档结束后程序退出



### 将S3或者DynamoDB中一个库中的所有Key的落地信息回档到Redis

   - restoreall	回档对应数据源中的所有用户存档

   回档结束后程序退出

### All in One 模式

   - allinone	开启所有功能

开启监听，以服务的模式启动。可通过命令完成所有功能。
比较monitor模式，多了通指TCP指令监听端口发送命令，触发回档独立存档以及全面回档的功能。


## TCP指令
程序在allinone下会监听Command_Addr中的地址，通过tcp协议可以向其发送命令，格式为字符串：

```
{op} {param}\n
```

支持的命令如下：

- restore {key}
  将key对应的落地信息回档到redis。allinone服务模式下生效。
- restore_all 将所有落地信息回档到redis。allinone服务模式下生效。
- sync {key}
  落地redis中key对应的信息。allinone|monitor服务模式下生效。
- sync_all 落地redis中所有key。allinone|monitor服务模式下生效。

可以通过telnet向工具发送指令

比较monitor模式，多了通过TCP指令监听端口发送指令，触发落地独立存档以及落地所有数据的功能。

## 配置

所有库配置都是在配置文件中指定的，配置分4个部分，分别是通用配置、落地数据源配置、落地配置、回档目标配置。

具体请参考onland.toml
