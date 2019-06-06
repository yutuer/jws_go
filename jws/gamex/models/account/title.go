package account

import (
	"math"

	"time"

	"sort"

	"vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/csrob"
	"vcs.taiyouxi.net/jws/gamex/modules/gvg"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/jws/gamex/modules/title_rank"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const title_update_time_offset = 2 // 称号更新时间延迟2s

type PlayerTitle struct {
	TitleCanActivate map[string]struct{} // 除了有实效的称号之外的，所有可激活称号
	TitleHadActivate map[string]struct{} // 除了有实效的称号之外的，所有已激活称号

	TitleSimplePvp           string
	NextSimplePvpRefreshTime int64
	TitleTeamPvp             string
	NextTeamPvpRefreshTime   int64
	TitleGVG                 string
	NextGVGRefreshTime       int64
	TitleWuShuang            string
	NextWuShuangRefreshTime  int64

	CSRobTitle     bool
	CSRobTitleTime int64

	TitleTakeOn    string
	TitleForClient map[string]struct{}

	FirstTitle bool
}

// 因为bilog所以大写public
type PlayerTitleInDB struct {
	TitleCanActivate []string `json:"title_can_act"`
	TitleHadActivate []string `json:"title_act"`

	TitleSimplePvp           string   `json:"title_spvp"`
	NextSimplePvpRefreshTime int64    `json:"next_sp_ref_t"`
	TitleTeamPvp             string   `json:"title_tpvp"`
	NextTeamPvpRefreshTime   int64    `json:"next_tp_ref_t"`
	TitleGVG                 string   `json:"title_gvg"`
	NextGVGRefreshTime       int64    `json:"next_gvg_ref_t"`
	TitleWuShuagn            string   `json:"title_wu_shuagn"`
	NextWushuangRefreshTime  int64    `json:"next_wushuang_refresh_time"`
	TitleTakeOn              string   `json:"title_on"`
	TitleForClient           []string `json:"title_cli"`

	CSRobTitle     bool
	CSRobTitleTime int64 `json:""csrob_t`

	FirstTitle bool `json:"first_title"`
}

func (pt *PlayerTitle) UpdateTitle(p *Account, now_time int64) {
	titleOnChg, oldTitleOn := pt.updateSimplePvpTitle(p, now_time)
	titleOnChg1, oldTitleOn1 := pt.updateTeamPvpTitle(p, now_time)
	titleOnChg2, oldTitleOn2 := pt.updateGVGTitle(p, now_time)
	titleOnChg3, oldTitleOn3 := pt.updateWuShuangTitle(p, now_time)
	pt.Update7DayTitle(p)
	// ok ?
	pt.OnOneWorld(p)
	// handle
	if titleOnChg || titleOnChg1 || titleOnChg2 || titleOnChg3 {
		var old string
		if oldTitleOn != "" {
			old = oldTitleOn
		}
		if oldTitleOn1 != "" {
			old = oldTitleOn1
		}
		if oldTitleOn2 != "" {
			old = oldTitleOn2
		}
		if oldTitleOn3 != "" {
			old = oldTitleOn3
		}
		p.GetHandle().OnTitleOnChg(old, pt.TitleTakeOn)
	}

	got, effectTime := pt.checkCSRobTitle(p)
	if true == pt.CSRobTitle {
		if got {
			// pt.firstTitleOn(gamedata.CSRobTitleWeekReward)
			if pt.CSRobTitleTime < effectTime {
				pt.TitleForClient[gamedata.CSRobTitleWeekReward] = struct{}{}
				p.Profile.GetData().SetNeedCheckMaxGS() // gs变化 10) 称号
			}
		} else {
			pt.CSRobTitle = false
			delete(pt.TitleForClient, gamedata.CSRobTitleWeekReward)
			if gamedata.CSRobTitleWeekReward == pt.TitleTakeOn {
				pt.SetTitleOn("")
			}
			p.Profile.GetData().SetNeedCheckMaxGS() // gs变化 10) 称号
		}
		pt.CSRobTitleTime = effectTime + 1
	} else {
		if got {
			pt.CSRobTitle = true
			pt.firstTitleOn(gamedata.CSRobTitleWeekReward)
			pt.TitleForClient[gamedata.CSRobTitleWeekReward] = struct{}{}
			p.Profile.GetData().SetNeedCheckMaxGS() // gs变化 10) 称号
		} else {
			delete(pt.TitleForClient, gamedata.CSRobTitleWeekReward)
		}
		pt.CSRobTitleTime = effectTime + 1
	}
}

func (pt *PlayerTitle) GetTitles() []string {
	res := make([]string, 0, len(pt.TitleHadActivate)+2)
	for t, _ := range pt.TitleHadActivate {
		res = append(res, t)
	}
	if pt.TitleSimplePvp != "" {
		res = append(res, pt.TitleSimplePvp)
	}
	if pt.TitleTeamPvp != "" {
		res = append(res, pt.TitleTeamPvp)
	}
	if pt.TitleWuShuang != "" {
		res = append(res, pt.TitleWuShuang)
	}
	if pt.TitleGVG != "" {
		res = append(res, pt.TitleGVG)
	}
	if true == pt.CSRobTitle {
		res = append(res, gamedata.CSRobTitleWeekReward)
	}
	return res
}

func (pt *PlayerTitle) IsCanActivate(title string) bool {
	_, ok := pt.TitleCanActivate[title]
	return ok
}

func (pt *PlayerTitle) ActivateTitle(title string) bool {
	if _, ok := pt.TitleHadActivate[title]; ok {
		return false
	}
	pt.TitleHadActivate[title] = struct{}{}
	delete(pt.TitleCanActivate, title)
	pt.firstTitleOn(title)
	return true
}

func (pt *PlayerTitle) GetNextRefTime() int64 {
	return int64(math.Min(
		float64(pt.NextSimplePvpRefreshTime),
		float64(pt.NextTeamPvpRefreshTime)))
}

func (pt *PlayerTitle) updateSimplePvpTitle(p *Account, now_time int64) (titleOnChg bool, oldTitleOn string) {
	if now_time < pt.NextSimplePvpRefreshTime {
		return
	}
	pt.NextSimplePvpRefreshTime = util.DailyBeginUnixByStartTime(now_time,
		gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypPVPBalance))
	if now_time >= pt.NextSimplePvpRefreshTime {
		pt.NextSimplePvpRefreshTime += util.DaySec
	}
	pt.NextSimplePvpRefreshTime += title_update_time_offset
	// 检查排行榜称号是否还有效
	old := pt.TitleSimplePvp
	delete(pt.TitleForClient, old)
	_chg := pt.TitleTakeOn == pt.TitleSimplePvp
	pt.TitleSimplePvp = ""
	rank := title_rank.GetModule(p.AccountID.ShardId).GetSimplePvpRank(p.AccountID.String())
	if rank > 0 {
		cRank := gamedata.TitleSimplePvpRank(rank)
		if cRank != nil {
			pt.TitleSimplePvp = cRank.GetTitleID()
			pt.TitleForClient[pt.TitleSimplePvp] = struct{}{}
			pt.firstTitleOn(pt.TitleSimplePvp)
		}
	}
	if _chg && pt.TitleTakeOn != pt.TitleSimplePvp {
		titleOnChg = true
		oldTitleOn = pt.TitleTakeOn
		pt.SetTitleOn("")
	}
	if old != pt.TitleSimplePvp {
		p.Profile.GetData().SetNeedCheckMaxGS() // gs变化 10) 称号
		logiclog.LogTitleChange(
			p.AccountID.String(),
			p.Profile.GetCurrAvatar(),
			p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId,
			old,
			pt.TitleSimplePvp,
			0,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
			"")
	}

	return
}

