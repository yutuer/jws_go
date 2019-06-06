package controllers

import (
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strconv"

	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/secure"
	"vcs.taiyouxi.net/platform/x/auth/config"

	"time"

	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/gate"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
	"vcs.taiyouxi.net/platform/x/auth/errorctl"
	"vcs.taiyouxi.net/platform/x/auth/limit"
	"vcs.taiyouxi.net/platform/x/auth/logiclog"
	"vcs.taiyouxi.net/platform/x/auth/models"
)

// LoginController is in charge of login process.
type LoginController struct {
	PlanXController
}

// FetchShardsInfo 客户端获取最新服务器配置情况表
// 目前从redis中读取实时，考虑缓存在内存中，通过运维指令再reload数据(T2141 DONE)
// TODO 更进一步的想法是用cdn进行负载平衡
// gid: game id, ios:1
func (lc *LoginController) FetchShardsInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		gids := c.Param("gid")
		logs.Info("gid: %s", gids)
		gid, err := strconv.Atoi(gids)
		if err != nil {
			errorctl.CtrlErrorReturn(c, "[Login.FetchShards]", err, errorctl.ClientErrorFormatVerifyFailed)
			return
		}

		shards, err := models.FetchShardsInfo(gid)
		if err != nil {
			errorctl.CtrlErrorReturn(c, "[Login.FetchShards]", err, errorctl.ClientErrorMaybeDBProblem)
			return
		}
		// 应用宝大区的也能看到200的服务器列表
		if gid == 203 {
			_shards, err := models.FetchShardsInfo(200)
			if err == nil {
				shards = append(shards, _shards...)
			}
		}

		// auth在体验服状态
		// ip白名单的玩家可以看到体验服
		// 限制ip的玩家还能看到正常的服务器，但进入时提示服务器维护中
		if config.Cfg.InnerExperience {
			res_shards := make([]models.ShardInfoForClient, 0, len(shards))
			ip := c.ClientIP()
			logs.Info("FetchShardsInfo clientip %s", ip)
			if !limit.IsLimitedByRemoteIP(ip) { // 限制ip的玩家只能看到正常的服务器
				for _, shard := range shards {
					show_state, err := strconv.Atoi(shard.ShowState)
					if err != nil {
						show_state = etcd.ShowState_Count // 没配showstate的话，就给个不存在的值
					}
					if show_state != etcd.ShowState_Experience {
						res_shards = append(res_shards, shard)
					}
				}
			} else { // ip白名单的玩家只能看到体验服
				for _, shard := range shards {
					show_state, err := strconv.Atoi(shard.ShowState)
					if err != nil {
						show_state = etcd.ShowState_Count // 没配showstate的话，就给个不存在的值
					}
					if show_state == etcd.ShowState_Experience {
						res_shards = append(res_shards, shard)
					}
				}
			}
			shards = res_shards
		}

		r := struct {
			Result string                      `json:"result"`
			Shards []models.ShardInfoForClient `json:"shards"`
		}{
			"ok",
			shards,
		}

		c.JSON(200, r)
	}
}

