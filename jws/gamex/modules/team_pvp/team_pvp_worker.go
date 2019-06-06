package team_pvp

import (
	"fmt"
	"math/rand"
	"time"

	"encoding/json"
	"strconv"

	"github.com/astaxie/beego/cache"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/safecache"
)

const (
	_ = iota
	TeamPvp_Cmd_GetRank
	TeamPvp_Cmd_GetEnemy
	TeamPvp_Cmd_Fight
	TeamPvp_Cmd_Update
	TeamPvp_Cmd_MyRank
	TeamPvp_Cmd_Debug_Exchange

	TeamPvp_Cmd_LockPlayerAndBegin
	TeamPvp_Cmd_UnlockPlayer
	TeamPvp_Cmd_UnlockPlayerAndEnd
)

const (
	Redis_Name = "TPVPLOCK"
)

const (
	_ = iota
	LockState
	InvalidState
	SuccessState
)

type TeamPvpCmd struct {
	Typ           int
	Acid          string
	AcidGs        int
	ARand         *rand.Rand
	AcidInfo      *helper.AccountSimpleInfo
	EnemyId       string
	EnemyRank     int
	DebugOperRank []int
	resChan       chan TeamPvpRet
	IsWin         bool // 客户端传送过来是否胜利
}

type TeamPvpRet struct {
	Enemy     helper.AccountSimpleInfo
	Enemies   []TPEnemy
	IsWin     bool
	WinRate   float32
	MyRnd     float32
	MyNewRank int
	Success   bool
	RetState  int
	Params    []string // some param for test
}

type playerInfo struct {
	rank     int                       `json:"r"`
	info     *helper.AccountSimpleInfo `json:"p"`
	islocked bool                      `json:"l"`
}

type worker struct {
	waitter    util.WaitGroupWrapper
	cmd_chan   chan TeamPvpCmd
	playerInfo map[string]*playerInfo // acid->info
	rankInfo   []string               // rank->acid
	lockCache  cache.Cache
}

func (w *worker) start(sid uint) {
	w.cmd_chan = make(chan TeamPvpCmd, 2048)
	w.playerInfo = make(map[string]*playerInfo, gamedata.GetTPvpRankMax())
	w.rankInfo = make([]string, gamedata.GetTPvpRankMax()+1)
	cache, err := safecache.NewSafeCache("teampvp_cache", `{"internal:"60"}`)
	if err != nil {
		logs.Critical("TeamPvp Startup Failed, due to Cache: %s", err.Error())
		panic("TeamPvp Startup Failed, by lockCache init failed")
	}
	w.lockCache = cache
	w.loadRank(sid)

	w.waitter.Wrap(func() {
		for cmd := range w.cmd_chan {
			func() {
				//by YZH 这个让parent never dead, 应该如此吗？
				defer logs.PanicCatcherWithInfo("TeamPvp Worker Panic")
				w.processCommand(sid, &cmd)
			}()
		}
	})
}

func (w *worker) stop() {
	close(w.cmd_chan)
	w.waitter.Wait()
}

func (w *worker) processCommand(sid uint, cmd *TeamPvpCmd) {
	switch cmd.Typ {
	case TeamPvp_Cmd_GetRank:
		w.getRank(sid, cmd)
	case TeamPvp_Cmd_GetEnemy:
		w.getEnemies(sid, cmd)
	case TeamPvp_Cmd_Fight:
		w.fightEnemy(sid, cmd)
	case TeamPvp_Cmd_Update:
		w.updateInfo(sid, cmd)
	case TeamPvp_Cmd_MyRank:
		w.getMyRank(sid, cmd)
	case TeamPvp_Cmd_Debug_Exchange:
		w.debugExchange(sid, cmd)
	case TeamPvp_Cmd_UnlockPlayerAndEnd:
		w.unlockPlayerAndEnd(sid, cmd)
	case TeamPvp_Cmd_LockPlayerAndBegin:
		w.lockPlayerAndBegin(sid, cmd)
	case TeamPvp_Cmd_UnlockPlayer:

	}
}

