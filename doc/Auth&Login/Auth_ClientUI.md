# Auth阶段的客户端UI flow和 错误码



```uml
@startuml


state 首次认证界面{
认证0:OnClickAuth
快速游戏:OnClickQuickGame
注册0:OnClickRegister
}

note left of 首次认证界面
首次进入游戏的认证界面
提供了[用户名]/[密码]的输入框
以及[认证]/[注册]/[快速游戏]三个按钮
end note


state 一般认证界面{
认证1:OnClickAuth
注册1:OnClickRegister
}

note left of 一般认证界面
常规进入游戏的认证界面
提供了一个历史认证过的账号下拉框
以及[认证]/[注册]两个按钮
end note

state 账户注册界面{
注册:OnClickRegisterConfirm	
注册取消:OnClickRegisterBack
}

note left of 账户注册界面
使用[用户名]和[密码]注册新用户
提供[注册]/[取消]两个按钮
end note

注册请求:Register
注册请求:这是个网络请求，需要出错处理

认证请求:Auth
认证请求:这是个网络请求，需要出错处理

获取服务器列表:GetShardList
获取服务器列表:这是个网络请求，需要出错处理

state 服务器列表界面{
确认选择:OnClickLogin	
}

note left of 服务器列表界面
从服务器列表下拉框中选择想认证的服务器
提供[Login]按钮
end note

登录:EnterShard
登录:这是个网络请求，需要出错处理


认证0 -> 认证请求
快速游戏 -> 认证请求
认证1 -> 认证请求

注册0 -down-> 账户注册界面
注册1 -down-> 账户注册界面

账户注册界面 -down-> 注册请求

认证请求 -down-> 获取服务器列表:Success
注册请求 -down-> 获取服务器列表:Success

获取服务器列表 -down-> 服务器列表界面:Success

确认选择 -down-> 登录
登录 -down-> Connect:Success

@enduml
```

# 网络错误码

所有网络请求都会返回result字段。

多数result返回值是
- "ok"-代表成功，伴随其他可用字段。
- "no"-代表失败，伴随error字段和forclient字段。

error字段通常在生产环境中是加密字段，请记录到客户端出错本地log中。

forclient的使用的错误码，就是接下来文档说明的，配合客户端逻辑处理的错误代码。

NOTE:

客户端传输到服务器的密码是md5后二次加密的。因此服务器没有用户原始的密码。存入数据库的密码又是经过salt
md5加密的。所以是是无法反向还原的。

因此客户端应该检测用户密码是否长度合适，是否足够安全。当然越安全就越繁琐，因此需要取一个平衡点.

## 通用错误码－客户端直接获取

此处错误码生成来自Unity客户端对http底层库的二次封装。

| 错误码   | 说明                     | 
| ----- | ---------------------- | 
| 10000 | resp文本不是协议的hashtable格式 | 
| 10001 | http底层error错误          | 
| 10002 | http由客户端发起abort退出      | 
| 10003 | http连接超时               | 
| 10004 | http获取内容超时             | 

## 来自服务器请求的错误码

所有错误代码都在文件x/authapix/errorctl/client_errorcode.go中维护。

### 通用错误码

| 错误码  | 说明         | 
| ---- | ---------- | 
| 100  | 不应该出现的情况   | 
| 101  | 请求的加密格式非法  | 
| 102  | 数据库无法返回正确值 | 
| 103  | 用户名不合法小于6个大于128个，必须字母开头，由字母和数字组成，区分大小写 |



### 认证请求

| 错误码  | 说明                               | API                 | 
| ---- | -------------------------------- | ------------------- | 
| 201  | 设备ID认证：玩家设备绑定了帐号密码，强制要求使用用户名密码登录 | /auth/v1/device/:id | 
| 202  | 用户名密码认证：用户名称未找到                  | /auth/v1/user/login | 
| 203  | 用户名密码认证：用户密码不正确                  | /auth/v1/user/login | 
| 204  | 设备ID认证： 注册码验证错误                    | /auth/v1/deviceWithCode | 
| 205  | 设备ID认证： 注册码已被使用过                   | /auth/v1/deviceWithCode | 
| 206  | 用户名密码认证：用户被禁止登陆                  | /auth/v1/user/login | 
| 210  | SDK: 英雄sdk查token返回错误                   | /auth/v1/deviceWithHero | 
| 211  | SDK: 英雄sdk查user信息返回错误                 | /auth/v1/deviceWithHero | 
| 220  | SDK: Quicksdk checkuser错误                  | /auth/v1/deviceWithQuick |

### 注册请求

| 错误码  | 说明                      | API                   | 
| ---- | ----------------------- | --------------------- | 
| 301  | 用户名称已存在                 | /auth/v1/user/reg/:id | 
| 302  | 该设备关联的存档已经绑定过用户名，无法再次绑定 | /auth/v1/user/reg/:id | 

### 登录分服

| 错误码  | 说明              | API               | 
| ---- | --------------- | ----------------- | 
| 401  | 无法获取网关，因为认证数据无效 | /login/v1/getgate | 
| 402  | 当前分服无可用网关       | /login/v1/getgate | 

