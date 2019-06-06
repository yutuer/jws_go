package chatserver

import (
	"sync/atomic"
	"time"

	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/chat/chatserver/proto/gen"
)

func Rec_RegReq(player *Player, data *gen.RegReq) {
	roleInfo := data.RoleInfo(nil)
	account, err := db.ParseAccount(string(roleInfo.AccountId()))
	if err != nil {
		logs.Error("player reg account format err %v", string(roleInfo.AccountId()))
		player.session.Send(Gen_ErrMsg("Err:AccountErr"))
		player.session.Close()
		return
	}

	logs.Debug("rec RegReq %s %s", roleInfo.Name(), roleInfo.AccountId())
	// 认证
	if string(data.ClientAuth()) != "robot" || CommonCfg.RunMode == "prod" {
		curAuth, lastAuth := GetAuthSource(string(roleInfo.AccountId()))
		if string(data.ClientAuth()) != curAuth && string(data.ClientAuth()) != lastAuth {
			logs.Warn("player Reg Auth err, name: %s, clientAuth %s curAuth %s lastAuth %s %s",
				roleInfo.Name(), data.ClientAuth(), curAuth, lastAuth, roleInfo.AccountId())
			player.session.Send(Gen_ErrMsg("Err:AuthFail"))
			go func() {
				time.Sleep(time.Second)
				player.session.Close()
			}()
			return
		}
	}

	if player.room == nil {
		if atomic.CompareAndSwapInt32(&player.isReg, 0, 1) { // 设置reg状态
			player.cityName = string(data.Town())
			player.Name = string(roleInfo.Name())
			player.roleInfo = roleInfo
			player.setPos(roleInfo.RolePos(nil))
			logs.Debug("Player reg %s %s", roleInfo.Name(), roleInfo.AccountId())
			enterRoomByCity(account.ServerString(), string(data.Town()), player.Name, player)
			//			logs.Debug("reg name:%s room:%s pos: %v %v %v", player.Name, player.room.Name, player.getPos().Pos(0),
			//				player.getPos().Pos(1), player.getPos().Pos(2))
			if player.room == nil {
				logs.Error("Rec_RegReq player enter room failed, %v %v %v",
					data.Town(), roleInfo.Name(), roleInfo.AccountId())
				player.session.Close()
				return
			}
			player.room.broadcast2OtherMsg(player, Gen_RoleEnterNotify(player.roleInfo, player.getPos())) // 发给其他人本人的信息
			// 返回房间已有角色的信息, 防止取到信息过程中房间中某玩家已经离开，所以加锁
			player.room.mutex.RLock()
			roles := make([]*gen.RoleInfo, 0, len(player.room.sid2Players))
			rolesPos := make([]*gen.RolePos, 0, len(player.room.sid2Players))
			for _, v := range player.room.sid2Players {
				if v.Name == player.Name {
					continue
				}

				roles = append(roles, v.roleInfo)
				rolesPos = append(rolesPos, v.getPos())
				logs.Debug("send to name:%s mem:%s", player.Name, v.Name)
			}
			player.session.Send(GenRegResp(player.room.Name, roles, rolesPos)) // 发给本人已在房间其他人的信息
			player.room.mutex.RUnlock()
		}
	} else {
		logs.Error("player %s reg but already has room", player.Name)
		player.session.Close()
	}
}

func Rec_PosNotify(player *Player, data *gen.PosNotify) {
	if player.room != nil {
		logs.Debug("rec PosNotify %v", data.Name())
		player.setPos(data.Pos(nil))
		player.room.broadcast2OtherMsg(player, Gen_PosNotify(data))
	}
}

func Rec_UnRegNotify(player *Player, accountId, name string) {
	if player.room != nil {
		logs.Debug("UnRegNotify name:%v room:%s %s %s",
			name, player.room.Name, accountId, name)
		player.room.broadcast2OtherMsg(player, Gen_UnRegNotify(accountId, name))
		player.room.OnPlayerUnReg(player)
		player.session.Close()
	}
}
