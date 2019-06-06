package account

import (
	"fmt"

	"time"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
	"vcs.taiyouxi.net/jws/gamex/modules/Expedition"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const OtherInfoForPlayerPLen = 3
const EXPEDITION_ENMY_NUM = 9

type ExpeditionInfo struct {
	ExpeditionIds           []string                                     `json:"etid"`    //被远征的的玩家ID
	ExpeditionNames         []string                                     `json:"etna"`    //被远征的玩家名称
	ExpeditionState         int32                                        `json:"ets"`     //当前最远关卡
	ExpeditionAward         int32                                        `json:"eta"`     //当前最远宝箱
	ExpeditionNum           int32                                        `json:"etnu"`    //远征通关总数
	ExpeditionStep          bool                                         `json:"ed_step"` //远征通关状态 true:表示已打败9人
	ExpeditionREstNum       int32                                        `json:"etrn"`    //远征重置次数
	ExpeditionLvl           uint32                                       `json:"expedition_lvl"`
	ExpeditionMyHero        [helper.AVATAR_NUM_CURR]ExpeditinHeroInfo    `json:"ed_myhero"`
	ExpeditionEnmyInfo      [EXPEDITION_ENMY_NUM]ExpeditionEnmyInfo      `json:"ed_enmy_info"`
	ExpeditionEnmyDetail    [EXPEDITION_ENMY_NUM]ExpeditionEnmyDetail    `json:"ed_enmy_d"`
	ExpeditionEnmySkillInfo [EXPEDITION_ENMY_NUM]ExpeditionEnmySkillInfo `json:"ed_enmy_skill_info"`

	IsActive         bool  `json:"is_active"` //是否第一次进入远征
	GetEnemyNextTime int64 `json:"n_t"`       // 下次更新敌人信息时间戳
}

type ExpeditinHeroInfo struct {
	HeroIslive  int     `json:"hero_live" codec:"live"`   //远征英雄状态 0:活着 1:死亡
	HeroHp      float32 `json:"hero_hp" codec:"hreohp"`   //主将剩余血量
	HeroWuSkill float32 `json:"wu_skill" codec:"wuskill"` //主将无双技能槽
	HeroExSkill float32 `json:"ex_skill" codec:"exskill"` //主将普通技能槽
}

type ExpeditionEnmyInfo struct {
	Acid     string   `json:"id" codec:"id"`             //玩家id
	Name     string   `json:"n" codec:"n"`               //玩家姓名
	Gs       int      `json:"gs" codec:"gs"`             //玩家战力
	HeroId   []int    `json:"hid" codec:"hid"`           //玩家主将
	FAs      []uint32 `json:"fas" codec:"fas"`           // 玩家主将等级
	FAStarLv []uint32 `json:"fastarlv" codec:"fastarlv"` // 主将星级
	HeroGs   []int    `json:"hero_gs" codec:"hero_gs"`   //主将战力
	CorpLv   uint32   `json:"corp_lv" codec:"corp_lv"`
}

type ExpeditionEnmyDetail struct {
	Enemies [OtherInfoForPlayerPLen]helper.Avatar2Client
}
type ExpeditionEnmySkillInfo struct {
	State   int64     `json:"state" codec:"state"` //敌人的状态 0没战斗过 1战斗过
	Hp      []float32 `json:"hp" codec:"hp"`
	WuSkill []float32 `json:"ws" codec:"ws"`
	ExSkill []float32 `json:"es" codec:"es"`
}

// 玩家首次打开界面和重置的时候调用，并保存到profile里，一天只能调一次
// 若返回值的acid为空时，请用机器人填补
func (ep *ExpeditionInfo) GetEnemyToday(acid string, now_t int64) (bool, *ExpeditionDbInfo) {
	info := &ExpeditionDbInfo{
		Acid: acid,
	}
	err := info.DBLoad(false)
	if err != nil {
		logs.Error("ExpeditionInfo GetEnemyToday DBLoad err %v", err)
		return false, nil
	}
	if info.TimeStamp < gamedata.GetCommonDayBeginSec(now_t) {
		return false, nil
	}

	for i, s := range info.ExpeditionEnmySimple {
		ds := info.ExpeditionEnmyDetail[i]
		logs.Debug("GetEnemyToday simple %s %s %d", s.Acid, s.Name, s.HeroId)
		for _, d := range ds.Enemies {
			logs.Debug("GetEnemyToday detail %s %s %d", d.Acid, d.Name, d.AvatarId)
		}
	}
	return true, info
}

// 生成数据 + 存库
func (ep *ExpeditionInfo) LoadEnemyToday(acid string, gsMax int64, now_t int64) {
	if now_t > ep.GetEnemyNextTime {
		ep.GetEnemyNextTime = util.GetNextDailyTime(
			gamedata.GetCommonDayBeginSec(now_t), now_t)

		account, err := db.ParseAccount(acid)
		if err != nil {
			logs.Error("ExpeditionInfo db.ParseAccount err %s %v", acid, err)
			return
		}
		go func() {
			defer logs.PanicCatcherWithInfo("ExpeditionInfo LoadEnemyToday Panic")

			b := time.Now().UnixNano()

			err, _, simples, infos := GetExpeditionEnemy(acid, account.ShardId, gsMax)
			if err != nil {
				logs.Error("ExpeditionInfo Expedition.GetExpeditionEnemy err %v", err)
				return
			}
			dbInfo := ExpeditionDbInfo{
				Acid:      acid,
				TimeStamp: now_t,
			}
			for i, info := range infos {
				if info == nil {
					continue
				}
				dbInfo.ExpeditionEnmyDetail[i] = *info
				dbInfo.ExpeditionEnmySimple[i] = *simples[i]
			}

			conn := driver.GetDBConn()
			defer conn.Close()
			if conn.IsNil() {
				logs.Error("ExpeditionInfo GetDBConn Err conn is nil")
				return
			}

			cb := redis.NewCmdBuffer()
			dbInfo.DBSave(cb)
			conn.DoCmdBuffer(cb, true)

			logs.Debug("LoadEnemyToday go cost %v", time.Now().UnixNano()-b)

		}()
	}
}

//已开启远征的老号,存档升级
func (ep *ExpeditionInfo) UpExpedition() {
	for idx, _ := range ep.ExpeditionMyHero {
		n := &ep.ExpeditionMyHero[idx]
		if n.HeroHp == 0 && n.HeroIslive == 0 {
			n.HeroExSkill = 0.5
			n.HeroHp = 1
			n.HeroIslive = 0
			n.HeroWuSkill = 0
		}
	}
}

const (
	Table_ExpeditionDbInfo = "expeditenemy"
)

type ExpeditionDbInfo struct {
	Acid                 string                                    `json:"acid"`
	TimeStamp            int64                                     `json:"st"`
	ExpeditionEnmySimple [EXPEDITION_ENMY_NUM]ExpeditionEnmyInfo   `json:"ed_enmy_s"`
	ExpeditionEnmyDetail [EXPEDITION_ENMY_NUM]ExpeditionEnmyDetail `json:"ed_enmy_d"`
}

func (p *ExpeditionDbInfo) DBName() string {
	return fmt.Sprintf("%s:%s", Table_ExpeditionDbInfo, p.Acid)
}

func (p *ExpeditionDbInfo) DBSave(cb redis.CmdBuffer) error {
	key := p.DBName()
	return driver.DumpToHashDBCmcBuffer(cb, key, p)
}

func (p *ExpeditionDbInfo) DBLoad(logInfo bool) error {
	key := p.DBName()

	_db := driver.GetDBConn()
	defer _db.Close()

	err := driver.RestoreFromHashDB(_db.RawConn(), key, p, false, logInfo)
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		return err
	}
	return err
}

