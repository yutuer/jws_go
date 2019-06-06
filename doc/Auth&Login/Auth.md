### 匿名登录（访客模式）

这里要考虑两个问题：

1. 快速创建游戏的匿名帐号, 匿名帐号的保密性
2. 匿名帐号如何转换成注册用户

玩家能够迅速进入游戏，利用设备ID或者服务器随机分配ID。

* 设备ID：
	* Apple iOS [IFA/IFV]@productid
	* Android: [TODO]@productid
	* Sina Weibo ID： 181817@sina
	* QQ ID： [32 Characters]@qq
* 服务器ID：相当于服务器当前玩家数量

玩家初次登录，通过发送过来的“设备ID”，结合当前服务器的ID，生成一个配对。
这就是默认用户的登录方式。设备ID是保存在用户的Preference里面的。
productid 在后续文档中统一认为是game id

#### 访客模式下用户的帐号是有可能丢失的

**对于iOS用户可能丢失帐号的情况包括但不限于**:

* 换设备
* 重置广告追踪ID
* 重置手机但不使用iTune/iCloud恢复
* 重装App
* 盗版渠道app换成正版渠道app

**对于Android用户可能丢失帐号的情况包括但不限于**:

* TODO

**对于使用第三方登录的帐号更换设备不存在丢失风险**，因为平台提供用户接入的唯一ID作为设备ID记录在系统中。而这个唯一ID又是可以在任何设备上用相同的第三方帐号登录找到的

**安全性考量**

* Apple iOS [IFA/IFV]@productid 位数较多，不容易猜到。
* Android: [TODO]@productid，需要进行唯一标识的MD5+salt后才能保存，否则有一定程度盗号危险
* Sina Weibo ID： 181817@sina@productid，危险！因为这种数字是可以通过浏览对方微博能够在URL中看到的。
* QQ ID： [32 Characters]@qq@productid 有待研究，但感觉这个32个UUID是腾讯内部给用户配置的社交ID。

### 用户名密码登录

对于用户名密码登录，是匿名登录（访客模式）的更高安全级别设定。
有了用户名密码，用户就可以在其他设备上登录游戏帐号了。
所以数据库需要有一个额外的KV表或者Hash表。