func (pt *PlayerTitle) updateTeamPvpTitle(p *Account, now_time int64) (titleOnChg bool, oldTitleOn string) {
	if now_time < pt.NextTeamPvpRefreshTime {
		return
	}
	pt.NextTeamPvpRefreshTime = util.DailyBeginUnixByStartTime(now_time,
		gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypTeamPVPBalance))
	if now_time > pt.NextTeamPvpRefreshTime {
		pt.NextTeamPvpRefreshTime += util.DaySec
	}
	pt.NextTeamPvpRefreshTime += title_update_time_offset
	// 检查排行榜称号是否还有效
	old := pt.TitleTeamPvp
	delete(pt.TitleForClient, old)
	_chg := pt.TitleTakeOn == pt.TitleTeamPvp
	pt.TitleTeamPvp = ""
	rank := title_rank.GetModule(p.AccountID.ShardId).GetTeamPvpRank(p.AccountID.String())
	if rank > 0 {
		cRank := gamedata.TitleTeamPvpRank(rank)
		if cRank != nil {
			pt.TitleTeamPvp = cRank.GetTitleID()
			pt.TitleForClient[pt.TitleTeamPvp] = struct{}{}
			pt.firstTitleOn(pt.TitleTeamPvp)
		}
	}
	if _chg && pt.TitleTakeOn != pt.TitleTeamPvp {
		titleOnChg = true
		oldTitleOn = pt.TitleTakeOn
		pt.SetTitleOn("")
	}
	if old != pt.TitleTeamPvp {
		p.Profile.GetData().SetNeedCheckMaxGS() // gs变化 10) 称号
		logiclog.LogTitleChange(
			p.AccountID.String(),
			p.Profile.GetCurrAvatar(),
			p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId,
			old,
			pt.TitleTeamPvp,
			0,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
			"")
	}
	return
}

