## RunMode dev和prod的使用

### Gamex和Auth

项目 | dev | prod|test
-----|------|----
【gamex】debug操作    | 有    |  | 有
【gamex】向etcd定时写ip    |     | 有 | 有
【gamex】玩家随机数不同 | |有|
【gamex】dataver更新频率| 30s|5min|30s
【auth】的gin使用ReleaseMode||有|有
【auth】的DebugKick|有||有
【auth】的错误信息加密||有|
【auth】的fetchshards从etcd取||有|有
【auth】的/auth/testsentry|有||有