func (w *worker) getRank(sid uint, cmd *TeamPvpCmd) {
	ret := TeamPvpRet{
		Success: true,
		Enemies: make([]TPEnemy, 0, TeamPvp_Rank_Show_Count),
	}
	for i := 1; i <= TeamPvp_Rank_Show_Count; i++ {
		if i >= len(w.rankInfo) {
			break
		}

		id := w.rankInfo[i]
		info := w.playerInfo[id]
		fas := make([]int, 0, len(info.info.TeamPvpAvatar))
		fastarlv := make([]int, 0, len(info.info.TeamPvpAvatarLv))
		fas = append(fas, info.info.TeamPvpAvatar[:]...)
		fastarlv = append(fastarlv, info.info.TeamPvpAvatarLv[:]...)
		ret.Enemies = append(ret.Enemies, TPEnemy{
			Acid:     info.info.AccountID,
			Name:     info.info.Name,
			Gs:       info.info.TeamPvpGs,
			Rank:     info.rank,
			FAs:      fas,
			FAStarLv: fastarlv,
		})
	}
	cmd.resChan <- ret
}
func (w *worker) getMyRank(sid uint, cmd *TeamPvpCmd) {
	ret := TeamPvpRet{
		Success: true,
	}
	info, ok := w.playerInfo[cmd.Acid]
	if ok {
		ret.MyNewRank = info.rank
	}
	cmd.resChan <- ret
}

func (w *worker) getEnemies(sid uint, cmd *TeamPvpCmd) {
	ret := TeamPvpRet{
		Success: true,
		Enemies: make([]TPEnemy, 0, gamedata.SumEnemyCountInList),
	}
	var enemiesIds []int
	if info, ok := w.playerInfo[cmd.Acid]; ok { // 在榜里
		enemiesIds = gamedata.TPvpMatchEnemies(uint32(info.rank))
	} else { // 不在榜里
		enemiesIds = gamedata.TPvpMatchEnemies(gamedata.GetTPvpRankMax())
	}

	for _, erid := range enemiesIds {
		eInfo := w.playerInfo[w.rankInfo[erid]]
		fas := make([]int, 0, len(eInfo.info.TeamPvpAvatar))
		fastarlv := make([]int, 0, len(eInfo.info.TeamPvpAvatarLv))
		fas = append(fas, eInfo.info.TeamPvpAvatar[:]...)
		fastarlv = append(fastarlv, eInfo.info.TeamPvpAvatarLv[:]...)
		ret.Enemies = append(ret.Enemies, TPEnemy{
			Acid:     eInfo.info.AccountID,
			Name:     eInfo.info.Name,
			Gs:       eInfo.info.TeamPvpGs,
			Rank:     eInfo.rank,
			FAs:      fas,
			FAStarLv: fastarlv,
		})
	}
	cmd.resChan <- ret
}

func mkGsByTPVPRarityImpact(gs int,
	avatarIdxLvs,
	avatarIdxs,
	enavatarIdxLvs,
	enavatarIdxs []int) int {

	var (
		radio                  float32
		rRadio, sRadio, tRadio [helper.TeamPvpAvatarsCount]float32
		num                    float32 = float32(len(avatarIdxs))
	)

	// 第1步：计算品质影响
	for idx, heroIdx := range avatarIdxs {
		data := gamedata.GetHeroData(heroIdx)
		if data == nil {
			continue
		}

		rRadio[idx] = data.GsAddon
	}

	// 第2步：计算被动技能修正影响
	for idx, heroIdx := range avatarIdxs {
		data := gamedata.GetHeroData(heroIdx)
		if data == nil {
			continue
		}

		ldata := data.LvData[avatarIdxLvs[idx]]

		idData := gamedata.GetIdentityData(ldata.IdId)
		if idData == nil {
			continue
		}

		if idData.PassiveSkill == nil {
			continue
		} else {
			sRadio[idx] = idData.PassiveSkill.GetModifyValue() *
				idData.PassiveSkill.GetCorrectionRatio()
		}

	}

	// 第3步：计算克制影响
	for idx, heroIdx := range avatarIdxs {
		for enIdx, enHeroIdx := range enavatarIdxs {
			selfData := gamedata.GetHeroData(heroIdx)
			enemData := gamedata.GetHeroData(enHeroIdx)
			if selfData == nil || enemData == nil {
				continue
			}

			selfLvData := selfData.LvData[avatarIdxLvs[idx]]
			enemLvData := enemData.LvData[enavatarIdxLvs[enIdx]]

			selfIDData := gamedata.GetIdentityData(selfLvData.IdId)
			enemIDData := gamedata.GetIdentityData(enemLvData.IdId)

			if selfIDData == nil || selfIDData.CounterSkill == nil {
				continue
			}

			if enemIDData == nil || enemIDData.CounterSkill == nil {
				continue
			}

			// 本身对对方造成的影响 CounterSkill表Counter3v3Target列控制，0自己，1对方
			if selfIDData.CounterSkill.GetCounter3V3Target() == 0 &&
				selfIDData.CounterSkill.IfTakeEffect(enemIDData) {
				tRadio[idx] += selfIDData.CounterSkill.GetModifyValue() *
					selfIDData.CounterSkill.GetCorrectionRatio()
			}

			// 对方对自己生效的
			if enemIDData.CounterSkill.GetCounter3V3Target() == 1 &&
				enemIDData.CounterSkill.IfTakeEffect(selfIDData) {
				tRadio[idx] += enemIDData.CounterSkill.GetModifyValue() *
					enemIDData.CounterSkill.GetCorrectionRatio()
			}
		}
	}

	for i := 0; i < helper.TeamPvpAvatarsCount; i++ {
		radio += (rRadio[i] + 1.0) * (sRadio[i] + 1.0) * (tRadio[i] + 1.0)
	}

	logs.Trace("mkGsByTPVPRarityImpact %v %v %v %v", radio, rRadio, sRadio, tRadio)

	return int((float32(gs) / num) * radio)
}

