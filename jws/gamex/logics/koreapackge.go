package logics

import (
	"vcs.taiyouxi.net/platform/planx/servers"
)

// GetPackageInfo : 获取兑礼包信息
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetPackageInfo 获取兑礼包信息请求消息定义
type reqMsgGetPackageInfo struct {
	Req
}

// rspMsgGetPackageInfo 获取兑礼包信息回复消息定义
type rspMsgGetPackageInfo struct {
	SyncResp
	PackagePropInfo [][]byte `codec:"package_prop_info"` // 礼包信息
}

// GetPackageInfo 获取兑礼包信息:
func (p *Account) GetPackageInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGetPackageInfo)
	rsp := new(rspMsgGetPackageInfo)

	initReqRsp(
		"Attr/GetPackageInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetPackageInfoHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	//rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// PackagePropInfo 获取兑礼包信息
type PackagePropInfo struct {
	PackageID     int64    `codec:"package_id"`           // 礼包ID
	PackageType   int64    `codec:"package_type"`         // 礼包类型
	PackageName   string   `codec:"package_name"`         // 礼包名称
	IapId         string   `codec:"iap_id"`               // 购买iapid
	VipLevel      int64    `codec:"vip_level"`            // vip等级
	LimitType     int64    `codec:"limit_type"`           // 限购类型
	Limitcount    int64    `codec:"limit_buy_t"`          // 限购数量
	Count         int64    `codec:"have_count"`           // 已经购买数量
	PackageItem   []string `codec:"package_item"`         // 包内物品id
	PackageCount  []int64  `codec:"package_count"`        // 包内物品数量
	StartTime     int64    `codec:"start_time"`           // 开始时间
	SubPackageId  int64    `codec:"sub_package_id"`       // 子礼包id
	EndTime       int64    `codec:"end_time"`             // 结束时间
	BackRatio     int64    `codec:"back_ratio"`           // 返还比例
	HCValue       int64    `codec:"hc_value"`             // 钻石数量
	ShowValue     int64    `codec:"show_value"`           // 原价
	ConditionType int64    `codec:"condition_type"`       // 条件类型
	ConditionIp1  int64    `codec:"condition_ip_1"`       // 条件数值参数1
	ConditionIp2  int64    `codec:"condition_ip_2"`       // 条件数值参数2
	ConditionSp1  string   `codec:"condition_sp_1"`       // 条件数值参数1
	ConditionSp2  string   `codec:"condition_sp_2"`       // 条件数值参数2
	CanBuyPackage int64    `codec:"can_buy_this_package"` // 是否满足条件
	CurrentPos    int64    `codec:"cur_pos"`              // 阶梯礼包购买阶段
	QuestName     string   `codec:"quest_name"`           // 任务名称
	QuestDes      string   `codec:"quest_des"`            // 任务描述
	BuyPackage    int64    `codec:"have_buy_package"`     // 是否购买条件礼包
	Progress      int64    `codec:"quest_progress"`       // 当前进度
	All           int64    `codec:"quest_all"`            // 总进度
	BackImage     int64    `codec:"back_image"`           // 背景模板
}

// GetSpecialPackageInfo : 获取特殊礼包信息
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgGetSpecialPackageInfo 获取特殊礼包信息请求消息定义
type reqMsgGetSpecialPackageInfo struct {
	Req
	SpackageNum int64 `codec:"special_package_number"` // 特殊礼包号
}

// rspMsgGetSpecialPackageInfo 获取特殊礼包信息回复消息定义
type rspMsgGetSpecialPackageInfo struct {
	SyncResp
	ContinueShow int64 `codec:"continue_show"` // 是否继续
}

// GetSpecialPackageInfo 获取特殊礼包信息:
func (p *Account) GetSpecialPackageInfo(r servers.Request) *servers.Response {
	req := new(reqMsgGetSpecialPackageInfo)
	rsp := new(rspMsgGetSpecialPackageInfo)

	initReqRsp(
		"Attr/GetSpecialPackageInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.GetSpecialPackageInfoHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// ReceiveConditionPackage : 领取条件礼包
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgReceiveConditionPackage 领取条件礼包请求消息定义
type reqMsgReceiveConditionPackage struct {
	Req
	PackageId    int64 `codec:"package_id"`     // 礼包号
	SubPackageId int64 `codec:"sub_package_id"` // 礼包号
}

// rspMsgReceiveConditionPackage 领取条件礼包回复消息定义
type rspMsgReceiveConditionPackage struct {
	SyncRespWithRewards
}

// ReceiveConditionPackage 领取条件礼包:
func (p *Account) ReceiveConditionPackage(r servers.Request) *servers.Response {
	req := new(reqMsgReceiveConditionPackage)
	rsp := new(rspMsgReceiveConditionPackage)

	initReqRsp(
		"Attr/ReceiveConditionPackageRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.ReceiveConditionPackageHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// CloseSendInfo : 关闭页面发送信息
//

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgCloseSendInfo 关闭页面发送信息请求消息定义
type reqMsgCloseSendInfo struct {
	Req
	PackageId    []int64 `codec:"package_id"`     // 礼包号
	SubPackageId []int64 `codec:"sub_package_id"` // 礼包号
}

// rspMsgCloseSendInfo 关闭页面发送信息回复消息定义
type rspMsgCloseSendInfo struct {
	SyncResp
}

// CloseSendInfo 关闭页面发送信息:
func (p *Account) CloseSendInfo(r servers.Request) *servers.Response {
	req := new(reqMsgCloseSendInfo)
	rsp := new(rspMsgCloseSendInfo)

	initReqRsp(
		"Attr/CloseSendInfoRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	warnCode := p.CloseSendInfoHandler(req, rsp)
	if warnCode != 0 {
		return rpcWarn(rsp, warnCode)
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}
