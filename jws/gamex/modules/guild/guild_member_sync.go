package guild

import (
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	. "vcs.taiyouxi.net/jws/gamex/modules/guild/info"
)

func (r *GuildModule) AddMemSyncReceiver(guildUUID string, mr helper.IGuildMemSyncReceiver) {
	r.guildCommandExecAsyn(guildCommand{
		Type: Command_AddMemSyncReceiver,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		memSyncReceiver: mr,
	})
}

func (r *GuildModule) DelMemSyncReceiver(guildUUID string, id int) GuildRet {
	res := r.guildCommandExec(guildCommand{
		Type: Command_DelMemSyncReceiver,
		BaseInfo: GuildSimpleInfo{
			GuildUUID: guildUUID,
		},
		memSyncReceiverID: id,
	})
	return res.ret
}
