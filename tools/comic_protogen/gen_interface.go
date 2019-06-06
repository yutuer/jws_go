package main

import "vcs.taiyouxi.net/tools/comic_protogen/util"

type GenInterface interface {
	GenCode(infos []*util.ProtoInfo)
}