func (pt *PlayerTitle) GetNextResetGVGTitleTime(now_t int64, gi gamedata.GVGInfo2Client) int64 {
	if now_t < gi.GVGResetTime {
		return gi.GVGResetTime
	} else if now_t >= gi.GVGResetTime && now_t <= gi.GVGBalanceEndTime {
		return gi.GVGBalanceEndTime
	} else {
		return 0
	}
}

func (pt *PlayerTitle) updateGVGTitle(p *Account, now_time int64) (titleOnChg bool, oldTitleOn string) {
	// 依据军团战时间线,而不依据自身
	now_t := gvg.GetModule(p.AccountID.ShardId).GetNowTime()
	//if now_t < pt.NextGVGRefreshTime {
	//	return
	//}
	gvgInfo, _ := gamedata.GetHotDatas().GvgConfig.GetGVGTime(p.AccountID.ShardId, now_t)
	nextRefreshTime := pt.GetNextResetGVGTitleTime(now_t, gvgInfo)
	logs.Debug("update gvg title, now_time: %d, nextRefreshTime: %d, new Time: %d", now_t, pt.NextGVGRefreshTime,
		nextRefreshTime)
	pt.NextGVGRefreshTime = nextRefreshTime + title_update_time_offset

	// 检查GVG长安城称号变化
	cityLeader := gvg.GetModule(p.AccountID.ShardId).GetLastChangAnLeader()

	gvgTitle := gamedata.TitleGVG()
	if gvgTitle == nil {
		logs.Error("Fatal Error No GVG Title GameData")
		return

	}

	logs.Debug("ChangAn leader is %s, %s", cityLeader, pt.TitleGVG)
	// 之前拥有，现在依然拥有
	if cityLeader == p.AccountID.String() && pt.TitleGVG == gvgTitle.GetTitleID() {
		return
	}

	// 之前未拥有，现在依然未拥有
	if cityLeader != p.AccountID.String() && pt.TitleGVG == "" {
		return
	}
	old := pt.TitleGVG
	delete(pt.TitleForClient, old)
	_chg := pt.TitleTakeOn == pt.TitleGVG
	pt.TitleGVG = ""
	if cityLeader == p.AccountID.String() {
		pt.TitleGVG = gvgTitle.GetTitleID()
		pt.TitleForClient[pt.TitleGVG] = struct{}{}
		pt.firstTitleOn(pt.TitleGVG)
	}

	if _chg && pt.TitleTakeOn != pt.TitleGVG {
		titleOnChg = true
		oldTitleOn = pt.TitleTakeOn
		pt.SetTitleOn("")
	}
	if old != pt.TitleGVG {
		p.Profile.GetData().SetNeedCheckMaxGS() // gs变化 10) 称号
		logiclog.LogTitleChange(
			p.AccountID.String(),
			p.Profile.GetCurrAvatar(),
			p.Profile.GetCorp().GetLvlInfo(),
			p.Profile.ChannelId,
			old,
			pt.TitleGVG,
			0,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
			"")
	}

	return
}

