package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/account/gs"
	"vcs.taiyouxi.net/jws/gamex/modules/csrob"
)

type CSRobPlayerInfo struct {
	CurrGrade       uint32 `codec:"curr_grade"`        // 当前选择的车子品质, 0为未选择
	CurrRefresh     uint32 `codec:"curr_refresh"`      // 当前已刷车次数
	CurrRefreshCost uint32 `codec:"curr_refresh_cost"` // 当前刷车已消耗

	CountBuild uint32 `codec:"count_build"` // 已出车数量
	CountHelp  uint32 `codec:"count_help"`  // 已援助数量
	CountRob   uint32 `codec:"count_rob"`   // 已抢劫数量

	Formation   []int `codec:"formation"`    // 阵容
	FormationGS int64 `codec:"formation_gs"` // 阵容战力

	HasCurrCar bool   `codec:"has_curr_car"` // 当前正在开车
	CurrCar    []byte `codec:"curr_car"`     // 当前车子状态
}

type CSRobCarInfo struct {
	CarID      uint32   `codec:"carid"`      // 粮车ID
	Grade      uint32   `codec:"grade"`      // 粮车品质
	Acid       string   `codec:"acid"`       // 出车的人
	Name       string   `codec:"name"`       // 出车的人名
	GuildID    string   `codec:"guild_id"`   // 公会ID
	GuildName  string   `codec:"guild_name"` // 公会名字
	GuildPos   int      `codec:"guild_pos"`  // 公会职务
	Team       [][]byte `codec:"team"`       // 粮车防守阵容 []CSRobAvatarInfo
	StartStamp int64    `codec:"startstamp"` // 粮车开始时间
	EndStamp   int64    `codec:"endstamp"`   // 粮车结束时间
	Robbing    bool     `codec:"robbing"`    // 粮车正在被打劫
	RobCount   int      `codec:"robcount"`   // 粮车已被打劫次数

	HelperHas  bool     `codec:"helper_has"`  //有援助人
	HelperAcid string   `codec:"helper_acid"` //援助人的ID
	HelperName string   `codec:"helper_name"` //援助人的姓名
	Helper     [][]byte `codec:"helper"`      // 粮车援助阵容 []CSRobAvatarInfo

	AppealNum     uint32   `codec:"appeal_num"`     //已求援次数
	AlreadyAppeal []string `codec:"already_appeal"` //已发给这些人求援
	AppealLeast   uint32   `codec:"appeal_least"`   //剩余求援次数
}

type CSRobAvatarInfo struct {
	Idx            int      `codec:"idx"`              // 武将数字ID
	Attr           []byte   `codec:"attr"`             // 武将属性
	Gs             int64    `codec:"gs"`               // 武将战力
	Skills         []int64  `codec:"skills"`           // 技能等级
	Fashions       []string `codec:"fashions"`         // 时装ID
	PassiveSkillId []string `codec:"p_skill_id"`       // 被动技能ID
	CounterSkillId []string `codec:"c_skill_id"`       // 被动技能ID
	TriggerSkillId []string `codec:"t_skill_id"`       // 被动技能ID
	Star           int      `codec:"star"`             // 武将星级
	HeroWing       int      `codec:"hero_wing"`        // 翅膀ID
	MagicPetfigure uint32   `codec:"magic_pet_figure"` // 灵宠外观
	DesSkills      []int64  `codec:"dgss"`             // 神兽技能
}

type CSRobRecord struct {
	Type       int      `codec:"type"`      //日志类型
	Timestamp  int64    `codec:"timestamp"` //日志时间
	DriverID   string   `codec:"driver_id"`
	DriverName string   `codec:"driver_name"`
	RobberID   string   `codec:"robber_id"`
	RobberName string   `codec:"robber_name"`
	HelperID   string   `codec:"helper_id"`
	HelperName string   `codec:"helper_name"`
	Grade      uint32   `codec:"grade"`
	GoodID     []string `codec:"good_id"`
	GoodNum    []int    `codec:"good_num"`
}

