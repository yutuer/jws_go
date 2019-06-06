package gvg

import (
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/jws/gamex/modules/player_msg"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type gvgModule struct {
	sid            uint
	world          *GVGWorld
	cmdChan        chan GVGCmd
	matchTimerChan <-chan time.Time
	timerChan      <-chan time.Time
	waitter        util.WaitGroupWrapper
	state          int
	battleRes      map[string]*battleRes
}

func (m *gvgModule) AfterStart(g *gin.Engine) {
	m.regETCD()
	g.POST(gvg_stop_url, func(c *gin.Context) {
		logs.Info("Get GVG stop info from multiplay")
		info := battleRes{}
		err := c.Bind(&info)
		if err != nil {
			c.String(400, err.Error())
			return
		}
		ret := m.CommandExec(GVGCmd{
			Typ:         Cmd_Typ_MutiplayEnd,
			MutiplayRet: &info,
		})
		if ret.ErrCode != nil {
			c.String(400, err.Error())
			return
		}
		c.String(200, "success")
	})
}

func (m *gvgModule) BeforeStop() {

}

func (m *gvgModule) Start() {
	m.battleRes = make(map[string]*battleRes, 100)
	m.cmdChan = make(chan GVGCmd, 2048)
	m.matchTimerChan = time.After(time.Second * MATCH_POLL_TIME)
	m.loadFromDB()
	mergeBalance(m.sid)
	m.waitter.Wrap(func() {
		timerChan := uutil.TimerSec.After(time.Second)
		needReset := true

		for {
			select {
			case command, ok := <-m.cmdChan:
				if !ok {
					logs.Warn("gvg cmdChan already closed")
					return
				}
				func() {
					defer logs.PanicCatcherWithInfo("gvg command fatal error")
					switch command.Typ {
					case Cmd_Typ_PrepareFight:
						m.prepareFight(&command)
					case Cmd_Typ_CancelMatch:
						m.cancelMatch(&command)
					case Cmd_Typ_EnterCity:
						m.enterCity(&command)
					case Cmd_Typ_EndFight:
						m.endFight(&command)
					case Cmd_Typ_LeaveCity:
						m.leaveCity(&command)
					case Cmd_Typ_MutiplayEnd:
						m.mutiplayEnd(&command)

					case Cmd_Typ_Get_GuildRank:
						m.getGuildRank(&command)
					case Cmd_Typ_Get_SelfGuildInfo:
						m.getSelfGuildInfo(&command)
					case Cmd_Typ_Get_PlayerInfo:
						m.getPlayerInfo(&command)
					case Cmd_Typ_Get_SelfGuildAllInfo:
						m.getSelfGuildAllInfo(&command)
					case Cmd_Typ_Get_CityLeader:
						m.getCityLeader(&command)
					case Cmd_Typ_Get_GuildWorldRank:
						m.getGuildWorldRank(&command)
					case Cmd_Typ_Get_PlayerWorldInfo:
						m.getPlayerWorldInfo(&command)
					case Cmd_Typ_Remove_Player:
						m.removePlayer(&command)
					case Cmd_Typ_Remove_Guild:
						m.removeGuild(&command)
					case Cmd_Typ_Rename_Guild:
						m.renameGuild(&command)
					case Cmd_Type_Rename_Player:
						m.renamePlayer(&command)
					default:
						logs.Error("gvgModule switch default %d", command.Typ)
						command.retChan <- GVGRet{}
					}
				}()

			case <-m.matchTimerChan:
				m.matchFight()
				m.matchTimerChan = time.After(MATCH_POLL_TIME * time.Second)
			case <-timerChan:
				// 活动结束结算
				func() {
					defer logs.PanicCatcherWithInfo("gvg balance fatal error")
					now_t := m.GetNowTime()

					// 返回上次活动结束时间, 方便结算
					timeInfo, lastEndTime := gamedata.GetHotDatas().GvgConfig.GetGVGTime(m.sid, now_t)
					if m.isActive() {
						if now_t > lastEndTime && m.world.LastBalanceTime != lastEndTime {
							logs.Warn("Rank Balance")

							// 算名次
							m.rankWinner()
							// 结算
							m.balance()
							m.transScoreToGuild(timeInfo)
							gvgLogiclog(m.world)

							m.world.cleanWorld()

							m.saveToDB()

							m.world.LastBalanceTime = lastEndTime
						}
						// 日结算
						// Warn: 每日都会给城池里分数最高的军团发奖
						dayBalanceTime := util.DailyBeginUnixByStartTime(now_t,
							gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypCommon))
						if now_t > dayBalanceTime && dayBalanceTime != m.world.LastDayBalanceTime {
							m.world.LastDayBalanceTime = dayBalanceTime
							m.dayBalance()
							m.saveToDB()
						}

						// 军团仓库结算
						// Warn: 每日都会给城池里分数最高的军团发奖
						dayGuildBalanceTime := util.DailyBeginUnixByStartTime(now_t,
							gamedata.GetBeginTimeByTyp(gamedata.DailyStartTypGVGGuildGiftGet))
						if now_t > dayGuildBalanceTime && dayGuildBalanceTime != m.world.LastGuildDayBalanceTime {
							m.world.LastGuildDayBalanceTime = dayGuildBalanceTime
							m.dayGuildBalance()
							//删除的税收结算
							for _, value := range m.world.removeGuild {
								if value.Guildid != "" {
									guild.GetModule(m.sid).GVGBalanceForInventory(value.Guildid, value.Cityid, 1)
									logs.Debug("Send remove reward city:%d to guild:%d", value.Cityid, value.Guildid)
								}
							}
							m.saveToDB()
						}
						needReset = true

					} else {
						if needReset {
							needReset = false
							logs.Warn("Clean GVG Data by GMTools")
							gvgLogiclog2(m.world)
							m.reset()
						}
					}

					// 重置
					// Warn: 每日都会给城池里分数最高的军团发奖
					if now_t > timeInfo.GVGResetTime && timeInfo.GVGResetTime != m.world.LastResetTime {
						m.world.LastResetTime = timeInfo.GVGResetTime
						m.reset()
						// 通知上一任长安城主卸任
						logs.Debug("LastChanganLeader is %s", m.world.getLastChangAnLeader())
						if m.world.getLastChangAnLeader() != "" {
							logs.Warn("Send Title to leader")
							player_msg.Send(m.world.getLastChangAnLeader(), player_msg.PlayerMsgTitleCode,
								player_msg.DefaultMsg{})
							m.world.setLastChangAnLeader("")
						}
					}
				}()
				timerChan = uutil.TimerSec.After(time.Second)
			}
		}

	})

}

