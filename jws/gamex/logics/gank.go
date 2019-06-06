package logics

import (
	"encoding/json"

	"fmt"
	"time"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/account/gs"
	"vcs.taiyouxi.net/jws/gamex/models/account/update/data_update"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/message"
	"vcs.taiyouxi.net/jws/gamex/models/sysnotice"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

// 切磋

func (p *Account) gankTownFightBegin(r servers.Request) *servers.Response {
	req := &struct {
		Req
		Acid    string `codec:"acid"`
		RecTime int64  `codec:"rec_t"` // 记录的时间戳，大于0有效，证明是从记录里复仇的
		LogIDS  int    `codec:"ids"`
	}{}
	resp := &struct {
		SyncResp
		Avatar []byte `codec:"avatar"`
	}{}

	initReqRsp(
		"PlayerAttr/GankTownFightBeginRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_param
	)

	dbAccountID, err := db.ParseAccount(req.Acid)
	if err != nil {
		return rpcErrorWithMsg(resp, Err_param, fmt.Sprintf("Err_param acid %s", req.Acid))
	}

	enemy_account, err := account.LoadPvPAccount(dbAccountID)
	if err != nil {
		logs.SentryLogicCritical(req.Acid, "LoadAccount %s Err By %s",
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

	_, _, _, _, bestHero, _, _ := gs.GetCurrAttr(account.NewAccountGsCalculateAdapter(enemy_account))
	a := helper.Avatar2Client{}
	err = account.FromAccount(&a, enemy_account, bestHero[0])
	if err != nil {
		logs.SentryLogicCritical(p.AccountID.String(),
			"FromAccount Err by %s",
			err.Error())
		return rpcError(resp, 2)
	}

	resp.Avatar = encode(a)

	// save to tmp
	p.Tmp.SetGankFightInfo(req.Acid, a.Name, req.RecTime, req.LogIDS)
	return rpcSuccess(resp)
}

func (p *Account) gankTownFightEnd(r servers.Request) *servers.Response {
	req := &struct {
		Req
		IsWin         bool `codec:"win"`
		SendSysNotice bool `codec:"sysnotc"`
		LogIDS        int  `codec:"ids"`
	}{}
	resp := &struct {
		SyncResp
		RecTimes            []int64  `codec:"rts"`
		RecFighterRoleIds   []string `codec:"rfids"`
		RecFighterRoleNames []string `codec:"rfnms"`
		RecIDSs             []int    `codec:"idss"`
	}{}

	initReqRsp(
		"PlayerAttr/GankTownFightEndRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_Enemy_Not_Found
		Err_Hc_Not_Enough
		Err_Json
		Err_DB
		Err_Param
	)

	eAcid, eName, eTS, eIDS := p.Tmp.GetGankFightInfo()
	if eAcid == "" || eName == "" {
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	if req.IsWin {
		// 跑马灯
		if req.SendSysNotice {
			cfg := gamedata.GetGankConf()
			cost := &account.CostGroup{}
			if !cost.AddHc(p.Account, int64(cfg.GetBroadcastCost())) ||
				!cost.CostBySync(p.Account, resp, "GankWinSysNotice") {
				logs.Warn("GankTownFightEndRsp Err_Hc_Not_Enough")
				return rpcWarn(resp, errCode.ClickTooQuickly)
			}
			if req.LogIDS >= gamedata.GankIDSCount() {
				return rpcErrorWithMsg(resp, Err_Param, "Err_Param")
			}
			sysnotice.NewSysRollNotice(p.AccountID.ServerString(),
				int32(gamedata.IDS_GANK_WIN_1+req.LogIDS)).
				AddParam(sysnotice.ParamType_RollName, eName).
				AddParam(sysnotice.ParamType_RollName, p.Profile.Name).Send()
		}
		// 是否是复仇
		if eTS > 0 {
			r := account.GankRecord{
				Time:            eTS,
				IDS:             eIDS,
				FighterRoleId:   eAcid,
				FighterRoleName: eName,
			}
			bb, err := json.Marshal(r)
			if err != nil {
				return rpcErrorWithMsg(resp, Err_Json, fmt.Sprintf("Err_Json err %v", err))
			}
			acid := p.AccountID.String()
			ok := message.RemPlayerMsg(acid,
				account.GankMsgTableKey,
				message.PlayerMsg{
					Params: []string{string(bb)},
				})
			if !ok {
				message.RemGankMsgById(acid, account.GankMsgTableKey,
					eIDS, eAcid, func(eId int, eAcid, record string) bool {
						gankRec := &account.GankRecord{}
						err = json.Unmarshal([]byte(record), gankRec)
						if err != nil {
							logs.SentryLogicCritical(acid, "remplayermsgbyid record json err %s", err.Error())
							return false
						}
						if gankRec.IDS == eId && gankRec.FighterRoleId == eAcid {
							return true
						}
						return false
					})
			}
		}
		// 对方记录记加上
		r := account.GankRecord{
			Time:            time.Now().Unix(),
			IDS:             req.LogIDS,
			FighterRoleId:   p.AccountID.String(),
			FighterRoleName: p.Profile.Name,
		}
		bb, err := json.Marshal(r)
		if err != nil {
			return rpcErrorWithMsg(resp, Err_Json, fmt.Sprintf("Err_Json2 err %v", err))
		}
		message.SendPlayerMsgs(eAcid,
			account.GankMsgTableKey, account.GankMsgCount,
			message.PlayerMsg{
				Params: []string{string(bb)},
			})
		// red point
		player_msg.Send(eAcid, player_msg.PlayerMsgGank,
			player_msg.PlayerGank{
				LogTS: r.Time,
			})
		// log
		p.Profile.GetGank().WinLog()
	}

	logs.Trace("acid %s gank %v", p.AccountID.String(), p.Profile.GetGank())

	// 带回log列表
	msgs, err := message.LoadPlayerMsgs(p.AccountID.String(),
		account.GankMsgTableKey, account.GankMsgCount)
	if err != nil {
		return rpcErrorWithMsg(resp, Err_DB, "Err_DB")
	}

	recs := make([]account.GankRecord, 0, account.GankMsgCount)
	for _, msg := range msgs {
		m := account.GankRecord{}
		err := json.Unmarshal([]byte(msg.Params[0]), &m)
		if err != nil {
			continue
		}
		recs = append(recs, m)
	}

	resp.RecTimes = make([]int64, 0, len(recs))
	resp.RecFighterRoleIds = make([]string, 0, len(recs))
	resp.RecFighterRoleNames = make([]string, 0, len(recs))
	resp.RecIDSs = make([]int, 0, len(recs))
	for _, m := range recs {
		resp.RecTimes = append(resp.RecTimes, m.Time)
		resp.RecFighterRoleIds = append(resp.RecFighterRoleIds, m.FighterRoleId)
		resp.RecFighterRoleNames = append(resp.RecFighterRoleNames, m.FighterRoleName)
		resp.RecIDSs = append(resp.RecIDSs, m.IDS)
	}

	resp.mkInfo(p)

	// log
	logiclog.LogGank(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId,
		eAcid, req.IsWin,
		eTS > 0, req.SendSysNotice,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	return rpcSuccess(resp)
}

func (p *Account) gankGetRecord(r servers.Request) *servers.Response {
	req := &struct {
		Req
	}{}
	resp := &struct {
		SyncResp
		RecTimes            []int64  `codec:"rts"`
		RecFighterRoleIds   []string `codec:"rfids"`
		RecFighterRoleNames []string `codec:"rfnms"`
		RecIDSs             []int    `codec:"idss"`
	}{}

	initReqRsp(
		"PlayerAttr/GankGetRecordRsp",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_DB
	)

	err, recs := p.Profile.GetGank().GetMsgLogs(p.AccountID.String())
	if err != nil {
		return rpcErrorWithMsg(resp, Err_DB, "Err_DB")
	}

	resp.RecTimes = make([]int64, 0, len(recs))
	resp.RecFighterRoleIds = make([]string, 0, len(recs))
	resp.RecFighterRoleNames = make([]string, 0, len(recs))
	resp.RecIDSs = make([]int, 0, len(recs))
	for _, m := range recs {
		resp.RecTimes = append(resp.RecTimes, m.Time)
		resp.RecFighterRoleIds = append(resp.RecFighterRoleIds, m.FighterRoleId)
		resp.RecFighterRoleNames = append(resp.RecFighterRoleNames, m.FighterRoleName)
		resp.RecIDSs = append(resp.RecIDSs, m.IDS)
		p.Profile.GetGank().SetLastReviewLogTS(m.Time)
	}

	return rpcSuccess(resp)
}

type AccountGsCalculateAdapter struct {
	acc *Account
}

func NewAccountGsCalculateAdapter(a *Account) *AccountGsCalculateAdapter {
	return &AccountGsCalculateAdapter{a}
}