func (w *worker) fightEnemy(sid uint, cmd *TeamPvpCmd) {
	ret := TeamPvpRet{
		Success: true,
	}
	eInfo, ok := w.playerInfo[cmd.EnemyId]
	// 排名变了，需要刷新列表
	if !ok || eInfo.rank != cmd.EnemyRank {
		// 排名上的人变了，刷新全部列表
		w.getEnemies(sid, cmd)
		return
	}
	// 算胜率
	ret.Enemy = *eInfo.info
	aGs := mkGsByTPVPRarityImpact(
		cmd.AcidGs,
		cmd.AcidInfo.TeamPvpAvatarLv[:],
		cmd.AcidInfo.TeamPvpAvatar[:],
		eInfo.info.TeamPvpAvatarLv[:],
		eInfo.info.TeamPvpAvatar[:])

	eGs := mkGsByTPVPRarityImpact(
		eInfo.info.TeamPvpGs,
		eInfo.info.TeamPvpAvatarLv[:],
		eInfo.info.TeamPvpAvatar[:],
		cmd.AcidInfo.TeamPvpAvatarLv[:],
		cmd.AcidInfo.TeamPvpAvatar[:])

	ret.IsWin, ret.WinRate, ret.MyRnd = CalcTeamPvpResult(aGs, eGs, cmd.ARand)
	ret.Params = make([]string, 0, 2)
	ret.Params = append(ret.Params, strconv.Itoa(aGs))
	ret.Params = append(ret.Params, strconv.Itoa(eGs))
	myOld := w.playerInfo[cmd.Acid]
	if myOld != nil {
		ret.MyNewRank = myOld.rank // 先填上老名次
	}
	// 胜利了要交换
	resDB := make(map[int]helper.AccountSimpleInfo, 2)
	enemyRank := eInfo.rank
	if ret.IsWin {
		if myOld != nil { // 我在榜上
			myRank := myOld.rank
			if enemyRank < myRank { // 敌人名次比我高
				w.playerInfo[cmd.Acid].rank = enemyRank
				w.playerInfo[cmd.EnemyId].rank = myRank
				w.rankInfo[enemyRank] = cmd.Acid
				w.rankInfo[myRank] = cmd.EnemyId
				myNew := w.playerInfo[cmd.Acid]
				enemyNew := w.playerInfo[cmd.EnemyId]
				ret.MyNewRank = myNew.rank
				// 更新敌人的rank
				if !IsTPvpRobotId(cmd.EnemyId) {
					player_msg.Send(cmd.EnemyId, player_msg.PlayerMsgTeamPvpRankChgCode,
						player_msg.PlayerMsgTeamPvpRankChg{
							Rank: enemyNew.rank,
						})
				}
				resDB[myNew.rank] = *myNew.info
				resDB[enemyNew.rank] = *enemyNew.info
			}
		} else { // 我不在榜上
			w.playerInfo[cmd.Acid] = &playerInfo{
				rank: enemyRank,
				info: cmd.AcidInfo,
			}
			delete(w.playerInfo, cmd.EnemyId)
			w.rankInfo[enemyRank] = cmd.Acid
			myNew := w.playerInfo[cmd.Acid]
			ret.MyNewRank = myNew.rank
			resDB[myNew.rank] = *myNew.info
			// 更新敌人的rank
			if !IsTPvpRobotId(cmd.EnemyId) {
				player_msg.Send(cmd.EnemyId, player_msg.PlayerMsgTeamPvpRankChgCode,
					player_msg.PlayerMsgTeamPvpRankChg{
						Rank: 0,
					})
			}
		}
	}
	cmd.resChan <- ret
	// 存db
	if ret.IsWin {
		GetModule(sid).CommandSaveExec(dbCmd{
			typ: DB_Cmd_Save,
			chg: resDB,
		})
	}
}

