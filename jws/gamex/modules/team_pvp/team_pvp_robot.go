package team_pvp

import (
	"fmt"
	"strconv"
	"strings"

	"encoding/json"

	"math/rand"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules"
	metricsModules "vcs.taiyouxi.net/platform/planx/metrics/modules"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
)

const (
	tpvp_flag = "TPvp"
)

// 机器人id
type TPvpRobot struct {
	Gid        uint
	Sid        uint
	InitRank   int
	RobotCfgId int32
}

func (r *TPvpRobot) getRobotId() string {
	return fmt.Sprintf("%d:%d:%s:%d:%d", r.Gid, r.Sid, tpvp_flag, r.InitRank, r.RobotCfgId)
}

func IsTPvpRobotId(robotId string) bool {
	_, err := ParseTPvpRobotId(robotId)
	return err == nil
}

func ParseTPvpRobotId(robotId string) (TPvpRobot, error) {
	ss := strings.SplitN(robotId, ":", 5)
	if len(ss) < 5 {
		return TPvpRobot{}, fmt.Errorf("TPvp robotId format err %s", robotId)
	}
	gid, err := strconv.Atoi(ss[0])
	if err != nil {
		return TPvpRobot{}, fmt.Errorf("TPvp robotId format err %s", robotId)
	}
	sid, err := strconv.Atoi(ss[1])
	if err != nil {
		return TPvpRobot{}, fmt.Errorf("TPvp robotId format err %s", robotId)
	}
	if ss[2] != tpvp_flag {
		return TPvpRobot{}, fmt.Errorf("TPvp robotId format err %s", robotId)
	}
	rank, err := strconv.Atoi(ss[3])
	if err != nil {
		return TPvpRobot{}, fmt.Errorf("TPvp robotId format err %s", robotId)
	}
	robotCfgId, err := strconv.Atoi(ss[4])
	if err != nil {
		return TPvpRobot{}, fmt.Errorf("TPvp robotId format err %s", robotId)
	}
	return TPvpRobot{
		Gid:        uint(gid),
		Sid:        uint(sid),
		InitRank:   rank,
		RobotCfgId: int32(robotCfgId),
	}, nil
}

func initNewRobots(w *worker, sid uint, robotRanks []int) {
	cbNames := redis.NewCmdBuffer()
	cbRank := redis.NewCmdBuffer()

	rng := &util.Kiss64Rng{}
	rng.Seed(time.Now().Unix())
	r := rand.New(rng)

	// 初始化机器人
	nameIndex := 0
	var nameCount int
	if robotRanks == nil {
		nameCount = int(gamedata.GetTPvpRankMax())
	} else {
		nameCount = len(robotRanks)
	}
	names := gamedata.RandRobotNames(nameCount)
	mCfg := gamedata.GetTPvpMatchCfg()
	for _, mc := range mCfg {
		for i := mc.MatchCfg.GetStart(); i <= mc.MatchCfg.GetEnd(); i++ {
			if containsRobotRanks(robotRanks, int(i)) {
				rid := TPvpRobot{
					Gid:        uint(game.Cfg.Gid),
					Sid:        sid,
					InitRank:   int(i),
					RobotCfgId: mc.MatchCfg.GetRobotID(),
				}
				rCfg := gamedata.GetDroidForTeamPvp(uint32(mc.MatchCfg.GetRobotID()))
				if rCfg == nil {
					panic(fmt.Errorf("TeamPvp initAllRobot robotCfgId %d not found", mc.MatchCfg.GetRobotID()))
				}
				gs := int32(rCfg.CorpGs)
				gs_up := gs + mc.MatchCfg.GetGSRandRage_Upper()
				gs_low := gs - mc.MatchCfg.GetGSRandRage_Lower()
				gs = r.Int31n(gs_up-gs_low+1) + gs_low
				teamPvpAvatars := [helper.TeamPvpAvatarsCount]int{}
				teamPvpAvatarLvs := [helper.TeamPvpAvatarsCount]int{}
				for i, a := range util.Shuffle1ToN(3) {
					teamPvpAvatars[i] = a
					teamPvpAvatarLvs[i] = 1
				}
				p := &playerInfo{
					rank: int(i),
					info: &helper.AccountSimpleInfo{
						Name:            names[nameIndex],
						AccountID:       rid.getRobotId(),
						CorpLv:          rCfg.CorpLv,
						CurrCorpGs:      int(gs),
						TeamPvpGs:       int(gs),
						TeamPvpAvatar:   teamPvpAvatars,
						TeamPvpAvatarLv: teamPvpAvatarLvs,
					},
				}
				nameIndex++
				w.playerInfo[p.info.AccountID] = p
				w.rankInfo[i] = p.info.AccountID

				jInfo, err := json.Marshal(*p.info)
				if err != nil {
					panic(fmt.Errorf("TeamPvp initAllRobot Marshal err %s", err.Error()))
				}
				cbNames.Send("HSET", driver.TableChangeName(sid), p.info.Name, p.info.AccountID)
				cbRank.Send("HSET", TableTeamPvpRank(sid), p.rank, string(jInfo))
			}
		}
	}

	// 存db
	_db := modules.GetDBConn()
	defer _db.Close()

	if _, err := metricsModules.DoCmdBufferWrapper(TeamPvp_DB_Counter_Name, _db, cbRank, true); err != nil {
		panic(fmt.Errorf("TeamPvp initAllRobot DoCmdBufferWrapper err %s", err.Error()))
	}

	db := driver.GetDBConn()
	defer db.Close()
	if _, err := metricsModules.DoCmdBufferWrapper(TeamPvp_DB_Counter_Name, db, cbNames, true); err != nil {
		panic(fmt.Errorf("TeamPvp initAllRobot names DoCmdBufferWrapper err %s", err.Error()))
	}
}

func containsRobotRanks(robotRanks []int, rid int) bool {
	if robotRanks == nil {
		return true
	}
	for _, rank := range robotRanks {
		if rank == rid {
			return true
		}
	}
	return false
}
