package sys_roll_notice

import (
	"vcs.taiyouxi.net/platform/x/gm_tools/common/plan_job"
)

var (
	jobManager plan_job.PlanJobManager
)

func Start() error {
	jobManager.Name = "sysRollNotice"
	return jobManager.Start()
}

func Stop() {
	jobManager.Stop()
}
