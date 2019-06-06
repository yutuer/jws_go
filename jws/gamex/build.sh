#!/bin/bash

export VERSION=vcs.taiyouxi.net/planx/version.Version
export BUILDCOUNTER=vcs.taiyouxi.net/planx/version.BuildCounter
export BUILDTIME=vcs.taiyouxi.net/planx/version.BuildTime
export GITHASH=vcs.taiyouxi.net/planx/version.GitHash
export counter=0
go build  -ldflags "-X ${VERSION}=1.0 -X ${GITHASH}=`git rev-parse HEAD` -X ${BUILDTIME}=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X ${BUILDCOUNTER}=${counter}"

