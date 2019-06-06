### Redis数据库设计

#### 匿名登录数据：user_id的生成和device_id对应关系

``INCR device_total 用户生成user_id``

存储device id和user id的关系有两种方案
 - 利用Hash Set
 - 数据库完全独立出来，直接使用K(device_id)=V(userid)的列散模式


**独立数据库列散KV模式**

K(device_id)=V(user_id)

`` SET {device_id} {"user_id":{user_id}, "username":"timesking"} ``

{user_id} 可以是一个json字符串，包含user_id和username。 如果存在username字段，可以根据产品需要决定是否强制客户端使用用户名密码登录。



#### 用户名密码登录数据

这条数据是方便客户端直接发起用户名密码登录请求时使用。

```
SET un:{usersname} {user_id} 用来支持用户名密码登录模式[^3]
```

[^3]: 如果一个设备ID存在用户名，则是否强制使用用户名密码登录取决于产品设定

#### 认证系统数据

* HSET uid:{user_id}
	* name {username}
	* device {device_id}
	* password md5(salt+{client_md5(password)})
	* email {email}
	* authtoken {generatedtoken}[^1]， 使用uuid生成{generatedtoken}，生成后应该RPC到对应的游戏登录（login）系统中。客户端使用authtoken去login系统进行服务器登录。需要设置有效时间。
所以authtoken会在login系统的数据库中有一个独立的有效时间。

[^1]: 登录过程参考http://redis.io/topics/twitter-clone,

### AuthToken

第一种方式：`HSET auths {authtoken} {user_id}`

用于根据玩家最后登录的authtoken进行快速链接登录， Hash数据有可能变得很大，需要找办法优化。或者单独一个数据库，直接使用KV模式。
而且authtoken如果需要过期时间，这个方式也不合适

第二种方式(正在使用)：`SETEX at:{authtoken} 60*3 {user_id}`



