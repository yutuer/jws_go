package game

import "vcs.taiyouxi.net/platform/planx/servers"

type PreparePlayer func(accountid, ip string) Player

type Server interface {
	Stop()
	Start(mux *servers.Mux)
}