func (lc *LoginController) FetchShardsInfoV2() gin.HandlerFunc {
	return func(c *gin.Context) {
		gids := c.Param("gid")
		logs.Info("<featch shard info> %s", gids)
		ver := lc.GetString(c, "ver")
		authToken := lc.GetString(c, "at") // AuthToken UUID
		sp := lc.GetString(c, "sp")
		gid, err := strconv.Atoi(gids)
		if err != nil {
			errorctl.CtrlErrorReturn(c, "[Login.FetchShards]", err, errorctl.ClientErrorFormatVerifyFailed)
			return
		}

		logs.Debug("FetchShardsInfoV2 param %v %v", gids, ver)

		//验证authToken有效，不区分game id，不区分shard id
		uid, err := models.LoginVerifyAuthToken(authToken)
		if err != nil {
			errorctl.CtrlErrorReturn(c, "[Login.FetchShardsInfo]", err, errorctl.ClientErrorLoginAuthtokenNotReady)
			return
		}

		isSuper := false
		if config.SuperUidsCfg.IsSuperUid(uid.String()) {
			isSuper = true
		}
		if !isSuper && len(ver) > 0 {
			ggInfo := config.GetGonggao(gids, ver)
			logs.Debug("FetchShardsInfoV2 Gonggao %v", ggInfo)
			if len(sp) > 0 && len(ggInfo.WhitelistPwd) > 0 {
				if sp == ggInfo.WhitelistPwd {
					isSuper = true
					logs.Debug("FetchShardsInfoV2 got super")
				}
			}
		}

		shards, err := models.FetchShardsInfo(gid)
		if err != nil {
			errorctl.CtrlErrorReturn(c, "[Login.FetchShards]", err, errorctl.ClientErrorMaybeDBProblem)
			return
		}
		logs.Info("<featch shard info> %v", shards)
		// 应用宝大区的也能看到200的服务器列表
		if gid == 203 {
			_shards, err := models.FetchShardsInfo(200)
			if err == nil {
				shards = append(shards, _shards...)
			}
		}

		// 若不是超级用户，只能看到正常上线了的shard
		if !isSuper {
			_shards := make([]models.ShardInfoForClient, 0, len(shards))
			for _, s := range shards {
				// 不是超级用户，看不到ShowState_Super的服
				show_state, err := strconv.Atoi(s.ShowState)
				if err != nil {
					show_state = etcd.ShowState_Count // 没配showstate的话，就给个不存在的值
				}
				if show_state == etcd.ShowState_Super {
					continue
				}
				if s.GetState() == etcd.StateOnline {
					_shards = append(_shards, s)
				}
			}
			shards = _shards
		}

		// auth在体验服状态
		// ip白名单的玩家可以看到体验服
		// 限制ip的玩家还能看到正常的服务器，但进入时提示服务器维护中
		if config.Cfg.InnerExperience {
			res_shards := make([]models.ShardInfoForClient, 0, len(shards))
			ip := c.ClientIP()
			logs.Info("FetchShardsInfo clientip %s", ip)
			if !isSuper && !limit.IsLimitedByRemoteIP(ip) { // 限制ip的玩家只能看到正常的服务器
				for _, shard := range shards {
					show_state, err := strconv.Atoi(shard.ShowState)
					if err != nil {
						show_state = etcd.ShowState_Count // 没配showstate的话，就给个不存在的值
					}
					if show_state != etcd.ShowState_Experience {
						res_shards = append(res_shards, shard)
					}
				}
			} else { // ip白名单的玩家只能看到体验服
				for _, shard := range shards {
					show_state, err := strconv.Atoi(shard.ShowState)
					if err != nil {
						show_state = etcd.ShowState_Count // 没配showstate的话，就给个不存在的值
					}
					if show_state == etcd.ShowState_Experience {
						res_shards = append(res_shards, shard)
					}
				}
			}
			shards = res_shards
		}

		r := struct {
			Result string                      `json:"result"`
			Shards []models.ShardInfoForClient `json:"shards"`
		}{
			"ok",
			shards,
		}
		logs.Info("<featch shard info> %v", r)
		c.JSON(200, r)
	}
}

//防止同一个帐号多重login, 通知上一个登录的设备退出
func avoidMultipleLogin(gid uint, uid db.UserID, reason string, banTime int64) (status string) {
	OldAccountID, OldLoginToken,
		OldRpc, ok := models.GetLastLoginStatus(gid, uid)

	if ok { //如果这个账户已有登录信息，尝试通知退出
		var oldConn *rpc.Client
		var reply int
		var err error
		if oldConn, err = jsonrpc.Dial("tcp", OldRpc); err == nil {
			defer oldConn.Close()
			err = oldConn.Call("GateServer.KickItOffline",
				&gate.KickOfflineParam{
					LoginToken: OldLoginToken,
					AccountID:  OldAccountID,
					Reason:     reason,

					AfterDuration:   5,
					NoLoginDuration: int(banTime),
				},
				&reply)

			if reply == 1 { // RPC通知成功，并已通知相关客户端下线, 请等一段时间后重试
				return "RETURN"
			} else {
				return "Gate told login server, not found that player"
			}
		}

		if err != nil {
			//相关服务器上RPC通知失败，则允许继续登录
			logs.Warn("[Login.getGate.call] kick out last login failed. rpc conn", err.Error())
			models.DeleteLoginStatus(gid, uid)
			return "Gate not exist"
		}

	}
	return "CONTINUE"
}

func notifyInfo(gid uint, uid db.UserID, info gate.InfoNotifyToClient) {
	OldAccountID, OldLoginToken,
		OldRpc, ok := models.GetLastLoginStatus(gid, uid)

	if ok { //如果这个账户已有登录信息，尝试通知退出
		var oldConn *rpc.Client
		var reply int
		var err error
		if oldConn, err = jsonrpc.Dial("tcp", OldRpc); err == nil {
			defer oldConn.Close()
			err = oldConn.Call("GateServer.NotifyInfo",
				&gate.NotifyInfoParam{
					LoginToken: OldLoginToken,
					AccountID:  OldAccountID,
					Info:       info,
				},
				&reply)
		}

		if err != nil {
			//相关服务器上RPC通知失败，则允许继续登录
			logs.Warn("[Login.getGate.call] notifyInfo last failed. rpc conn", err.Error())
		}

	}
}

