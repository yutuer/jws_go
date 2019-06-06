package clienttag

import (
	"fmt"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	Tag_GuildIn = iota
	Tag_DailyNotice
	Tag_ShowFashion
	Tag_Count
)

type ClientTag struct {
	TagState []int `json:"tag"`
}

func (tag *ClientTag) SetTag(index, val int) error {
	if index >= len(tag.TagState) {
		logs.Error("ClientTag SetTag index > len(tag)")
		return fmt.Errorf("ClientTag SetTag index > len(tag)")
	}
	tag.TagState[index] = val
	return nil
}

func (tag *ClientTag) Init() {
	if tag.TagState == nil {
		tag.TagState = make([]int, Tag_Count)
	}
	if len(tag.TagState) < Tag_Count {
		for i := 0; i < Tag_Count-len(tag.TagState); i++ {
			tag.TagState = append(tag.TagState, 0)
		}
	}
}