func (w *worker) updateInfo(sid uint, cmd *TeamPvpCmd) {
	ret := TeamPvpRet{
		Success: true,
	}
	o, ok := w.playerInfo[cmd.Acid]
	if ok {
		w.playerInfo[cmd.Acid].info = cmd.AcidInfo
	}
	cmd.resChan <- ret

	if ok {
		resDB := make(map[int]helper.AccountSimpleInfo, 1)
		resDB[o.rank] = *cmd.AcidInfo
		GetModule(sid).CommandSaveExec(dbCmd{
			typ: DB_Cmd_Save,
			chg: resDB,
		})
	}
}

func (w *worker) loadRank(sid uint) {
	_db := modules.GetDBConn()
	defer _db.Close()

	m, err := redis.StringMap(_do(_db, "HGETALL", TableTeamPvpRank(sid)))
	if err != nil && err != redis.ErrNil {
		logs.Error("TeamPvp loadRank err %v", err)
		panic(fmt.Errorf("TeamPvp loadRank err %v", err))
	}

	if err == redis.ErrNil || len(m) <= 0 {
		initNewRobots(w, sid, nil)
		logs.Info("Teampvp initAllRobot %d", len(w.playerInfo))
	} else {
		duplicateMap := make(map[string]int) // acid, rid 最佳排名
		robotRanks := make([]int, 0)         // 要填充的机器人
		for r, info := range m {
			rid, err := strconv.Atoi(r)
			if err != nil {
				logs.Error("TeamPvp loadRank Atoi err %s", err.Error())
				panic(fmt.Errorf("TeamPvp loadRank Atoi err %v", err))
			}
			sm := &helper.AccountSimpleInfo{}
			if err := json.Unmarshal([]byte(info), sm); err != nil {
				logs.Error("TeamPvp loadRank Unmarshal err %s	%d %s", err.Error(), rid, info)
				panic(fmt.Errorf("TeamPvp loadRank Unmarshal err %v", err))
			}
			if bestRank, ok := duplicateMap[sm.AccountID]; ok {
				if rid < bestRank {
					robotRanks = append(robotRanks, bestRank)

					w.setPlayerInfo(rid, sm)
					duplicateMap[sm.AccountID] = rid
				} else {
					robotRanks = append(robotRanks, rid)
				}
			} else {
				w.setPlayerInfo(rid, sm)
				duplicateMap[sm.AccountID] = rid
			}
		}
		w.initRobots(sid, robotRanks)
		logs.Info("Teampvp loadRank %d", len(w.playerInfo))
	}
}

func (w *worker) setPlayerInfo(rid int, sm *helper.AccountSimpleInfo) {
	w.rankInfo[rid] = sm.AccountID
	w.playerInfo[sm.AccountID] = &playerInfo{
		rank: rid,
		info: sm,
	}
}

func (w *worker) initRobots(sid uint, robotRanks []int) {
	if len(robotRanks) <= 0 {
		return
	}

	initNewRobots(w, sid, robotRanks)
	logs.Warn("team pvp change rank, %v", robotRanks)
}

