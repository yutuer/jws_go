package worship

import (
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

func (r *module) processCmd(c *cmd) *cmd {
	switch c.Type {
	case cmdTypeGetWorshipData:
		return r.getWorshipData()
	case cmdTypeWorship:
		return r.worship(c)
	case cmdTypeRefershTop:
		return r.refershTop(c)
	case cmdTypeGetWorship:
		return r.getWorship(c)
	}

	return nil
}

func (r *module) GetWorshipData(ctx context.Context) []worshipAccData {
	res := r.sendCmd(ctx, &cmd{
		Type: cmdTypeGetWorshipData,
	})

	if res == nil {
		return nil
	}

	return res.Top[:]
}

func (r *module) getWorshipData() *cmd {
	return &cmd{
		Top: r.top.copyTop(),
	}
}

func (r *module) Worship(
	ctx context.Context,
	accountID string) {
	r.sendCmd(ctx, &cmd{
		Type:    cmdTypeWorship,
		Account: accountID,
	})
}

func (r *module) worship(c *cmd) *cmd {
	r.top.worship(c.Account)
	return nil
}

func (r *module) GetWorship(
	ctx context.Context,
	accountID string) int {

	res := r.sendCmd(ctx, &cmd{
		Type:    cmdTypeWorship,
		Account: accountID,
	})

	if res == nil {
		return 0
	} else {
		return res.Type
	}
}

func (r *module) getWorship(c *cmd) *cmd {
	res := new(cmd)
	res.Type = r.top.getWorship(c.Account)
	return res
}

func (r *module) RefershTop(
	ctx context.Context,
	top [WorshipAccountCount]helper.AccountSimpleInfo) {
	r.sendCmd(ctx, &cmd{
		Type: cmdTypeRefershTop,
	})
	return
}

func (r *module) refershTop(c *cmd) *cmd {
	r.top.clean()
	return nil
}