func (m *gvgModule) Stop() {
	close(m.cmdChan)
	m.waitter.Wait()
	m.saveToDB()
}

func (m *gvgModule) CommandExec(cmd GVGCmd) *GVGRet {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	cmd.retChan = make(chan GVGRet, 1)

	select {
	case m.cmdChan <- cmd:
	case <-ctx.Done():
		logs.Error("GVG cmdChan is full")
	}

	select {
	case ret := <-cmd.retChan:
		logs.Debug("CommandExec success")
		return &ret
	case <-ctx.Done():
		logs.Error("GVG CommandExec apply <-retChan timeout")
		return &GVGRet{}
	}
}

func (m *gvgModule) CommandExecAsync(cmd GVGCmd) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	cmd.retChan = make(chan GVGRet, 1)

	select {
	case m.cmdChan <- cmd:
	case <-ctx.Done():
		logs.Error("GVG cmdChan is full")
	}
}

func (m *gvgModule) saveToDB() {
	dbWorld := &Gvg2DB{}
	m.world.SaveDataToDB(dbWorld)
	logs.Warn("GVG save to DB: %v", *dbWorld)
	err := dbWorld.dbSave(m.sid)
	if err != nil {
		logs.Error("gvg save error: %v", err)
	}
}

func (m *gvgModule) loadFromDB() {
	dbWorld := &Gvg2DB{}
	err := dbWorld.dbLoad(m.sid)
	if err != nil {
		logs.Error("gvg load error: %v", err)
	}
	logs.Warn("GVG load from DB: %v", *dbWorld)
	m.world = &GVGWorld{}
	m.world.LoadDataFromDB(dbWorld)
}
