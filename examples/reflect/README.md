# Go 反射的需求

客户端需求

有很多中情况下

- 客户端会发送一个请求列表过来，列表中包含了所有需要获取的属性name，服务器根据名称内容返回指定key的value
- 客户端发送一个kv dict到服务器，更新相关k的v。

在一个所有数据都在内存中的静态类型语言，需要独立的使用反射才能实现这些功能。

为了缓解映射带来的性能损失有两个解决方案

- 手动缓存相关类型的数据到一个map，像redigo的做法
- gocodec的方式进行go generate后期代码生成加速


解决方案：
- struct -> map[string]interface, http://godoc.org/github.com/fatih/structs
- map[string]interface -> struct, http://godoc.org/github.com/mitchellh/mapstructure