func (pt *PlayerTitle) updateWuShuangTitle(p *Account, now_time int64) (titleOnChg bool, oldTitleOn string) {
	if now_time < pt.NextWuShuangRefreshTime {
		return
	}
	pt.NextWuShuangRefreshTime = util.DailyBeginUnixByStartTime(now_time,
		gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypeWspvpRefresh))
	if now_time > pt.NextWuShuangRefreshTime {
		pt.NextWuShuangRefreshTime += util.DaySec
	}
	pt.NextWuShuangRefreshTime += title_update_time_offset
	// 检查排行榜称号是否还有效
	old := pt.TitleWuShuang
	delete(pt.TitleForClient, old)
	_chg := pt.TitleTakeOn == pt.TitleWuShuang
	pt.TitleWuShuang = ""

	rank := title_rank.GetModule(p.AccountID.ShardId).GetWuShuangRank(p.AccountID.String())

	if rank > 0 {
		cRank := gamedata.TitleWushuangRank(rank)
		if cRank != nil {
			pt.TitleWuShuang = cRank.GetTitleID()
			pt.TitleForClient[pt.TitleWuShuang] = struct{}{}
			pt.firstTitleOn(pt.TitleWuShuang)
		}
	}
	if _chg && pt.TitleTakeOn != pt.TitleWuShuang {
		titleOnChg = true
		oldTitleOn = pt.TitleTakeOn
		pt.SetTitleOn("")
	}
	if old != pt.TitleWuShuang {
		p.Profile.GetData().SetNeedCheckMaxGS() // gs变化 10) 称号
	}

	logiclog.LogTitleChange(
		p.AccountID.String(),
		p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId,
		old,
		pt.TitleWuShuang,
		0,
		func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) },
		"")
	return
}

func (pt *PlayerTitle) Update7DayTitle(p *Account) {
	rank := title_rank.GetModule(p.AccountID.ShardId).Get7DayGsRank(p.AccountID.String())
	if rank > 0 {
		title := gamedata.Title7DayGsRank(rank).GetTitleID()
		if pt.ActivateTitle(title) {
			pt.TitleForClient[title] = struct{}{}
		}
	}
}

func (pt *PlayerTitle) SetTitleOn(title string) {
	pt.TitleTakeOn = title
}

func (pt *PlayerTitle) GetTitleOnShowForOther(acc *Account) string {
	now_t := time.Now().Unix()
	if pt.TitleTakeOn == pt.TitleSimplePvp {
		if now_t > pt.NextSimplePvpRefreshTime {
			return ""
		}
	}
	if pt.TitleTakeOn == pt.TitleTeamPvp {
		if now_t > pt.NextTeamPvpRefreshTime {
			return ""
		}
	}
	if pt.TitleTakeOn == pt.TitleGVG {
		if now_t > pt.NextGVGRefreshTime {
			return ""
		}
	}
	if pt.TitleTakeOn == pt.TitleWuShuang {
		if now_t > pt.NextWuShuangRefreshTime {
			return ""
		}
	}
	if pt.TitleTakeOn == gamedata.CSRobTitleWeekReward {
		if got, _ := pt.checkCSRobTitle(acc); false == got {
			return ""
		}
	}
	return pt.TitleTakeOn
}

