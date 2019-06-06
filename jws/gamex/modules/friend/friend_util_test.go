package friend

import (
	"fmt"
	"strconv"
	"testing"
)

func TestRandomFriendList(t *testing.T) {
	friendList := make([]FriendSimpleInfo, 50)
	for i := 0; i < 50; i++ {
		if i%2 == 0 {
			friendList[i] = FriendSimpleInfo{AcID: strconv.Itoa(i), LastActTime: 0}
		} else {
			friendList[i] = FriendSimpleInfo{AcID: strconv.Itoa(i), LastActTime: 1}
		}
	}
	for i := 0; i < 1000; i++ {
		list := SelectFriendInfo(friendList)
		if len(list) < 10 {
			fmt.Println(len(list), list)
		}
	}
}
