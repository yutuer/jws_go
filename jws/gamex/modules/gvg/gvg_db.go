package gvg

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

const (
	db_counter_key = "GVG_DB"
)

type Gvg2DB struct {
	LastWorldInfo
	CityInfo []GvgCity2DB `json:"city_info"`
	/*---------------------*/
	Players []gvgPlayer2DB `json:"players"`
	Guilds  []gvgGuild2DB  `json:"guilds"`
}

type GvgCity2DB struct {
	ID              int                      `json:"id"`
	LeaderGuildID   string                   `json:"leader_guild_id"`
	LeaderGuildName string                   `json:"leader_guild_name"`
	TopNLeader      [gamedata.GVGTopN]string `json:"top_n_leader"`
}

type gvgPlayer2DB struct {
	AcID      string              `json:"acid"`
	CityScore []gvgPlayerScore2DB `json:"city_score"`
	Name      string              `json:"name"`
	GuildID   string              `json:"guild_id"`
}

type gvgPlayerScore2DB struct {
	CityID int   `json:"city_id"`
	Score  int64 `json:"score"`
}

type gvgGuild2DB struct {
	GuildID   string             `json:"guild_id"`
	CityScore []gvgGuildScore2DB `json:"city_score"`
	Name      string             `json:"name"`
}

type gvgGuildScore2DB struct {
	CityID int   `json:"city_id"`
	Score  int64 `json:"score"`
}

func (m *Gvg2DB) dbSave(shardId uint) error {
	cb := redis.NewCmdBuffer()

	if err := driver.DumpToHashDBCmcBuffer(cb, TableGVG(shardId), m); err != nil {
		return fmt.Errorf("DumpToHashDBCmcBuffer err %v", err)
	}

	db := driver.GetDBConn()
	defer db.Close()
	if db.IsNil() {
		return fmt.Errorf("cant get redis conn")
	}

	if _, err := modules.DoCmdBufferWrapper(
		db_counter_key, db, cb, true); err != nil {
		return fmt.Errorf("DoCmdBuffer error %s", err.Error())
	}
	return nil
}

func (m *Gvg2DB) dbLoad(shardId uint) error {
	_db := driver.GetDBConn()
	defer _db.Close()

	err := driver.RestoreFromHashDB(_db.RawConn(),
		TableGVG(shardId), m, false, false)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	return nil
}

func (m *Gvg2DB) InitPG() {
	m.Players = make([]gvgPlayer2DB, 0)
	m.Guilds = make([]gvgGuild2DB, 0)
}
