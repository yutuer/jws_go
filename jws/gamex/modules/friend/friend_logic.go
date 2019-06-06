package friend

import (
	"time"

	"fmt"

	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (m *friendModule) scanPlayer(cursor int) ([]string, int) {
	//defer timeTrack(time.Now(), "ScanPlayer_V1")
	p := make([]string, 0, countPerScan)
	conn := modules.GetDBConn()
	defer conn.Close()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by nil error")
		return p, cursor
	}
	reply := []interface{}{}
	r, err := redis.Values(conn.Do("ZSCAN", rank.TableRankCorpGs(m.sid), cursor, "COUNT", countPerScan))
	if err != nil {
		logs.Error("redis value zscan err by %v", err)
		return p, cursor
	}
	reply = r
	if len(reply) < 2 {
		logs.Error("redis value zscan err by reply illegal(len < 2)")
		return p, cursor
	}
	newCur, err := redis.Int(reply[0], nil)
	if err != nil {
		logs.Error("redis Int reply err by %v", err)
		return p, cursor
	}
	info, err := redis.Strings(reply[1], nil)
	if err != nil {
		logs.Error("redis Strings reply err by %v", err)
		return p, cursor
	}
	for i := 0; i+1 < len(info); i += 2 {
		key := info[i]
		p = append(p, key)
	}
	return p, newCur
}

func (m *friendModule) InitGiftInfo() {
	m.GiftChan = make(chan GiftCmd, 100)
	m.initTimer()
	m.GiftInfo.Init()
	err := m.GiftInfo.loadGiftInfo(m.sid)
	if err != nil {
		panic(fmt.Sprintf("start friend module load db error by %v", err))
	}
}

func (m *friendModule) initTimer() {
	m.timeChan = time.After(save_interval)
}

func (m *friendModule) Init() {
	defer timeTrack(time.Now(), "InitCache_V1", m.sid)
	m.friendCache.Init()
	//m.friendCache.InitGauge(m.sid)
	p := make([]string, 0, 2000)
	conn := modules.GetDBConn()
	if conn.IsNil() {
		logs.Error("GetDBConn Err by nil error")
		return
	}
	reply := []interface{}{}
	r, err := redis.Values(conn.Do("ZRANGE", rank.TableRankCorpGs(m.sid), 0, -1, "WITHSCORES"))
	if err != nil {
		logs.Error("ZRANGE err by %b", err)
		return
	}
	conn.Close()
	reply = r
	if len(reply) <= 0 {
		logs.Warn("no data in corp rank")
		return
	}
	if len(reply)%2 != 0 {
		logs.Error("ZRANGE err by illegal format")
		return
	}
	info, err := redis.Strings(reply, nil)
	if err != nil {
		logs.Error("redis Strings reply err by %v", err)
		return
	}
	for i := 0; i+1 < len(info); i += 2 {
		key := info[i]
		p = append(p, key)
	}
	friends := GenFriendInfo(p)
	logs.Debug("get friend info count: %d", len(friends))
	m.friendCache.putInCache(friends)
}

func (m *friendModule) GetGSClosePlayer(gs int) []FriendSimpleInfo {
	id := m.getCloseID(gs)
	logs.Debug("get close gs id: %v", id)
	_, info := m.friendCache.queryFriendInfoAndRet(id)
	return info
}

func (m *friendModule) getCloseID(gs int) []string {
	conn := modules.GetDBConn()
	defer conn.Close()
	gs *= rank.RankByCorpDelayPowBase
	if conn.IsNil() {
		logs.Error("GetDBConn Err by nil error")
		return nil
	}
	r1, err := redis.Strings(conn.Do("ZRANGEBYSCORE", rank.TableRankCorpGs(m.sid), gs, "+inf", "LIMIT", 0, selectGSCloseNum/2))
	if err != nil {
		logs.Error("ZRANGE err by %", err)
		return nil
	}
	r2, err := redis.Strings(conn.Do("ZREVRANGEBYSCORE", rank.TableRankCorpGs(m.sid), gs, "-inf", "LIMIT", 0, selectGSCloseNum/2))
	if err != nil {
		logs.Error("ZRANGE err by %", err)
		return nil
	}
	return append(r1, r2[:]...)
}

