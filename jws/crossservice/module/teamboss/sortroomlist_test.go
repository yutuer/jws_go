package teamboss

import (
	"fmt"
	"testing"

	"vcs.taiyouxi.net/jws/helper"

	"github.com/stretchr/testify/assert"
)

func TestSortRoomList(t *testing.T) {
	// 目前排序规则：人数不满且公开 -> 不满且私密 -> 满
	infos := make([]helper.RetRoomList, 0)

	// 遍历等级、人数、房间状态
	for level := 90; level < 111; level += 10 {
		for playerCount := 1; playerCount < 3; playerCount++ {
			for roomStatus := 0; roomStatus < 3; roomStatus++ {
				room := new(helper.RetRoomList)
				room.PlayerCount = playerCount
				room.RoomStatus = roomStatus
				room.Level = uint32(level)

				infos = append(infos, *room)
			}
		}
	}

	before := "Before(Player Num//Status//Level): "
	for _, info := range infos {
		before += fmt.Sprintf("[%v %v %v] ", info.PlayerCount, info.RoomStatus, info.Level)
	}

	sortedList := sortRoomList(infos)

	after := "After(Player Num//Status//Level): "
	for _, info := range sortedList {
		// 滤掉已经开战的房间
		if info.RoomStatus == 2 {
			continue
		}
		after += fmt.Sprintf("[%v %v %v] ", info.PlayerCount, info.RoomStatus, info.Level)
	}

	// 显示结果
	t.Logf(before)
	t.Logf(after)

	// 前三应该是单人公开房间
	for i := 0; i < 3; i++ {
		assert.Equal(t, 1, sortedList[i].PlayerCount)
		assert.Equal(t, 0, sortedList[i].RoomStatus)
	}

	// 前四到前六是单人私密房间
	for i := 3; i < 6; i++ {
		assert.Equal(t, 1, sortedList[i].PlayerCount)
		assert.Equal(t, 1, sortedList[i].RoomStatus)
	}

	// 后面就无所谓了
}