func (pt *PlayerTitle) GetTitlesForOther(acc *Account) []string {
	now_t := time.Now().Unix()
	res := make([]string, 0, len(pt.TitleHadActivate)+2)
	for t, _ := range pt.TitleHadActivate {
		res = append(res, t)
	}
	if pt.TitleSimplePvp != "" {
		if now_t > pt.NextSimplePvpRefreshTime {
			rank := title_rank.GetModule(acc.AccountID.ShardId).
				GetSimplePvpRank(acc.AccountID.String())
			if rank > 0 {
				cRank := gamedata.TitleSimplePvpRank(rank)
				if cRank != nil {
					res = append(res, cRank.GetTitleID())
				}
			}
		} else {
			res = append(res, pt.TitleSimplePvp)
		}
	}
	if pt.TitleTeamPvp != "" {
		if now_t > pt.NextTeamPvpRefreshTime {
			rank := title_rank.GetModule(acc.AccountID.ShardId).
				GetTeamPvpRank(acc.AccountID.String())
			if rank > 0 {
				cRank := gamedata.TitleTeamPvpRank(rank)
				if cRank != nil {
					res = append(res, cRank.GetTitleID())
				}
			}
		} else {
			res = append(res, pt.TitleTeamPvp)
		}
	}
	if pt.TitleWuShuang != "" {
		if now_t > pt.NextWuShuangRefreshTime {
			rank := title_rank.GetModule(acc.AccountID.ShardId).
				GetWuShuangRank(acc.AccountID.String())
			if rank > 0 {
				cRank := gamedata.TitleWushuangRank(rank)
				if cRank != nil {
					res = append(res, cRank.GetTitleID())
				}
			}
		} else {
			res = append(res, pt.TitleWuShuang)
		}
	}
	if pt.TitleGVG != "" {
		cityLeader := gvg.GetModule(acc.AccountID.ShardId).GetLastChangAnLeader()
		if cityLeader == acc.AccountID.String() {
			res = append(res, pt.TitleGVG)
		}
	}
	if true == pt.CSRobTitle {
		if got, _ := pt.checkCSRobTitle(acc); true == got {
			res = append(res, gamedata.CSRobTitleWeekReward)
		}
	}
	return res
}

func (pt *PlayerTitle) checkCSRobTitle(acc *Account) (bool, int64) {
	return csrob.GetModule(acc.AccountID.ShardId).Ranker.CheckMeHasTitle(acc.AccountID.String())
}

func (pt *PlayerTitle) OnVip(p *Account) {
	pt.onCond(p, COND_TYP_VIP)
}

func (pt *PlayerTitle) OnGs(p *Account) {
	pt.onCond(p, COND_TYP_Max_Avatar_GS)
}

func (pt *PlayerTitle) OnEatBaozi(p *Account) {
	pt.onCond(p, COND_TYP_EatBaoziCount)
}

func (pt *PlayerTitle) OnFestivalBoss(p *Account) {
	pt.onCond(p, COND_TYP_GUILD_FESTIVALBOSS)

}

func (pt *PlayerTitle) OnExchangeShop(p *Account) {
	typ := COND_TYP_ExchangeShop
	for _, cfg := range gamedata.GetTitleCond(typ) {
		_, ok := pt.TitleCanActivate[cfg.GetTitleID()]
		_, ok2 := pt.TitleHadActivate[cfg.GetTitleID()]
		if !ok && !ok2 {
			pt.TitleCanActivate[cfg.GetTitleID()] = struct{}{}
			pt.TitleForClient[cfg.GetTitleID()] = struct{}{}
			player_msg.Send(p.AccountID.String(), player_msg.PlayerMsgTitleCode,
				player_msg.DefaultMsg{})
			logs.Debug("PlayerTitle cond_type activate %s %d %s",
				p.AccountID.String(), typ, cfg.GetTitleID())
		}
	}
}

func (pt *PlayerTitle) OnOneWorld(p *Account) {
	pt.onCond(p, COND_TYP_GVG_ONEWORLD)
	// 自动激活
	for _, cfg := range gamedata.GetTitleCond(COND_TYP_GVG_ONEWORLD) {
		titleID := cfg.GetTitleID()
		if pt.IsCanActivate(titleID) {
			pt.ActivateTitle(titleID)
		}
	}
}

