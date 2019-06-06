package friend

import (
	"encoding/json"
	"math/rand"
	"sort"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func getLvFromJson(v interface{}) (int, error) {
	jsonStr, err := redis.String(v, nil)
	if err != nil {
		return 0, err

	}

	lv := struct {
		Level int `json:"lv"`
	}{}
	err = json.Unmarshal([]byte(jsonStr), &lv)
	if err != nil {
		return 0, err
	}
	return lv.Level, nil
}

func getVIPLvFromJson(v interface{}) (int, error) {
	jsonStr, err := redis.String(v, nil)
	if err != nil {
		return 0, err
	}

	lv := struct {
		Level int `json:"v"`
	}{}
	err = json.Unmarshal([]byte(jsonStr), &lv)
	if err != nil {
		return 0, err
	}
	return lv.Level, nil
}

func getNameFromJson(v interface{}) (string, error) {
	jsonStr, err := redis.String(v, nil)
	if err != nil {
		return "", err
	}
	return jsonStr, nil
}

func getLogOutTimeFromJson(v interface{}) (int64, error) {
	jsonInt, err := redis.Int64(v, nil)
	if err != nil {
		return -1, err
	}
	return jsonInt, nil
}

func getAvatarFromJson(v interface{}) (int, error) {
	jsonInt, err := redis.Int(v, nil)
	if err != nil {
		return -1, err
	}
	return jsonInt, nil
}

func getGSFromJson(v interface{}) (int, error) {
	jsonStr, err := redis.String(v, nil)
	if err != nil {
		return 0, err
	}

	gsInfo := struct {
		GS int `json:"corp_gs"`
	}{}
	err = json.Unmarshal([]byte(jsonStr), &gsInfo)
	if err != nil {
		return 0, err
	}
	return gsInfo.GS, nil
}

type ti struct {
	Opt  string  `json:"opt"`
	Cost float64 `json:"costTime"`
}

func timeTrack(start time.Time, name string, sid uint) {
	elapsed := time.Since(start).Seconds()
	//ti := ti{
	//	Opt:  name,
	//	Cost: elapsed,
	//}
	//logiclog.Error("", 0, 0, "", name, ti, "", "[BI]")

	//metrics.SimpleSend(fmt.Sprintf(name + ".%d.%d.%s.%s", game.Cfg.Gid, sid, name, "time"),
	//	fmt.Sprintf("%d",elapsed))
	logs.Info(name+" cost: %f", elapsed)
}

func GenFriendInfo(players []string) []FriendSimpleInfo {
	//defer timeTrack(time.Now(), "GenFriendInfo_V1")
	retInfo := make([]FriendSimpleInfo, 0, len(players))
	index := 0
	for {
		conn := driver.GetDBConn()
		if conn.IsNil() {
			logs.Error("GetDBConn Err by nil error")
			return retInfo
		}

		cb := redis.NewCmdBuffer()
		nextIndex := index + countPerInit
		if nextIndex > len(players) {
			nextIndex = len(players)
		}
		for i := index; i < nextIndex; i++ {
			cb.Send("HMGET", "profile:"+players[i], "corp", "v", "name", "data", "logouttime", "curr_avatar")
		}
		reply := []interface{}{}
		r, err := redis.Values(conn.DoCmdBuffer(cb, false))
		if err != nil {
			logs.Error("Do Buffer err by %v for player count: %d", err, len(players))
			conn.Close()
			return retInfo
		}
		conn.Close()
		reply = r
		l := nextIndex - index
		if len(reply) != l {
			logs.Error("genFriendInfo err by illegal format, len(reply) != len(players)")
			return retInfo
		}
		for i := 0; i < l; i++ {
			values, err := redis.Values(reply[i], nil)
			if err != nil {
				logs.Debug("convert type err by %v for player: %s", err, players[i])
				continue
			}
			if len(values) < 6 {
				logs.Error("redis.Value err by illegal format(len < 2) for player: %s", players[i])
				continue
			}
			lv, err := getLvFromJson(values[0])
			if err != nil {
				logs.Warn("getLvInfoFromJson err by %v for player: %s", err, players[i])
				continue
			}
			vipLV, err := getVIPLvFromJson(values[1])
			if err != nil {
				logs.Warn("getVIPLvFromJson err by %v for player: %s", err, players[i])
				continue
			}
			name, err := getNameFromJson(values[2])
			if err != nil {
				logs.Warn("getNameFromJson err by %v for player: %s", err, players[i])
				continue
			}
			gs, err := getGSFromJson(values[3])
			if err != nil {
				logs.Warn("getGSFromJson err by %v for player: %s", err, players[i])
				continue
			}
			lastLogOutTime, err := getLogOutTimeFromJson(values[4])
			if err != nil {
				logs.Warn("getLogOutTimeFromJson err by %v for player: %s", err, players[i])
				continue
			}
			currAvatar, err := getAvatarFromJson(values[5])
			if err != nil {
				logs.Warn("getAvatarFromJson err by %v for player: %s", err, players[i])
				continue
			}
			retInfo = append(retInfo, FriendSimpleInfo{
				AcID:        players[index+i],
				Name:        name,
				GS:          gs,
				Lv:          lv,
				VIPLv:       vipLV,
				LastActTime: lastLogOutTime,
				Avatar:      currAvatar,
			})
		}
		index = nextIndex
		if index == len(players) {
			break
		}
	}

	return retInfo
}

/*
	随机选择
*/

func SelectFriendInfo(friends []FriendSimpleInfo) (ret [FriendCountPerRet]FriendSimpleInfo) {
	if len(friends) <= FriendCountPerRet {
		for i, f := range friends {
			ret[i] = f
		}
		return
	}
	weightArray := make([]int, len(friends))
	weightArray[0] = getWeight(friends[0])
	for i, friend := range friends {
		if i != 0 {
			weightArray[i] = weightArray[i-1] + getWeight(friend)
		}
	}
	randList := randomListNList(weightArray[len(friends)-1], FriendCountPerRet)
	j := 0
	for i := 0; i < len(friends); i++ {
		if j == FriendCountPerRet {
			break
		}
		if randList[j] <= weightArray[i] {
			ret[j] = friends[i]
			j++
		}
	}
	return
}

func getWeight(friend FriendSimpleInfo) int {
	if friend.LastActTime == 0 {
		return 7
	} else {
		return 3
	}
}

func randomListNList(n, m int) []int {
	randomList := make([]int, n)
	for i := 0; i < n; i++ {
		randomList[i] = i
	}
	for i := 0; i < m; i++ {
		rand := rand.Int31n(int32(n))
		randomList[i], randomList[rand] = randomList[rand], randomList[i]
	}
	sort.Ints(randomList[0:m])
	return randomList[0:m]
}
