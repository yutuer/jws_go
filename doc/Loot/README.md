# 掉落


## 掉落实现

1. Template
2. ItemGroup
3. Stage
    - Reward Limit
    - Reward xxx


```sequence
Client->Server: PrepareLootForLevel\n(levelid, [敌兵ID1x10， 敌兵ID2x1])
Server->Server: 生成掉落信息
note over Server:掉落信息保存到临时变量\nlastLevelLoot中\n辞典dict{掉落ID：内容...}

Server->Client: 返回客户端使用掉落信息Response
note over Client: Reponse内容 (\n[敌兵ID1, [(掉落ID1,内容)， (掉落ID2,内容)]], \n[敌兵ID2, [(掉落ID1,内容)， (掉落ID2,内容)]]] ...)

note over Client,Server: ...玩家进入游戏，客户端搜集玩家实际搜集的掉落ID...
Client-> Server: DeclareLoot\n(掉落ID1， 掉落ID2...)

Server->Server: 验证和保存，清空临时变量
note over Server: 对比lastLevelLoot中的信息,确认掉落有效性,\n全部有效则保存相关掉落（软通，材料和武器填入背包）\n无效则返回错误信息
Server->Client: 返回验证保存结果

```

### 客户端传递敌兵信息的意义

服务器可以校验，这些敌兵信息是否和服务器的数量一致。
因为敌兵，坛子的码放是在Unity中进行，数量往往可能我配置的不一样。
因此，客户端传递这些信息给服务器后，服务器可以发现问题，并记录运营Log。
如果必要，运营团队可以使用数据更新的方法修正这个问题。

此外，客户端在掉落的产生上要有一定的容错。如果服务器返回50个坛子的掉落。

 - 客户端Unity的版本实际上有55个坛子，那么必然有5个坛子是一定没有掉落的。
 - 客户端Unity的版本实际上有40个坛子，那么必然有10个掉落信息不会被使用到。

### 什么都没有掉落的情况是否还要生成掉落ID？

目前的决定是生成掉落ID。

TODO:
如果使用M/N的计数模式，则不生成这些ID也是没有问题的。N是服务器返回的有掉落的数量，
M是Unity客户端当前关卡中实际的坛子数量。





