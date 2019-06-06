package market_activity

import (
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func (ma *MarketActivityModule) NotifyHotDataUpdate(ver *gamedata.DataVerConf) {
	logs.Debug("[MarketActivityModule] NotifyHotDataUpdate")
	cmd := makeMarketCommand(Command_HotDataUpdate, marketCommandParam{HotDataVer: ver})
	ma.commandExecAsyn(cmd)
}

func (ma *MarketActivityModule) cmdHotDataUpdate(cmd *marketCommand) {
	logs.Debug("[MarketActivityModule] cmdHotDataUpdate execute")

	ma.maTimeSet.reloadHotData()
}

func (mt *MarketTimeSet) reloadHotData() {
	now_t := time.Now().Unix()

	//检查热数据版本变化是否涉及模块的更新
	activityCfg := gamedata.GetHotDatas().Activity
	for pid, ids := range ActivityList {
		pcfgs := activityCfg.GetActivitySimpleInfo(pid)
		if 0 == len(pcfgs) {
			continue
		}
		pcfg := pcfgs[0]

		//如果父节点失效, 清楚所有的触发器
		if pcfg.EndTime < now_t || pcfg.Cfg.GetActivityValid() <= 0 {
			for _, id := range ids {
				mt.freeTrigger(id)
			}
		} else {
			//先通知父节点ID, 可能会清除老数据, 这个接口只能在线程串化队列中调用,否则有并发问题(可能)
			mt.ma.notifyRankParentID(pid, pcfg.ActivityId)
			for _, id := range ids {
				cfgs := activityCfg.GetActivitySimpleInfo(id)

				//有效数据触发刷新, 无效数据释放触发器
				if 0 != len(cfgs) && cfgs[0].Cfg.GetActivityValid() > 0 {
					mt.refresh(cfgs[0], now_t)
				} else {
					mt.freeTrigger(id)
				}
			}
		}
	}
}

type MarketTimeSet struct {
	ma        *MarketActivityModule
	setStatus map[uint32]MarketTimeStatus
}

type MarketTimeStatus struct {
	startTime    int64
	endTime      int64
	activityType uint32
	activityID   uint32

	triggerTimer *time.Timer
}

func (mt *MarketTimeSet) refresh(cfg *gamedata.HotActivityInfo, now_t int64) {
	activity := cfg.ActivityType

	//判断信息是否被更新
	oldStatus, exist := mt.setStatus[activity]
	if true == exist {
		//如果未被更新, 返回
		if oldStatus.activityID == cfg.ActivityId && oldStatus.activityType == cfg.ActivityType && oldStatus.endTime == cfg.EndTime {
			return
		}

		//如果被更新, 先清除旧触发器
		if nil != oldStatus.triggerTimer {
			oldStatus.triggerTimer.Stop()
		}
	}

	//构造新的状态
	status := MarketTimeStatus{
		startTime:    cfg.StartTime,
		endTime:      cfg.EndTime,
		activityType: cfg.ActivityType,
		activityID:   cfg.ActivityId,
	}

	//构造新的触发器
	status.triggerTimer = time.AfterFunc(
		time.Duration(cfg.EndTime-now_t)*time.Second,
		func() {
			mt.handleTrigger(&status)
		})

	mt.setStatus[activity] = status
	logs.Debug("[MarketActivityModule] refresh activity [%d], status [%v]", activity, mt.setStatus[activity])
}

func (mt *MarketTimeSet) handleTrigger(status *MarketTimeStatus) {
	logs.Debug("[MarketActivityModule] trigger activity [%d], status [%v]", status.activityType, status)
	if false == mt.ma.maRank.checkHadSnapShoot(status.activityType) {
		mt.ma.notifyMakeSnapShoot(status.activityType, status.activityID)
		mt.ma.notifySendReward(status.activityType, status.activityID)
	}
}

func (mt *MarketTimeSet) freeTrigger(activity uint32) {
	st, exist := mt.setStatus[activity]
	if false == exist || nil == st.triggerTimer {
		return
	}
	st.triggerTimer.Stop()
	delete(mt.setStatus, activity)
}
