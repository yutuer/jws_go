package market_activity

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"sync"

	"vcs.taiyouxi.net/jws/gamex/models/account/simple_info"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/market_activity"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/jws/gamex/modules/rank"
	"vcs.taiyouxi.net/jws/gamex/modules/title_rank"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

func (ma *MarketActivityModule) notifyMakeSnapShoot(activity uint32, activityID uint32) {
	logs.Debug("[MarketActivityModule] notifyMakeSnapShoot")
	cmd := makeMarketCommand(Command_MakeSnapShoot, marketCommandParam{ActivityType: activity, ActivityID: activityID})
	ma.commandExecAsyn(cmd)
}
func (ma *MarketActivityModule) cmdMakeSnapShoot(cmd *marketCommand) {
	logs.Debug("[MarketActivityModule] cmdMakeSnapShoot execute")
	ma.maRank.makeSnapShoot(cmd.ActivityType, cmd.ActivityID)
}

func (ma *MarketActivityModule) notifySendReward(activityType uint32, activityID uint32) {
	logs.Debug("[MarketActivityModule] notifySendReward")
	cmd := makeMarketCommand(Command_SendReward, marketCommandParam{ActivityType: activityType, ActivityID: activityID})
	ma.commandExecAsyn(cmd)
}
func (ma *MarketActivityModule) cmdSendReward(cmd *marketCommand) {
	logs.Debug("[MarketActivityModule] cmdSendReward execute")
	ma.maRank.sendReward(cmd.ActivityType, cmd.ActivityID)
}

func (ma *MarketActivityModule) notifyRankParentID(activityType uint32, activityID uint32) {
	logs.Debug("[MarketActivityModule] notifyRankParentID")
	cmd := makeMarketCommand(Command_RankParentID, marketCommandParam{ActivityType: activityType, ActivityID: activityID})

	//因为notify的来源也是模块的线程消息队列,所以不发消息直接处理
	//ma.commandExecAsyn(cmd)
	ma.cmdRankParentID(cmd)
}
func (ma *MarketActivityModule) cmdRankParentID(cmd *marketCommand) {
	logs.Debug("[MarketActivityModule] cmdRankParentID execute")
	ma.maRank.refreshRankParentID(cmd.ActivityType, cmd.ActivityID)
}

type MarketRank struct {
	RankTopN     map[uint32]*RankTopN
	RankTopNLock sync.RWMutex

	record MarketRankRecord

	ma *MarketActivityModule

	sid uint
	db  *MarketRankDB
}

func (mr *MarketRank) Init(ma *MarketActivityModule) {
	mr.ma = ma
	mr.sid = ma.sid
	mr.db = &MarketRankDB{
		sid: ma.sid,
	}
	mr.record.RankBatch = make(map[string]uint32)
	mr.record.RewardRecord = make(map[string]int64)

	mr.RankTopN = make(map[uint32]*RankTopN)
}

type MarketRankRecord struct {
	RankBatch    map[string]uint32 `json:"rank_batch,omitempty"`
	RankParentID uint32            `json:"parentID,omitempty"`
	RewardRecord map[string]int64  `json:"reward,omitempty"`
}

func (mr *MarketRankRecord) checkAlreadyReward(actID uint32) bool {
	if nil == mr.RewardRecord || 0 == mr.RewardRecord[fmt.Sprint(actID)] {
		return false
	}

	return true
}

func (mr *MarketRankRecord) setRewardRecord(actID uint32) {
	if nil == mr.RewardRecord {
		mr.RewardRecord = map[string]int64{}
	}
	mr.RewardRecord[fmt.Sprint(actID)] = time.Now().Unix()
}

type RankTopN struct {
	TopN []rank.CorpDataInRank `json:"topN"`
	//MinScoreToTopN int64                       `json:"min"`
}

func (ma *MarketActivityModule) GetRank(activity uint32, acid string, info *helper.AccountSimpleInfo) *rank.RankByCorpGetRes {
	snapShoot := ma.maRank.checkHadSnapShoot(activity)
	var res *rank.RankByCorpGetRes

	logs.Debug("[MarketActivityModule] GetRank activity [%d], snapShoot [%v]", activity, snapShoot)
	// 判断活动时间, 是否取动态数据
	if snapShoot {
		res = ma.maRank.getRankSnapShoot(activity, acid)
	} else {
		res = ma.maRank.getRankDynamic(activity, acid, info)
	}

	return res
}

