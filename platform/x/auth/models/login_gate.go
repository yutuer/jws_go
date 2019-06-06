package models

import (
	"encoding/json"
	"fmt"
	"time"

	"vcs.taiyouxi.net/platform/planx/servers/db"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/auth/config"
)

// GateInfo infomation structure...
type GateInfo struct {
	GID     uint
	ShardID uint
	CCU     int64
	// GameIPAddrPort 通常是公网ip:port
	GameIPAddrPort string
	// RPCIPAddrPort 通常是内网ip:port !千万不要开放成公网IP!
	RPCIPAddrPort string
}

const gateZSetKey string = "gates:%d"    //sid shard id
const gateIPExpKey string = "gate:%d:%s" //sid, ip
const gateInfoKey string = "gateinfo:%s" //ip
func makeGateInfoKey(addrPort string) string {
	return fmt.Sprintf(gateInfoKey, addrPort)
}

// AddOneGate 注册当前的Gate主机信息，由3个Key实现
// SETEX gate:0:127.0.0.1:8866 5mins CCU 用来实现Gate注册心跳。必须每5分钟注册一次。
// ZADD gates:0 CCU 127.0.0.1:8866 用来进行CCU排序，方便login服务能够快速获取负载最小的gate ip给client
// HMSET gateinfo:127.0.0.1:8866 gameip:ipaddr rpcip:rpcip first:uuid用于服务发现
func AddOneGate(gate GateInfo) (bool, error) {

	db := loginRedisPool.Get()
	defer db.Close()

	zsetKey := fmt.Sprintf(gateZSetKey, gate.ShardID)
	key := fmt.Sprintf(gateIPExpKey, gate.ShardID, gate.GameIPAddrPort)
	infoKey := makeGateInfoKey(gate.GameIPAddrPort)
	db.Send("SETEX", key, 60*5, gate.CCU)
	db.Send("ZADD", zsetKey, gate.CCU, gate.GameIPAddrPort)
	db.Send("HMSET", redis.Args{}.Add(infoKey).
		Add("gameip").Add(gate.GameIPAddrPort).
		Add("rpcip").Add(gate.RPCIPAddrPort)...)
	//只有这个键值不存在的时候才会生成新的。方便得知这个Gate如果意外重启后，
	//通过这个键值能够得知是不是这个Gate的状态都应该重置
	//db.Send("HSETNX", redis.Args{}.Add(infoKey).
	//Add("first").Add(uuid.NewV4().String()))...)
	_, err := db.Do("")
	if err != nil {
		logs.Error("models Add One Gate Failed", err.Error())
		return false, err
	}

	//_, err = db.Do("HMSET", infoKey, redis.Args{}.
	//Add("gameip").Add(gate.GameIPAddrPort).
	//Add("rpcip").Add(gate.RPCIPAddrPort))
	//if err != nil {
	//logs.Error("models Add One Gate Failed", err.Error())
	//return false, err
	//}

	return true, nil
}

func CleanGate(gate *GateInfo, reason string) error {
	if !config.Cfg.IsRunModeLocal() {
		keyIp := fmt.Sprintf("%s/%d/%d/%s", config.Cfg.EtcdRoot, gate.GID, gate.ShardID, etcd.KeyIp)
		if err := etcd.Delete(keyIp); err != nil {
			logs.Error("Remove a Gate from etcd key %s err %v", keyIp, err)
		}
		keyInternalIp := fmt.Sprintf("%s/%d/%d/%s", config.Cfg.EtcdRoot, gate.GID, gate.ShardID, etcd.KeyInternalIp)
		if err := etcd.Delete(keyInternalIp); err != nil {
			logs.Error("Remove a Gate from etcd key %s err %v", keyInternalIp, err)
		}
	} else {
		db := loginRedisPool.Get()
		defer db.Close()

		zsetKey := fmt.Sprintf(gateZSetKey, gate.ShardID)
		key := fmt.Sprintf(gateIPExpKey, gate.ShardID, gate.GameIPAddrPort)
		infoKey := makeGateInfoKey(gate.GameIPAddrPort)
		//infoKey := fmt.Sprintf(gateInfoKey, gate.GameIPAddrPort)

		db.Send("DEL", key, infoKey)
		db.Send("ZREM", zsetKey, gate.GameIPAddrPort)
		logs.Info("models, CleanGate ", gate, " due to %s", reason)
		_, err := db.Do("")
		if err != nil {
			logs.Error("Remove a Gate, but got db error, %s", err.Error())
		}
	}
	return nil
}

