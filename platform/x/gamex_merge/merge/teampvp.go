package merge

import (
	"encoding/json"
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/team_pvp"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func MergeTeamPvp() error {
	ARankDB := GetARankDBConn()
	defer ARankDB.Close()
	BRankDB := GetBRankDBConn()
	defer BRankDB.Close()
	ResRankDB := GetResRankDBConn()
	defer ResRankDB.Close()

	AM, err := redis.StringMap(ARankDB.Do("HGETALL", team_pvp.TableTeamPvpRank(Cfg.AShardId)))
	if err != nil {
		logs.Error("MergeTeamPvp HGETALL A err %v", err)
		return err
	}

	BM, err := redis.StringMap(BRankDB.Do("HGETALL", team_pvp.TableTeamPvpRank(Cfg.BShardId)))
	if err != nil {
		logs.Error("MergeTeamPvp HGETALL B err %v", err)
		return err
	}

	sumCount := len(AM)
	idx := 1
	res := make(map[int]string, sumCount)
	for i := 1; i <= sumCount; i++ {
		// A
		AInfo, ok := AM[fmt.Sprintf("%d", i)]
		if !ok {
			logs.Error("MergeTeamPvp A not has rank %d", i)
			return fmt.Errorf("MergeTeamPvp rank not found")
		}
		Asm := &helper.AccountSimpleInfo{}
		if err := json.Unmarshal([]byte(AInfo), Asm); err != nil {
			logs.Error("MergeTeamPvp json.Unmarshal err %s", err.Error())
			return err
		}
		if len(res) >= sumCount {
			break
		}
		if !isAcidDel(Asm.AccountID) {
			res[idx] = AInfo
			idx++
		} else {
			logs.Debug("TeamPvp rank del account %s", Asm.AccountID)
		}
		// B
		BInfo, ok := BM[fmt.Sprintf("%d", i)]
		if !ok {
			logs.Error("MergeTeamPvp B not has rank %d", i)
			return fmt.Errorf("MergeTeamPvp rank not found")
		}

		Bsm := &helper.AccountSimpleInfo{}
		if err := json.Unmarshal([]byte(BInfo), Bsm); err != nil {
			logs.Error("MergeTeamPvp json.Unmarshal err %s", err.Error())
			return err
		}
		if isAcidDel(Bsm.AccountID) {
			logs.Debug("TeamPvp rank del account %s", Bsm.AccountID)
			continue
		}
		newBName, ok := account_name_old2new[Bsm.Name]
		if ok {
			Bsm.Name = newBName
			logs.Debug("TeamPvp mem name -> %s", newBName)
			_info, err := json.Marshal(*Bsm)
			if err != nil {
				logs.Error("MergeTeamPvp json.Marshal err %s", err.Error())
				return err
			}
			BInfo = string(_info)
		}

		if len(res) >= sumCount {
			break
		}
		res[idx] = BInfo
		idx++
	}

	// save
	cb := redis.NewCmdBuffer()
	for k, v := range res {
		cb.Send("HSET", team_pvp.TableTeamPvpRank(Cfg.ResShardId), k, v)
	}
	if _, err := ResRankDB.DoCmdBuffer(cb, true); err != nil {
		logs.Error("MergeTeamPvp save res err %s", err.Error())
		return err
	}
	return nil
}
