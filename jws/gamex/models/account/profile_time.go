package account

import (
	"time"
	"vcs.taiyouxi.net/platform/planx/servers/game"
)

func (p *Profile) GetRegTimeUnix() int64 {
	return game.GetUnixTimeAfterRegister(p.CreateTime, p.GetProfileNowTime())
}

func (p *Profile) GetRegDayCountUnix() int64 {
	return game.GetDayAfterRegister(p.CreateTime, p.GetProfileNowTime())
}

func (p *Profile) GetProfileNowTime() int64 {
	//	return p.GetRegTimeUnix() + p.CreateTime
	return time.Now().Unix() + p.DebugAbsoluteTime
}