func (w *worker) debugExchange(sid uint, cmd *TeamPvpCmd) {
	ret := TeamPvpRet{
		Success: true,
	}
	if len(cmd.DebugOperRank) < 2 {
		cmd.resChan <- ret
		return
	}
	r1 := cmd.DebugOperRank[0]
	r2 := cmd.DebugOperRank[1]
	if r1 <= 0 || r2 <= 0 || r1 == r2 ||
		r1 >= len(w.rankInfo) || r2 >= len(w.rankInfo) {
		cmd.resChan <- ret
		return
	}

	r1Id := w.rankInfo[r1]
	r2Id := w.rankInfo[r2]
	// exchange
	w.rankInfo[r1] = r2Id
	w.rankInfo[r2] = r1Id
	w.playerInfo[r1Id].rank = r2
	w.playerInfo[r2Id].rank = r1

	cmd.resChan <- ret

	resDB := make(map[int]helper.AccountSimpleInfo, 2)
	nr1 := w.playerInfo[r1Id]
	nr2 := w.playerInfo[r2Id]
	resDB[nr1.rank] = *nr1.info
	resDB[nr2.rank] = *nr2.info
	GetModule(sid).CommandSaveExec(dbCmd{
		typ: DB_Cmd_Save,
		chg: resDB,
	})

	if !IsTPvpRobotId(r1Id) {
		player_msg.Send(r1Id, player_msg.PlayerMsgTeamPvpRankChgCode,
			player_msg.PlayerMsgTeamPvpRankChg{
				Rank: nr1.rank,
			})
	}
	if !IsTPvpRobotId(r2Id) {
		player_msg.Send(r2Id, player_msg.PlayerMsgTeamPvpRankChgCode,
			player_msg.PlayerMsgTeamPvpRankChg{
				Rank: nr2.rank,
			})
	}
}

func CalcTeamPvpResult(myGs, enemyGs int, r *rand.Rand) (isWin bool, winRate, myRnd float32) {
	YourGSMax := (float32)(gamedata.GetTPvpCommonCfg().GetYourGSMax())
	YourGSMin := (float32)(gamedata.GetTPvpCommonCfg().GetYourGSMin())
	HopeLimitMax := gamedata.GetTPvpCommonCfg().GetHopeLimitMax()
	HopeLimitMin := gamedata.GetTPvpCommonCfg().GetHopeLimitMin()
	G0 := (float32)(myGs)
	G1 := (float32)(enemyGs)
	//R := HopeLimitMin - (G0-YourGSMin)/(YourGSMax-YourGSMin)*(HopeLimitMin-HopeLimitMax)

	a := (HopeLimitMax - HopeLimitMin) * YourGSMax * YourGSMin / (YourGSMin - YourGSMax)
	b := (YourGSMax*HopeLimitMax - YourGSMin*HopeLimitMin) / (YourGSMax - YourGSMin)
	R := a/G0 + b

	Gmax := (1 + R) * G0
	Gmin := 1 / (1 + R) * G0

	if G1 >= Gmax { // 必能胜你的玩家最低战力是
		logs.Trace("CalcTeamPvpResult definitely fail")
		return false, 0.0, 0.0
	}
	if G1 <= Gmin { // 你能战胜的玩家的最高战力P
		logs.Trace("CalcTeamPvpResult definitely win")
		return true, 1.0, 1.0
	}
	logs.Trace("CalcTeamPvpResult definitely rate ...")
	rate := (Gmax - G1) / (Gmax - Gmin) // 玩家本局的胜率为
	rndF := r.Float32()
	logs.Trace("teampvp randF %f rate %f isWin %v myGs %f enemyGs %f", rndF, rate, isWin, G0, G1)
	if rndF <= rate {
		return true, rate, rndF
	}
	return false, rate, rndF
}

func (w *worker) lockPlayerAndBegin(sid uint, cmd *TeamPvpCmd) {
	ret := TeamPvpRet{}

	eInfo, ok := w.playerInfo[cmd.EnemyId]
	// 排名变了，需要刷新列表
	if !ok || eInfo.rank != cmd.EnemyRank {
		// 排名上的人变了，刷新全部列表
		w.getEnemies(sid, cmd)
		return
	}
	key := getCacheKey(sid, cmd.EnemyId)
	isLock := w.lockCache.IsExist(key)
	if isLock {
		logs.Info("The enemy you want fight is locked now")
		ret.RetState = LockState
		ret.Success = false
		cmd.resChan <- ret
		return
	}

	systemTime := gamedata.GetTPvpCommonCfg().GetServerServeTime()

	err := w.lockCache.Put(key, cmd.Acid, time.Duration(int64(systemTime)))
	if err != nil {
		logs.Error("Cache Error by: %v", err)
		ret.Success = false
	} else {
		ret.Success = true
	}
	ret.Enemy = *eInfo.info
	cmd.resChan <- ret
}

