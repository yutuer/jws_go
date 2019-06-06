package logics

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/modules/friend"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

/*
	好友系统
	好友信息由内存维护，更新排行榜时同时更新好友信息
*/

// AddBlackList : 将指定玩家添加入黑名单
//

// reqMsgAddBlackList 将指定玩家添加入黑名单请求消息定义
type reqMsgAddBlackList struct {
	Req
	ID string `codec:"id"` // 玩家ID
}

// rspMsgAddBlackList 将指定玩家添加入黑名单回复消息定义
type rspMsgAddBlackList struct {
	SyncResp
}

// AddBlackList 将指定玩家添加入黑名单:
func (p *Account) AddBlackList(r servers.Request) *servers.Response {
	req := new(reqMsgAddBlackList)
	rsp := new(rspMsgAddBlackList)

	initReqRsp(
		"Attr/AddBlackListRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	checkAccountIDInShard(req.ID, p.AccountID.ShardId, fmt.Sprintf("add blacklist by %s", p.AccountID.String()))
	if req.ID == p.AccountID.String() {
		return rpcWarn(rsp, errCode.AddSelfBlackList)
	}
	friendInfo := friend.GetModule(p.AccountID.ShardId).GetOneFriendInfo(req.ID, false)
	if friendInfo.IsNil() {
		return rpcWarn(rsp, errCode.AddBlackListError)
	}
	if !p.Friend.AddBlack(friendInfo) {
		return rpcWarn(rsp, errCode.AddBlackListError)
	}
	// 如果在好友列表中，需要移除
	p.Friend.RemoveFriend(req.ID)
	rsp.OnChangeFriendList()
	rsp.OnChangeBlackList()
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// AddFriend : 添加好友
//

// reqMsgAddFriend 添加好友请求消息定义
type reqMsgAddFriend struct {
	Req
	ID string `codec:"id"` // 玩家ID
}

// rspMsgAddFriend 添加好友回复消息定义
type rspMsgAddFriend struct {
	SyncResp
}

// AddFriend 添加好友:
func (p *Account) AddFriend(r servers.Request) *servers.Response {
	req := new(reqMsgAddFriend)
	rsp := new(rspMsgAddFriend)

	initReqRsp(
		"Attr/AddFriendRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	checkAccountIDInShard(req.ID, p.AccountID.ShardId, fmt.Sprintf("add friend by %s", p.AccountID.String()))

	if req.ID == p.AccountID.String() {
		return rpcWarn(rsp, errCode.AddSelfFriend)
	}
	friendInfo := friend.GetModule(p.AccountID.ShardId).GetOneFriendInfo(req.ID, false)
	if friendInfo.IsNil() {
		return rpcWarn(rsp, errCode.AddFriendError)
	}
	if !p.Friend.AddFriend(friendInfo) {
		return rpcWarn(rsp, errCode.AddFriendError)
	}

	rsp.OnChangeFriendList()

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// FindFriend : 根据名字查找好友

// reqMsgFindFriend 根据名字查找好友请求消息定义
type reqMsgFindFriend struct {
	Req
	Name string `codec:"name"` // 玩家名字
}

// rspMsgFindFriend 根据名字查找好友回复消息定义
type rspMsgFindFriend struct {
	SyncResp
	PlayerInfo []byte `codec:"player_info"` // 玩家信息
}

// FindFriend 根据名字查找好友: 打开限时名将ui，获取排行等信息
func (p *Account) FindFriend(r servers.Request) *servers.Response {
	req := new(reqMsgFindFriend)
	rsp := new(rspMsgFindFriend)

	initReqRsp(
		"Attr/FindFriendRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	sid := p.AccountID.ShardId
	ID, err := p.Friend.GetFriendIDByName(sid, req.Name)
	if err != nil {
		logs.Warn("getfriendidbyname err by %v", err)
		return rpcWarn(rsp, errCode.NoFindPlayer)
	}
	checkAccountIDInShard(ID, p.AccountID.ShardId, fmt.Sprintf("find friend by %s", p.AccountID.String()))

	friendInfo := friend.GetModule(sid).GetOneFriendInfo(ID, false)
	if friendInfo.IsNil() {
		return rpcWarn(rsp, errCode.NoFindPlayer)
	}
	if player_msg.GetModule(p.AccountID.ShardId).IsOnline(friendInfo.AcID) {
		friendInfo.LastActTime = 0
	}

	rsp.PlayerInfo = encode(friendInfo)

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// RemoveBlackList : 将指定玩家移除黑名单
//

// reqMsgRemoveBlackList 将指定玩家移除黑名单请求消息定义
type reqMsgRemoveBlackList struct {
	Req
	ID string `codec:"id"` // 玩家ID
}

// rspMsgRemoveBlackList 将指定玩家移除黑名单回复消息定义
type rspMsgRemoveBlackList struct {
	SyncResp
}

// RemoveBlackList 将指定玩家移除黑名单:
func (p *Account) RemoveBlackList(r servers.Request) *servers.Response {
	req := new(reqMsgRemoveBlackList)
	rsp := new(rspMsgRemoveBlackList)

	initReqRsp(
		"Attr/RemoveBlackListRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	p.Friend.RemoveBlack(req.ID)
	rsp.OnChangeBlackList()
	// logic imp end
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// RemoveFriend : 删除好友
//

// reqMsgRemoveFriend 删除好友请求消息定义
type reqMsgRemoveFriend struct {
	Req
	ID string `codec:"id"` // 玩家ID
}

// rspMsgRemoveFriend 删除好友回复消息定义
type rspMsgRemoveFriend struct {
	SyncResp
}

// RemoveFriend 删除好友:
func (p *Account) RemoveFriend(r servers.Request) *servers.Response {
	req := new(reqMsgRemoveFriend)
	rsp := new(rspMsgRemoveFriend)

	initReqRsp(
		"Attr/RemoveFriendRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	p.Friend.RemoveFriend(req.ID)
	rsp.OnChangeFriendList()
	// logic imp end
	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// UpdateRecentPlayer : 更新指定最近联系人的状态
//

// reqMsgUpdateRecentPlayer 更新指定最近联系人的状态请求消息定义
type reqMsgUpdateRecentPlayer struct {
	Req
	PlayerID []string `codec:"player_id"` // 联系人ID
}

// rspMsgUpdateRecentPlayer 更新指定最近联系人的状态回复消息定义
type rspMsgUpdateRecentPlayer struct {
	SyncResp
	PlayerInfo [][]byte `codec:"player_info"` // 联系人信息
}

// UpdateRecentPlayer 更新指定最近联系人的状态:
func (p *Account) UpdateRecentPlayer(r servers.Request) *servers.Response {
	req := new(reqMsgUpdateRecentPlayer)
	rsp := new(rspMsgUpdateRecentPlayer)

	initReqRsp(
		"Attr/UpdateRecentPlayerRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	if req.PlayerID == nil {
		return rpcSuccess(rsp)
	}
	rsp.PlayerInfo = make([][]byte, 0, len(req.PlayerID))
	nowT := p.GetProfileNowTime()
	needNew := nowT-p.Friend.LastUpdateRecentPlayerTime > friend.UpdateRecentPlayerInterval
	for _, id := range req.PlayerID {
		checkAccountIDInShard(id, p.AccountID.ShardId, fmt.Sprintf("update recent player by %s", p.AccountID.String()))
	}
	if needNew {
		friendInfo := friend.GetModule(p.AccountID.ShardId).GetFriendInfo(req.PlayerID, false)
		for _, info := range friendInfo {
			if player_msg.GetModule(p.AccountID.ShardId).IsOnline(info.AcID) {
				info.LastActTime = 0
			}
			rsp.PlayerInfo = append(rsp.PlayerInfo, encode(info))
		}
		p.Friend.LastUpdateRecentPlayerTime = nowT
	}

	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

// GetRecommendPlayer : 获得推荐玩家列表
//

// reqMsgGetRecommendPlayer 获得推荐玩家列表请求消息定义
type reqMsgGetRecommendPlayer struct {
	Req
}

// rspMsgGetRecommendPlayer 获得推荐玩家列表回复消息定义
type rspMsgGetRecommendPlayer struct {
	SyncResp
	PlayerInfo [][]byte `codec:"player_info"` // 玩家信息
}

// GetRecommendPlayer 获得推荐玩家列表:
func (p *Account) GetRecommendPlayer(r servers.Request) *servers.Response {
	req := new(reqMsgGetRecommendPlayer)
	rsp := new(rspMsgGetRecommendPlayer)

	initReqRsp(
		"Attr/GetRecommendPlayerRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	if p.Friend.NeedUpdateRecommendPlayer(p.GetProfileNowTime()) {
		p.Friend.UpdateRecommendPlayer(p.AccountID.String(),
			friend.GetModule(p.AccountID.ShardId).GetGSClosePlayer(p.Profile.GetData().CorpCurrGS))
		p.Friend.LastUpdateRecommendPlayerTime = p.GetProfileNowTime()
	}

	friends := p.Friend.GetRecommendPlayer()
	logs.Debug("recommendPlayer: %v", friends)
	for i, f := range friends {
		if player_msg.GetModule(p.AccountID.ShardId).IsOnline(f.AcID) {
			friends[i].LastActTime = 0
		}
	}
	rsp.PlayerInfo = make([][]byte, 0, friend.FriendCountPerRet)
	ret := friend.SelectFriendInfo(friends)
	logs.Debug("select recommendPlayer: %v", ret)
	for _, f := range ret {
		if !f.IsNil() {
			rsp.PlayerInfo = append(rsp.PlayerInfo, encode(f))
		}
	}
	// logic imp end

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

func checkAccountIDInShard(id string, sid uint, reason string) {
	acc, err := db.ParseAccount(id)
	if err != nil || game.Cfg.GetShardIdByMerge(acc.ShardId) != game.Cfg.GetShardIdByMerge(sid) {
		logs.Warn("parse account error for id: %d  cause: %s", id, reason)
	}
}
