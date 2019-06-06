package teamboss

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"

	"github.com/stretchr/testify/assert"
)

// GetRoom的简单功能测试
func TestRoomInfo_GetRooms(t *testing.T) {
	ri := new(RoomInfo)
	room1 := new(Room)
	room2 := new(Room)
	ri.Rooms = map[uint32][]*Room{111: {room1, room2}, 222: {}}

	// 有
	r := ri.GetRooms(111)
	assert.Equal(t, []*Room{room1, room2}, r)

	// 空
	r2 := ri.GetRooms(222)
	assert.Empty(t, r2)

	// 无
	r3 := ri.GetRooms(333)
	assert.Empty(t, r3)
}

func TestRoomInfo_CreateRoom(t *testing.T) {
	ri := new(RoomInfo)
	ri.Rooms = make(map[uint32][]*Room)

	// 功能
	ri.CreateRoom(level)

	assert.Equal(t, 1, len(ri.Rooms))
	assert.Equal(t, 1, len(ri.Rooms[level]))

	// 并行建立房间
	wg.Add(20)

	go func() {
		for lvl := 90; lvl < 110; lvl++ {
			lvl := lvl
			go func() {
				defer wg.Done()
				ri.CreateRoom(uint32(lvl))
			}()
		}
	}()

	wg.Add(50)

	go func() {
		for i := 0; i < 50; i++ {
			go func() {
				defer wg.Done()
				ri.CreateRoom(100)
			}()
		}
	}()

	wg.Wait()

	assert.Equal(t, 20, len(ri.Rooms))
	assert.Equal(t, 52, len(ri.Rooms[100]))
}

// GetRoom的并行测试
func TestRoomInfo_GetRooms2(t *testing.T) {
	ri := new(RoomInfo)
	ri.Rooms = make(map[uint32][]*Room)

	ri.CreateRoom(level)

	// 并行读写
	wg.Add(50)

	go func() {
		for i := 0; i < 50; i++ {
			go func() {
				defer wg.Done()
				ri.CreateRoom(level)
			}()
		}
	}()

	go func() {
		for j := 0; j < 100; j++ {
			ri.CreateRoom(uint32(96 + j%7))
		}
	}()

	go func() {
		for k := 0; k < 10; k++ {
			rooms := ri.GetRooms(100)
			if len(rooms) < 1 || len(rooms) > 50 {
				t.Error(fmt.Sprintf("GetRooms(exist) error, actual amount %v", len(rooms)))
			}
			erooms := ri.GetRooms(10)
			if len(erooms) != 0 {
				t.Error(fmt.Sprintf("GetRooms(empty) error, actual amount %v", len(erooms)))
			}
		}
	}()

	wg.Wait()

	rooms := ri.GetRooms(level)
	assert.True(t, len(rooms) >= 50)
}

func TestRoomInfo_GetRoom(t *testing.T) {
	ri := new(RoomInfo)
	ri.Rooms = make(map[uint32][]*Room)

	// 有
	r := ri.CreateRoom(100)
	r2 := ri.GetRoom(r.ID)
	assert.Equal(t, r, r2)

	// 无
	r3 := ri.GetRoom("233_233_0")
	assert.Nil(t, r3)

	// 非法参数
	r6 := ri.GetRoom("Which is not exist")
	assert.Nil(t, r6)

	// 异步
	var wg sync.WaitGroup
	wg.Add(1)
	roomID := r.ID

	go func() {
		defer wg.Done()
		ri.DelRoom(roomID)
	}()

	r4 := ri.GetRoom(roomID)
	assert.Equal(t, r, r4)

	wg.Wait()

	r5 := ri.GetRoom(roomID)
	assert.Nil(t, r5)
}

