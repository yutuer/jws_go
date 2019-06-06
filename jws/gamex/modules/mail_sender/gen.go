package mail_sender

import (
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

var (
	mInstance map[uint]*MailSenderModule
)

func init() {
	mInstance = make(map[uint]*MailSenderModule, 6)
	modules.RegModule(modules.Module_MailSender, newMailSenderModule)
}

func GetModule(shard uint) *MailSenderModule {
	return mInstance[game.Cfg.GetShardIdByMerge(shard)]
}

func newMailSenderModule(sid uint) modules.ServerModule {
	m := genMailSenderModule(sid)
	mInstance[sid] = m
	return m
}