func (m *friendModule) GetFriendInfo(players []string, needNew bool) []FriendSimpleInfo {
	defer timeTrack(time.Now(), "GetFriendInfo", m.sid)
	if needNew {
		return GenFriendInfo(players)
	} else {
		leaveID, retInfo := m.friendCache.queryFriendInfoAndRet(players)
		if len(leaveID) > 0 {
			info := GenFriendInfo(leaveID)
			retInfo = append(retInfo, info[:]...)
		}
		return retInfo
	}
}

func (m *friendModule) GetOneFriendInfo(player string, needNew bool) FriendSimpleInfo {
	defer timeTrack(time.Now(), "GetOneFriendInfo", m.sid)
	info := m.friendCache.findFriendInfo(player)
	if !info.IsNil() {
		return info
	} else {
		info := GenFriendInfo([]string{player})
		if len(info) > 0 {
			return info[0]
		}
	}
	logs.Debug("No find FriendInfo, id: %s", player)
	return FriendSimpleInfo{}
}

func (m *friendModule) UpdateFriendInfo(info *helper.AccountSimpleInfo, logOutTime int64) {
	defer timeTrack(time.Now(), "UpdateFriendInfo", m.sid)
	friendInfo := FriendSimpleInfo{
		AcID:        info.AccountID,
		Name:        info.Name,
		GS:          info.CurrCorpGs,
		Lv:          int(info.CorpLv),
		VIPLv:       int(info.Vip),
		Avatar:      info.CurrAvatar,
		LastActTime: logOutTime,
	}

	m.friendCache.putInCache([]FriendSimpleInfo{friendInfo})
}

func (m *friendModule) UpdateFriendList(friendInfo []FriendSimpleInfo, needNew bool) []FriendSimpleInfo {
	ids := make([]string, 0, len(friendInfo))
	for _, info := range friendInfo {
		ids = append(ids, info.AcID)
	}
	ret := m.GetFriendInfo(ids, needNew)
	if len(ret) != len(friendInfo) {
		logs.Warn("have friend info no find")
	}
	return ret
}

func (m *friendModule) GiftCommandExec(cmd GiftCmd) *GiftRet {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()
	cmd.Ret = make(chan GiftRet)
	select {
	case m.GiftChan <- cmd:
	case <-ctx.Done():
		logs.Error("friend moduel gift cmdChan is full")
	}
	select {
	case ret := <-cmd.Ret:
		return &ret
	case <-ctx.Done():
		logs.Error("friend moduel gift cmdChan apply <-retChan timeout")
		return nil
	}
}

func (m *friendModule) GiftCommandExecAsync(cmd GiftCmd) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()
	select {
	case m.GiftChan <- cmd:
	case <-ctx.Done():
		logs.Error("friend moduel gift cmdChan is full")
	}
}

func (m *friendModule) handleGiftCmd(cmd *GiftCmd) {
	switch cmd.Typ {
	case GiftCmd_Give:
		if v, ok := cmd.Param.([]string); ok {
			m.GiftInfo.PutGift(v[0], v[1])
		} else {
			logs.Error("fatal error convert type")
		}
	case GiftCmd_Receive:
		if v, ok := cmd.Param.([]string); ok {
			cmd.Ret <- GiftRet{
				Ret: m.GiftInfo.ReceiveGift(v[0], v[1]),
			}

		} else {
			logs.Error("fatal error convert type")
		}
	case GiftCmd_GetInfo:
		if v, ok := cmd.Param.([]string); ok {
			var removeItem []string
			if len(v) > 1 {
				removeItem = v[2:]
			}
			cmd.Ret <- GiftRet{
				Ret: m.GiftInfo.GetGiftInfo(v[0], len(v) > 1, removeItem),
			}
		} else {
			logs.Error("fatal error convert type")
		}
	}
}