func TestRoomInfo_DelRoom(t *testing.T) {
	ri := new(RoomInfo)
	ri.Rooms = make(map[uint32][]*Room)
	roomIDmap := make(map[string]bool)

	// 生成100个房间
	for i := 0; i < 100; i++ {
		room := ri.CreateRoom(uint32(95 + i%10))
		roomIDmap[room.ID] = true
	}

	// 删除不存在的ID
	ri.DelRoom("233_233_23333")

	// 删除非法的ID
	ri.DelRoom("NotExistID")

	// 删除正常ID
	roomID := ri.Rooms[100][0].ID
	assert.NotNil(t, ri.GetRoom(roomID))
	ri.DelRoom(roomID)
	assert.Nil(t, ri.GetRoom(roomID))

	// 并行
	wg.Add(1001) // 500 * 2 建立加删除， 1 删除之前的100个

	roomIDchan := make(chan string)

	// 创建房间
	go func() {
		for i := 0; i < 500; i++ {
			i := i
			go func() {
				defer wg.Done()
				roomIDchan <- ri.CreateRoom(uint32(95 + i%10)).ID
			}()
		}
	}()

	// 删除上一步创建的房间
	go func() {
		for i := 0; i < 500; i++ {
			go func() {
				defer wg.Done()
				ri.DelRoom(<-roomIDchan)
			}()
		}
	}()

	// 删除已有的100个
	go func() {
		for i := 0; i < 100; i++ {
			for roomID := range roomIDmap {
				ri.DelRoom(roomID)
			}
		}

		defer wg.Done()
	}()

	defer close(roomIDchan)

	wg.Wait()

	// 期望：所有房间都删掉了
	for level, rooms := range ri.Rooms {
		for index, room := range rooms {
			if room != nil {
				t.Errorf("Room del failed; level:%v, index:%v", level, index)
			}
		}
	}
}

// GenRoomInfo没有办法mock time.Now()来测试
// 只能退求其次，用另一个用例验证一周七天都可以获得TeamTypeID
// 并验证所有TeamTypeID都有正确的配置
func TestRoom_GenRoomInfo(t *testing.T) {
	teamTypeMap := make(map[uint32]bool, 0)
	typeMap := gamedata.GetTBossHeroTypeMap()

	// 遍历所有的TeamTypeID
	for _, val := range typeMap {
		table := val.GetType_Table()
		for _, typ := range table {
			typeId := typ.GetTeamTypeID()
			if typeId != 0 {
				teamTypeMap[typeId] = true
			}
		}
	}

	for teamTypeID := range teamTypeMap {
		room := &Room{}
		room.TeamTypID = teamTypeID
		room.BossID = gamedata.GetTBBossID(room.TeamTypID)
		room.SceneID = gamedata.GetTBSceneID(room.TeamTypID)

		assert.NotEqual(t, 0, room.TeamTypID)
		assert.NotEqual(t, 0, room.BossID)
		assert.NotEqual(t, 0, room.SceneID)
	}
}

// 下面这些都是TestDebugGetTBTeamTypeID的Setup
type TBossTypeForRand ProtobufGen.TBOSSHEROTYPE

func (tt TBossTypeForRand) GetWeight(index int) int {
	return int(tt.Type_Table[index].GetChooseChance())
}

func (tt TBossTypeForRand) Len() int {
	return len(tt.Type_Table)
}

// 这是改写了gamedata.GetTBTeamTypeID()
func debugGetTBTeamTypeID(date uint32) uint32 {
	//date := uint32(time.Now().Weekday())
	gdTBossHeroTypeMap := gamedata.GetTBossHeroTypeMap()
	for day, item := range gdTBossHeroTypeMap {
		if date == uint32(time.Sunday) && day == 7 {
			randIndex := util.RandomItem(TBossTypeForRand(*item))
			return item.GetType_Table()[randIndex].GetTeamTypeID()
		} else if date == day {
			randIndex := util.RandomItem(TBossTypeForRand(*item))
			return item.GetType_Table()[randIndex].GetTeamTypeID()
		}
	}
	return 0
}

// 验证一周7天都能正确获得RoomID
func TestDebugGetTBTeamTypeID(t *testing.T) {
	now := time.Now()

	for days := 0; days < 31; days++ {
		testTime := now.Add(time.Duration(24*days) * time.Hour)
		day := testTime.Weekday()
		teamTypeID := debugGetTBTeamTypeID(uint32(day))

		assert.NotEqual(t, 0, teamTypeID)
	}
}

