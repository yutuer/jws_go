package logics

import (
	"time"

	"fmt"

	"encoding/json"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/account/update/data_update"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/jws/gamex/modules/simple_pvp_rander"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (p *Account) GetSimplePvpEnemy(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Type int `codec:"type"`
	}{}
	resp := &struct {
		SyncResp
		Avatar      [][]byte `codec:"avatar"`
		Pos         int      `codec:"pos"`
		Score       int64    `codec:"score"`
		ShowWarn    bool     `codec:"warn"`
		SwitchCount int      `codec:"swcount"`
	}{}

	initReqRsp(
		"PlayerAttr/GetSimplePvpEnemyRsp",
		r.RawBytes,
		req, resp, p)

	const (
		SwitchError = 10
		CostError   = 11
		ArgError    = 12
	)

	const (
		GetEnemy    = int(3)
		SwitchEnemy = int(2)
	)
	if req.Type != GetEnemy && req.Type != SwitchEnemy {
		return rpcError(resp, ArgError)
	}
	acid := p.AccountID.String()
	bs := time.Now().UnixNano()
	resp.Pos, resp.Score = rank.GetModule(p.AccountID.ShardId).RankSimplePvp.GetPos(acid)
	uutil.MetricRedisByAccount(p.AccountID, "GetSimplePvp-RankGetPos", fmt.Sprintf("%d", time.Now().UnixNano()-bs))

	resp.Score /= rank.SimplePvpScorePow
	p.Profile.GetSimplePvp().Score = resp.Score
	enemyCount := 1
	resp.Avatar = make([][]byte, 0, enemyCount)
	simplePvpState := p.Tmp.GetSimplePvpState()
	simplePvp := p.Profile.GetSimplePvp()
	if req.Type == GetEnemy {
		// Enemys 是数组,不会访问不到index 0
		if simplePvpState.Enemys[0].GetAcId() != "" {
			canAutoSwitch := simplePvp.CanAutoSwitch(p.Profile.GetProfileNowTime())
			if canAutoSwitch {
				// 超时,需要重新生成敌人信息,由系统自动生成,不需要花费钻石
				resp.ShowWarn = true
			} else {
				// 不需要重新生成敌人信息,直接从Profile里拉取即可
				for _, enemy := range simplePvpState.Enemys {
					resp.Avatar = append(resp.Avatar, encode(enemy))
				}
				resp.SwitchCount = simplePvp.SwitchCount
				resp.mkInfo(p)
				return rpcSuccess(resp)
			}
		}
	}

	// 重新选择,需要扣除相关资源
	if req.Type == SwitchEnemy {
		simplePvp.UpdateSwitchCount(p.Profile.GetProfileNowTime())
		costData := gamedata.GetPvpSwitchCostData(simplePvp.SwitchCount)
		if costData == nil {
			return rpcError(resp, SwitchError)
		}
		costs := account.CostGroup{}

		if !costs.AddCostData(p.Account, costData.Gives()) {
			return rpcError(resp, CostError)
		}
		if !costs.CostBySync(p.Account, resp, "Switch1V1Enemy") {
			return rpcError(resp, CostError)
		}
		simplePvp.SwitchCount += 1
	}

	// 生成敌人信息
	enemys := [account.SimplePvpEnemyCount]helper.Avatar2Client{}
	enemySimpleInfos := [account.SimplePvpEnemyCount]helper.AccountSimpleInfo{}

	// 前3局都是机器人
	if p.Profile.GetSimplePvp().IsDroid() {
		// 当第一次或者没活人时打pvp时随机机器人
		droidEnemys, error := p.mkDroidForSimplePvp(enemyCount)
		if error != nil {
			return rpcError(resp, 2)
		}
		for i, droidEnemy := range droidEnemys {
			enemys[i] = droidEnemy
			resp.Avatar = append(resp.Avatar, encode(droidEnemy))
		}
		logs.Debug("It's Droid")
	} else {
		bs := time.Now().UnixNano()
		enemyIDInRanks := sPvpRander.RandSimplePvpEnemy(
			p.AccountID.ShardId,
			acid,
			enemyCount,
			resp.Pos)
		uutil.MetricRedisByAccount(p.AccountID, "GetSimplePvp-RandSimplePvpEnemy", fmt.Sprintf("%d", time.Now().UnixNano()-bs))

		if enemyIDInRanks != nil && len(enemyIDInRanks) >= enemyCount {
			for idx, enemyIDInRank := range enemyIDInRanks {
				enemyAcID := enemyIDInRank
				dbAccountID, err := db.ParseAccount(enemyAcID)
				if err != nil {
					return rpcError(resp, 6)
				}
				bs := time.Now().UnixNano()
				enemy_account, err := account.LoadPvPAccount(dbAccountID)
				uutil.MetricRedisByAccount(p.AccountID, "GetSimplePvp-LoadPvPAccount", fmt.Sprintf("%d", time.Now().UnixNano()-bs))
				if err != nil {
					logs.SentryLogicCritical(acid, "LoadAccount %s Err By %s",
						dbAccountID, err.Error())

					return rpcError(resp, 1)
				}

				// 数据更新不涉及结构变化
				// 此处处理数据更新, version最后的更新也是在这里做的
				err = data_update.Update(enemy_account.Profile.Ver, true, enemy_account)
				if err != nil {
					logs.SentryLogicCritical(enemy_account.AccountID.String(),
						"data_update Err By %s",
						err.Error())
					return rpcError(resp, 1)
				}

				a := helper.Avatar2Client{}
				err = account.FromAccount(
					&a,
					enemy_account,
					enemy_account.Profile.GetSimplePvp().PvpDefAvatar)
				if err != nil {
					logs.SentryLogicCritical(p.AccountID.String(),
						"FromAccount Err by %s",
						err.Error())
					return rpcError(resp, 2)
				}

				bs = time.Now().UnixNano()
				resPos, resScore := rank.GetModule(p.AccountID.ShardId).RankSimplePvp.GetPos(enemyAcID)
				uutil.MetricRedisByAccount(p.AccountID, "GetSimplePvp-RankGetPos", fmt.Sprintf("%d", time.Now().UnixNano()-bs))
				resScore /= rank.SimplePvpScorePow
				a.SimplePvpScore = resScore
				a.SimplePvpRank = resPos

				enemys[idx] = a
				enemySimpleInfos[idx] = enemy_account.GetSimpleInfo()
				resp.Avatar = append(resp.Avatar, encode(a))
			}
		}
		if enemys[0].GetAcId() == "" {
			droidEnemys, error := p.mkDroidForSimplePvp(enemyCount)
			if error != nil {
				return rpcError(resp, 2)
			}
			for i, droidEnemy := range droidEnemys {
				enemys[i] = droidEnemy
				resp.Avatar = append(resp.Avatar, encode(droidEnemy))
			}
			logs.Debug("It's Droid")
		}
		logs.Trace("enemyIDInRanks %v", enemyIDInRanks)
	}

	simplePvpState.AddEnemy(enemys, enemySimpleInfos)
	simplePvp.LastGetTime = p.Profile.GetProfileNowTime()
	resp.SwitchCount = simplePvp.SwitchCount
	//更新1v1信息
	resp.OnChangeSimplePvp()
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

func (p *Account) BeginSimplePvp(r servers.Request) *servers.Response {
	req := &struct {
		Req
		EnemyIdx int `codec:"enemy_idx"`
	}{}
	resp := &struct {
		SyncResp
	}{}

	initReqRsp(
		"PlayerAttr/BeginSimplePvpRsp",
		r.RawBytes,
		req, resp, p)
	simplePvp := p.Profile.GetSimplePvp()
	canAutoSwitch := simplePvp.CanAutoSwitch(p.Profile.GetProfileNowTime())
	if canAutoSwitch {
		// 超时,需要重新生成敌人信息,不可以开始战斗
		return rpcWarn(resp, errCode.SimplePvpMatchOutOfTime)
	} else {

		simplePvpState := p.Tmp.GetSimplePvpState()
		p.Tmp.SetLevelEnterTime(time.Now().Unix())

		if !p.Profile.GetCounts().UseJustDayBegin(
			gamedata.CounterTypeSimplePvp,
			p.Account) {

			return rpcError(resp, 2)
		}

		if !p.Profile.GetCounts().Use(
			gamedata.CounterTypeSimplePvp,
			p.Account) {
			return rpcError(resp, 2)
		} else {
			resp.OnChangeGameMode(
				gamedata.CounterTypeSimplePvp)
		}

		acid := p.AccountID.String()

		err := p.Profile.GetSimplePvp().OnPvpBegin(
			simplePvp.PvpDefAvatar,
			req.EnemyIdx,
			simplePvpState)
		if err != nil {
			logs.SentryLogicCritical(acid, "OnPvpBegin Err By %s", err.Error())
			return rpcError(resp, 1)
		}
		// 记次数
		p.Profile.GetData().Times_1V1++

		p.updateCondition(account.COND_TYP_SimplePvp,
			1, 0, "", "", resp)

		resp.OnChangeSimplePvp()
	}
	resp.mkInfo(p)

	return rpcSuccess(resp)
}

func (p *Account) EndSimplePvp(r servers.Request) *servers.Response {
	req := &struct {
		Req
		IsSuccess int    `codec:"s"` //TODO 加密 by FanYang
		Hackjson  string `codec:"hackjson"`
	}{}
	resp := &struct {
		SyncResp
		Pos          int   `codec:"pos"`
		Score        int64 `codec:"score"`
		CheatedIndex []int `codec:"cheat_idx"` // 哪项作弊检查未通过，-1无效
	}{}

	const (
		_ = iota
		CodeErrPosDataGetErr
		CodeHackUnmarshalErr
	)

	initReqRsp(
		"PlayerAttr/EndSimplePvpRsp",
		r.RawBytes,
		req, resp, p)

	acid := p.AccountID.String()

	//gs := GetCurrGS(p.Account)
	nowT := time.Now().Unix()
	costTime := nowT - p.Tmp.GetLevelEnterTime()
	simplePvp := p.Profile.GetSimplePvp()

	// 反作弊检查
	if req.Hackjson != "" {
		hacks := []float32{}
		if err := json.Unmarshal([]byte(req.Hackjson), &hacks); err != nil {
			return rpcErrorWithMsg(resp, CodeHackUnmarshalErr, fmt.Sprintf("hack unmarshal err %s", err.Error()))
		}
		resp.CheatedIndex = p.AntiCheat.CheckFightRelAll(
			acid,
			hacks,
			p.Account,
			account.Anticheat_Typ_SimplePVP,
			costTime)
	} else {
		resp.CheatedIndex = []int{}
		logs.Info("[Antichest-Empty] SimplePvp   acid %s req.Hackjson is empty",
			p.AccountID.String())
	}
	if len(resp.CheatedIndex) > 0 {
		return rpcWarn(resp, errCode.YouCheat)
	}
	rankModule := rank.GetModule(p.AccountID.ShardId)
	// old value
	bs := time.Now().UnixNano()
	oldPos, oldScore := rankModule.RankSimplePvp.GetPos(acid)
	uutil.MetricRedisByAccount(p.AccountID, "EndSimplePvp-RankGetPos", fmt.Sprintf("%d", time.Now().UnixNano()-bs))

	oldScore /= rank.SimplePvpScorePow
	// pvp
	simplePvpState := p.Tmp.GetSimplePvpState()

	curr_avatar := simplePvp.PvpDefAvatar
	curr_enemy, curr_enemy_info := simplePvpState.GetCurrEnemy()

	isDroid := gamedata.IsAccountIsAnDroid(curr_enemy.GetAcId())
	var enemyPos, enemyOldPos int
	var enemyScore, enemyOldScore int64

	if !isDroid {
		bs := time.Now().UnixNano()
		enemyOldPos, enemyOldScore = rankModule.RankSimplePvp.GetPos(curr_enemy.GetAcId())
		uutil.MetricRedisByAccount(p.AccountID, "EndSimplePvp-RankGetPos", fmt.Sprintf("%d", time.Now().UnixNano()-bs))

		enemyOldScore /= rank.SimplePvpScorePow
	}

	sa_d, sb_d, err := simplePvp.OnPvpEnd(req.IsSuccess, simplePvpState)
	if err != nil {
		logs.Warn("OnPvpEnd Err By %s %s", acid, err.Error())
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	simpleInfo := p.Account.GetSimpleInfo()

	bs = time.Now().UnixNano()
	rankModule.RankSimplePvp.AddDeta(&simpleInfo, sa_d)
	if !isDroid {
		rankModule.RankSimplePvp.AddEnemyDeta(&curr_enemy_info, sb_d)
	}
	uutil.MetricRedisByAccount(p.AccountID, "EndSimplePvp-UpdateRank", fmt.Sprintf("%d", time.Now().UnixNano()-bs))

	bs = time.Now().UnixNano()
	resp.Pos, resp.Score = rank.GetModule(p.AccountID.ShardId).RankSimplePvp.GetPos(acid)
	uutil.MetricRedisByAccount(p.AccountID, "EndSimplePvp-RankGetPos", fmt.Sprintf("%d", time.Now().UnixNano()-bs))

	resp.Score /= rank.SimplePvpScorePow
	p.Profile.GetFirstPassRank().OnRank(
		gamedata.FirstPassRankTypSimplePvp,
		int(resp.Score))

	if !isDroid {
		bs := time.Now().UnixNano()
		enemyPos, _ = rankModule.RankSimplePvp.GetPos(curr_enemy.GetAcId())
		uutil.MetricRedisByAccount(p.AccountID, "EndSimplePvp-RankGetPos", fmt.Sprintf("%d", time.Now().UnixNano()-bs))
	}

	// Record
	simplePvp.AddSimplePvpHistory(acid, account.SimplePvpFightRecord{
		IsSuccess:         req.IsSuccess > 0,
		RankChange:        oldPos - resp.Pos,
		AvatarID:          curr_avatar,
		AvatarStarLv:      int(simpleInfo.AvatarStarLvl[curr_avatar]),
		EnemyAvatarID:     curr_enemy.AvatarId,
		EnemyAvatarStarLv: int(curr_enemy.HeroStarLv[curr_enemy.AvatarId]),
		EnemyLv:           curr_enemy.CorpLv,
		EnemyID:           curr_enemy.GetAcId(),
		EnemyName:         curr_enemy.Name,
		EnemyGS:           curr_enemy.Gs,
		EnemyRank:         curr_enemy.SimplePvpRank,
		Time:              nowT,
		IsAttack:          true,
	})

	if !isDroid {
		simplePvp.AddSimplePvpHistory(curr_enemy.GetAcId(), account.SimplePvpFightRecord{
			IsSuccess:         req.IsSuccess <= 0,
			RankChange:        enemyOldPos - enemyPos,
			AvatarID:          curr_enemy.AvatarId,
			AvatarStarLv:      int(curr_enemy.HeroStarLv[curr_enemy.AvatarId]),
			EnemyAvatarID:     curr_avatar,
			EnemyAvatarStarLv: int(simpleInfo.AvatarStarLvl[curr_avatar]),
			EnemyLv:           simpleInfo.CorpLv,
			EnemyID:           acid,
			EnemyName:         p.Profile.Name,
			EnemyGS:           p.Profile.GetData().HeroGs[curr_avatar],
			EnemyRank:         oldPos,
			Time:              nowT,
			IsAttack:          false,
		})
	}

	// market activity
	p.Profile.GetMarketActivitys().OnGameMode(p.AccountID.String(),
		gamedata.CounterTypeSimplePvp,
		1,
		p.Profile.GetProfileNowTime())

	resp.OnChangeSimplePvp()
	resp.mkInfo(p)

	// 记log
	if !isDroid {
		enemyPos, enemyScore = rankModule.RankSimplePvp.GetPos(curr_enemy.GetAcId())
		enemyScore /= rank.SimplePvpScorePow
	}
	logiclog.LogPvPFinish(acid, curr_avatar, p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		p.Profile.Data.CorpCurrGS,
		int(oldScore), int(resp.Score),
		oldPos, resp.Pos,
		curr_enemy.GetAcId(), curr_enemy.AvatarId, curr_enemy.Gs,
		int(enemyOldScore), int(enemyScore),
		enemyOldPos, enemyPos,
		req.IsSuccess > 0, costTime,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")

	return rpcSuccess(resp)
}

func (p *Account) GetSimplePvpRecord(r servers.Request) *servers.Response {
	const (
		_           = iota
		CODE_ID_Err // 失败:ID不存在
	)

	req := &struct {
		Req
	}{}

	resp := &struct {
		Resp
		SimplePvpRecord [][]byte `codec:"spvprecord"`
	}{}

	initReqRsp(
		"PlayerAttr/GetSimplePvpRecordRsp",
		r.RawBytes,
		req, resp, p)

	logs.Trace("[%s]GetSimplePvpRecordRsp", p.AccountID)

	simplePvpRecord := p.Profile.GetSimplePvp().GetSimplePvpHistory(p.AccountID.String())
	resp.SimplePvpRecord = make([][]byte, 0, len(simplePvpRecord))

	for _, s := range simplePvpRecord {
		resp.SimplePvpRecord = append(resp.SimplePvpRecord, encode(s))
	}
	if len(resp.SimplePvpRecord) > account.SimplePvpFightRecordCount2Client {
		resp.SimplePvpRecord = resp.SimplePvpRecord[0:account.SimplePvpFightRecordCount2Client]
	}

	return rpcSuccess(resp)
}

// SetSimplePvpDefAvatar : 设置1v1pvp防守者
// 1v1竞技场主界面有个入口打开所有主将列表，选择你的出战者（同时也是你1v1的防守者）

// TODO 由tiprotogen工具生成, 需要实现具体的逻辑

// reqMsgSetSimplePvpDefAvatar 设置1v1pvp防守者请求消息定义
type reqMsgSetSimplePvpDefAvatar struct {
	Req
	DefAvatarID int64 `codec:"aid"` // 防守主将Id(0-关平 1-张飞 。。。)
}

// rspMsgSetSimplePvpDefAvatar 设置1v1pvp防守者回复消息定义
type rspMsgSetSimplePvpDefAvatar struct {
	SyncResp
}

// SetSimplePvpDefAvatar 设置1v1pvp防守者: 1v1竞技场主界面有个入口打开所有主将列表，选择你的出战者（同时也是你1v1的防守者）
func (p *Account) SetSimplePvpDefAvatar(r servers.Request) *servers.Response {
	req := new(reqMsgSetSimplePvpDefAvatar)
	rsp := new(rspMsgSetSimplePvpDefAvatar)

	initReqRsp(
		"Attr/SetSimplePvpDefAvatarRsp",
		r.RawBytes,
		req, rsp, p)

	// logic imp begin
	if p.Profile.GetHero().GetStar(int(req.DefAvatarID)) <= 0 {
		return rpcError(rsp, 1)
	}
	p.Profile.GetSimplePvp().PvpDefAvatar = int(req.DefAvatarID)
	// logic imp end

	rsp.OnChangeSimplePvp()

	rsp.mkInfo(p)
	return rpcSuccess(rsp)
}

func (p *Account) mkDroidForSimplePvp(enemyCount int) ([]helper.Avatar2Client, error) {
	ret := make([]helper.Avatar2Client, 0, 1)
	droid := gamedata.GetRandDroidForSimplePvp()
	for i := 0; i < enemyCount; i++ {
		a := helper.Avatar2Client{}
		err := account.FromAccountByDroid(&a, droid, -1)
		if err != nil {
			logs.SentryLogicCritical(p.AccountID.String(),
				"FromAccount droid Err by %s",
				err.Error())
			return ret, err
		}
		a.SimplePvpScore = rank.SimplePvpInitScoreReal
		a.SimplePvpRank = 9999
		ret = append(ret, a)
	}
	return ret, nil
}