func (mr *MarketRank) getRankDynamic(activity uint32, acid string, info *helper.AccountSimpleInfo) *rank.RankByCorpGetRes {
	var res *rank.RankByCorpGetRes

	logs.Debug("[MarketActivityModule] getRankDynamic activity [%d]", activity)
	switch activity {
	case Activity_RankPlayerGs:
		res = rank.GetModule(mr.sid).RankCorpGs.Get(acid)
		res.Score = rank.GetModule(mr.sid).RankCorpGs.GetCorpInfo(info).Score
	case Activity_RankHeroStar:
		res = rank.GetModule(mr.sid).RankByHeroStar.Get(acid)
		res.Score = rank.GetModule(mr.sid).RankByHeroStar.GetCorpInfo(info).Score
	case Activity_RankDestiny:
		res = rank.GetModule(mr.sid).RankByDestiny.Get(acid)
		res.Score = rank.GetModule(mr.sid).RankByDestiny.GetCorpInfo(info).Score
	case Activity_RankStone:
		res = rank.GetModule(mr.sid).RankByJade.Get(acid)
		res.Score = rank.GetModule(mr.sid).RankByJade.GetCorpInfo(info).Score
	case Activity_RankPlayerLevel:
		res = rank.GetModule(mr.sid).RankByCorpLv.Get(acid)
		res.Score = rank.GetModule(mr.sid).RankByCorpLv.GetCorpInfo(info).Score
	case Activity_RankArmStar:
		res = rank.GetModule(mr.sid).RankByEquipStarLv.Get(acid)
		res.Score = rank.GetModule(mr.sid).RankByEquipStarLv.GetCorpInfo(info).Score
	case Activity_RankHeroDestiny:
		res = rank.GetModule(mr.sid).RankByHeroDestiny.Get(acid)
		res.Score = rank.GetModule(mr.sid).RankByHeroDestiny.GetCorpInfo(info).Score
	case Activity_RankHeroSwingStarLv:
		res = rank.GetModule(mr.sid).RankByWingStar.Get(acid)
		res.Score = rank.GetModule(mr.sid).RankByWingStar.GetCorpInfo(info).Score
	case Activity_RankHeroJadeTwo:
		res = rank.GetModule(mr.sid).RankByHeroJadeTwo.Get(acid)
		res.Score = rank.GetModule(mr.sid).RankByHeroJadeTwo.GetCorpInfo(info).Score
	case Activity_RankWuShuangGs:
		res = rank.GetModule(mr.sid).RankByHeroWuShuangGs.Get(acid)
		res.Score = rank.GetModule(mr.sid).RankByHeroWuShuangGs.GetCorpInfo(info).Score
	case Activity_RankExclusiveWeapon:
		res = rank.GetModule(mr.sid).RankByExclusiveWeapon.Get(acid)
		res.Score = rank.GetModule(mr.sid).RankByExclusiveWeapon.GetCorpInfo(info).Score
	case Activity_RankAstrology:
		res = rank.GetModule(mr.sid).RankByAstrology.Get(acid)
		res.Score = rank.GetModule(mr.sid).RankByAstrology.GetCorpInfo(info).Score
	default:
		logs.Warn("[MarketActivityModule] getRankDynamic unkown activity [%d]", activity)
		res = &rank.RankByCorpGetRes{}
	}
	return res
}

func (mr *MarketRank) getRankSnapShoot(actType uint32, acid string) *rank.RankByCorpGetRes {
	res := &rank.RankByCorpGetRes{}

	logs.Debug("[MarketActivityModule] getRankSnapShoot activity [%d]", actType)

	mr.RankTopNLock.RLock()
	topN := mr.RankTopN[actType]
	if nil == topN {
		logs.Warn("[MarketActivityModule] getRankSnapShoot unSnapShoot activity [%d]", actType)
		res.TopN = []rank.CorpDataInRank{}
	} else {
		res.TopN = topN.TopN[:]
	}
	mr.RankTopNLock.RUnlock()

	res.Pos, res.Score = mr.getPosAndScoreByAcid(actType, acid)
	logs.Debug("[MarketActivityModule] getRankSnapShoot activity [%d], Res: TopN [%d], Pos [%d], Score [%d] ", actType, len(res.TopN), res.Pos, res.Score)

	return res
}

