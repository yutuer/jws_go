package allinone

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/gin-gonic/gin"

	csCfg "vcs.taiyouxi.net/jws/crossservice/config"
	"vcs.taiyouxi.net/jws/crossservice/util/discover/exclusion"
	"vcs.taiyouxi.net/jws/crossservice/util/http_util"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

var handle *exclusion.Handle

type groupStatus struct {
	myOccupy    map[int]bool
	moduleGroup map[string][]int
}

var curGroupStatus groupStatus

func init() {
	curGroupStatus.myOccupy = make(map[int]bool)
	curGroupStatus.moduleGroup = make(map[string][]int)
}

func occupyGroup() []uint32 {
	handle = exclusion.NewHandle("crossservice/group")

	interval := time.Second * 10
	handle.SetTTL(interval * 10)

	go func() {
		ticker := time.NewTicker(interval)
		for nil != handle {
			handle.Publish()
			<-ticker.C
		}
	}()

	list := getGroupIDList()
	exclusionNum := csCfg.Cfg.ExclusionNum
	if 0 == exclusionNum {
		exclusionNum = 1
	}
	num := int(math.Ceil(float64(len(list)) / float64(exclusionNum)))

	occList := []uint32{}
	for _, id := range list {
		if true == handle.AddNode(fmt.Sprintf("%d/%d", csCfg.Cfg.Gid, id), csCfg.GetIndex()) {
			occList = append(occList, uint32(id))
			curGroupStatus.myOccupy[id] = true
			if len(occList) >= num {
				break
			}
		}
	}

	http_util.GET("/ShowOccupyGroup", ginShowOccupyGroup)
	http_util.GET("/ShowGamedataGroup", ginShowGamedataGroup)
	return occList
}

func releaseGroup() {
	h := handle
	handle = nil
	h.UnPublish()
}

func getGroupIDList() []int {
	arrRange := getShardIDRange()
	list := make([]int, 0)
	for _, pair := range arrRange {
		WBGroupID := gamedata.GetWBGroupIDRange(pair[0], pair[1])
		list = append(list, WBGroupID...)
		curGroupStatus.moduleGroup["WBGroupID"] = append(curGroupStatus.moduleGroup["WBGroupID"], WBGroupID...)

		TBGroupID := gamedata.GetTBGroupIDRange(pair[0], pair[1])
		list = append(list, TBGroupID...)
		curGroupStatus.moduleGroup["TBGroupID"] = append(curGroupStatus.moduleGroup["TBGroupID"], WBGroupID...)
	}

	curGroupStatus.moduleGroup["WBGroupID"] = unique(curGroupStatus.moduleGroup["WBGroupID"])
	curGroupStatus.moduleGroup["TBGroupID"] = unique(curGroupStatus.moduleGroup["TBGroupID"])
	return unique(list)
}

func getShardIDRange() [][]uint32 {
	out := [][]uint32{}
	for _, l := range csCfg.Cfg.ShardRange {
		if 2 != len(l) {
			continue
		}
		out = append(out, l)
	}
	return out
}

func unique(list []int) []int {
	if nil == list {
		return []int{}
	}
	out := []int{}
	sort.Ints(list)

	last := -1
	for _, i := range list {
		if last == i {
			continue
		}
		last = i
		out = append(out, i)
	}
	return out
}

func ginShowOccupyGroup(c *gin.Context) {
	out := "Occupy Group\n"
	col := 5
	index := 0
	for id := range curGroupStatus.myOccupy {
		index++
		out += fmt.Sprintf("%10d", id)
		if 0 == index%col {
			out += "\n"
		}
	}
	out += "\n"

	c.String(200, out)
}

func ginShowGamedataGroup(c *gin.Context) {
	out := "Gamedata Group\n"
	col := 5
	for key, list := range curGroupStatus.moduleGroup {
		out += key + "\n"
		for index, id := range list {
			out += fmt.Sprintf("%10d", id)
			if 0 == (index+1)%col {
				out += "\n"
			}
		}
		out += "\n"
	}

	c.String(200, out)
}