// GetGate is used for client to get gate server on specific shards servers
// 如果返回值中有retry，则需要客户端等待ms毫秒后再次重试
func (lc *LoginController) GetGate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken := lc.GetString(c, "at") // AuthToken UUID
		//game id maybe used later for shard of redis cluster
		//gID := lc.GetString("gid")      // 渠道 game id ex. ios: 1
		shardName := lc.GetString(c, "sn") // 子服的ID -- shard name
		//Shard name 换算shard id(sid)
		sid, gid, showState, err := models.GetShardID(shardName)
		if err != nil {
			//数据库级错误, 键值不存在或者格式不对等，或者连接不上
			errorctl.CtrlErrorReturn(c, "[Login.getGate]", err, errorctl.ClientErrorMaybeDBProblem)
			return
		}

		ver := lc.GetString(c, "ver")
		sp := lc.GetString(c, "sp")
		logs.Debug("GetGate ver %v sp %v", ver, sp)

		show_state, err := strconv.Atoi(showState)
		if err != nil {
			show_state = etcd.ShowState_Count // 没配showstate的话，就给个不存在的值
		}

		//验证authToken有效，不区分game id，不区分shard id
		uid, err := models.LoginVerifyAuthToken(authToken)
		if err != nil {
			errorctl.CtrlErrorReturn(c, "[Login.getGate]", err, errorctl.ClientErrorLoginAuthtokenNotReady)
			return
		}
		if uid == db.InvalidUserID {
			errorctl.CtrlErrorReturn(c, "[Login.getGate]", fmt.Errorf("timeout"), errorctl.ClientErrorLoginAuthtokenNotReady)
			//errorctl.CtrlErrorReturn(c, "[Login.getGate]",
			//	fmt.Errorf("uid should not be 0, this should not happen!"),
			//	errorctl.ClientErrorUnknown)
			return
		}

		isSuper := false
		if config.SuperUidsCfg.IsSuperUid(uid.String()) {
			isSuper = true
		}
		if !isSuper && len(ver) > 0 {
			ggInfo := config.GetGonggao(fmt.Sprintf("%d", gid), ver)
			// 检查白名单密码
			if len(ggInfo.WhitelistPwd) > 0 && sp == ggInfo.WhitelistPwd {
				logs.Debug("GetGate got super")
				isSuper = true
			}
			if ggInfo.MaintainStartTime > 0 && ggInfo.MaintainEndTime > 0 {
				now_t := time.Now().Unix()
				if now_t >= ggInfo.MaintainStartTime &&
					now_t < ggInfo.MaintainEndTime { // 在阻挡公告时间内
					// 只有白名单的能进
					if !isSuper {
						errorctl.CtrlErrorReturn(c, "[Login.getGate]",
							fmt.Errorf("maintance"), errorctl.ClientErrorGetGateMaintenance)
						return
					}
				}
			}
		}

		// 根据auth是否体验服状态，和shard的显示状态，决定：
		// 服务器是否可见
		// 服务器是否限制ip
		if config.Cfg.InnerExperience { // auth 体验服
			if show_state != etcd.ShowState_Experience {
				errorctl.CtrlErrorReturn(c, "[Login.getGate]", fmt.Errorf("server invalid"), errorctl.ClientErrorGetGateMaintenance)
				return
			}
			ip := c.ClientIP()
			logs.Info("GetGate clientip %s", ip)
			if !isSuper && !limit.IsLimitedByRemoteIP(ip) {
				errorctl.CtrlErrorReturn(c, "[Login.getGate]", fmt.Errorf("iplimit %s", ip), errorctl.ClientErrorGetGateMaintenance)
				return
			}
		} else { // 正常状态
			if show_state == etcd.ShowState_Experience {
				errorctl.CtrlErrorReturn(c, "[Login.getGate]", fmt.Errorf("server invalid"), errorctl.ClientErrorGetGateMaintenance)
				return
			}
			if show_state == etcd.ShowState_Maintenance {
				ip := c.ClientIP()
				logs.Info("GetGate clientip %s", ip)
				if !isSuper {
					errorctl.CtrlErrorReturn(c, "[Login.getGate]", fmt.Errorf("iplimit %s", ip), errorctl.ClientErrorGetGateMaintenance)
					return
				}
			}
		}

		//生成LoginToken
		loginToken := uuid.NewV4().String()

		//防止同一个帐号多重login, 通知上一个登录的设备退出
		if st := avoidMultipleLogin(gid, uid, gate.ReLogin_Kick, 10); st == "RETURN" {
			c.JSON(200, map[string]interface{}{
				"result": "retry",
				"ms":     2000, //2000 ms后重试
			})
			return
		}

		var gateInfo *models.GateInfo
		for {
			//确认有效后，根据数据库记录给玩家返回有效的（人数少的服务器）Gate IP
			ob, err := models.GetOneGate(gid, sid)
			if ob == nil || err != nil {
				if err == models.XErrGetGateNotExist {
					errorctl.CtrlErrorReturn(c, "[Login.getGate]", err, errorctl.ClientErrorGetGateNotExist)
				} else {
					errorctl.CtrlErrorReturn(c, "[Login.getGate]", err, errorctl.ClientErrorUnknown)
				}
				return
			}
			logs.Trace("[Login.getGate] ob: %v, err:%v", ob, err)
			// loginToken, user id 发送到相关Gate server api， 由Gate Server保存在内存中。
			// FIXME loginToken在后续操作中会变成一个密钥，所以需要使用更高级的加密算法加密传输！
			// 怎么处理失败，要进行指数级别重试么？
			var conn *rpc.Client
			if conn, err = jsonrpc.Dial("tcp", ob.RPCIPAddrPort); err != nil {
				//CtrlErrorReturn(c, "[Login.getGate.notifygate]", err, ob)
				//beego.Error("[Login.getGate.notifygate] error %s, %v", err.Error(), ob)
				reason := fmt.Sprintf("[Login.getGate.notifygate] error %s, %v", err.Error(), ob)
				models.CleanGate(ob, reason)
				continue
			}
			var reply bool
			// get number ID
			account := db.Account{
				UserId:  uid,
				GameId:  gid,
				ShardId: sid,
			}
			numberID, err := models.GetNumberIDForPlayer(account.String())
			if err != nil {
				logs.Error("get number id for player err by %v", err)
			}
			if err = conn.Call("GateServer.RegisterLoginToken",
				&gate.LoginNotify{
					Account:    account,
					LoginToken: loginToken,
					NumberID:   numberID,
				},
				&reply); err != nil {
				//beego.Error("[Login.getGate.call] error %s, %v", err.Error(), ob)
				reason := fmt.Sprintf("[Login.getGate.call] error %s, %v", err.Error(), ob)
				models.CleanGate(ob, reason)
				//CtrlErrorReturn(c, "[Login.getGate RPC call]", err, ob)
				continue
			}
			conn.Close()
			err = nil

			gateInfo = ob
			break

		}
		team, cfgteam := GetServerTeam(uid.Bytes(), gid, sid)
		enc_ip := secure.Encode64ForNet([]byte(gateInfo.GameIPAddrPort))
		enc_token := secure.Encode64ForNet([]byte(loginToken))
		logiclog.LogLoginTeam(uid.String(), team, "", cfgteam, gid, sid)
		c.JSON(200, map[string]interface{}{
			"ip":         enc_ip,
			"result":     "ok",
			"logintoken": enc_token, //Auth token 返回给客户端
			"Team":       team,
		})
	}
}