type CSRobAppeal struct {
	Acid       string `codec:"acid"`
	Name       string `codec:"name"`
	CarID      uint32 `codec:"car_id"`
	Grade      uint32 `codec:"grade"`        //粮车品质
	AppealTime int64  `codec:"appeal_time"`  //请求时间
	EndStamp   int64  `codec:"end_stamp"`    //粮车结束时间戳
	HasHelper  bool   `codec:"has_helper"`   //已有护送
	HelperIsMe bool   `codec:"helper_is_me"` //就是我护送的
	RobedNum   int    `codec:"robed_num"`    //已抢劫人数
}

type CSRobEnemy struct {
	Acid  string `codec:"acid"`
	Name  string `codec:"name"`
	Count uint32 `codec:"count"`

	HasCurrCar bool   `codec:"has_curr_car"` //当前正在开车
	CurrCar    []byte `codec:"curr_car"`     // 当前车子状态
}

type CSRobGuildInfo struct {
	GuildID   string `codec:"guild_id"`
	GuildName string `codec:"guild_name"`
	BestGrade uint32 `codec:"best_grade"`
}

type CSRobGuildEnemy struct {
	GuildID   string `codec:"guild_id"`
	GuildName string `codec:"guild_name"`
	Count     uint32 `codec:"count"`
	BestGrade uint32 `codec:"best_grade"`
}

type CSRobGuildTeam struct {
	Acid       string   `codec:"acid"`
	Name       string   `codec:"name"`
	Team       [][]byte `codec:"team"` //[]CSRobAvatarInfo
	TeamGS     int64    `codec:"teamgs"`
	IsOnline   bool     `codec:"is_online"`   //是否在线
	AutoAccept []int64  `codec:"auto_accept"` //自动接受
}

//CSRobGuildRankElem 排行榜元素
type CSRobGuildRankElem struct {
	GuildID     string `codec:"guild_id"`
	GuildName   string `codec:"guild_name"`
	GuildMaster string `codec:"guild_master"`
	Rank        uint32 `codec:"rank"`
	RobCount    uint32 `codec:"rob_count"`
}

//CSRobAvatarForRank ..
type CSRobAvatarForRank struct {
	Idx  int   `codec:"idx"`  // 武将数字ID
	Gs   int64 `codec:"gs"`   // 武将战力
	Star int   `codec:"star"` // 武将星级
}

//CSRobNationalityRankElem ..
type CSRobNationalityRankElem struct {
	Pos  uint32   `codec:"pos"`
	Name string   `codec:"name"`
	Team [][]byte `codec:"team"` //[]CSRobAvatarForRank
}

func buildCSRobPlayerInfo(src *csrob.PlayerInfo) *CSRobPlayerInfo {
	ret := &CSRobPlayerInfo{}

	ret.CurrGrade = src.GradeRefresh.CurrGrade
	ret.CurrRefresh = src.GradeRefresh.CarSumGradeRefresh
	ret.CurrRefreshCost = src.GradeRefresh.CarSumGradeCost

	ret.CountBuild = src.Count.Build
	ret.CountRob = src.Count.Rob
	ret.CountHelp = src.Count.Help

	ret.Formation = src.CurrFormation[:]

	return ret
}

func buildCSRobAvatarInfo(src *csrob.HeroInfo) *CSRobAvatarInfo {
	ret := &CSRobAvatarInfo{}

	ret.Idx = src.Idx
	ret.Attr = src.Attr[:]
	ret.Gs = src.Gs
	ret.Skills = src.AvatarSkills[:]
	ret.Fashions = src.AvatarFashion[:]
	ret.PassiveSkillId = src.PassiveSkillId[:]
	ret.CounterSkillId = src.CounterSkillId[:]
	ret.TriggerSkillId = src.TriggerSkillId[:]
	ret.Star = src.StarLevel
	ret.HeroWing = src.HeroSwing
	ret.MagicPetfigure = src.MagicPetfigure
	ret.DesSkills = src.DesSkills[:]

	return ret
}

