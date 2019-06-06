package logics

import (
	"encoding/json"

	"fmt"

	"time"

	gamelog "vcs.taiyouxi.net/jws/gamex/logiclog"
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/models/stage_star"
	"vcs.taiyouxi.net/jws/gamex/modules/global_info"
	"vcs.taiyouxi.net/platform/planx/servers"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const MAX_STAR = 7
const MAX_STAR_Count = 3

type RequestPrepareLootForLevelEnemy struct {
	Req
	LevelID   string `codec:"levelid"`
	AvatarIDs []int  `codec:"avatar_id"`
}

type ResponsePrepareLootForLevelEnemy struct {
	SyncResp
	//返回敌兵ID1， byte[]:代表[(掉落ID1,内容)， (掉落ID2,内容)]]
	Result map[string]interface{} `codec:"result"`
	//返回敌兵ID1, int32:代表[(掉落ID1,敌兵最大数量)...]
	ResultN map[string]int `codec:"resultN"`
	ResultS []byte         `codec:"resultS"`
}

type loot struct {
	ID    uint16 `codec:"id"`
	Data  string `codec:"data"`
	Count uint32 `codec:"count"`
}

type lootContent_t struct {
	ID   string `codec:"id"`
	Info []loot `codec:"info"`
}

func (p *Account) PrepareLootForLevelEnemy(r servers.Request) *servers.Response {
	//GetLvlEnmyLootResponse
	/*
		- 生成掉落ID的算法uint16
		- 读取关卡表获取敌人数量N
		- 根据敌人ID获取掉落规则 需要临时处理代码
		- 循环N次，生成掉落和掉落ID
	*/
	const (
		_                 = iota
		CODE_StageID_Err  // 失败:关卡ID不存在
		CODE_LootInfo_Err // 失败:关卡掉落信息存储错误
	)

	const (
		CODE_MIN = 20
	)

	req := &RequestPrepareLootForLevelEnemy{}
	resp := &ResponsePrepareLootForLevelEnemy{}

	initReqRsp(
		"PlayLevel/GetLvlEnmyLootResponse",
		r.RawBytes,
		req, resp, p)

	p.Tmp.Last_Level_Prepare = req.LevelID
	p.Tmp.Last_Level_Prepare_TS = time.Now().Unix()

	code, warncode := p.IsStageCanPlay(req.LevelID, req.AvatarIDs, false)
	if warncode != 0 {
		return rpcWarn(resp, uint32(warncode))
	}
	if code != 0 {
		return rpcError(resp, CODE_MIN+code)
	}

	loot_data_ok := mkLootData(p, req.LevelID, req.AvatarIDs[:], resp)
	if !loot_data_ok {
		logs.Warn("PrepareLootForLevelEnemy mkLootData err %s", req.LevelID)
		return rpcWarn(resp, errCode.ClickTooQuickly)
	}

	reward_data_ok := mkStageRewards(p, req.LevelID, req.AvatarIDs[:], resp)
	if !reward_data_ok {
		return rpcError(resp, CODE_StageID_Err)
	}

	// 若是活动
	stage_data := gamedata.GetStageData(req.LevelID)
	if stage_data.GameModeId > 0 {
		p.Profile.GameMode.EnterGameMode(stage_data.GameModeId)
		resp.OnChangeGameMode(stage_data.GameModeId)
	} else {
		p.Profile.GetStage().SetLastStageId(req.LevelID)
		resp.OnChangeLastStage()
	}

	resp.mkInfo(p)

	now_time := time.Now().Unix()
	p.Tmp.SetLevelEnterTime(now_time)

	// log
	gamelog.LogStage_c(
		p.AccountID.String(),
		p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(),
		p.Profile.ChannelId,
		stage_data.Id,
		"EnterStage", now_time,
		p.Profile.GetData().CorpCurrGS,
		"")
	return rpcSuccess(resp)

}

type RequestDeclareLootForLvlEnmy struct {
	Req
	LootIDs   []uint16 `codec:"lootids"`
	AvatarIDs []int    `codec:"avatar_id"`
	KillNums  []int    `codec:"killn"`
	KillTyps  []string `codec:"killt"`
	IsSuccess bool     `codec:"is_success"`
	Star      int      `codec:"star"`
	Hackjson  string   `codec:"hackjson"`
	LevelId   string   `codec:"levelid"`
	SeqId     string   `codec:"seqid"`
}

type RequestDeclareLootForLvlEnmyV2 struct {
	ReqWithAnticheat
	LootIDs   []uint16 `codec:"lootids"`
	AvatarIDs []int    `codec:"avatar_id"`
	KillNums  []int    `codec:"killn"`
	KillTyps  []string `codec:"killt"`
	IsSuccess bool     `codec:"is_success"`
	Star      int      `codec:"star"`
	LevelId   string   `codec:"levelid"`
	SeqId     string   `codec:"seqid"`
}

type StageReward struct {
	Item_id string `codec:"id"`
	Count   uint32 `codec:"count"`
	Data    string `codec:"data"`
}

type ResponseDeclareLootForLvlEnmy struct {
	RespWithAnticheat
	StageRewards      [][]byte `codec:"rewards"`
	IsSuccess         bool     `codec:"is_success"` // 关卡是成功还是失败，原样返回给客户端
	ScType            int      `codec:"sc_t"`       // TBD 如果确认是发钱的话 可以移除
	ScValue           int64    `codec:"sc_v"`
	AvatarXp          []uint32 `codec:"avatar_xp"`            // xp是发给参战的武将的
	HeroXpAdd         uint32   `codec:"hero_xp_add"`          // 主将经验
	CorpXpAdd         uint32   `codec:"corp_xp_add"`          // 战队经验
	GoldLevelPoint    uint32   `codec:"gold_level_point"`     // 金币关卡积分
	GoldLevelPointAdd uint32   `codec:"gold_level_point_add"` // 金币关卡积分增加
	ExpLevelPoint     uint32   `codec:"exp_level_point"`      // 经验关积分
	ExpLevelPointAdd  uint32   `codec:"exp_level_point_add"`  // 经验关积分增加
	DCLevelPoint      uint32   `codec:"dc_level_point"`       // 天命关积分
	DCLevelPointAdd   uint32   `codec:"dc_level_point_add"`   // 天命关积分增加
}

const NORMAL_REWARD_MAX = 5

func (r *ResponseDeclareLootForLvlEnmy) initData() {
	r.AvatarXp = make([]uint32,
		gamedata.AVATAR_NUM_CURR,
		gamedata.AVATAR_NUM_CURR)
}

func (r *ResponseDeclareLootForLvlEnmy) AddResReward(g *gamedata.CostData2Client) {
	for i := 0; i < g.Len(); i++ {
		ok, it, c, _, ds := g.GetItem(i)
		if ok {
			r.addReward(it, c, &ds)
		}
	}
	r.ScType = gamedata.SC_Money
	r.ScValue = int64(g.AllSc0)
	r.CorpXpAdd = g.AllCorpXp
	r.HeroXpAdd = g.EachHeroXp
}

func (r *ResponseDeclareLootForLvlEnmy) MergeReward() {}

func (r *ResponseDeclareLootForLvlEnmy) addReward(item_id string, count uint32, data *gamedata.BagItemData) {
	// Virtual Item 不发给客户端
	is_goodwill, _, _ := gamedata.IsGeneralGoodwillItem(item_id)
	is_hero, _, _, _ := gamedata.IsHeroPieceItem(item_id)
	is_buffer := gamedata.IsItemToBuffWhenAdd(item_id)
	is_2_sc, itemID, countc := gamedata.IsItemToSCWhenAdd(item_id)
	is_2_c, itemID1, countc1 := gamedata.IsItemTreasurebox(item_id)

	//logs.Warn("StageRewardsss %s %d", item_id, count)

	if is_buffer && !is_goodwill {
		return
	}

	if item_id == gamedata.VI_CorpXP {
		return
	}
	if item_id == gamedata.VI_XP {
		return
	}

	if item_id == helper.MatEevoUniversalItemID {
		return
	}

	if item_id == gamedata.VI_Sc0 || itemID == gamedata.VI_Sc0 || itemID1 == gamedata.VI_Sc0 {
		return
	}

	// 客户端不支持转成GENERAL_SSHUANG形式的返回
	if is_goodwill || is_hero {
		r.StageRewards = append(r.StageRewards, encode(StageReward{item_id, count, ""}))
		return
	}

	if is_2_sc { // 金钱是算的最终值
		if itemID != helper.VI_Sc0 {
			//logs.Warn("StageReward %s %d", itemID, count*countc)
			r.StageRewards = append(r.StageRewards, encode(StageReward{itemID, count * countc, ""}))
		}
		return
	}
	if is_2_c {
		//logs.Warn("StageRewardsdasa %s %d", itemID1, count*countc1)
		r.StageRewards = append(r.StageRewards, encode(StageReward{itemID1, count * countc1, ""}))
		return
	}
	if data != nil {
		r.StageRewards = append(r.StageRewards, encode(StageReward{item_id, count, data.ToDataStr()}))
	} else {
		r.StageRewards = append(r.StageRewards, encode(StageReward{item_id, count, ""}))
	}

}

func (p *Account) DeclareLootForLevelEnemy(r servers.Request) *servers.Response {

	req := &RequestDeclareLootForLvlEnmy{}
	resp := &ResponseDeclareLootForLvlEnmy{}

	initReqRsp(
		"PlayLevel/DeclareLvlEnmyLootResponse",
		r.RawBytes,
		req, resp, p)

	resp.initData()
	resp.StageRewards = make([][]byte, 0, NORMAL_REWARD_MAX)
	resp.IsSuccess = req.IsSuccess

	var loot_content_data lootContent_t
	if err := json.Unmarshal(p.Tmp.LootContent, &loot_content_data); err != nil {
		logs.Warn("CODE_NoPreLootInfo last_getproto [%d] last_getinfo [%d] "+
			"last_prepare [%s %d] last_declare [%s %d] this_declare [%s %s]",
			p.Tmp.Last_GetProto_TS, p.Tmp.Last_GetInfo_TS,
			p.Tmp.Last_Level_Prepare, p.Tmp.Last_Level_Prepare_TS,
			p.Tmp.Last_Level_Declare, p.Tmp.Last_Level_Declare_TS,
			req.LevelId, req.SeqId)
		return rpcErrorWithMsg(resp, 1,
			fmt.Sprintf("CODE_NoPreLootInfo json.Unmarshal err %v", err))
		//logs.Warn("DeclareLootForLevelEnemy %s %s", acid, fmt.Sprintf("CODE_NoPreLootInfo err %v", err))
		//return rpcSuccess(resp)
	}

	stage_id := loot_content_data.ID
	p.Tmp.Last_Level_Declare = stage_id
	p.Tmp.Last_Level_Declare_TS = time.Now().Unix()

	return p.declareSuccess(req, resp, loot_content_data)
}

/*
	此协议用来保护客户端
	1、没发prepare就发Declare
	2、连发两次Declare
*/
func (p *Account) DeclareLootForLevelEnemyV2(r servers.Request) *servers.Response {

	req := &RequestDeclareLootForLvlEnmyV2{}
	resp := &ResponseDeclareLootForLvlEnmy{}

	initReqRsp(
		"PlayLevel/DeclareLvlEnmyLootResponse_v2",
		r.RawBytes,
		req, resp, p)

	const (
		_ = iota
		Err_IsStageCanPlay
		Err_mkLootData
		Err_mkStageRewards
	)

	resp.initData()
	resp.StageRewards = make([][]byte, 0, NORMAL_REWARD_MAX)
	resp.IsSuccess = req.IsSuccess

	var loot_content_data lootContent_t
	if err := json.Unmarshal(p.Tmp.LootContent, &loot_content_data); err != nil {
		logs.Warn("CODE_NoPreLootInfo V2 %s last_getproto [%d] last_getinfo [%d] "+
			"last_prepare [%s %d] last_declare [%s %d] this_declare [%s %s]",
			p.AccountID.String(),
			p.Tmp.Last_GetProto_TS, p.Tmp.Last_GetInfo_TS,
			p.Tmp.Last_Level_Prepare, p.Tmp.Last_Level_Prepare_TS,
			p.Tmp.Last_Level_Declare, p.Tmp.Last_Level_Declare_TS,
			req.LevelId, req.SeqId)

		// 补救措施：应对前端没法prepare就发declare和发两次declare两种情况
		// 再次尝试结算，失败就返回成功吧......
		_resp := &ResponsePrepareLootForLevelEnemy{}
		code, warncode := p.IsStageCanPlay(req.LevelId, req.AvatarIDs, false)
		if warncode != 0 || code != 0 {
			logs.Warn("%s DeclareLootForLevelEnemyV2 IsStageCanPlay %s warncode %d code %d",
				req.LevelId, p.AccountID.String(), warncode, code)
			p.declareOnChange(resp)
			return rpcSuccess(resp)
			//return rpcErrorWithMsg(resp, Err_IsStageCanPlay, fmt.Sprintf("Err_IsStageCanPlay %s", req.LevelId))
		}
		loot_data_ok := mkLootData(p, req.LevelId, req.AvatarIDs[:], _resp)
		if !loot_data_ok {
			return rpcErrorWithMsg(resp, Err_mkLootData, fmt.Sprintf("Err_mkLootData %s", req.LevelId))
		}
		reward_data_ok := mkStageRewards(p, req.LevelId, req.AvatarIDs[:], _resp)
		if !reward_data_ok {
			return rpcErrorWithMsg(resp, Err_mkStageRewards, fmt.Sprintf("Err_mkStageRewards %s", req.LevelId))
		}

		json.Unmarshal(p.Tmp.LootContent, &loot_content_data)
		req.LootIDs = make([]uint16, len(loot_content_data.Info))
		for id, _ := range loot_content_data.Info {
			req.LootIDs[id] = uint16(id) + p.Tmp.LootRand
		}
	}

	p.Tmp.Last_Level_Declare = req.LevelId
	p.Tmp.Last_Level_Declare_TS = time.Now().Unix()

	_req := &RequestDeclareLootForLvlEnmy{
		LootIDs:   req.LootIDs,
		AvatarIDs: req.AvatarIDs,
		KillNums:  req.KillNums,
		KillTyps:  req.KillTyps,
		IsSuccess: req.IsSuccess,
		Star:      req.Star,
		Hackjson:  req.Hackjson,
	}
	return p.declareSuccess(_req, resp, loot_content_data)
}

func (p *Account) declareSuccess(req *RequestDeclareLootForLvlEnmy,
	resp *ResponseDeclareLootForLvlEnmy, loot_content_data lootContent_t) *servers.Response {

	const (
		_                       = iota + 10
		CODE_NoPreLootInfo      // 失败:没有申请过开启关卡
		CODE_NoPreLootStageInfo // 失败:没有申请过开启关卡
		CODE_SendRewardErr      // 警告:奖励物品发送失败
		CODE_LootIdErr          // 警告:申请的掉落ID无效
		CODE_AntiCheat_Unmarshal_Err
		CODE_MiniLvl_Already_Pass
	)
	const (
		CODE_MIN = 20
	)
	now_t := p.Profile.GetProfileNowTime()
	acid := p.AccountID.String()
	var stage_id string
	needSysNotice := false

	levelCostTime := time.Now().Unix() - p.Tmp.GetLevelEnterTime()
	stage_id = loot_content_data.ID
	if stage_id == "" {
		return rpcError(resp, CODE_NoPreLootStageInfo)
	}
	// 清理关卡信息
	p.Tmp.CleanStageData()
	if req.IsSuccess && req.Star > 0 {
		// 反作弊检查
		// 反作弊检查
		if req.Hackjson != "" {
			hacks := []float32{}
			if err := json.Unmarshal([]byte(req.Hackjson), &hacks); err != nil {
				return rpcErrorWithMsg(resp, CODE_AntiCheat_Unmarshal_Err, fmt.Sprintf("hack unmarshal err %s", err.Error()))
			}
			resp.CheatedIndex = p.AntiCheat.CheckFightRelAll(
				acid,
				hacks,
				p.Account,
				account.Anticheat_Typ_LevelStage,
				levelCostTime)
		} else {
			resp.CheatedIndex = []int{}
			logs.Info("[Antichest-Empty] Stage %s acid %s req.Hackjson is empty", stage_id, p.AccountID.String())
		}
		if len(resp.CheatedIndex) > 0 {
			if isTimeCheat(resp.CheatedIndex) {
				return rpcWarn(resp, errCode.YouTimeCheat)
			}
			return rpcWarn(resp, errCode.YouCheat)
		}
		stage_data := gamedata.GetStageData(stage_id)

		// 检查装备物品数量, 此处由于前端应该在进关卡时检查，所以只记log
		if p.BagProfile.GetEquipCount() >= gamedata.GetEquipCountUpLimit() {
			logs.Error("[DeclareLevel-BagFull] equip %s %s", p.AccountID.String(), stage_id)
		}
		if p.Profile.GetJadeBag().GetJadeSumCount() >= gamedata.GetJadeCountUpLimit() {
			logs.Error("[DeclareLevel-BagFull] jade %s %s", p.AccountID.String(), stage_id)
		}

		// 消耗
		code, warnCode := p.CostStagePay(stage_id, req.AvatarIDs, false, resp)
		if warnCode != 0 {
			return rpcWarn(resp, warnCode)
		}
		if code != 0 || req.Star < 0 || req.Star > MAX_STAR {
			return rpcErrorWithMsg(resp, CODE_MIN+code, fmt.Sprintf("%s", stage_id))
		}

		// 星级计算
		player_stage_info := p.Profile.GetStage().GetStageInfo(
			gamedata.GetCommonDayBeginSec(p.Profile.GetProfileNowTime()),
			stage_id,
			p.GetRand())

		star := int32(req.Star)

		is_first_pass := player_stage_info.MaxStar == 0

		oldStar := player_stage_info.MaxStar
		oldStarCount := stage_star.GetStarCount(oldStar)
		newStarCount := stage_star.GetStarCount(star)
		if oldStarCount < newStarCount {
			if stage_data.Type != gamedata.LEVEL_TYPE_MINILEVEL {
				p.Profile.GetStage().AddChapterStarFromStage(
					stage_id,
					newStarCount-oldStarCount)
			}
		}
		player_stage_info.MaxStar = stage_star.AddStar(oldStar, star)

		player_stage_info.T_count += 1
		player_stage_info.Sum_count += 1

		// 条件结算更新
		for i := 0; i < len(req.KillTyps) && i < len(req.KillNums); i++ {
			// 杀敌数量
			p.updateCondition(
				account.COND_TYP_Kill,
				req.KillNums[i], // p1 代表数量
				0,
				req.KillTyps[i], // p3 代表
				stage_id, resp)  // p3 代表类型
		}
		p.updateCondition(account.COND_TYP_Stage_Pass,
			1, int(newStarCount), stage_id, "", resp)

		// 主线关卡通过次数条件
		if stage_data.Type == gamedata.LEVEL_TYPE_MAIN ||
			stage_data.Type == gamedata.LEVEL_TYPE_MINILEVEL {
			p.updateCondition(account.COND_TYP_Any_Stage_Pass,
				1, 1, "", "", resp)
			needSysNotice = true
			if stage_data.LevelIndex > p.Profile.GetData().FarthestStageIndex {
				p.Profile.GetData().FarthestStageIndex = stage_data.LevelIndex
			}
			p.Profile.GetMarketActivitys().OnGameMode(acid, gamedata.CounterTypeMain, 1, now_t)

		}
		// 精英关卡通过次数条件
		if stage_data.Type == gamedata.LEVEL_TYPE_ELITE {
			p.updateCondition(account.COND_TYP_Any_Stage_Pass,
				1, 2, "", "", resp)
			needSysNotice = true
			if stage_data.EliteLevelIndex > p.Profile.GetData().FarthestEliteStageIndex {
				p.Profile.GetData().FarthestEliteStageIndex = stage_data.EliteLevelIndex
			}
			p.Profile.GetMarketActivitys().OnGameMode(acid, gamedata.CounterTypeElite, 1, now_t)
			p.Profile.GetHmtActivityInfo().AddDungeonJy(p.GetProfileNowTime(), 1)
		}
		// 地狱关卡通过次数条件
		if stage_data.Type == gamedata.LEVEL_TYPE_HELL {
			p.updateCondition(account.COND_TYP_Any_Stage_Pass,
				1, 3, "", "", resp)
			needSysNotice = true
			if stage_data.HellLevelIndex > p.Profile.GetData().FarthestHellStageIndex {
				p.Profile.GetData().FarthestHellStageIndex = stage_data.HellLevelIndex
			}
			p.Profile.GetMarketActivitys().OnGameMode(acid, gamedata.CounterTypeHell, 1, now_t)
			p.Profile.GetHmtActivityInfo().AddDungeonDy(p.GetProfileNowTime(), 1)
		}

		// 任意关卡通过次数条件
		p.updateCondition(account.COND_TYP_Any_Stage_Pass,
			1, 0, "", "", resp)

		if is_first_pass {
			p.Account.GetHandle().OnFirstPassStage(stage_id)
		}

		// TBD 这个数据量后面可能很大，需要考虑客户端自己做模拟
		resp.OnChangeStageAll()
		// TBD 章节数据，如果过大，考虑优化
		resp.OnChangeChapterAll()

		datalist := loot_content_data.Info
		idstart := int(p.Tmp.LootRand - 1)
		idend := len(datalist) + int(p.Tmp.LootRand)
		logs.Trace("loot pre : %v", datalist)

		// 伪掉落真掉落
		rewardShouldGive := p.Tmp.StageRewards.Mk2One()

		for _, id := range req.LootIDs {
			iid := int(id)
			if iid > idstart && iid < idend {
				idx := iid - int(p.Tmp.LootRand)
				lootitem := datalist[idx]
				logs.Trace("loot %v", lootitem)

				if !gamedata.IsFixedIDItemID(lootitem.Data) {
					item_data := gamedata.MakeItemData(p.AccountID.String(), p.Account.GetRand(), lootitem.Data)
					rewardShouldGive.AddItemWithData(lootitem.Data, *item_data, lootitem.Count)
				} else {
					rewardShouldGive.AddItem(lootitem.Data, lootitem.Count)
				}
			} else {
				logs.Warn("%s DeclareLootForLevelEnemy stage %s got some unknown id declared! %d, loot_content %v",
					acid, stage_id, id, loot_content_data)
				return rpcWarn(resp, errCode.ClickTooQuickly)
			}
		}
		p.Tmp.StageRewards.Init(8)

		data := rewardShouldGive.Gives()
		data.AddAvatars(req.AvatarIDs)
		// 兑换商店道具掉落
		giveExchangeData := p.getExchangeShopLoot(stage_id, uint32(stage_data.Type))
		data.AddGroup(giveExchangeData)
		ok := account.GiveBySync(p.Account, data, resp, "StageReward")
		if !ok {
			logs.SentryLogicCritical(acid, "Give Loot Item Err")
		}
		// game mode 奖励
		p.gameModeAward(acid, stage_id, resp)

		// 全局信息
		if needSysNotice {
			global_info.OnLevelFinish(p.AccountID.ShardId, stage_id, acid, p.Profile.Name)
		}
		// logiclog
		gamelog.LogStageFinish(acid, p.Profile.GetCurrAvatar(), stage_id, true, newStarCount, 1, false, levelCostTime,
			p.Profile.GetCorp().Level, p.Profile.ChannelId, p.Profile.Data.CorpCurrGS, p.Profile.GetDestinyGeneral().SkillGenerals,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	} else { // 失败逻辑
		// 目前的逻辑是什么也不做
		logs.Trace("Stage failed!")
		// logiclog
		gamelog.LogStageFinish(acid, p.Profile.GetCurrAvatar(), stage_id, false, 0, 1, false, levelCostTime,
			p.Profile.GetCorp().Level, p.Profile.ChannelId, p.Profile.Data.CorpCurrGS, p.Profile.GetDestinyGeneral().SkillGenerals,
			func(last string) string { return p.Profile.GetLastSetCurLogicLog(last) }, "")
	}

	p.Profile.GetHero().SetNeedSync()
	resp.OnChangeSC()
	resp.OnChangeEnergy()
	resp.OnChangeAvatarExp()
	resp.OnChangeMarketActivity()

	//resp.OnChangeBoss()
	resp.MsgOK = "ok"
	logs.Trace("resp %v", resp)
	resp.mkInfo(p)

	// log
	gamelog.LogStage_c(p.AccountID.String(), p.Profile.GetCurrAvatar(),
		p.Profile.GetCorp().GetLvlInfo(), p.Profile.ChannelId, stage_id, "LeaveStage", p.Tmp.GetLevelEnterTime(),
		p.Profile.GetData().CorpCurrGS, "")
	return rpcSuccess(resp)
}

func (p *Account) getExchangeShopLoot(levelID string, stageType uint32) (giveData *gamedata.CostData) {
	giveData = &gamedata.CostData{}
	hotData := gamedata.GetHotDatas()
	acts := hotData.Activity.GetActivityInfoValid(gamedata.ActStageExchangePropLoot,
		p.Profile.ChannelQuickId, p.Profile.GetProfileNowTime())
	logs.Debug("cur valid exchange shop info: %v", acts)
	if len(acts) > 0 {
		if len(acts) > 1 {
			logs.Warn("multi valid exchange shop exists")
		}
		exchangeReward := gamedata.GetHotDatas().HotStageLootExchangeData.GetStageLootExchangeProp(
			acts[0].ActivityId, levelID, stageType, p.Profile.GetCorp().GetLvlInfo())
		if exchangeReward != nil {
			giveData.AddItem(exchangeReward.GetItemID(), exchangeReward.GetItemCount())
		}
	}
	return giveData
}

func (p *Account) gameModeAward(acid, stage_id string, resp *ResponseDeclareLootForLvlEnmy) {
	now_t := p.Profile.GetProfileNowTime()
	// 金币关
	if p.Tmp.GoldLevelPoint >= 0 && gamedata.IsGoldLevel(stage_id) {
		p.updateCondition(account.COND_TYP_GoldLevel,
			1, 0, "", "", resp)
		p.updateCondition(account.COND_TYP_Try_Test,
			1, 0, "", "", resp)

		sc := gamedata.GetGoldLevelMinReward(stage_id) + p.Tmp.GoldLevelPoint
		v, _ := p.Profile.GetVip().GetVIP()
		vipCfg := gamedata.GetVIPCfg(int(v))
		var vipAdd uint32
		if vipCfg != nil {
			vipAdd = uint32(vipCfg.GoldLevelAdd * float32(sc))
		}
		sc = sc + vipAdd
		p.Tmp.GoldLevelPoint = 0

		gold_give := account.GiveGroup{}
		giveReward(p, &gold_give, gamedata.VI_Sc0, sc, nil)
		resp.addReward(gamedata.VI_Sc0, sc, nil)
		ok := gold_give.GiveBySyncAuto(p.Account, resp, "StageGoldLevel")
		if !ok {
			logs.SentryLogicCritical(acid, "Give StageGoldLevel sc Err")
		}
		resp.GoldLevelPoint = sc
		resp.GoldLevelPointAdd = vipAdd
		// market activity
		p.Profile.GetMarketActivitys().OnGameMode(acid, gamedata.CounterTypeGoldLevel, 1, now_t)

	}
	// 精铁关
	if minReward, isSuccess := gamedata.GetExpLevelMinAward(stage_id); isSuccess {
		p.updateCondition(account.COND_TYP_FiLevel,
			1, 0, "", "", resp)
		p.updateCondition(account.COND_TYP_Try_Test,
			1, 0, "", "", resp)
		resp.ExpLevelPoint = p.Tmp.ExpLevelPoint + minReward

		v, _ := p.Profile.GetVip().GetVIP()
		vipCfg := gamedata.GetVIPCfg(int(v))
		if vipCfg != nil {
			resp.ExpLevelPointAdd = uint32(vipCfg.IronLevelAdd * float32(resp.ExpLevelPoint))
		}
		fi := resp.ExpLevelPoint + resp.ExpLevelPointAdd
		p.Tmp.ExpLevelPoint = 0

		exp_give := account.GiveGroup{}
		giveReward(p, &exp_give, helper.VI_Sc1, fi, nil)
		resp.addReward(helper.VI_Sc1, fi, nil)

		ok := exp_give.GiveBySyncAuto(p.Account, resp, "StageExpLevel")
		if !ok {
			logs.SentryLogicCritical(acid, "Give StageExpLevel Item Err")
		}
		// market activity
		p.Profile.GetMarketActivitys().OnGameMode(acid, gamedata.CounterTypeFineIronLevel, 1, now_t)

	}
	// 天命关
	if minReward, isSuccess := gamedata.GetDCLevelMinAward(stage_id); isSuccess {
		p.updateCondition(account.COND_TYP_ExpLevel,
			1, 0, "", "", resp)
		p.updateCondition(account.COND_TYP_Try_Test,
			1, 0, "", "", resp)

		resp.DCLevelPoint = p.Tmp.DCLevelPoint + minReward
		v, _ := p.Profile.GetVip().GetVIP()
		vipCfg := gamedata.GetVIPCfg(int(v))
		if vipCfg != nil {
			resp.DCLevelPointAdd = uint32(vipCfg.DcLevelAdd * float32(resp.DCLevelPoint))
		}
		dc := resp.DCLevelPoint + resp.DCLevelPointAdd
		p.Tmp.DCLevelPoint = 0

		dc_give := account.GiveGroup{}
		giveReward(p, &dc_give, helper.VI_DC, dc, nil)
		resp.addReward(helper.VI_DC, dc, nil)

		ok := dc_give.GiveBySyncAuto(p.Account, resp, "StageDCLevel")
		if !ok {
			logs.SentryLogicCritical(acid, "Give StageDCLevel Item Err")
		}
		// market activity
		p.Profile.GetMarketActivitys().OnGameMode(acid, gamedata.CounterTypeDCLevel, 1, now_t)

	}
}

func (p *Account) declareOnChange(resp *ResponseDeclareLootForLvlEnmy) {
	p.Profile.GetHero().SetNeedSync()
	resp.OnChangeSC()
	resp.OnChangeAllGameMode()
	resp.OnChangeEnergy()
	resp.OnChangeAvatarExp()
	resp.OnChangeMarketActivity()
	resp.OnChangeStageAll()
	resp.OnChangeChapterAll()
	resp.OnChangeBag()
	resp.OnChangeQuestAll()
	resp.OnChangeCorpExp()
	resp.mkInfo(p)
}