func (mr *MarketRank) makeSnapShoot(activity uint32, actID uint32) {
	if true == mr.checkHadSnapShoot(activity) || true == mr.record.checkAlreadyReward(actID) {
		logs.Debug("[MarketActivityModule] makeSnapShoot ignore, activity [%d]", activity)
		return
	}

	list := mr.getRankContent(activity)
	if nil == list {
		logs.Error("[MarketActivityModule] makeSnapShoot getRankContent nil, activity [%d:%d]", activity, actID)
		return
	}
	logs.Debug("[MarketActivityModule] makeSnapShoot getRankContent, activity [%d], list length [%d]", activity, len(list))

	//排行数据进行快照
	err := mr.db.setSnapShoot(activity, list)
	if nil != err {
		logs.Error("[MarketActivityModule] Set SnapShoot Failed, %v", err)
		return
	}
	logs.Debug("[MarketActivityModule] makeSnapShoot setSnapShoot over, activity [%d]", activity)

	//读取TopN的玩家摘要信息并寄存
	sortlist := sortFromAcid2score(list)
	mr.makeTopN(activity, actID, sortlist)
	logs.Debug("[MarketActivityModule] makeSnapShoot makeTopN over, activity [%d]", activity)
}

func (mr *MarketRank) getPosAndScoreByAcid(activity uint32, acid string) (int, int64) {
	rank, redisScore := mr.db.getPosAndRedisScore(activity, acid)
	logs.Debug("[MarketActivityModule] getPosAndScoreByAcid getPosAndRedisScore over, activity [%d], pos [%d], redisScore [%f]", activity, rank, redisScore)

	return rank, mr.redisScoreToScore(activity, redisScore)
}

func (mr *MarketRank) makeTopN(actType uint32, actID uint32, list []pair) {
	topN := &RankTopN{
		TopN: make([]rank.CorpDataInRank, 0),
	}

	n := RankTopSize
	if n > len(list) {
		n = len(list)
	}
	for i := 0; i < n; i++ {
		pair := list[i]
		ac, err := db.ParseAccount(pair.Acid)
		if nil != err {
			logs.Error("[MarketActivityModule] makeTopN ParseAccount Failed %v", err)
			continue
		}
		info, err := simple_info.LoadAccountSimpleInfoProfile(ac)
		if nil != err {
			logs.Error("[MarketActivityModule] makeTopN LoadAccountSimpleInfoProfile Failed %v", err)
			continue
		}

		data := &rank.CorpDataInRank{}
		data.SetDataFromAccount(info, mr.redisScoreToScore(actType, pair.Score))

		topN.TopN = append(topN.TopN, *data)
	}
	logs.Debug("[MarketActivityModule] makeTopN collect data over, activity [%d]", actType)

	err := mr.db.setTopN(actType, actID, topN)
	if nil != err {
		logs.Error("[MarketActivityModule] makeTopN setTopN Failed, %v", err)
	}
	logs.Debug("[MarketActivityModule] makeTopN setTopN over, activity [%d]", actType)

	mr.setTopNSafe(actType, topN)
	logs.Debug("[MarketActivityModule] makeTopN setTopNSafe over, activity [%d]", actType)
}