// GetOneGate 通常情况下，找出方式应该是找出CCU最少的那个Gate给玩家使用。
// 在下线某服务器的情况下，可以使得指定Gate上的用户越来越少，实现Gate服务器下线的平滑运维。
func GetOneGate(gid, sid uint) (gate *GateInfo, err error) {
	if !config.Cfg.IsRunModeLocal() {
		keyIp := fmt.Sprintf("%s/%d/%d/%s", config.Cfg.EtcdRoot, gid, sid, etcd.KeyIp)
		ip, err := etcd.Get(keyIp)
		if err != nil {
			return nil, fmt.Errorf("models.login.GetOneGate get ip from etcd err %v", err)
		}
		keyInternalIp := fmt.Sprintf("%s/%d/%d/%s", config.Cfg.EtcdRoot, gid, sid, etcd.KeyInternalIp)
		internalip, err := etcd.Get(keyInternalIp)
		if err != nil {
			return nil, fmt.Errorf("models.login.GetOneGate get internalip from etcd err %v", err)
		}
		var g GateInfo
		g.GID = gid
		g.ShardID = sid
		g.GameIPAddrPort = ip
		g.RPCIPAddrPort = internalip
		logs.Debug("models.login.GetOneGate %v", g)
		return &g, nil
	} else {
		db := loginRedisPool.Get()
		defer db.Close()

		logs.Trace("models.login.GetOneGate get sid: %d", sid)
		zsetKey := fmt.Sprintf(gateZSetKey, sid)
	find:
		values, err := redis.Values(db.Do("ZRANGE", zsetKey, 0, 0, "WITHSCORES"))
		if err == nil && len(values) > 0 {
			logs.Trace("models.login.GetOneGate GetOne %d", len(values))
			var g GateInfo
			g.GID = gid
			g.ShardID = sid
			values, err = redis.Scan(values, &g.GameIPAddrPort, &g.CCU)
			if err == nil {
				//TODO IP可能因为AutoScaling等原因有可能不存在了，需要根据另外一个键值校验是否gate:ip过期
				//这里如果为transaction会更加安全 T4190

				logs.Trace("models.login.GetOneGate values: %v", g)
				key := fmt.Sprintf(gateIPExpKey, sid, g.GameIPAddrPort)
				exists, _ := redis.Bool(db.Do("EXISTS", key))
				if !exists {
					CleanGate(&g, "gate TTL key not exist")
					//db.Do("ZREM", zsetKey, g.GameIPAddrPort)
					goto find
				}
			} else {
				logs.Error("models.login.GetOneGate Failed, %s", err.Error())
			}

			//get rpc address fill into GateInfo
			infoKey := makeGateInfoKey(g.GameIPAddrPort)
			//infoKey := fmt.Sprintf(gateInfoKey, g.GameIPAddrPort)
			g.RPCIPAddrPort, err = redis.String(db.Do("HGET", infoKey, "rpcip"))
			if err != nil {
				return nil, fmt.Errorf("models.login.GetOneGate, not found rpcip")
			}
			logs.Debug("models.login.GetOneGate %v, %d", g, len(values))
			return &g, nil
		}
	}
	return nil, XErrGetGateNotExist
}

type LoginStatueCode int

const (
	LSC_GETGATE LoginStatueCode = 1 + iota
	LSC_LOGIN
	LSC_LOGOFF
)

type loginStatus struct {
	Sid           uint
	RPCGateIPAddr string
	LoginToken    string
	AccountID     string
	Timestamp     int64
	Status        int
}

func makeLoginStatusKey(gid uint, uid db.UserID) string {
	return fmt.Sprintf("ls:%d:%s", gid, uid)
}

func GetLastLoginStatus(gid uint, uid db.UserID) (
	OldAccountID, OldToken, OldRPC string, HasOldValue bool) {

	db := loginRedisPool.Get()
	defer db.Close()

	rk_game_login_status := makeLoginStatusKey(gid, uid)
	oldvalue, err := redis.Bytes(db.Do("GET", rk_game_login_status))
	if err != nil {
		logs.Warn("models.GetLastLoginStatus err: %v %s", err.Error(), rk_game_login_status)
		return
	}

	if oldvalue == nil {
		HasOldValue = false
		return
	}
	var oldls loginStatus
	err = json.Unmarshal(oldvalue, &oldls)
	if err != nil {
		logs.Warn("models.GetLastLoginStatus unmarshal oldvalue failed: ", err)
		return
	} else {
		//TODO 根据旧值，需要踢出前面登录的玩家
		OldToken = oldls.LoginToken
		OldRPC = oldls.RPCGateIPAddr
		OldAccountID = oldls.AccountID
		HasOldValue = true
		logs.Debug("models.GetLastLoginStatus old status should logout, gid: %v, uid:%s", gid, uid.String())
	}
	return
}

// DeleteLoginStatus 通常因为某种原因，需要主动删除错误信息
func DeleteLoginStatus(gid uint, uid db.UserID) {

	db := loginRedisPool.Get()
	defer db.Close()

	rk_game_login_status := makeLoginStatusKey(gid, uid)

	logs.Debug("LoginStatus del %s", rk_game_login_status)
	if _, err := db.Do("DEL", rk_game_login_status); err != nil {
		logs.Warn("models.DeleteLoginStatus DEL key error: ", rk_game_login_status, err.Error())
	}
}

// UpdateLoginStatus 用于控制玩家在一个gameid下，同时只能玩一个帐号
// 使用gid:uid为唯一标识，存储上次玩家的登录信息
func UpdateLoginStatus(account db.Account,
	gateRPCAddr, loginToken string,
	status LoginStatueCode) {

	db := loginRedisPool.Get()
	defer db.Close()

	rk_game_login_status := makeLoginStatusKey(account.GameId, account.UserId)

	ls := loginStatus{
		Sid:           account.ShardId,
		RPCGateIPAddr: gateRPCAddr,
		LoginToken:    loginToken,
		AccountID:     account.String(),
		Status:        int(status),
		Timestamp:     time.Now().Unix(),
	}
	if value, err := json.Marshal(&ls); err != nil {
		logs.Error("models.updateloginstatus err json Marshal with:", err, ls)
	} else {
		//loginStatus 设置，24小时过期，长期不回来玩的用户的登录状态不维护
		logs.Debug("UpdateLoginStatus SETEX %s", rk_game_login_status)
		_, err := db.Do("SETEX", rk_game_login_status, 60*60*24, value)
		if err != nil {
			logs.Error("models.updateloginstatus err with db with :", err)
		} else {

		}
	}
	return
}
