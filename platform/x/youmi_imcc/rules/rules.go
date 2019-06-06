package rules

import "vcs.taiyouxi.net/platform/x/youmi_imcc/base"

/*
	0 represent pass, -1 represent infinite max
*/

func init() {
	sensitiveWordInitial()
	similarityInitial()
}

func Reg() {
	base.AddMsgHandler(checkSensitiveWord)
	base.AddTimerMsgHandler(checkSimilarity)
	//base.AddTimerMsgHandler(queryInterval)
}