func buildCSRobCarInfo(src *csrob.PlayerRob) *CSRobCarInfo {
	ret := &CSRobCarInfo{}

	ret.CarID = src.CarID
	ret.Grade = src.Info.Grade
	ret.StartStamp = src.Info.StartStamp
	ret.EndStamp = src.Info.EndStamp
	ret.Robbing = src.Robbing
	ret.RobCount = len(src.Robbers)
	ret.Acid = src.Acid
	ret.Name = src.Name
	ret.GuildID = src.GuildID
	ret.GuildName = src.GuildName
	ret.GuildPos = src.GuildPos
	ret.AppealNum = uint32(len(src.AlreadyAppeal))
	ret.AlreadyAppeal = src.AlreadyAppeal

	ret.Team = [][]byte{}
	for _, h := range src.Info.Team {
		ret.Team = append(ret.Team, encode(buildCSRobAvatarInfo(&h)))
	}

	ret.Helper = [][]byte{}
	if nil != src.Helper {
		ret.HelperHas = true
		ret.HelperAcid = src.Helper.Acid
		ret.HelperName = src.Helper.Name
		for _, h := range src.Helper.Team {
			ret.Helper = append(ret.Helper, encode(buildCSRobAvatarInfo(&h)))
		}
	} else {
		ret.HelperHas = false
	}

	return ret
}

func buildCSRobRecord(src *csrob.PlayerRecord) *CSRobRecord {
	ret := &CSRobRecord{}

	ret.Type = src.Type
	ret.DriverID = src.DriverID
	ret.DriverName = src.DriverName
	ret.RobberName = src.RobberName
	ret.RobberID = src.RobberID
	ret.HelperID = src.HelperID
	ret.HelperName = src.HelperName
	ret.Grade = src.Grade
	ret.Timestamp = src.Timestamp
	for id, num := range src.Goods {
		ret.GoodID = append(ret.GoodID, id)
		ret.GoodNum = append(ret.GoodNum, int(num))
	}

	return ret
}

func buildCSRobAppeal(src *csrob.PlayerAppeal) *CSRobAppeal {
	ret := &CSRobAppeal{}

	ret.CarID = src.CarID
	ret.Name = src.Name
	ret.Acid = src.Acid
	ret.Grade = src.Grade
	ret.AppealTime = src.AppealTime
	ret.EndStamp = src.EndStamp
	ret.HasHelper = src.HasHelper
	ret.HelperIsMe = src.HelperIsMe
	ret.RobedNum = len(src.Robbers)

	return ret
}

func buildCSRobEnemy(src *csrob.PlayerEnemy) *CSRobEnemy {
	ret := &CSRobEnemy{}

	ret.Count = src.Count
	ret.Name = src.Name
	ret.Acid = src.Acid

	if nil != src.CurrCar {
		ret.HasCurrCar = true
		ret.CurrCar = encode(buildCSRobCarInfo(src.CurrCar))
	} else {
		ret.HasCurrCar = false
	}

	return ret
}

