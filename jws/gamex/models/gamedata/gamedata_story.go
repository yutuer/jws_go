package gamedata

import (
	"github.com/golang/protobuf/proto"
	"strconv"
	"strings"
	ProtobufGen "vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	gdStoryDetailed   []*ProtobufGen.STORY
	gdStorysBySection [][]int
)

func loadStoryDetailed(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.STORY_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdStoryDetailed = make(
		[]*ProtobufGen.STORY, len(data)+1, len(data)+1)

	for _, c := range data {
		id := int(c.GetQuestID())
		if id < 0 || id >= len(gdStoryDetailed) {
			logs.Error("gdStoryDetailed id Err By %d", id)
		}
		gdStoryDetailed[id] = c
	}
	//logs.Trace("gdStoryDetailed %v", gdStoryDetailed)
}

func loadStorySection(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.SECTION_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdStorysBySection = make(
		[][]int, len(data)+1, len(data)+1)

	for _, c := range data {
		id := int(c.GetSectionID())
		if id < 0 || id >= len(gdStorysBySection) {
			logs.Error("loadStorySection id Err By %d", id)
		}

		storys := strings.Split(c.GetConten(), ",")
		story_ints := make([]int, 0, len(storys))
		for _, story_s := range storys {
			story_int, err := strconv.Atoi(story_s)
			if err != nil {
				logs.Error("loadStorySection story_int conv By Err %d %s", id, err.Error())
			} else {
				story_ints = append(story_ints, story_int)
			}
		}

		gdStorysBySection[id] = story_ints
	}
	//logs.Trace("gdStorysBySection %v", gdStorysBySection)
}

func GetStorys() []*ProtobufGen.STORY {
	return gdStoryDetailed[:]
}

func GetStorysBySection(section_id int) []int {
	if section_id < 0 || section_id >= len(gdStorysBySection) {
		return nil
	} else {
		return gdStorysBySection[section_id][:]
	}
}

//
//
//
//
//
//