func (mr *MarketRank) getRankContent(activity uint32) map[string]float64 {
	switch activity {
	case Activity_RankPlayerGs:
		return rank.GetModule(mr.sid).GetRankContent(rank.Ex_RankId_RankByCorpGS)
	case Activity_RankHeroStar:
		return rank.GetModule(mr.sid).GetRankContent(rank.Ex_RankId_RankByHeroStar)
	case Activity_RankStone:
		return rank.GetModule(mr.sid).GetRankContent(rank.Ex_RankId_RankByJade)
	case Activity_RankArmStar:
		return rank.GetModule(mr.sid).GetRankContent(rank.Ex_RankId_RankByEquipStarLv)
	case Activity_RankPlayerLevel:
		return rank.GetModule(mr.sid).GetRankContent(rank.Ex_RankId_RankByCorpLv)
	case Activity_RankDestiny:
		return rank.GetModule(mr.sid).GetRankContent(rank.Ex_RankId_RankByDestiny)
	case Activity_RankHeroDestiny:
		return rank.GetModule(mr.sid).GetRankContent(rank.Ex_RankId_HeroDestinyLv)
	case Activity_RankHeroSwingStarLv:
		return rank.GetModule(mr.sid).GetRankContent(rank.Ex_RankId_RankBySwingStarLv)
	case Activity_RankHeroJadeTwo:
		return rank.GetModule(mr.sid).GetRankContent(rank.Ex_RankId_HeroByJadeTwo)
	case Activity_RankWuShuangGs:
		return rank.GetModule(mr.sid).GetRankContent(rank.Ex_RankId_HeroByWuShuangGs)
	case Activity_RankExclusiveWeapon:
		return rank.GetModule(mr.sid).GetRankContent(rank.Ex_RankId_ExclusiveWeapon)
	case Activity_RankAstrology:
		return rank.GetModule(mr.sid).GetRankContent(rank.Ex_RankId_Astrology)
	}

	return nil
}

func (mr *MarketRank) setTopNSafe(actType uint32, data *RankTopN) {
	topN := &RankTopN{
		TopN: data.TopN,
	}

	mr.RankTopNLock.Lock()
	defer mr.RankTopNLock.Unlock()
	mr.RankTopN[actType] = topN
}

func (mr *MarketRank) checkHadSnapShoot(actType uint32) bool {
	mr.RankTopNLock.RLock()
	defer mr.RankTopNLock.RUnlock()
	return nil != mr.RankTopN[actType]
}

func (mr *MarketRank) redisScoreToScore(activity uint32, rs float64) int64 {
	switch activity {
	case Activity_RankPlayerGs:
		return rank.GetModule(mr.sid).RankCorpGs.RebaseScore(int64(rs))
	case Activity_RankHeroStar:
		return rank.GetModule(mr.sid).RankByHeroStar.RebaseScore(rs)
	case Activity_RankStone:
		return rank.GetModule(mr.sid).RankByJade.RebaseScore(rs)
	case Activity_RankDestiny:
		return rank.GetModule(mr.sid).RankByDestiny.RebaseScore(rs)
	case Activity_RankPlayerLevel:
		return rank.GetModule(mr.sid).RankByCorpLv.RebaseScore(rs)
	case Activity_RankArmStar:
		return rank.GetModule(mr.sid).RankByEquipStarLv.RebaseScore(rs)
	case Activity_RankHeroDestiny:
		return rank.GetModule(mr.sid).RankByHeroDestiny.RebaseScore(rs)
	case Activity_RankHeroSwingStarLv:
		return rank.GetModule(mr.sid).RankByWingStar.RebaseScore(rs)
	case Activity_RankHeroJadeTwo:
		return rank.GetModule(mr.sid).RankByHeroJadeTwo.RebaseScore(rs)
	case Activity_RankWuShuangGs:
		return rank.GetModule(mr.sid).RankByHeroWuShuangGs.RebaseScore(rs)
	case Activity_RankAstrology:
		return rank.GetModule(mr.sid).RankByAstrology.RebaseScore(rs)
	case Activity_RankExclusiveWeapon:
		return rank.GetModule(mr.sid).RankByExclusiveWeapon.RebaseScore(rs)
	}
	return 0
}

func (mr *MarketRank) ReloadAll() {
	//load record
	record := mr.db.getRankRecord()
	if nil == record {
		return
	}
	mr.record = *record

	activityCfg := gamedata.GetHotDatas().Activity
	now_t := time.Now().Unix()

	for pType, sTypes := range ActivityList {
		pCfgs := activityCfg.GetActivityInfoFilterTime(pType, now_t)
		if 0 == len(pCfgs) {
			continue
		}
		pID := pCfgs[0].ActivityId
		strType := fmt.Sprintf("%d", pType)
		if pID != mr.record.RankBatch[strType] {
			//新的值等待refresh来刷新
			logs.Debug("[MarketActivityModule] ReloadAll ignore load, type:%d parentID [%d]", pType, mr.record.RankBatch[strType])
			continue
		}

		for _, t := range sTypes {
			subCfgs := activityCfg.GetActivitySimpleInfo(t)
			if 0 == len(subCfgs) || subCfgs[0].ActivityParentID != pID {
				continue
			}
			mr.reload(t, subCfgs[0].ActivityId)
		}
	}
}

