package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// GetTBTeamList : 获取组队BOSS的队伍列表
// 获取组队BOSS的队伍信息

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgGetTBTeamList 获取组队BOSS的队伍列表请求消息定义
type reqMsgGetTBTeamList struct {
	Req
	DifficultyId int64 `codec:"diff_id"` // 请求的难度ID
}

// rspMsgGetTBTeamList 获取组队BOSS的队伍列表回复消息定义
type rspMsgGetTBTeamList struct {
	SyncResp
	TeamList [][]byte `codec:"team_list"` // 队伍列表
	TBTeamDayOfWeek int64 `codec:"tb_tdow"` // 组队boss星期几 0开始周一
}

// GetTBTeamList 获取组队BOSS的队伍列表: 获取组队BOSS的队伍信息
func (p *Account) GetTBTeamList(r servers.Request) *servers.Response {
	req := new(reqMsgGetTBTeamList)
	rsp := new(rspMsgGetTBTeamList)

	initReqRsp(
		"Attr/GetTBTeamListRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetTBTeamListHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// TBTeamSimple 获取组队BOSS的队伍列表
type TBTeamSimple struct {
	
	TeamId string `codec:"tm_id"` // 队伍id
	TeamMemberCount int64 `codec:"tm_c"` // 队伍人数
	LeaderPlayerName string `codec:"leader_pn"` // 队长名字
	LeaderPlayerSid int64 `codec:"leader_sid"` // 队长区服
	FightAvatarIds []int64 `codec:"f_ava_ids"` // 出战的角色avatarId
	AvatarLevel []int64 `codec:"f_ava_lv"` // 出战的主将等级
	AvatarStarLevel []int64 `codec:"f_av_slv"` // 出战的主将星级
	TeamState int64 `codec:"t_state"` // 队伍状态 0=开放;1=无法加入；2=满员
}

// TBTeamReady : 在队中准备
// 在队中准备

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgTBTeamReady 在队中准备请求消息定义
type reqMsgTBTeamReady struct {
	Req
	ReadyTeamDiff int64 `codec:"t_rd_tdf"` // 准备进入的难度
	ReadyTeamId string `codec:"t_rd_tid"` // 准备的队伍id
	ReadyNowStatus int64 `codec:"t_rd_st"` // 队伍目前准备状态 0未准备 1已准备
	ReadyHeroId int64 `codec:"t_rd_hid"` // 准备的武将id
}

// rspMsgTBTeamReady 在队中准备回复消息定义
type rspMsgTBTeamReady struct {
	SyncResp
}

// TBTeamReady 在队中准备: 在队中准备
func (p *Account) TBTeamReady(r servers.Request) *servers.Response {
	req := new(reqMsgTBTeamReady)
	rsp := new(rspMsgTBTeamReady)

	initReqRsp(
		"Attr/TBTeamReadyRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.TBTeamReadyHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}


// CreatTBTeam : 创建组队BOSS队伍
// 创建组队BOSS队伍

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgCreatTBTeam 创建组队BOSS队伍请求消息定义
type reqMsgCreatTBTeam struct {
	Req
	CreateDifficultyId int64 `codec:"c_diff_id"` // 请求的难度ID
}

// rspMsgCreatTBTeam 创建组队BOSS队伍回复消息定义
type rspMsgCreatTBTeam struct {
	SyncResp
	MyTeamInfo []byte `codec:"my_team"` // 个人队伍信息
}

// CreatTBTeam 创建组队BOSS队伍: 创建组队BOSS队伍
func (p *Account) CreatTBTeam(r servers.Request) *servers.Response {
	req := new(reqMsgCreatTBTeam)
	rsp := new(rspMsgCreatTBTeam)

	initReqRsp(
		"Attr/CreatTBTeamRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.CreatTBTeamHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// TBTeamDetail 创建组队BOSS队伍
type TBTeamDetail struct {
	
	MyTeamId string `codec:"my_tm_id"` // 队伍id
	MyTeamDifficulty int64 `codec:"my_t_diff"` // 个人队伍难度
	LeaderIndex int64 `codec:"my_lead_ind"` // 个人队伍队长索引 0队长是本人 1队长是队友
	MyTeamSetting int64 `codec:"my_t_s"` // 个人队伍当前加入设定
	RedBoxTickState int64 `codec:"my_rb_t"` // 房间必得红宝箱勾选状态
	TBTeamTypeId int64 `codec:"tb_ttid"` // TB阵容组合id
	LevelInfoId string `codec:"tb_lviid"` // 战斗场景id
	TBWbossId string `codec:"tb_wbid"` // 组队bossId
	MyTeamMember [][]byte `codec:"my_t_mems"` // 个人队伍成员信息
}
// TeamMember 创建组队BOSS队伍
type TeamMember struct {
	
	PosId int64 `codec:"t_mem_pos"` // 选择上阵武将的位置
	HeroId int64 `codec:"t_mem_hid"` // 个人队伍成员选择武将id
	HeroMagicPet int64 `codec:"t_mem_hmp"` // 武将灵宠形象id，0无，1~6不同形象
	HeroSwing int64 `codec:"t_mem_hsw"` // 个人队伍成员武将翅膀
	HeroFashion []string `codec:"t_mem_hfa"` // 个人队伍成员武将时装
	HeroGlory string `codec:"t_mem_g"` // 个人队伍成员神兵id
	TeamMemAcid string `codec:"t_mem_acid"` // 个人队伍成员acid
	TBMemLevel int64 `codec:"tb_m_liid"` // 角色等级
	TBMemHeroStar int64 `codec:"tb_m_hs"` // 角色星级
	TeamMemGs int64 `codec:"t_mem_gs"` // 个人队伍成员战力
	TeamMemSid int64 `codec:"t_mem_sid"` // 个人队伍成员服务器id
	TeamMemName string `codec:"t_mem_name"` // 个人队伍成员名字
	TeamMemReadyState int64 `codec:"t_mem_ready"` // 个人队伍成员准备状态
	TeamMemCompressGs int64 `codec:"t_mem_cgs"` // 个人队伍成员压缩过的战力
	TeamMemVIPLvl int64 `codec:"t_mem_vip"` // 个人队伍成员vip等级
	TeamMemIcon int64 `codec:"t_mem_ic"` // 个人队伍成员头像
}

// TBChooseHero : 房间中选将界面确认选将
// 房间中选将界面确认选将

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgTBChooseHero 房间中选将界面确认选将请求消息定义
type reqMsgTBChooseHero struct {
	Req
	TbChooseHeroId int64 `codec:"t_chava"` // 选将界面选择武将的id
	TbChooseHeroTeamId string `codec:"t_cha_tid"` // 选将的所在房间id
	TbChooseHeroPos int64 `codec:"t_cha_p"` // 选将界面选择武将的位置
	TbChooseHeroDiff int64 `codec:"t_cha_di"` // 选将所在房间的难度
}

// rspMsgTBChooseHero 房间中选将界面确认选将回复消息定义
type rspMsgTBChooseHero struct {
	SyncResp
	ChooseHeroId int64 `codec:"t_c_hid"` // 成员选择武将id
	ChooseHeroMagicPet int64 `codec:"t_c_hmp"` // 选择武将灵宠形象id，0无，1~6不同形象
	ChooseHeroSwing int64 `codec:"t_c_hsw"` // 选择武将翅膀
	ChooseHeroFashion []string `codec:"t_c_hfa"` // 选择武将时装
	ChooseHeroPost int64 `codec:"t_c_post"` // 选择武将选择后的位置
	ChooseHeroCompressGs int64 `codec:"t_c_gs"` // 选择武将的压缩战力
}

// TBChooseHero 房间中选将界面确认选将: 房间中选将界面确认选将
func (p *Account) TBChooseHero(r servers.Request) *servers.Response {
	req := new(reqMsgTBChooseHero)
	rsp := new(rspMsgTBChooseHero)

	initReqRsp(
		"Attr/TBChooseHeroRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.TBChooseHeroHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}


// TBTeamJoinSetting : 队中邀请加入设定 0开放 1仅限邀请
// 队中邀请加入设定

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgTBTeamJoinSetting 队中邀请加入设定 0开放 1仅限邀请请求消息定义
type reqMsgTBTeamJoinSetting struct {
	Req
	TeamJoinSetting int64 `codec:"t_set"` // 队中邀请加入设定
	TeamSettingTeamId string `codec:"t_se_tid"` // 邀请设定更改的队伍id
}

// rspMsgTBTeamJoinSetting 队中邀请加入设定 0开放 1仅限邀请回复消息定义
type rspMsgTBTeamJoinSetting struct {
	SyncResp
	TeamJoinSetting int64 `codec:"t_set"` // 队中邀请加入设定
}

// TBTeamJoinSetting 队中邀请加入设定 0开放 1仅限邀请: 队中邀请加入设定
func (p *Account) TBTeamJoinSetting(r servers.Request) *servers.Response {
	req := new(reqMsgTBTeamJoinSetting)
	rsp := new(rspMsgTBTeamJoinSetting)

	initReqRsp(
		"Attr/TBTeamJoinSettingRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.TBTeamJoinSettingHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}


// JoinTBTeam : 加入组队BOSS的队伍
// 加入组队BOSS的队伍

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgJoinTBTeam 加入组队BOSS的队伍请求消息定义
type reqMsgJoinTBTeam struct {
	Req
	JoinTeamId string `codec:"j_id"` // 想加入队伍的ID
	JoinLeaderId string `codec:"j_l_id"` // 想加入队伍的队长ID
}

// rspMsgJoinTBTeam 加入组队BOSS的队伍回复消息定义
type rspMsgJoinTBTeam struct {
	SyncResp
	MyTeamInfo []byte `codec:"my_team"` // 个人队伍信息
	JoinTBTeamResult int64 `codec:"j_t_ret"` // 加入房间结果，0成功，1房间不存在，2距离上次离开房间时间过短，3房间已满，4内部错误
}

// JoinTBTeam 加入组队BOSS的队伍: 加入组队BOSS的队伍
func (p *Account) JoinTBTeam(r servers.Request) *servers.Response {
	req := new(reqMsgJoinTBTeam)
	rsp := new(rspMsgJoinTBTeam)

	initReqRsp(
		"Attr/JoinTBTeamRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.JoinTBTeamHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}


// LeaveTBTeam : 离开组队BOSS的队伍
// 离开组队BOSS的队伍

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgLeaveTBTeam 离开组队BOSS的队伍请求消息定义
type reqMsgLeaveTBTeam struct {
	Req
	LeaveTeamId string `codec:"l_id"` // 想退出队伍的ID
}

// rspMsgLeaveTBTeam 离开组队BOSS的队伍回复消息定义
type rspMsgLeaveTBTeam struct {
	SyncResp
}

// LeaveTBTeam 离开组队BOSS的队伍: 离开组队BOSS的队伍
func (p *Account) LeaveTBTeam(r servers.Request) *servers.Response {
	req := new(reqMsgLeaveTBTeam)
	rsp := new(rspMsgLeaveTBTeam)

	initReqRsp(
		"Attr/LeaveTBTeamRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.LeaveTBTeamHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}


// TBTeamKick : 踢出组队boss某人
// 踢出组队boss某人

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgTBTeamKick 踢出组队boss某人请求消息定义
type reqMsgTBTeamKick struct {
	Req
	BeKickedAcid string `codec:"tbk_ac"` // 被踢的人的acid
	BeKickedTeamId string `codec:"tbk_tid"` // 被踢的人所在的队伍id
}

// rspMsgTBTeamKick 踢出组队boss某人回复消息定义
type rspMsgTBTeamKick struct {
	SyncResp
}

// TBTeamKick 踢出组队boss某人: 踢出组队boss某人
func (p *Account) TBTeamKick(r servers.Request) *servers.Response {
	req := new(reqMsgTBTeamKick)
	rsp := new(rspMsgTBTeamKick)

	initReqRsp(
		"Attr/TBTeamKickRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.TBTeamKickHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}


// GetRedBoxCostHC : 房间中勾选花费钻石一定获得红宝箱
// 房间中勾选花费钻石一定获得红宝箱

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgGetRedBoxCostHC 房间中勾选花费钻石一定获得红宝箱请求消息定义
type reqMsgGetRedBoxCostHC struct {
	Req
	TBCostTeamID string `codec:"t_rb_tid"` // 勾选红宝箱的teamid
	IsTickRedBox int64 `codec:"t_rb"` // 是勾选还是放弃勾选红宝箱 0不勾 1勾
}

// rspMsgGetRedBoxCostHC 房间中勾选花费钻石一定获得红宝箱回复消息定义
type rspMsgGetRedBoxCostHC struct {
	SyncResp
}

// GetRedBoxCostHC 房间中勾选花费钻石一定获得红宝箱: 房间中勾选花费钻石一定获得红宝箱
func (p *Account) GetRedBoxCostHC(r servers.Request) *servers.Response {
	req := new(reqMsgGetRedBoxCostHC)
	rsp := new(rspMsgGetRedBoxCostHC)

	initReqRsp(
		"Attr/GetRedBoxCostHCRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetRedBoxCostHCHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}


// GetTBMemberInfo : 获取组队BOSS的队友信息
// 获取组队BOSS的队友信息

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑


// reqMsgGetTBMemberInfo 获取组队BOSS的队友信息请求消息定义
type reqMsgGetTBMemberInfo struct {
	Req
	TeamMemAcid string `codec:"t_m_id"` // 队友的acid
	TeamMemTeamId string `codec:"t_m_tid"` // 队友的TeamId
}

// rspMsgGetTBMemberInfo 获取组队BOSS的队友信息回复消息定义
type rspMsgGetTBMemberInfo struct {
	SyncResp
	TBTeammate []byte `codec:"t_tm"` // 队友信息
}

// GetTBMemberInfo 获取组队BOSS的队友信息: 获取组队BOSS的队友信息
func (p *Account) GetTBMemberInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGetTBMemberInfo)
	rsp := new(rspMsgGetTBMemberInfo)

	initReqRsp(
		"Attr/GetTBMemberInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetTBMemberInfoHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}


