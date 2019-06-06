package guild_info

import (
	"fmt"
	"math"
	"sort"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

type GuildWorshipInfo struct {
	WorshipMember      []MemberInfo     `json:"worship_member"`
	WorshipLog         []worshipLogInfo `json:"worship_log"`
	WorshipIndex       int64            `json:"worship_index"`
	IsOpen             bool             `json:"is_open"`
	LastDailyResetTime int64            // 每日重置时间
}

type MemberInfo struct {
	MemberAccountId string    //玩家id
	MemberNames     string    // 玩家姓名
	MemberGs        int       // 战力
	HeroId          int       // 主将ID
	HeroSwing       int       // 翅膀ID
	MagicPetfigure  uint32    // 灵宠外观
	HeroFashion     [2]string // 时装
	MemberSign      int64     // 签 0 , 1 , 2
	WorshipNum      int       // 被膜拜的次数
	CorpLv          uint32
	Vip             uint32
}

type worshipLogInfo struct {
	PlayerNames string // 玩家姓名
	SignId      int64  // 签Id
	Worshiptiem int64  // 膜拜时间
}

func (e *GuildWorshipInfo) AddWorshipLogInfo(pn string, sid int64, time int64) {
	e.WorshipLog = append(e.WorshipLog, worshipLogInfo{pn, sid, time})
}

func (e *GuildWorshipInfo) GetWorshipLogInfo() []worshipLogInfo {
	return e.WorshipLog[:]
}

func (e *GuildWorshipInfo) UpdateWorshipIndex(i int) {
	e.WorshipIndex += 1
	e.WorshipMember[i].WorshipNum += 1
}

func (e *GuildWorshipInfo) GetWorshipIndex() int64 {
	return e.WorshipIndex
}

func (e *GuildWorshipInfo) CheckDailyReset(now int64) bool {
	//logs.Debug("guild Worship CheckDailyReset, %d, %d", now, e.LastDailyResetTime)
	if !util.IsSameUnixByStartTime(e.LastDailyResetTime, now,
		gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypGuildWorshipReset)) {
		//发放奖励邮件
		e.SendWorshipMail()

		e.DailyReset(now)
		logs.Debug("guild Worship daily reset")
		return true
	}
	return false
}

func (e *GuildWorshipInfo) DailyReset(nowTime int64) {
	e.LastDailyResetTime = nowTime
	e.WorshipLog = nil
	e.WorshipIndex = 0
	for _, x := range e.WorshipMember {
		x.WorshipNum = 0
	}
}

func (e *GuildWorshipInfo) SendWorshipMail() {
	if len(e.WorshipMember) != 0 {
		for _, x := range e.WorshipMember {
			if x.WorshipNum != 0 {
				mail_sender.BatchSendMail2Account(x.MemberAccountId,
					timail.Mail_send_By_Guild,
					mail_sender.IDS_MAIL_WORSHIP_TITLE,
					[]string{fmt.Sprintf("%d", x.WorshipNum)},
					gamedata.GetWorshipReward(uint32(x.WorshipNum)),
					"GuildWorshipMail", false)
			}
		}
	}
}

const MAX_WORSHIP_MEMBER = 5

const (
	bad    = 0
	normal = 1
	good   = 2
)

func (g *GuildInfoBase) getGuildMemberTop5() {
	memberdates := make(ByGs, g.Base.MemNum) // 军团所有成员
	for i, member := range g.Members[:g.Base.MemNum] {
		memberdates[i].MemberNames = member.Name
		memberdates[i].MemberGs = member.CurrCorpGs
		memberdates[i].HeroId = member.CurrAvatar
		memberdates[i].HeroSwing = member.Swing
		memberdates[i].MagicPetfigure = member.MagicPetfigure
		memberdates[i].HeroFashion = member.FashionEquips
		memberdates[i].MemberId = member.AccountID
		memberdates[i].CorpLv = member.CorpLv
		memberdates[i].Vip = member.CorpLv

	}
	if len(memberdates) < 1 {
		logs.Debug("guuildWorship: fail to find member, guilduuid: %s", g.Base.GuildUUID)
	} else {
		logs.Debug("guuildWorship: find %v ", memberdates)
	}
	sort.Sort(memberdates)

	gw := &g.GuildWorship
	gw.WorshipMember = make([]MemberInfo, int(math.Min(MAX_WORSHIP_MEMBER, float64(len(memberdates)))))
	for i := 0; i < len(gw.WorshipMember); i++ {
		gw.WorshipMember[i].MemberNames = memberdates[i].MemberNames
		gw.WorshipMember[i].MemberGs = memberdates[i].MemberGs
		gw.WorshipMember[i].HeroId = memberdates[i].HeroId
		gw.WorshipMember[i].HeroSwing = memberdates[i].HeroSwing
		gw.WorshipMember[i].MagicPetfigure = memberdates[i].MagicPetfigure
		gw.WorshipMember[i].HeroFashion = memberdates[i].HeroFashion
		gw.WorshipMember[i].MemberAccountId = memberdates[i].MemberId
		gw.WorshipMember[i].CorpLv = memberdates[i].CorpLv
		gw.WorshipMember[i].Vip = memberdates[i].Vip

	}
	sign := getSign(len(gw.WorshipMember))
	for i, x := range sign {
		gw.WorshipMember[i].MemberSign = int64(x)
	}

}

func (g *GuildInfoBase) TryResetGuildWorship(nowT int64) {
	if g.Base.Level < gamedata.GetGuildActivityLvl(gamedata.GuildActivity_GUILD_WORSHIPCRIT_NAME) {
		return
	}
	if gamedata.IsSameDayGuildWorship(g.Base.Guild2LvlTimes, nowT) {
		return
	}
	refresh := g.GuildWorship.CheckDailyReset(nowT)
	if refresh {
		g.GuildWorship.IsOpen = true
		g.getGuildMemberTop5()
	}
}

func getSign(pnum int) []int {
	originalSing := []int{bad, bad, normal, normal, good}
	return util.ShuffleArray(originalSing[:pnum])
}

type Person struct {
	MemberNames    string
	MemberGs       int
	HeroId         int
	HeroSwing      int
	MagicPetfigure uint32
	HeroFashion    [2]string
	MemberSign     int64
	MemberId       string
	CorpLv         uint32
	Vip            uint32
}

// ByGs implements sort.Interface for []Person based on
// the gs field.
type ByGs []Person

func (a ByGs) Len() int           { return len(a) }
func (a ByGs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByGs) Less(i, j int) bool { return a[i].MemberGs > a[j].MemberGs }