func (mr *MarketRank) reload(actType, actID uint32) {
	//取快照TopN
	data, err := mr.db.getTopN(actType, actID)
	if nil != err {
		logs.Error("[MarketActivityModule] reload getTopN Failed %v", err)
		return
	}
	logs.Debug("[MarketActivityModule] reload getTopN over, activity [%d]", actType)

	if nil != data {
		mr.setTopNSafe(actType, data)
		logs.Debug("[MarketActivityModule] reload setTopNSafe over, activity [%d]", actType)
	}
}

func (mr *MarketRank) clear(actType uint32) {
	mr.RankTopNLock.Lock()
	delete(mr.RankTopN, actType)
	mr.RankTopNLock.Unlock()

	logs.Debug("[MarketActivityModule] clear over, activity [%d]", actType)
}

type rewardParamCond struct {
	rankTop     uint32
	rankBottom  uint32
	scoreBottom uint32
	reward      map[string]uint32
}
type rewardParam struct {
	mailIdx int
	count   uint32
	conds   []rewardParamCond
}

func (mr *MarketRank) sendReward(activityType uint32, activityID uint32) {
	param := rewardParam{}
	logs.Debug("[MarketActivityModule] sendReward, activity [%d:%d]", activityType, activityID)
	if !mr.isActivityValid(activityID) {
		return
	}
	//检查是否发过奖
	if true == mr.record.checkAlreadyReward(activityID) {
		logs.Warn("[MarketActivityModule] sendReward already set record, activity %d, %v", activityID)
		return
	}
	//记录已发奖
	mr.record.setRewardRecord(activityID)
	err := mr.db.setRankRecord(&mr.record)
	if nil != err {
		logs.Error("[MarketActivityModule] sendReward set record failed, activity %d, %v", activityID, err)
		return
	}

	//准备奖励的相关参数
	activityCfg := gamedata.GetHotDatas().Activity
	subCfg := activityCfg.GetMarketActivitySubConfig(activityID)
	if subCfg == nil {
		return
	}
	for _, cfg := range subCfg {
		score, err := strconv.Atoi(cfg.GetSFCValue1())
		if nil != err {
			logs.Error("[MarketActivityModule] sendReward parse reward failed, activity %d, %v", activityID, err)
			continue
		}
		cond := rewardParamCond{
			rankTop:     cfg.GetFCValue1(),
			rankBottom:  cfg.GetFCValue2(),
			scoreBottom: uint32(score),
			reward:      make(map[string]uint32),
		}
		for _, r := range cfg.GetItem_Table() {
			cond.reward[r.GetItemID()] = r.GetItemCount()
		}
		param.conds = append(param.conds, cond)

		if param.count < cfg.GetFCValue2() {
			param.count = cfg.GetFCValue2()
		}
	}

	sort.Sort(param)
	logs.Debug("[MarketActivityModule] sendReward param over, activity [%d-%d], param {%v}", activityType, activityID, param)

	//取相应的排行数据
	list, err := mr.db.getTopRange(activityType, param.count)
	if nil != err {
		logs.Error("[MarketActivityModule] sendReward getTopRange failed, activity %d, %v", activityID, err)
	}
	logs.Debug("[MarketActivityModule] sendReward getTopRange over, activity [%d-%d], list length [%v]", activityType, activityID, len(list))

	//发奖
	sort.Sort(pairlist(list))
	for i_order, m := range list {
		score := mr.redisScoreToScore(activityType, m.Score)
		order := i_order + 1

		logs.Debug("[MarketActivityModule] sendReward check activity %d, accout[%s], player [%d:%d] [%d]", activityType, m.Acid, score, order)
		for condindex, cond := range param.conds {
			if cond.rankBottom >= uint32(order) && int64(cond.scoreBottom) <= score {
				//发这个档次的奖
				logs.Debug("[MarketActivityModule] sendReward sendMail activity %d, accout[%s], sub index [%d]", activityType, m.Acid, condindex)
				err := mail_sender.BatchSendMail2Account(
					m.Acid,
					timail.Mail_Send_By_Market_Activity,
					getRewardMailIds(activityType),
					[]string{
						fmt.Sprintf("%d", score),
						fmt.Sprintf("%d", order),
					},
					cond.reward,
					fmt.Sprintf("MarketActivityModule: sendReward by activity %d", activityType), false)
				if nil != err {
					logs.Error("[MarketActivityModule] sendReward Err, %v", err)
				}
				break
			}
		}
	}
	if activityType == gamedata.ActCorpGsActivityRank {
		// 发称号
		acids := make([]string, 0)
		for _, item := range list {
			acids = append(acids, item.Acid)
		}
		title_rank.GetModule(mr.sid).Set7DayGsRank(acids)
	}
}

