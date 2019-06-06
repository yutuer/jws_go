### 数据库封装

目前Auth同时支持利用Redis和DynamoDB作数据存储，为了方便切换DB实现和减少重复代码，
将数据库的读写操作封装为一系列API：

```
type DBInterface interface {
    Init() error
    GetDeviceInfo(deviceID string) (*deviceUserInfo, error)
    SetDeviceInfo(deviceID, name string, uid db.UserID) (*deviceUserInfo, error)
    IsNameExist(name string) (int, error)
    GetUnKey(name string) (db.UserID, error)
    SetUnKey(name string, uid db.UserID) error
    GetUnInfo(uid db.UserID) (string, string, error)
    UpdateUnInfo(uid db.UserID, deviceID, authToken string) error
    SetUnInfo(uid db.UserID, name, deviceID, passwd, email, authToken string) error
    IncrDeviceTotal() (db.UserID, error)
    GetAuthToken(authToken string) (db.UserID, error)
    SetAuthToken(authToken string, userID db.UserID, time_out int64) error
}
```

分别有DBByRedis和DBByDynamoDB两种实现。

Redis实现和DynamoDB实现中数据的结构是一致，有如下数据：

|     名称      |索引类型| 索引 | 索引格式        |   说明             |
|--------------|------|------|----------------|-------------------|
| Device       | HASH | Id   | {device_id}    |匿名登录数据         |
| Name         | HASH | Name | un:{usersname} |用户名密码登录数据    |
| UserInfo     | HASH | UId  | uid:{user_id}  |认证系统数据         |
| DeviceTotal  | HASH | Id   | -              |device_total       |
| AuthToken    | HASH | Token| {authtoken}    |AuthToken          |

#### 1 Device API
读取和设置匿名登录数据
```
    GetDeviceInfo(deviceID string) (*deviceUserInfo, error)
    SetDeviceInfo(deviceID, name string, uid db.UserID) (*deviceUserInfo, error)
```

#### 2 Name API
读取和设置用户名密码登录数据，检查用户名是否已被注册
```
    IsNameExist(name string) (int, error)
    GetUnKey(name string) (db.UserID, error)
    SetUnKey(name string, uid db.UserID) error
```

#### 3 UserInfo API
读取和设置认证系统数据，更新数据中的AuthToken
```
    GetUnInfo(uid db.UserID) (string, string, error)
    UpdateUnInfo(uid db.UserID, deviceID, authToken string) error
    SetUnInfo(uid db.UserID, name, deviceID, passwd, email, authToken string) error
```

#### 4 DeviceTotal API
将device_total加一并返回，用于生成user_id
```
    IncrDeviceTotal() (db.UserID, error)
```

#### 5 AuthToken API
读取和设置AuthToken
```
    GetAuthToken(authToken string) (db.UserID, error)
    SetAuthToken(authToken string, userID db.UserID, time_out int64) error
```