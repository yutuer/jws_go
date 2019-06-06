#Cheat_Beholder工具
##什么是Cheat_Beholder?
    Cheat_Beholder是一款为了临时解决《极无双》作弊外挂的一款服务器程序，该程序会全面扫描玩家的
    存档信息，记录数据出现异常的玩家信息
    包含："玩家昵称，ID，玩家战力，最远关卡id，最远关卡推荐战力，VIP等级"
    并将这些信息存储在一个名为cheat_02-01-2006_15:04:05_PM.csv的文件中
    
##我如何使用Cheat_Beholder?
    进入cheat_Beholder的文件目录下，通过命令./cheat_beholder直接运行程序即可。
    
##使用者注意：
    在conf/data下的level_trial.data文件中存储着用于判断玩家数据异常的表格生成的二进制文件
    不同地区的数据异常表格内容有所差异，使用者在对不同地区的玩家存档进行扫描时应注意更换对应
    的level_trial.data文件。（越南的表格推荐战力数据是国服的四倍）
##作弊认证方式：
    在认证玩家是否作弊时，Cheat_Beholder工具使用account.ProfileData结构体中的
    CorpCurrGS_HistoryMax字段作为玩家历史最高战力，
    如果当前战力小于level_trial表格中最远关卡推荐战力的70%，则认定玩家作弊