func GetExpeditionEnemy(acid string, sid uint, gs int64) (error,
	[]string, []*ExpeditionEnmyInfo, []*ExpeditionEnmyDetail) {

	b := time.Now().UnixNano()
	err, enemies_acid := Expedition.GetExpeditionEnemyId(acid, sid, gs)
	if err != nil {
		logs.Error("GetExpeditionEnemy _GetExpeditionEnemyId err %v", err)
		return err, nil, nil, nil
	}
	logs.Debug("GetExpeditionEnemy cost 1 %v", time.Now().UnixNano()-b)
	b = time.Now().UnixNano()
	err, simples, infos := _GetExpeditionEnemySimpleInfo(enemies_acid)
	if err != nil {
		logs.Error("GetExpeditionEnemy _GetExpeditionEnemySimpleInfo err %v", err)
		return err, nil, nil, nil
	}
	logs.Debug("GetExpeditionEnemy cost 2 %v", time.Now().UnixNano()-b)
	return err, enemies_acid, simples, infos
}

func _GetExpeditionEnemySimpleInfo(acids []string) (error,
	[]*ExpeditionEnmyInfo, []*ExpeditionEnmyDetail) {

	res_s := make([]*ExpeditionEnmyInfo, len(acids))
	res_d := make([]*ExpeditionEnmyDetail, len(acids))
	for i, acid := range acids {
		if acid == "" {
			continue
		}
		dbAccountID, err := db.ParseAccount(acid)
		if err != nil {
			logs.Error("_GetExpeditionEnemySimpleInfo db.ParseAccount %d %s %v", i, acid, err)
			continue
		}

		enemy_account, err := LoadPvPAccount(dbAccountID)
		if err != nil {
			logs.Error("_GetExpeditionEnemySimpleInfo LoadPvPAccount err %s %v",
				acid, err)
			continue
		}
		// simple info
		av_lv := make([]uint32, 0, 3)
		av_s := make([]uint32, 0, 3)
		av_gs := make([]int, 0, 3)
		enemy_data := enemy_account.Profile.Data
		enemy_profile := enemy_account.Profile
		for _, id := range enemy_data.BestHeroAvatar {
			av_lv = append(av_lv, enemy_profile.GetHero().HeroLevel[id])
			av_s = append(av_s, enemy_profile.GetHero().HeroStarLevel[id])
			av_gs = append(av_gs, enemy_data.HeroGs[id])
		}
		simple := &ExpeditionEnmyInfo{
			Acid:     acid,
			Name:     enemy_profile.Name,
			Gs:       enemy_data.CorpCurrGS,
			HeroId:   enemy_data.BestHeroAvatar,
			FAs:      av_lv,
			FAStarLv: av_s,
			HeroGs:   av_gs,
			CorpLv:   enemy_account.GetCorpLv(),
		}
		res_s[i] = simple
		// detial
		var enmy ExpeditionEnmyDetail
		j := 0
		for _, eav := range enemy_account.Profile.Data.BestHeroAvatar {
			a := helper.Avatar2Client{}
			err = FromAccount(&a, enemy_account, eav)
			if err != nil {
				logs.Error("_GetExpeditionEnemySimpleInfo account.FromAccount err %d %v", i, err)
				continue
			}
			enmy.Enemies[j] = a
			j++
			logs.Debug("_GetExpeditionEnemySimpleInfo %d %s %s %d", i, a.Acid, a.Name, eav)
		}

		res_d[i] = &enmy
	}
	return nil, res_s, res_d
}