func (p *Account) buildHeroList(formation []int) []csrob.HeroInfo {
	_, heroAttr, _, heroGs, _, _, _ := gs.GetCurrAttr(account.NewAccountGsCalculateAdapter(p.Account))

	team := make([]csrob.HeroInfo, 0, len(formation))
	for _, idx := range formation {
		hero := csrob.HeroInfo{}

		hero.Idx = idx
		hero.Attr = encode(heroAttr[idx])
		hero.Gs = int64(heroGs[idx])
		hero.StarLevel = int(p.Profile.GetHero().GetStar(idx))
		for _, skill := range p.Profile.GetAvatarSkill().GetByAvatar(idx) {
			hero.AvatarSkills = append(hero.AvatarSkills, int64(skill))
		}
		hero.AvatarFashion = p.getEquipFashionTids(idx)
		hero.HeroSwing = p.Profile.GetHero().GetSwing(idx).CurSwing
		hero.MagicPetfigure = p.Profile.GetHero().GetMagicPetFigure(idx)
		hero.PassiveSkillId = p.Profile.GetHero().HeroSkills[idx].PassiveSkill[:]
		hero.CounterSkillId = p.Profile.GetHero().HeroSkills[idx].CounterSkill[:]
		hero.TriggerSkillId = p.Profile.GetHero().HeroSkills[idx].TriggerSkill[:]
		for _, skill := range p.Profile.DestinyGenerals.SkillGenerals {
			hero.DesSkills = append(hero.DesSkills, int64(skill)-1)
		}

		//TODO 填充属性
		team = append(team, hero)
	}

	return team
}

func (p *Account) buildHeroListFromRank(formation []int) []csrob.HeroInfoForRank {
	_, _, _, heroGs, _, _, _ := gs.GetCurrAttr(account.NewAccountGsCalculateAdapter(p.Account))

	team := make([]csrob.HeroInfoForRank, 0, len(formation))
	for _, idx := range formation {
		hero := csrob.HeroInfoForRank{}

		hero.Idx = idx
		hero.Gs = int64(heroGs[idx])
		hero.StarLevel = int(p.Profile.GetHero().GetStar(idx))

		//TODO 填充属性
		team = append(team, hero)
	}

	return team
}

func buildCSRobGuildInfo(src *csrob.GuildInfo) *CSRobGuildInfo {
	ret := &CSRobGuildInfo{}

	ret.GuildID = src.GuildID
	ret.GuildName = src.GuildName
	ret.BestGrade = src.BestGrade

	return ret
}

func buildCSRobGuildEnemy(src *csrob.GuildEnemy) *CSRobGuildEnemy {
	ret := &CSRobGuildEnemy{}

	ret.GuildID = src.GuildID
	ret.Count = src.Count
	ret.GuildName = src.GuildName
	ret.BestGrade = src.BestGrade

	return ret
}

func buildCSRobGuildTeam(src *csrob.GuildTeam) *CSRobGuildTeam {
	ret := &CSRobGuildTeam{}

	ret.Acid = src.Acid
	ret.Name = src.Name
	ret.TeamGS = 0

	for _, hero := range src.Hero {
		ret.TeamGS += hero.Gs
	}
	ret.Team = make([][]byte, 0, len(src.Hero))
	for _, team := range src.Hero {
		ret.Team = append(ret.Team, encode(buildCSRobAvatarInfo(&team)))
	}
	ret.AutoAccept = []int64{}
	for _, g := range src.AutoAccept {
		ret.AutoAccept = append(ret.AutoAccept, int64(g))
	}

	return ret
}

func buildCSRobGuildRankElem(src *csrob.GuildRankElem) *CSRobGuildRankElem {
	ret := &CSRobGuildRankElem{}

	ret.GuildID = src.GuildID
	ret.GuildMaster = src.GuildMaster
	ret.GuildName = src.GuildName
	ret.Rank = src.Rank
	ret.RobCount = src.RobCount

	return ret
}

func buildCSRobNationalityRankElem(src *csrob.RankTeam, pos uint32) *CSRobNationalityRankElem {
	ret := &CSRobNationalityRankElem{}

	ret.Pos = pos
	ret.Name = src.Name
	ret.Team = make([][]byte, 0, len(src.Heros))
	for _, hero := range src.Heros {
		ret.Team = append(ret.Team, encode(buildCSRobAvatarForRank(&hero)))
	}

	return ret
}

func buildCSRobAvatarForRank(src *csrob.HeroInfoForRank) *CSRobAvatarForRank {
	ret := &CSRobAvatarForRank{}

	ret.Idx = src.Idx
	ret.Gs = src.Gs
	ret.Star = src.StarLevel

	return ret
}
