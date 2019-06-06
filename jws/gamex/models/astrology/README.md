# Astrology 星图系统

### Cheat 列表

- DebugOp:AstrologyGetInfo //查看信息

客户端不会有表现

- DebugOp:AstrologyInto,{int:武将ID},{int:孔ID},{int:占位符},{int:占位符},{string:星魂ID},{int:占位符} //镶嵌星魂

示例: DebugOp:AstrologyInto,9,1,0,0,SS_FLASHING_1_3,1

- DebugOp:AstrologyDestroyInHero,{int:武将ID},{int:孔ID} //分解武将身上的星魂

示例: DebugOp:AstrologyDestroyInHero,9,1

- DebugOp:AstrologyDestroyInBag,{int:占位符},{int:占位符},{int:占位符},{int:占位符},{string:星魂ID},{int:占位符} //分解背包里的一个星魂

示例: AstrologyDestroyInBag,0,0,0,0,SS_FLASHING_5_5,1

- DebugOp:AstrologyDestroySkip,{int:品质ID},{int:品质ID},{int:品质ID},{int:品质ID} //一键分解背包里的星魂

示例: AstrologyDestroySkip,1,2,3,4

- DebugOp:AstrologySoulUpgrade,{int:武将ID},{int:孔ID} //升级武将身上的星魂

示例: DebugOp:AstrologySoulUpgrade,9,1

- DebugOp:AstrologyAugur,{int:是否一键占星:0=否,1=是} //占星

示例: DebugOp:AstrologyAugur,0

- DebugOp:AstrologyFillBag  //向背包中增加所有星魂各100个

示例: DebugOp:AstrologyFillBag

- DebugOp:AstrologyClearMyAstrologyData //清除玩家的星魂数据

示例: DebugOp:AstrologyClearMyAstrologyData