// DebugKick是调试用API，协助客户端测试踢人操作
func (lc *LoginController) DebugKick() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch config.Cfg.Runmode {
		case "dev", "test":
			loginToken := lc.GetString(c, "lt")
			Reason := lc.GetString(c, "reason")
			var AfterDur, NoLogin int
			var err error
			AfterDur, err = lc.GetInt(c, "after")
			if err != nil {
				AfterDur = 5
			}
			NoLogin, err = lc.GetInt(c, "nologin")
			if err != nil {
				NoLogin = 5
			}

			accountId := lc.GetString(c, "account") //0:0:0001
			account, err := db.ParseAccount(accountId)

			if err != nil {
				c.String(500, "account Err")
				return
			}
			_, _,
				OldRpc, ok := models.GetLastLoginStatus(account.GameId, account.UserId)
			if ok {
				var reply int
				if Conn, err := jsonrpc.Dial("tcp", OldRpc); err == nil {
					err = Conn.Call("GateServer.KickItOffline",
						&gate.KickOfflineParam{
							LoginToken:      loginToken,
							AccountID:       accountId,
							Reason:          Reason,
							AfterDuration:   AfterDur,
							NoLoginDuration: NoLogin,
						},
						&reply)
					Conn.Close()
					result := map[string]interface{}{
						"rpc": true,
					}
					if err != nil {
						result["err"] = err.Error()
					}
					c.JSON(200, result)
					logs.Debug("DebugKick try to kick login status!", loginToken, accountId)
				}
			} else {
				logs.Debug("DebugKick did not find rpc login status!", loginToken, accountId)
			}

		default:
		}
		c.String(404, "Runmode Err")
	}
}
