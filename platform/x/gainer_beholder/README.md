#Gainer_Beholder工具
##什么是Gainer_Beholder?
    Gainer_Beholder是一款用于检测《极无双》韩国直购礼包订单异常的一款本地程序
    该程序扫描提供的包含订单信息的xlsx文件
    将出现问题的订单信息记录在一个新的名为gainer_02-01-2006_15:04:05_PM.xlsx的文件中
   
    
##我如何使用Gainer_Beholder?
    进入cheat_Beholder的文件目录下，运行main.go即可
    
##使用者注意：
    1.
    main.go文件下的
    const(
    	xlsxName string = "KoSuccessOrders.xlsx"
    	sheetName string = "orders"
    )
    保存着读取文件的名字及页签的名字，如需读取不同的文件/页签，记得更改其名字
    
    2.
    get_gainer.go下的
    var GIDS []string = []string{"207"}
    保存着需要判断的大区ID（GID）207表示韩国，
    使用时如要判断其他地区，需要修改此变量的值，
    如果GID为多个，存储为[]string{A,B,C}的形式即可。
    
    3.
    不同地区的礼包信息不同，若要将该程序用于其他地区订单异常情况的扫描，
    及时更换conf/data下的hotpackage.data。
    
##异常认证方式：
    在提供的订单结构中的pay_description字段下
    每条信息最后三个值为对应的IAPID|ID|SUBID
    如果IAPID的值在对应的ID|SUBID礼包下的IAPID字段没有找到
    则认定订单发生异常
##使用工具：
    Gainer_Beholder工具使用了planx/util的RW_xlsx包
    进入"vcs.taiyouxi.net/platform/planx/util/RW_xlsx"以查看更多信息