package logics

import "vcs.taiyouxi.net/platform/planx/servers"

func (p *Account) RegisterSystemRequestHandler(r *servers.Mux) {
	r.HandleFunc(servers.SYSTEM_TICK_30_CODE, p.SystemTick30)
}
func (p *Account) SystemTick30(r servers.Request) *servers.Response {
	return &servers.Response{
		Code:          servers.SYSTEM_TICK_30_CODE,
		RawBytes:      nil,
		ForceDBChange: true,
	}
}
