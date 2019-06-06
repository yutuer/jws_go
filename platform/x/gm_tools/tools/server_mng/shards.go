package server_mng

import (
	"encoding/json"
	"strconv"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GameShard struct {
	ID     int    `json:"ID"`
	GID    int    `json:"GID"`
	Name   string `json:"name"`
	Dn     string `json:"dn"`
	Level  int    `json:"lv"`
	IsAuto int    `json:"auto"`
}

type ShardInRedis struct {
	ID  int `json:"ID"`
	GID int `json:"GID"`
}

type ClientShardInRedis struct {
	Name   string `json:"name"`
	Dn     string `json:"dn"`
	Level  int    `json:"lv"`
	IsAuto int    `json:"auto"`
}

type GameShards struct {
	GID    int
	Shards []GameShard
}

/*

   0:[{"name":"4e7a46ff8e315a59ef1ad1c3af5ab98c","dn":"dev-shard10","lv":1,"auto":1},
      {"name":"adf3546119027cd18ab136a2824334cf","dn":"dev-shard11","lv":0,"auto":0},
      {"name":"0197d3267aec6845c2ff2f1c5d377c8f","dn":"dev-shard12","lv":0,"auto":0},
      {"name":"6fcc13952f508a7c926c24f3fef5d32e","dn":"load-shard20","lv":0,"auto":0}]

*/
func NewGameShardsFromRedis(gid, data string) (*GameShards, error) {
	gameShards := GameShards{}

	gidN, err := strconv.Atoi(gid)
	if err != nil {
		logs.Error("NewGameShardsFromRedis Err by GID %s", gid)
		return nil, err
	}

	gameShards.GID = gidN
	err = json.Unmarshal([]byte(data), &gameShards.Shards)
	if err != nil {
		logs.Error("NewGameShardsFromRedis Err Unmarshal by %s GID %s %s",
			err.Error(), gid, data)
		return nil, err
	}

	logs.Trace("NewGameShardsFromRedis %v", gameShards)
	for i := 0; i < len(gameShards.Shards); i++ {
		gameShards.Shards[i].GID = gidN
	}

	return &gameShards, nil
}

func (g *GameShards) LoadShardInRedis(data map[string]ShardInRedis) {
	for i := 0; i < len(g.Shards); i++ {
		n := g.Shards[i].Name
		info, ok := data[n]
		if !ok {
			logs.Error("LoadShardInRedis Err No Name %s", n)
		} else {
			g.Shards[i].ID = info.ID
		}
	}
}

/*

map[
	a0c18ef1f5c378cad7a615c366d09e33:{"ID":15,"GID":1}
	1132cb1511809fe680e393925e736aa3:{"ID":101,"GID":101}
	666fb1aca8a9bd7e145195090367dd47:{"ID":17,"GID":1}]

*/
func NewShardsFromRedis(data map[string]string) (map[string]ShardInRedis, error) {
	res := make(map[string]ShardInRedis, 32)

	for name, d := range data {
		shard := ShardInRedis{}
		err := json.Unmarshal([]byte(d), &shard)
		if err != nil {
			logs.Error("NewShardsFromRedis Err Unmarshal by %s GID %s %s",
				err.Error(), name, d)
			return res, err
		} else {
			res[name] = shard
		}
	}

	logs.Trace("NewShardsFromRedis %v", res)
	return res, nil
}
