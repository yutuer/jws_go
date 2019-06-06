package base

import "vcs.taiyouxi.net/jws/gamex/models/helper"

type ActCommand struct {
	ErrCode int

	ParamInts        []int64
	ParamStrs        []string
	ParamItemC       map[string]uint32
	ParamAccountInfo *helper.AccountSimpleInfo

	NeedPushs [GuildActCount]bool
}

func (c *ActCommand) SetNeedSync(t int) {
	if t <= 0 || t >= GuildActCount {
		return
	} else {
		c.NeedPushs[t] = true
	}
}

func (c *ActCommand) SetNeedAll() {
	for i := 0; i < len(c.NeedPushs); i++ {
		c.NeedPushs[i] = true
	}
}

func NewActCmd(str string, p1, p2 int64, acInfo *helper.AccountSimpleInfo) ActCommand {
	return ActCommand{
		ParamInts:        []int64{p1, p2},
		ParamStrs:        []string{str},
		ParamAccountInfo: acInfo,
	}
}

func NewActCmdInt(p1, p2 int64, acInfo *helper.AccountSimpleInfo) ActCommand {
	return ActCommand{
		ParamInts:        []int64{p1, p2},
		ParamAccountInfo: acInfo,
	}
}

func NewActCmdAcInfo(acInfo *helper.AccountSimpleInfo) ActCommand {
	return ActCommand{
		ParamAccountInfo: acInfo,
	}
}

func NewActCmdRes() ActCommand {
	return ActCommand{}
}

func ReturnActCmdError(errorCode int) *ActCommand {
	return &ActCommand{
		ErrCode:   errorCode,
		ParamInts: []int64{0, 0, 0, 0, 0},
		ParamStrs: []string{"", "", "", ""},
	}
}
