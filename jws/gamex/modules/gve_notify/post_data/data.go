package post_data

import "vcs.taiyouxi.net/jws/gamex/models/helper"

type StartGVEPostResData struct {
	Data      helper.Avatar2ClientByJson
	Reward    []string
	Count     []uint32
	IsUseHc   bool
	IsDouble  bool
	RobotData []*helper.Avatar2ClientByJson
}