func (mr *MarketRank) isActivityValid(activityID uint32) bool {
	act := gamedata.GetHotDatas().Activity.GetActivitySimpleInfoById(activityID)
	if act == nil {
		logs.Error("[MarketActivityModule], activity is invalid [%d]", activityID)
		return false
	}
	pAct := gamedata.GetHotDatas().Activity.GetActivitySimpleInfoById(act.ActivityParentID)
	typ := market_activity.GetHotTypeByActType(pAct.ActivityType)
	if !game.Cfg.GetHotActValidData(mr.sid, typ) {
		logs.Warn("[MarketActivityModule], activity is closed by gm [%d], sid %d", activityID, mr.sid)
		return false
	}
	return true
}

func getRewardMailIds(activity uint32) int {
	switch activity {
	case Activity_RankPlayerGs:
		return mail_sender.IDS_MAIL_CAPACITY_TITLE
	case Activity_RankHeroStar:
		return mail_sender.IDS_MAIL_HEROSTAR_TITLE
	case Activity_RankStone:
		return mail_sender.IDS_MAIL_JADE_TITLE
	case Activity_RankDestiny:
		return mail_sender.IDS_MAIL_DG_TITLE
	case Activity_RankPlayerLevel:
		return mail_sender.IDS_MAIL_GROUPLV_TITLE
	case Activity_RankArmStar:
		return mail_sender.IDS_MAIL_EQUIPSTAR_TITLE
	case Activity_RankAstrology:
		return mail_sender.IDS_MAIL_ASTROLOGY_TITLE
	case Activity_RankExclusiveWeapon:
		return mail_sender.IDS_MAIL_GWC_TITLE
	case Activity_RankHeroSwingStarLv:
		return mail_sender.IDS_MAIL_HWSTAR_TITLE
	case Activity_RankWuShuangGs:
		return mail_sender.IDS_MAIL_WS_CAPACITY_TITLE
	case Activity_RankHeroDestiny:
		return mail_sender.IDS_MAIL_HERODESTINY_TITLE
	case Activity_RankHeroJadeTwo:
		return mail_sender.IDS_MAIL_JADE_TITLE
	}
	return 0
}

func (mr *MarketRank) refreshRankParentID(actType, actID uint32) {
	logs.Debug("[MarketActivityModule] refreshRankParentID %d:%d", actType, actID)

	//记录活动父节点:作为版本判断
	strType := fmt.Sprintf("%d", actType)
	if mr.record.RankBatch[strType] != actID {
		logs.Debug("[MarketActivityModule] refreshRankParentID type:%d changed %d => %d", actType, mr.record.RankBatch[strType], actID)
		mr.record.RankBatch[strType] = actID
		err := mr.db.setRankRecord(&mr.record)
		if nil != err {
			logs.Error("[MarketActivityModule] makeSnapShoot setRankRecord Failed, %v", err)
		}

		logs.Debug("[MarketActivityModule] refreshRankParentID type:%d clear all", actType)
		if list, exist := ActivityList[actType]; exist {
			for _, subType := range list {
				mr.clear(subType)
			}
		}
	}
}
