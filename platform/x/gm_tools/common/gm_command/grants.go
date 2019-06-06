package gm_command

var grantByTyp map[string]int = map[string]int{
	"addUser":                   0,  //添加用户
	"changePass":                1,  //更改密码
	"getUser":                   2,  //获取用户
	"virtualIAP":                3,  //发送虚拟充值邮件
	"delVirtualIAP":             4,  //删除虚拟充值邮件
	"getVirtualIAP":             5,  //加载用户虚拟充值记录
	"getSysPublic":              6,  //加载公告
	"sendSysPublic":             7,  //公告子页面发布按钮
	"delSysPublic":              8,  //删除公告
	"releaseSysPublic":          9,  //公告页面发布公告按钮
	"getSysRollNotice":          10, //加载跑马灯公告
	"sendSysRollNotice":         11, //添加跑马灯页面提交按钮
	"delSysRollNotice":          12, //删除跑马灯公告
	"updateSysRollNotice":       13, //发送跑马灯公告
	"activitySysRollNotice":     14,
	"banAccount":                15, //封禁账号
	"getAllBanAccount":          16, //获取所有的封禁账号(未完成)
	"getEndpoint":               17, //获取Endpoint设置
	"setEndpoint":               18, //设置Endpoint
	"gagAccount":                19, //禁言账号
	"getInfoByNickName":         20, //昵称查询
	"getChannelUrl":             21, //渠道链接
	"updateChannelUrl":          22, //渠道链接确定并上传
	"getAllInfoByAccountID":     23, //账号转移Query按钮
	"setAllInfoByAccountID":     24, //账号转移上传账号信息
	"cleanUnEquipItems":         25, //账号转移红色按钮
	"getAccountByDeviceID":      26, //通过设备ID查询账号
	"getAccountByName":          27, //通过玩家昵称查询玩家账号
	"createNamePassForUID":      28, //
	"createDeviceForUID":        29, //
	"queryIAPInfo":              30, //支付查询查询log
	"queryIAPInfoMail":          31, //支付查询查询Mail
	"queryHCInfo":               32, //支付查询查询HC获得/消耗
	"getShards":                 33, //
	"modShard":                  34, //
	"syncShards":                35, //
	"getShardsFromEtcd":         36, //Shard管理获取Shards信息
	"setShard2Etcd":             37, //
	"delMail":                   38, //单人邮件删除邮件
	"getAllGids":                39, //Shard状态标签
	"getShardShowStateByGid":    40, //Shard状态标签设置Gid
	"setShardsShowState":        41, //Shard状态标签设置按钮
	"releaseSysPublicToRelease": 42, //公告页面发布公告按钮
	"getActValidInfo":           43, //活动开关加载
	"setActValidInfo":           44, //活动开关提交
	"queryGateWayInfo":          45, //支付查询Dynamo信息查询
	"setVersionHotC":            46, //数据版本提交按钮
	"getServerHotData":          47, //数据版本获取服务器数据信息
	"signalServerHotC":          48, //
	"getServerByVersion":        49, //数据版本获取指定版本的服务器
	"signalAllServerHotC":       50, //数据版本热更新
	"delRank":                   51, //删除排行榜
	"reloadRank":                52,
	"getAllServer":              53, // 跑马灯公告获取所有服务器
	"setShardsShowTeamAB":       54, // Shard状态标签设置TeamAB
	"transDevice":               55, // 设备转移
	"virtualtrueIAP":            56, // 虚拟充值发送真实虚拟充值邮件
	"restartShard":              57, // Shard管理
	"uploadKS3":                 58, // 公告页面上传金山云
	"getBatchMailDetail":        59, // 批量邮件下载邮件
	"getBatchMailIndex":         60, // 批量邮件批量号
	"setShardStateEtcd":         61, // 设置服务器状态
	"setGrant":                  62, // 设置权限
	"getGrant":                  63, // 获取权限
	"getNickNameFromRedisByACID":64, // 根据ACID获取玩家昵称
}
