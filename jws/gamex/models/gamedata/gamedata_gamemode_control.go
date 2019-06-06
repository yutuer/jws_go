package gamedata

import (
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"

	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type GameModeControlData struct {
	ModeID             int
	OpenLevel          uint32                  // 开启等级
	GetType            int                     // 获取次数类型 0.定时重置（每一次次数增加前都会清掉之前的次数）1.定时增加（可以累积次数）
	IsNeedGetDayInWeek [util.WeekDayCount]bool // 周几是否参与次数刷新/增长
	GetDailyTime       int64                   // 获取时刻
	GetCount           int                     // 获取次数
	CountMax           int                     // 1.为定时增加类型时，可设定累积次数的上限
	MinPreAdd          int64                   // 获取次数类型 2 时, 几分钟增加一点
}

func (g *GameModeControlData) FromData(data *ProtobufGen.MODECONTROL) {
	g.ModeID = int(data.GetModeID())
	g.OpenLevel = data.GetOpenLevel()
	g.GetType = int(data.GetGetTicketType())
	g.GetDailyTime = util.DailyTimeFromString(data.GetGetTicketTime())
	g.GetCount = int(data.GetGetTicketNumber())
	g.CountMax = int(data.GetGetTicketValue1())
	g.MinPreAdd = int64(data.GetGetTicketCD())

	days := strings.Split(data.GetGetTicketDay(), ",")
	logs.Trace("GameModeControlData %v %v", *g, days)

	for _, weekday := range days {
		dn, err := strconv.Atoi(weekday)
		if err != nil {
			logs.Error("GameModeControlData %v %s Err", days, data.GetGetTicketDay())
			panic(err)
		}
		dn = util.TimeWeekDayTranslateFromCfg(dn)
		if dn < 0 || dn >= len(g.IsNeedGetDayInWeek) {
			logs.Error("GameModeControlData weekday %v %s Err", days, data.GetGetTicketDay())
			panic(err)
		}
		g.IsNeedGetDayInWeek[dn] = true
	}
}

var (
	gdGameModeControlData [CounterTypeCountMax]GameModeControlData
)

func loadGameModeControlData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.MODECONTROL_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()

	for _, c := range data {
		if int(c.GetModeID()) >= len(gdGameModeControlData) {
			logs.Warn("loss game mode control id: %d", c.GetModeID())
			continue
		}
		gdGameModeControlData[int(c.GetModeID())].FromData(c)
	}
	logs.Warn("gdGameModeControlData %v", gdGameModeControlData)
}

func GetGameModeControlData(typ int) *GameModeControlData {
	return &gdGameModeControlData[typ]
}
