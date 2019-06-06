﻿##删除玩家账号的某些物品信息：
```
type PerInfo struct {
    VirtualMoney []string //虚拟货币
    Mat []string //各种石头
    Xp []string //经验丹
    Jad map[string]int64 //宝石
    Acid string //账号id
    RedisAddr string //redis的地址
    RedisNum int //数据库号
}
```
profile和PlayerBag都有相应的DBload和DBsave用于拉取和保存
###宝石：
直接删除profile中的jade信息
删除Jad[jadeid]个宝石、将身上和神兽上面的宝石置0
注：将宝石的经验清0可能会出现bug

###经验丹(Account.PlayerBag中)：
使用playerbag.RemoveWithFixedID删除

###包子等虚拟货币(profile.sc):
将对应等虚拟货币置0

###各种石头(Account.PlayerBag中)：
使用playerbag.RemoveWithFixedID删除