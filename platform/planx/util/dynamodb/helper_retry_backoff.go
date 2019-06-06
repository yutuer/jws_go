package dynamodb

import (
	"github.com/cenk/backoff"
	"vcs.taiyouxi.net/platform/planx/util/awshelper"
)

func NewExponentialBackOffSleepFirst() *backoff.ExponentialBackOff {
	return awshelper.NewExponentialBackOffSleepFirst()
}
