数据通用同步文档
========================================

客户端和服务器需要保持数据同步，服务器需要在数据更新时同步最新的数据给客户端。
很多客户端请求都会造成数据的更新，为了一劳永逸的解决问题，在消息的基类做了通用的数据更新逻辑

为了防止错误，现在所有的数据同步都是全量更新，服务器会将所有的内容给客户端。

### 1. 服务器SyncResp基类
考虑到减少包的大小，传输的数据中不会全部包含同步数据，
对于服务器端来说，是可以明确知道包是否需要携带更新数据。
新增了SyncResp基类，继承自Resp，当一个回包需要携带同步数据时，就继承SyncResp，而不是Resp。
SyncResp中包括各种同步逻辑，服务器通过OnChangeXXX()接口指定包是否要同步这类数据到客户端。
Sync的OnChangeXXX()接口仅仅是标记这些数据需要更新，因为同步需要将逻辑处理之后的结果发给客户端。
所以SyncResp提供了mkInfo()接口，这个接口调用时会收集需要同步的数据，将其写入包中。


### 2. 同步数据的选择
服务器端不会每次都同步所有类型的数据给客户端，
服务器端处理逻辑的过程中是明确知道哪一类的数据有变动，当明确数据变动的时候，
通过OnChangeXXX()接口标记数据已发生改变，这样调用mkInfo()时，数据就会存入包中供客户端同步。


### 3. 客户端AbstractRspMsg基类中的同步逻辑
客户端在AbstractRspMsg基类中实现同步的数据解析和处理，
客户端会假定所有的包都可能带有同步数据，这样不会带来很多负担。
因为服务器端并不会同步所有类型的数据，所以客户端需要确定某一类数据是否需要同步，这个可以通过isNeedSyncXXX()函数来确定。


### 4. 通用的更新请求接口
因为所有的同步都是全量更新，所以以前实现的各种GetXXXInfo()可以统一用一个接口实现，
这里实现了一下两个通用的更新请求：

获取部分信息

```
RpcHandler GetInfo (Connection.OnMessageCallback callback, 
                             bool need_bag, 
                             bool need_sc, 
                             bool need_avatar_exps, 
                             bool need_corp, 
                             bool need_energy, 
                             bool need_stage_all )
```

- need_bag 背包数据需要更新
- need_sc 软通数据需要更新
- need_avatar_exps 角色经验需要更新
- need_corp 战队信息（主要是等级）需要更新
- need_energy 体力信息需要更新
- need_stage_all 关卡信息需要全量更新


获取全部信息更新

```
RpcHandler GetAllInfo (Connection.OnMessageCallback callback)
```

客户端可以使用这两个接口取代各种独立的get接口


### 5. 注意事项

- 1.SyncResp类型的消息，不要忘记在最后调用mkInfo()来填充同步数据
- 2.当一类数据发生变动，不要忘记调用OnChangeXXX()来标记数据变动



### TODO