func (pt *PlayerTitle) onCond(p *Account, cond_typ int) {
	for _, cfg := range gamedata.GetTitleCond(cond_typ) {
		_, ok := pt.TitleCanActivate[cfg.GetTitleID()]
		_, ok2 := pt.TitleHadActivate[cfg.GetTitleID()]
		if !ok && !ok2 {
			if CheckCondition(p, cfg.GetFCType(), int64(cfg.GetFCValueIP1()),
				int64(cfg.GetFCValueIP2()), "", "") {
				pt.TitleCanActivate[cfg.GetTitleID()] = struct{}{}
				pt.TitleForClient[cfg.GetTitleID()] = struct{}{}
				player_msg.Send(p.AccountID.String(), player_msg.PlayerMsgTitleCode,
					player_msg.DefaultMsg{})
				logs.Debug("PlayerTitle cond_type activate %s %d %s",
					p.AccountID.String(), cond_typ, cfg.GetTitleID())
			}
		}
	}
}

func (pt *PlayerTitle) firstTitleOn(title string) {
	if !pt.FirstTitle {
		pt.FirstTitle = true
		pt.SetTitleOn(title)
	}
}

func (pt *PlayerTitle) ToDB() PlayerTitleInDB {
	db := PlayerTitleInDB{
		TitleSimplePvp:           pt.TitleSimplePvp,
		NextSimplePvpRefreshTime: pt.NextSimplePvpRefreshTime,
		TitleTeamPvp:             pt.TitleTeamPvp,
		NextTeamPvpRefreshTime:   pt.NextTeamPvpRefreshTime,
		TitleWuShuagn:            pt.TitleWuShuang,
		NextWushuangRefreshTime:  pt.NextWuShuangRefreshTime,
		TitleGVG:                 pt.TitleGVG,
		NextGVGRefreshTime:       pt.NextGVGRefreshTime,
		TitleTakeOn:              pt.TitleTakeOn,
		FirstTitle:               pt.FirstTitle,
	}
	db.TitleCanActivate = make([]string, 0, len(pt.TitleCanActivate))
	for t, _ := range pt.TitleCanActivate {
		db.TitleCanActivate = append(db.TitleCanActivate, t)
	}
	sort.Strings(db.TitleCanActivate)
	db.TitleHadActivate = make([]string, 0, len(pt.TitleHadActivate))
	for t, _ := range pt.TitleHadActivate {
		db.TitleHadActivate = append(db.TitleHadActivate, t)
	}
	sort.Strings(db.TitleHadActivate)
	db.TitleForClient = make([]string, 0, len(pt.TitleForClient))
	for t, _ := range pt.TitleForClient {
		db.TitleForClient = append(db.TitleForClient, t)
	}
	db.CSRobTitleTime = pt.CSRobTitleTime
	db.CSRobTitle = pt.CSRobTitle
	sort.Strings(db.TitleForClient)
	return db
}

func (pt *PlayerTitle) FromDB(data *PlayerTitleInDB) error {
	pt.TitleCanActivate = make(map[string]struct{}, len(data.TitleCanActivate))
	for _, t := range data.TitleCanActivate {
		pt.TitleCanActivate[t] = struct{}{}
	}
	pt.TitleHadActivate = make(map[string]struct{}, len(data.TitleHadActivate))
	for _, t := range data.TitleHadActivate {
		pt.TitleHadActivate[t] = struct{}{}
	}
	pt.TitleForClient = make(map[string]struct{}, len(data.TitleForClient))
	for _, t := range data.TitleForClient {
		pt.TitleForClient[t] = struct{}{}
	}
	pt.TitleSimplePvp = data.TitleSimplePvp
	pt.NextSimplePvpRefreshTime = data.NextSimplePvpRefreshTime
	pt.TitleTeamPvp = data.TitleTeamPvp
	pt.NextTeamPvpRefreshTime = data.NextTeamPvpRefreshTime
	pt.TitleWuShuang = data.TitleWuShuagn
	pt.NextWuShuangRefreshTime = data.NextWushuangRefreshTime
	pt.TitleGVG = data.TitleGVG
	pt.NextGVGRefreshTime = data.NextGVGRefreshTime
	pt.TitleTakeOn = data.TitleTakeOn
	pt.FirstTitle = data.FirstTitle
	pt.CSRobTitleTime = data.CSRobTitleTime
	pt.CSRobTitle = data.CSRobTitle
	return nil
}
