package general

import (
	"time"

	"reflect"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	Init_General_Count  = 4
	Init_Relation_Count = 1
)

type PlayerGenerals struct {
	dbkey        db.ProfileDBKey
	dirtiesCheck map[string]interface{}
	Ver          int64 `redis:"version"`

	generals       map[string]*General
	relations      map[string]*Relation
	general2ActRel map[string]map[string]struct{}
	//LastTime   int64 `redis:"lasttime"`
	CreateTime int64 `redis:"createtime"`

	GeneralAr            []General
	general_ar_old_size  int
	RelationAr           []Relation
	relation_ar_old_size int

	QuestList            []GeneralQuestInList
	QuestListNextRefTime int64
	QuestRec             []GeneralQuestRec
	QuestNextId          int64
}

func (p *PlayerGenerals) OnAfterLogin() {
	logs.Trace("PlayerGenerals OnAfterLogin")
	if p.GeneralAr == nil {
		p.GeneralAr = make([]General, 0, Init_General_Count)
	}
	if p.RelationAr == nil {
		p.RelationAr = make([]Relation, 0, Init_Relation_Count)
	}
	if p.QuestList == nil {
		p.QuestList = make([]GeneralQuestInList, 0, gamedata.GeneralQuestListMax)
	}
	if p.QuestRec == nil {
		p.QuestRec = make([]GeneralQuestRec, 0, gamedata.GeneralQuestSetting().GetNGQAccParallel())
	}
	// 第一次要建立map
	p.general_ar_old_size = -1
	p.mkGeneralMap()
	p.relation_ar_old_size = -1
	p.mkRelationMap()
}

func (p *PlayerGenerals) GetGeneral(idx string) *General {
	g, ok := p.generals[idx]
	if ok {
		return g
	} else {
		return nil
	}
}

func (p *PlayerGenerals) GetGeneralByIdx(idx int) *General {
	if idx < 0 || idx >= len(p.GeneralAr) {
		return nil
	}

	return &p.GeneralAr[idx]
}

func (p *PlayerGenerals) IsExistGeneral(idx string) bool {
	g := p.GetGeneral(idx)
	return g != nil && g.IsHas()
}

func (p *PlayerGenerals) mkGeneralMap() {
	if p.general_ar_old_size == len(p.GeneralAr) {
		// 只要append就要检查一下
		return
	}
	p.generals = make(map[string]*General, len(p.GeneralAr))
	for i := 0; i < len(p.GeneralAr); i++ {
		p.generals[p.GeneralAr[i].Id] = &p.GeneralAr[i]
	}
	p.general_ar_old_size = len(p.GeneralAr)
}

func (p *PlayerGenerals) mkRelationMap() {
	p.relations = make(map[string]*Relation, len(p.RelationAr))
	p.general2ActRel = make(map[string]map[string]struct{}, len(p.GeneralAr))
	for i := 0; i < len(p.RelationAr); i++ {
		rel := &p.RelationAr[i]
		p.relations[rel.Id] = rel
		if rel.Level > 0 {
			cfg := gamedata.GetGeneralRelationInfo(rel.Id)
			p.updateGen2ActRel(rel.Id, cfg)
		}
	}
	p.relation_ar_old_size = len(p.RelationAr)
}

func (p *PlayerGenerals) AddGeneralNum(idx string, v uint32, reason string) {
	g, ok := p.generals[idx]
	if !ok {
		l := len(p.GeneralAr)
		p.GeneralAr = append(p.GeneralAr, General{Id: idx})
		p.GeneralAr[l].AddGeneralNum(v)
		p.mkGeneralMap()
	} else {
		g.AddGeneralNum(v)
	}
}

func (p *PlayerGenerals) GetGeneralRelation(relationId string) *Relation {
	if gamedata.GetGeneralRelationInfo(relationId) == nil {
		return nil
	}
	rel, ok := p.relations[relationId]
	if !ok {
		p.RelationAr = append(p.RelationAr, Relation{Id: relationId})
		p.mkRelationMap()
		return &p.RelationAr[len(p.RelationAr)-1]
	}
	return rel
}

func (p *PlayerGenerals) GetAllGeneral() []General {
	return p.GeneralAr[:]
}

func (p *PlayerGenerals) GetAllExistGeneralCount() int {
	n := 0
	for _, g := range p.GeneralAr {
		if g.IsHas() {
			n++
		}
	}
	return n
}

func (p *PlayerGenerals) GetAllGeneralRel() []Relation {
	return p.RelationAr[:]
}

func (p *PlayerGenerals) GetGeneralActRelNum(g string) int {
	rels, ok := p.general2ActRel[g]
	if ok {
		return len(rels)
	}
	return 0
}

func (p *PlayerGenerals) updateGen2ActRel(relId string, cfg *gamedata.GeneralRelInfo) {
	for _, g := range cfg.Generals {
		rels, ok := p.general2ActRel[g]
		if !ok {
			rels = make(map[string]struct{}, 5)
		}
		rels[relId] = struct{}{}
		p.general2ActRel[g] = rels
	}
}

// DB ------------------------------------------------
func NewPlayerGenerals(account db.Account) *PlayerGenerals {
	now_t := time.Now().Unix()
	return &PlayerGenerals{
		dbkey: db.ProfileDBKey{
			Account: account,
			Prefix:  "general",
		},
		//Ver:        helper.CurrDBVersion,
		//LastTime:   now_t,
		CreateTime: now_t,
	}
}

func (p *PlayerGenerals) DBName() string {
	return p.dbkey.String()
}

func (p *PlayerGenerals) DBSave(cb redis.CmdBuffer, forceDirty bool) error {
	key := p.DBName()

	if forceDirty {
		p.dirtiesCheck = nil
	}
	err, newDirtyCheck, chged := driver.DumpToHashDBCmcBufferCheckDirty(
		cb, key, p, p.dirtiesCheck)
	if err != nil {
		return err
	}
	if !game.Cfg.IsRunModeProd() {
		if !reflect.DeepEqual(p.dirtiesCheck, newDirtyCheck) {
			logs.Trace("Save PlayerGenerals %s %v", p.dbkey.Account.String(), chged)
		} else {
			logs.Trace("Save PlayerGenerals clean %s", p.dbkey.Account.String())
		}
	}
	p.dirtiesCheck = newDirtyCheck
	return nil
}

func (p *PlayerGenerals) DBLoad(logInfo bool) error {
	key := p.DBName()

	_db := driver.GetDBConn()
	defer _db.Close()

	logs.Trace("PlayerGenerals DBLoad")

	err := driver.RestoreFromHashDB(_db.RawConn(), key, p, false, logInfo)

	// RESTORE_ERR_Profile_No_Data 表示玩家第一次登陆游戏，没有存档，这不视为Bug
	// 外面的逻辑需要根据此判断是否是第一次登陆游戏
	if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
		logs.Trace("PlayerGenerals DBLoad %v", err)
		return err
	}
	p.dirtiesCheck = driver.GenDirtyHash(p)
	p.OnAfterLogin()

	return err
}