func (w *worker) _validKey(key, curVal string) bool {
	if !w.lockCache.IsExist(key) {
		logs.Debug("team_pvp commit timeout(target has been unlocked)")
		return false
	}

	ret := w.lockCache.Get(key)
	if ret == nil {
		logs.Debug("team_pvp commit timeout(target has been unlocked)-next")
		return false
	}

	if ret.(string) != curVal {
		logs.Debug("team_pvp commit timeout(target is locked by other player)")
		return false
	}

	return true

}

func (w *worker) unlockPlayerAndEnd(sid uint, cmd *TeamPvpCmd) {
	ret := TeamPvpRet{}
	key := getCacheKey(sid, cmd.EnemyId)

	if !w._validKey(key, cmd.Acid) {
		ret.Success = false
		ret.RetState = InvalidState
		cmd.resChan <- ret
		logs.Debug("can't unlock enemy")
		return
	}

	eInfo, ok := w.playerInfo[cmd.EnemyId]
	if !ok {
		// 理论上不可能走到这一步
		ret.Success = false
		cmd.resChan <- ret
		logs.Error("Fatal Error, can't get enemyInfo")
		return
	}
	logs.Debug("Enemy Rank is %d", eInfo.rank)

	err := w.lockCache.Delete(key)
	if err != nil {
		logs.Error("Unock player failed by %v", err)
		ret.Success = false
		cmd.resChan <- ret
		logs.Error("Fetal Error, can't unlock enemy")
		return
	} else {
		ret.Success = true
	}
	ret.Enemy = *eInfo.info
	myOld := w.playerInfo[cmd.Acid]
	if myOld != nil {
		ret.MyNewRank = myOld.rank // 先填上老名次
	}
	// 胜利了要交换
	resDB := make(map[int]helper.AccountSimpleInfo, 2)
	enemyRank := eInfo.rank
	if cmd.IsWin {
		if myOld != nil { // 我在榜上
			myRank := myOld.rank
			if enemyRank < myRank { // 敌人名次比我高
				w.playerInfo[cmd.Acid].rank = enemyRank
				w.playerInfo[cmd.EnemyId].rank = myRank
				w.rankInfo[enemyRank] = cmd.Acid
				w.rankInfo[myRank] = cmd.EnemyId
				myNew := w.playerInfo[cmd.Acid]
				enemyNew := w.playerInfo[cmd.EnemyId]
				logs.Debug("Enemy Rank(in rank) is %d", ret.MyNewRank)
				ret.MyNewRank = myNew.rank
				// 更新敌人的rank
				if !IsTPvpRobotId(cmd.EnemyId) {
					player_msg.Send(cmd.EnemyId, player_msg.PlayerMsgTeamPvpRankChgCode,
						player_msg.PlayerMsgTeamPvpRankChg{
							Rank: enemyNew.rank,
						})
				}
				resDB[myNew.rank] = *myNew.info
				resDB[enemyNew.rank] = *enemyNew.info
			}
		} else { // 我不在榜上
			w.playerInfo[cmd.Acid] = &playerInfo{
				rank: enemyRank,
				info: cmd.AcidInfo,
			}
			delete(w.playerInfo, cmd.EnemyId)
			w.rankInfo[enemyRank] = cmd.Acid
			myNew := w.playerInfo[cmd.Acid]
			ret.MyNewRank = myNew.rank
			logs.Debug("Enemy Rank(not in rank) is %d", ret.MyNewRank)
			resDB[myNew.rank] = *myNew.info
			// 更新敌人的rank
			if !IsTPvpRobotId(cmd.EnemyId) {
				player_msg.Send(cmd.EnemyId, player_msg.PlayerMsgTeamPvpRankChgCode,
					player_msg.PlayerMsgTeamPvpRankChg{
						Rank: 0,
					})
			}
		}
	}

	cmd.resChan <- ret

	// 存db
	if cmd.IsWin {
		GetModule(sid).CommandSaveExec(dbCmd{
			typ: DB_Cmd_Save,
			chg: resDB,
		})
	}
}

//直接发送解锁命令回调
func (w *worker) unlockPlayer(sid uint, cmd *TeamPvpCmd) {
	ret := TeamPvpRet{}
	key := getCacheKey(sid, cmd.EnemyId)
	err := w.lockCache.Delete(key)
	if err != nil {
		logs.Error("Unock player failed by %v", err)
		ret.Success = false
	} else {
		ret.Success = true
	}
	cmd.resChan <- ret
}

func getCacheKey(sid uint, uid string) string {
	return fmt.Sprintf("%d_%s", sid, uid)
}
