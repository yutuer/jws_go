# 极无双

go版本1.6.3

## 调试接口

### pprof

运行文件目录，pprof文件里记录着的所使用的端口

### goroutine泄漏检测

运行文件目录，leaktest文件里记录着的所使用的端口

创建goroutine镜像基点

http://127.0.0.1:port/leaktest/snapshot

和基点计算差值

http://127.0.0.1:port/leaktest/check

结果输出到stderr中

### 给指定accountid发送push推送

运行文件目录，debugtest文件里记录着的所使用的端口

http://127.0.0.1:port/push/:acid/:content

需要acid对应的账号用手机登陆过服务器

### 全局表（即表名没有acid或guilduuid）的表名

统一在需要的包下建立xxx_dbname.go文件，并将获取表名的函数写在里面

方便搜索查找，以后合服也方便修改

可参考：modules/guild/guild_dbname.go





# 启动环境



```bash
brew install etcd redis
brew services start etcd
brew services start redis
cd {vcs.taiyouxi.net/tools}
go build
./tools newserv -c new_serv_local.toml

cd {vcs.taiyouxi.net/gamex}
go build
./gamex allinone
```

