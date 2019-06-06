package version

import "fmt"

/*
export VERSION=vcs.taiyouxi.net/platform/planx/version.Version
export BUILDCOUNTER=vcs.taiyouxi.net/platform/planx/version.BuildCounter
export BUILDTIME=vcs.taiyouxi.net/platform/planx/version.BuildTime
export GITHASH=vcs.taiyouxi.net/platform/planx/version.GitHash
export counter=0
go build  -ldflags "-X ${VERSION} 1.5 -X ${GITHASH} `git rev-parse HEAD` -X ${BUILDTIME} `date -u '+%Y-%m-%d_%I:%M:%S%p'` -X ${BUILDCOUNTER} ${counter}"
*/
var (
	Version      = "0.0.1"
	BuildCounter = "0"
	BuildTime    = "2015-08-01UTC"
	GitHash      = "None"
)

func GetVersion() string {
	return fmt.Sprintf("%s(%s) %s %s", Version, BuildCounter, GitHash, BuildTime)
}

func VerC() string {
	return fmt.Sprintf("%s(%s)", Version, BuildCounter)
}