func TestRoom_RemovePlayer(t *testing.T) {
	r := new(Room)
	p1 := new(Player)
	p1.SimpleInfo.AcID = "ACID_Player1"
	p1.SimpleInfo.Status = 1 // 是否点准备
	p2 := new(Player)
	p2.SimpleInfo.AcID = "ACID_Player2"

	r.AddPlayer(p1)
	r.AddPlayer(p2)

	// 各种设置
	r.LeadAcID = "ACID_Player1"      // 队长
	r.AdvanceCostID = "ACID_Player1" // 谁付钱买箱子
	r.BoxStatus = 1                  // 箱子状态
	r.RoomState = 1                  // 仅访问
	r.PositionAcID[0] = "ACID_Player1"
	r.PositionAcID[1] = "ACID_Player2"

	// 做个备份
	r0 := new(Room)
	*r0 = *r

	// Remove不存在的
	r.RemovePlayer("NoOne")
	assert.Equal(t, 2, r.PlayerCount())

	// Remove队长和付费的那个
	r.RemovePlayer("ACID_Player1")
	assert.Equal(t, 1, r.PlayerCount())
	assert.Equal(t, "ACID_Player2", r.LeadAcID)
	assert.Equal(t, 0, r.BoxStatus)
	assert.Equal(t, 1, r.RoomState)
	assert.Equal(t, "", r.PositionAcID[0])

	// 恢复回来，Remove另一只
	*r = *r0
	r.Players = []*Player{p1, p2}
	p1.SimpleInfo.Status = 1

	r.RemovePlayer("ACID_Player2")
	assert.Equal(t, 1, r.PlayerCount())
	assert.Equal(t, "ACID_Player1", r.LeadAcID)
	assert.Equal(t, 1, r.BoxStatus)
	assert.Equal(t, 1, r.RoomState)
	assert.Equal(t, "", r.PositionAcID[1])

	// 再Remove剩下的一只
	r.RemovePlayer("ACID_Player1")
	assert.Equal(t, 0, r.PlayerCount())
}

func TestRoom_GetExtraPlayer(t *testing.T) {
	r := new(Room)

	sid1 := uint(testGroupId1)
	sid2 := uint(testGroupId2)

	p1 := new(Player)
	p1.SimpleInfo.AcID = "ACID_Player1"
	p1.SimpleInfo.Sid = sid1
	p2 := new(Player)
	p2.SimpleInfo.AcID = "ACID_Player2"
	p2.SimpleInfo.Sid = sid2

	r.AddPlayer(p1)
	r.AddPlayer(p2)

	// 另一个人
	players := r.GetExtraPlayer("ACID_Player2")

	assert.Equal(t, 1, len(players))
	assert.Equal(t, "ACID_Player1", players[sid1][0])

	// 两个人
	players = r.GetExtraPlayer("")

	assert.Equal(t, 2, len(players))
	assert.Equal(t, "ACID_Player1", players[sid1][0])
	assert.Equal(t, "ACID_Player2", players[sid2][0])

	// 同服两人
	p2.SimpleInfo.Sid = sid1
	players = r.GetExtraPlayer("")

	assert.Equal(t, 1, len(players))
	assert.Equal(t, 2, len(players[sid1]))
}

// BenchmarkRoomInfo_CreateRoom 建房间的性能测试
func BenchmarkRoomInfo_CreateRoom(b *testing.B) {
	ri := new(RoomInfo)
	ri.Rooms = make(map[uint32][]*Room)

	for i := 0; i < b.N; i++ {
		ri.CreateRoom(uint32(96 + i%8))
		if i%4096 == 0 {
			ri.Rooms = make(map[uint32][]*Room)
		}
	}
}

// BenchmarkRoomInfo_CreateRoom 删房间的性能测试
func BenchmarkRoomInfo_DelRoom(b *testing.B) {
	ri := new(RoomInfo)
	ri.Rooms = make(map[uint32][]*Room)
	ri2 := new(RoomInfo)
	ri2.Rooms = make(map[uint32][]*Room)

	roomIDs := make([]string, 10000)

	for i := 0; i < 10000; i++ {
		room := ri.CreateRoom(uint32(96 + i%8))
		roomIDs[i] = room.ID
	}

	for i := 0; i < b.N; i++ {
		idx := i % 8192

		if idx == 0 {
			for k, v := range ri.Rooms {
				ri2.Rooms[k] = v
			}
		}

		ri2.DelRoom(roomIDs[idx])
	}
}