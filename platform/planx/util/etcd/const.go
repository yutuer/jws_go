package etcd

const (
	KeyIp                          = "ip"
	KeyInternalIp                  = "internalip"
	KeySName                       = "sn"
	KeyDName                       = "dn"
	KeyState                       = "state"
	KeyShowState                   = "ss"
	KeyServerStartTime             = "serstarttime"     // gm配置的开服时间
	KeyServerUsedStartTime         = "serusedstarttime" // 服务器用的开服时间
	KeyServerLaunchTime            = "serlaunchtime"    // gamex启动时间
	KeyGlobalClientDataVer         = "globalclientinfo/dataver_v2"
	KeyGlobalClientBundleVer       = "globalclientinfo/bundlever_v2"
	KeyEndPoint                    = "gonggao"
	KeyEndPoint_Whitelistpwd       = "whitelistpwd"
	KeyEndPoint_maintain_starttime = "maintain_starttime"
	KeyEndPoint_maintain_endtime   = "maintain_endtime"
	KeyActValid                    = "actvalid"
	KeyOrder                       = "order"
	DirHotData                     = "hotdata"
	KeyHotDataCurr                 = "curr"
	KeyBaseDataBuild               = "basedatabuild"
	KeyHotDataBuild                = "hotdatabuild"
	KeyHotDataSeq                  = "hotdataseq"
	KeyHotDataGid                  = "gid"
	KeyHotDataLastSignalTime       = "lastSignalTime"
	KeyMergedShard                 = "mergedshard"
	KeyShardVersion                = "version" // 服务器版本号
	KeyTeamAB                      = "teamAB"
	KeyGlobalClientDataVerMIn      = "globalclientinfo/dataver_min"
	KeyGlobalClientBundleVerMin    = "globalclientinfo/bundlever_min"
	StateOnline                    = "online"
	MultiLang                      = "multilang"
	Language                       = "language"
)

const ServerListVersion = "serverlistversion"

const (
	ShowState_NewSer      = iota // 0 新服
	ShowState_Fluent             // 1 流畅
	ShowState_Hot                // 2 火爆
	ShowState_Crowd              // 3 拥挤
	ShowState_Maintenance        // 4 维护，玩家看得到，但进不去
	ShowState_Experience         // 5 体验，玩家看不到
	ShowState_Super              // 6 玩家看不到
	ShowState_Count
)

type ShardInfo struct {
	Machine       string
	Gid           uint
	Sid           uint
	Ip            string
	SName         string
	DName         string
	State         string
	ShowState     string
	StartTime     string
	LaunchTime    string
	RealStartTime string
	Version       string
	MultiLang     string
	Language      string
}
