package logiclog

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/timesking/seelog"

	"time"

	"strings"

	"path/filepath"

	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/eslogger"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	txlumberjack "vcs.taiyouxi.net/platform/planx/util/lumberjack.v2"
)

const (
	writer_buff_size = 32768
	gameId           = "134"
)

type heroLog struct {
	logger         *txlumberjack.Logger
	loggerCurrency *txlumberjack.Logger
}

var hl *heroLog

func CloseHeroLog() {
	if hl != nil {
		hl.Close()
	}
}

func init() {
	hl = &heroLog{}
	seelog.RegisterReceiver("herologiclog", hl)
}

func (heroLog *heroLog) AfterParse(initArgs seelog.CustomReceiverInitArgs) error {
	filename := "hero.log"
	if initArgs.XmlCustomAttrs != nil && len(initArgs.XmlCustomAttrs) > 0 {
		filename = initArgs.XmlCustomAttrs["filename"]
	}

	heroLog.logger = &txlumberjack.Logger{
		FileTempletName: filename,
		MaxSize:         10000, // 10g
		BufSize:         32768, // 30k
		TimeLocal:       "Asia/Shanghai",
		GetUTCSec:       func() int64 { return time.Now().Unix() },
	}

	// herocurrencylog
	ext := filepath.Ext(filename)
	filename = filename[:len(filename)-len(ext)] + "_currency" + ext
	heroLog.loggerCurrency = &txlumberjack.Logger{
		FileTempletName: filename,
		MaxSize:         10000, // 10g
		BufSize:         32768, // 30k
		TimeLocal:       "Asia/Shanghai",
		GetUTCSec:       func() int64 { return time.Now().Unix() },
	}
	return nil
}

func (heroLog *heroLog) ReceiveMessage(message string, level seelog.LogLevel, context seelog.LogContextInterface) error {
	if level != seelog.ErrorLvl {
		return nil
	}
	var log eslogger.ESLoggerInfo
	err := json.Unmarshal([]byte(message), &log)
	if err != nil {
		logs.Error("reading standard input json.Unmarshal err %v line %s", err, message)
		return err
	}
	ok, _ := regexp.Match(BITag+".*", []byte(log.Extra))
	if !ok {
		return nil
	}
	var accountNameId, gidSid string
	var shard string
	strChan := game.Gid2Channel[game.Cfg.Gid] // channelid
	if log.AccountID != "" {
		ass := strings.Split(log.AccountID, ":")
		shard = ass[1]
		accountNameId = fmt.Sprintf("%s:%s", ass[0], ass[2]) // 账号ID
		strChan := game.Gid2Channel[game.Cfg.Gid]
		gidSid = fmt.Sprintf("%s%04s%06s", strChan, gameId, ass[1])
	} else if log.GuildID != "" {
		ass := strings.Split(log.GuildID, ":")
		shard = ass[1]
		gidSid = fmt.Sprintf("%s%04s%06s", strChan, gameId, ass[1])
	}

	var resLine string
	var resCurrency string

	switch log.Type {
	case LogicTag_Login:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_Login{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_Login)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$ACCOUNT_LOGIN$$区服ID$$账号ID$$设备ID$$账号名称$$账号当前元宝余额$$登录IP
		resLine = fmt.Sprintf("%s$$ACCOUNT_LOGIN$$%s$$%s$$%s$$%s$$%d$$%s\n", log.TimeUTC8,
			gidSid, accountNameId, info.DeviceId, info.AccountName, info.HcBuy, info.Ip)
		if info.IsReg {
			// 时间$$ACCOUNT_REGISTER$$区服ID$$账号ID$$设备ID$$账号名称$$客户端版本号$$客户端类型$$检测到的客户端手机号
			resLine += fmt.Sprintf("%s$$ACCOUNT_REGISTER$$%s$$%s$$%s$$%s$$%s$$%s$$%s\n", log.TimeUTC8,
				gidSid, accountNameId, info.DeviceId, info.AccountName, info.ClientVer, info.MachineType, info.PhoneNum)
		}
		profileInfo, _ := json.Marshal(&info.ProfileInfo)
		// 时间$$ROLE_LOGIN$$区服ID$$账号ID$$角色ID$$账户名$$账号当前余额$$设备ID$$角色名称$$角色信息$$登录IP
		resLine += fmt.Sprintf("%s$$ROLE_LOGIN$$%s$$%s$$%s$$%s$$%d$$%s$$%s$$%s$$%s", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.AccountName, info.HcBuy, info.DeviceId, info.ProfileName, profileInfo, info.Ip)
	case LogicTag_CreateRole:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_CreateProfile{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_CreateProfile)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$CREATE_ROLE$$区服ID$$账号ID$$角色ID$$角色名称$$性别
		resLine = fmt.Sprintf("%s$$CREATE_ROLE$$%s$$%s$$%s$$%s$$%s", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.Name, "")
	case LogicTag_Logout:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_LogOut{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_LogOut)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$ROLE_LOGOUT$$区服ID$$账号ID$$角色ID$$在线时长（秒）
		resLine = fmt.Sprintf("%s$$ROLE_LOGOUT$$%s$$%s$$%s$$%d", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.OnLineTime)
	case LogicTag_GiveItem:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_GiveItem{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_GiveItem)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err %s %v", log.Type, _log.Info)
			return err
		}
		strItems, _ := json.Marshal(info.Items)
		// 时间$$GET_ITEM$$区服ID$$账号ID$$角色ID$$获得物品$$原因$$vip等级
		resLine = fmt.Sprintf("%s$$GET_ITEM$$%s$$%s$$%s$$%s$$%s$$%d", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, string(strItems), info.Reason, info.VIP)
	case LogicTag_CostItem:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_CostItem{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_CostItem)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err %s %v", log.Type, _log.Info)
			return err
		}
		strItemss := make([]string, len(info.Items))
		for id, c := range info.Items {
			strItemss = append(strItemss, fmt.Sprintf("%s,%d", id, c))
		}
		strItems := strings.Join(strItemss, ";")
		// 时间$$REMOVE_ITEM$$区服ID$$账号ID $$角色ID$$扣除物品$$原因$$vip等级
		resLine = fmt.Sprintf("%s$$REMOVE_ITEM$$%s$$%s$$%s$$%s$$%s$$%d", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, string(strItems), info.Reason, info.VIP)
	case LogicTag_GiveCurrency:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_GiveCurrency{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_GiveCurrency)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$GET_MONEY$$区服ID$$账号ID$$角色ID$$获得之前的数值$$获得之后的数值$$代币类型$$原因$$vip等级
		resLine = fmt.Sprintf("%s$$GET_MONEY$$%s$$%s$$%s$$%d$$%d$$%s$$%s$$%d", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.BefValue, info.AftValue, info.Type, info.Reason, info.VIP)
		//$$游戏id$$渠道id$$区服id$$ 平台id$$cp账户id$$角色id $$币种$$消耗的免费一级货币 0 $$消耗的付费赠送的一级货币 0
		// $$消耗的付费获得的一级货币 0 $$原因任务:1, 活动:2, 充值:3 $$道具id $$道具数量 $$道具有效期
		//$$道具类型 $$角色等级 $$角色vip等级 $$ip $$操作时间 $$事件名称
		payTime2, _ := time.Parse("2006-01-02 15:04:05", log.TimeUTC8)
		payTime3 := payTime2.Format("20060102150405")
		var gameidx string
		if uutil.IsVNVer() {
			gameidx = "10013"
		} else {
			gameidx = "102"
		}

		if info.Type == "VI_HC_Buy" {
			resCurrency = fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\t"+
				"%s\t%d\t%s\t%s\t%d\t%s\t%s\t%d\t%d\t%s\t%s\titemobtain", gameidx, info.Channel, shard,
				info.Platform, log.AccountID, info.Avatar, "CNY", "0", "0", info.Value,
				info.Reason, info.Type, info.Value, "-1", info.ItemType, info.CorpLvl,
				info.VIP, info.Ip, payTime3)

		} else if info.Type == "VI_HC_Give" {
			resCurrency = fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\t"+
				"%d\t%s\t%s\t%s\t%d\t%s\t%s\t%d\t%d\t%s\t%s\titemobtain", gameidx, info.Channel, shard,
				info.Platform, log.AccountID, info.Avatar, "CNY", "0", info.Value, "0",
				info.Reason, info.Type, info.Value, "-1", info.ItemType, info.CorpLvl,
				info.VIP, info.Ip, payTime3)

		} else {
			resCurrency = fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\t"+
				"%s\t%s\t%s\t%s\t%d\t%s\t%s\t%d\t%d\t%s\t%s\titemobtain", gameidx, info.Channel, shard,
				info.Platform, log.AccountID, info.Avatar, "CNY", "0", "0", "0",
				info.Reason, info.Type, info.Value, "-1", info.ItemType, info.CorpLvl,
				info.VIP, info.Ip, payTime3)

		}

	case LogicTag_CostCurrency:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_CostCurrency{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_CostCurrency)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$REMOVE_MONEY$$区服ID$$账号ID$$角色ID$$扣除之前的数值$$扣除之后的数值$$代币类型$$原因
		resLine = fmt.Sprintf("%s$$REMOVE_MONEY$$%s$$%s$$%s$$%d$$%d$$%s$$%s", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.BefValue, info.AftValue, info.Type, info.Reason)
	case LogicTag_CorpExpChg:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_CorpExpChg{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_CorpExpChg)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$EXP_CHANGE$$区服ID$$账号ID$$角色ID$$获得之前的数值$$获得之后的数值$$原因
		resLine = fmt.Sprintf("%s$$EXP_CHANGE$$%s$$%s$$%s$$%d$$%d$$%s", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.BefValue, info.AftValue, info.Reason)
	case LogicTag_CorpLevelChg:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_CorpLevelChg{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_CorpLevelChg)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$LEVEL_CHANGE$$区服ID$$账号ID$$角色ID$$升级前级别$$升级前经验$$升级后级别$$升级后经验
		resLine = fmt.Sprintf("%s$$LEVEL_CHANGE$$%s$$%s$$%s$$%d$$%d$$%d$$%d", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.BefLevel, info.BefExp, info.AftLevel, info.AftExp)
	case LogicTag_QuestFinish:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_QuestFinish{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_QuestFinish)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		str_re := strings.Join(info.Rewards, ";")
		// 时间$$FINISH_QUEST$$区服ID$$账号ID$$角色ID$$任务ID$$奖励
		resLine = fmt.Sprintf("%s$$FINISH_QUEST$$%s$$%s$$%s$$%d$$%s", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.QuestId, str_re)
	case LogicTag_StageFinish:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_StageFinish{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_StageFinish)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		isSweep := 0
		if info.IsSweep {
			isSweep = 1
		}
		// 时间$$PVE_INSTANCE$$区服ID$$账号ID$$角色ID$$副本ID$$副本名称$$是否胜利$$次数$$是否扫荡$$消耗时间$$战队等级$$战力$$出战神将
		resLine = fmt.Sprintf("%s$$PVE_INSTANCE$$%s$$%s$$%s$$%s$$%s$$%d$$%d$$%d$$%d$$%d$$%d$$%v", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.StageId, info.StageId, info.IsWin, info.Times, isSweep,
			info.CostTime, info.CorpLvl, info.GS, info.SkillGenerals)
	case LogicTag_Pvp:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_Pvp{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_Pvp)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$PVP_INSTANCE$$区服ID$$账号ID$$角色ID$$是否胜利$$角色类型$$GS$$之前的积分$$之后的积分$$之前的排位$$之后的排位$$敌人ID$$敌人角色类型$$敌人的GS$$敌人之前的积分$$敌人之后的积分$$敌人之前的排位$$敌人之后的排位
		resLine = fmt.Sprintf("%s$$PVP_INSTANCE$$%s$$%s$$%s$$%d$$%s$$%d$$%d$$%d$$%d$$%d$$%s$$%d$$%d$$%d$$%d$$%d$$%d",
			log.TimeUTC8, gidSid, accountNameId, log.AccountID,
			info.IsWin, log.Avatar, info.MyGs, info.MyBefScore, info.MyAftScore, info.MyBefPos, info.MyAftPos,
			info.EnemyId, info.EnemyAvatar, info.EnemyGs, info.EnemyBefScore, info.EnemyAftScore, info.EnemyBefPos, info.EnemyAftPos)
	case LogicTag_StoreBuy:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_StoreBuy{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_StoreBuy)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$SHOPBUY$$区服ID$$账号ID$$角色ID$$物品ID$$物品数量$$代币类型$$代币数量$$商店类型
		resLine = fmt.Sprintf("%s$$SHOPBUY$$%s$$%s$$%s$$%s$$%d$$%s$$%d$$%s", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.ItemId, info.ItemCount, info.CoinType, info.CoinCount, info.StoreType)
	case LogicTag_ShopBuy:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_StoreBuy{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_StoreBuy)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$SHOPBUY$$区服ID$$账号ID$$角色ID$$物品ID$$物品数量$$花费代币类型$$花费代币数量$$商城类型
		resLine = fmt.Sprintf("%s$$SHOPBUY$$%s$$%s$$%s$$%s$$%d$$%s$$%d$$%s", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.ItemId, info.ItemCount, info.CoinType, info.CoinCount, info.StoreType)
	case LogicTag_Tutorial:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_Tutorial{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: err %v type %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_Tutorial)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$NEWG$$区服ID$$账号ID $$角色ID$$引导类型id$$步骤N
		resLine = fmt.Sprintf("%s$$NEWG$$%s$$%s$$%s$$$$%s", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.Step)
	case LogicTag_Gacha:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_Gacha{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: %v %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_Gacha)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		strItems, _ := json.Marshal(info.Items)
		// 时间$$CHOUJIANG$$区服ID$$账号ID$$角色ID$$抽奖类型$$获得物品$$花费代币类型$$扣除代币数
		resLine = fmt.Sprintf("%s$$CHOUJIANG$$%s$$%s$$%s$$%s$$%s$$%s$$%d", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.GachaType, strItems, info.CoinType, info.CoinCount)
	case LogicTag_GeneralAddNum:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_GeneralAddNum{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: %v %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_GeneralAddNum)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$GeneralAddNum$$区服ID$$账号ID$$角色ID$$副将id$$副将碎片数量$$原因
		resLine = fmt.Sprintf("%s$$GeneralAddNum$$%s$$%s$$%d$$%s", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.GeneralId, info.Count, info.Reason)
	case LogicTag_GeneralStarLvlUp:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_GeneralStar{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: %v %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_GeneralStar)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$GeneralStarLvlUp$$区服ID$$账号ID$$角色ID$$副将id$$副将星级$$原因
		resLine = fmt.Sprintf("%s$$GeneralStarLvlUp$$%s$$%s$$%d$$%s", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.GeneralId, info.Star_Aft, info.Reason)
	case LogicTag_GeneralRelLvlUp:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_GeneralRel{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: %v %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_GeneralRel)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$GeneralRelLvlUp$$区服ID$$账号ID$$角色ID$$羁绊id$$羁绊等级$$原因
		resLine = fmt.Sprintf("%s$$GeneralRelLvlUp$$%s$$%s$$%d$$%s", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.Relation, info.Level, info.Reason)
	case LogicTag_GuildCreate:
	case LogicTag_GuildAddMem:
	case LogicTag_GuildDelMem:
	case LogicTag_GuildDismiss:
	case LogicTag_GuildPosChg:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_Guild{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: %v %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_Guild)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		memsJson, err := json.Marshal(info.Mems)
		if err != nil {
			logs.Error("guild mem marshal err %v", err)
			return err
		}
		if log.Type != LogicTag_GuildPosChg {
			// 时间$$GuildCreate$$区服ID$$公会ID$$公会序号$$公会名称$$公会等级$$公会成员数量$$公会成员信息$$当事人ID
			resLine = fmt.Sprintf("%s$$%s$$%s$$%s$$%d$$%s$$%d$$%d$$%s$$%s", log.TimeUTC8,
				log.Type, gidSid, info.GuildUUID, info.GuildID, info.Name, info.Level, info.MemNum, memsJson, info.Acid)
		} else {
			// 时间$$GuildPosChg$$区服ID$$公会ID$$公会序号$$公会名称$$公会等级$$公会成员数量$$公会成员信息$$职位变动公会成员ID$$变动前职位$$变动后职位
			resLine = fmt.Sprintf("%s$$%s$$%s$$%s$$%d$$%s$$%d$$%d$$%s$$%s$$%s$$%s", log.TimeUTC8, log.Type,
				gidSid, info.GuildUUID, info.GuildID, info.Name, info.Level, info.MemNum, memsJson, info.Acid, info.BefPos, info.AftPos)
		}
	case LogicTag_GuildGateEnemyOver:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_GuildGateEnemy{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: %v %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_GuildGateEnemy)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$GuildGateEnemyOver$$区服ID$$公会ID$$公会序号$$公会名称$$活动参加人数$$总获得积分
		resLine = fmt.Sprintf("%s$$GuildGateEnemyOver$$%s$$%s$$%d$$%s$$%d$$%d", log.TimeUTC8,
			gidSid, info.GuildUUID, info.GuildID,
			info.Name, info.MemJoinCount, info.Point)
	case LogicTag_RedeemCode:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_RedeemCode{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: %v %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_RedeemCode)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		isNoLimit := 0
		if info.RedeemCode_IsNoLimit {
			isNoLimit = 1
		}
		// 时间$$RedeemCode$$区服ID$$账号ID$$角色ID$$角色名称$$礼品码$$批次$$是否有限制
		resLine = fmt.Sprintf("%s$$RedeemCode$$%s$$%s$$%s$$%s$$%s$$%d$$%d", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.Name, info.RedeemCode, info.RedeemCode_BatchId, isNoLimit)
	case LogicTag_TrialLvlFinish:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_TrialLvl{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: %v %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_TrialLvl)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$ TrialLvlFinish$$区服ID$$账号ID$$角色ID$$关卡id$$是否胜利$$战力$$消耗时间$$出战神将
		resLine = fmt.Sprintf("%s$$TrialLvlFinish$$%s$$%s$$%s$$%d$$%d$$%d$$%d$$%v", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.LevelId, info.IsWin, info.Gs, info.CostTime, info.SkillGenerals)
	case LogicTag_TrialReset:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_TrialReset{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: %v %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_TrialReset)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$ TrialReset$$区服ID$$账号ID$$角色ID$$最远关id
		resLine = fmt.Sprintf("%s$$TrialReset$$%s$$%s$$%s$$%d", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.MostLvl)
	case LogicTag_TrialSweep:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_TrialSweep{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: %v %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_TrialSweep)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$ TrialSweep $$区服ID$$账号ID $$角色ID$$扫荡事件
		resLine = fmt.Sprintf("%s$$TrialSweep$$%s$$%s$$%s$$%s", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.SweepEvent)
	case LogicTag_Phone:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_Phone{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: %v %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_Phone)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		// 时间$$ Phone$$区服ID$$账号ID $$角色ID$$角色名$$手机号
		resLine = fmt.Sprintf("%s$$Phone$$%s$$%s$$%s$$%s$$%s", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.Name, info.Phone)
	case LogicTag_HotActivityAward:
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_HotActivityAward{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: %v %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_HotActivityAward)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		//时间$$FINISH_ACTIVITY$$区服ID$$账号ID$$角色ID$$主任务ID$$奖励$$任务类型$$子任务id
		resLine = fmt.Sprintf("%s$$FINISH_ACTIVITY%s$$%s$$%s$$%d$$%v$$%d$$%d", log.TimeUTC8,
			gidSid, accountNameId, log.AccountID, info.Activityid, info.AwardItem, info.ActivityType, info.SubActivityId)
	case LogicTag_IAP: //playercharger,goldobtain
		_log := eslogger.ESLoggerInfo{Info: &LogicInfo_IAP{}}
		err := json.Unmarshal([]byte(message), &_log)
		if err != nil {
			logs.Error("hero ReceiveMessage type unmarshal: %v %s", err, log.Type)
			return err
		}
		info, ok := _log.Info.(*LogicInfo_IAP)
		if !ok {
			logs.Error("hero ReceiveMessage type cast err : %s %v", log.Type, _log.Info)
			return err
		}
		HeroTime2, _ := time.Parse("2006-01-02 15:04:05", log.TimeUTC8)
		HeroTime3 := HeroTime2.Format("20060102150405")
		//$$游戏id$$渠道id $$区服id$$ 平台id $$cp账户id $$角色id $$订单id $$第三方订单id
		// $$订单状态  1 成功 $$订单时间 $$订单成交时间$$订单金额 $$币种 $$付费获得的一级货币
		// $$付费赠送的一级货币 $$货币类型   0 钻石 $$档位类型 $$角色名称 $$角色等 $$角色vip等级$$账户总付费金额$$IP$$事件名称
		if uutil.IsVNVer() {
			resCurrency = fmt.Sprintf("10013\t%s\t%s\t%s\t%s\t%d\t%s\t%s\t1\t%s\t%s\t%d\tCNY\t%d\t%d\t0"+
				"\t%d\t%s\t%d\t%d\t%d\t%s\tplayercharger",
				info.Channel, shard,
				info.Platform, log.AccountID, info.Avatar, info.Order,
				info.GameOrderId, HeroTime3, info.PayTime, info.Money, info.HcBuy,
				info.HCGive, info.GoodIdx, info.Name, info.CorpLvl, info.VIP,
				info.Moneysum, info.Ip)

			//$$游戏id $$渠道id $$区服id$$ 平台id $$cp账户id $$角色id$$ 付费金额 $$币种 $$原因$$子原因
			//$$免费获得的一级货币增量 $$付费赠送的一级货币增量 $$付费获得的一级货币增量
			//$$当前免费获得的一级货币数量 $$当前付费赠送的一级货币数量 $$当前付费获得的一级货币数量
			//$$角色名称 $$角色等级 $$角色vip等级 $$账户总付费金额$$IP$$事件名称
			resCurrency1 := fmt.Sprintf("10013\t%s\t%s\t%s\t%s\t%d\t%d\tCNY\t3\t"+
				"%s\t%d\t%d\t%d\t%d\t%s\t%d\t%s\t%d\t%d\t%d\t%s\t%s\tgoldobtain",
				info.Channel, shard,
				info.Platform, log.AccountID, info.Avatar, info.Money,
				"0", info.HcBuy, info.HCGive, info.HasHcCp, info.HasHcGive,
				"0", info.HasHcBuy, info.Name, info.CorpLvl, info.VIP,
				info.Moneysum, info.Ip, info.PayTime)
			if len(resCurrency1) > 0 {
				if _, err := heroLog.loggerCurrency.Write([]byte(resCurrency1 + "\n")); err != nil {
					logs.Error("hero ReceiveMessage write %v", err)
					return err
				}
			}
		} else {

			resCurrency = fmt.Sprintf("102\t%s\t%s\t%s\t%s\t%d\t%s\t%s\t1\t%s\t%s\t%d\tCNY\t%d\t%d\t0"+
				"\t%d\t%s\t%d\t%d\t%d\t%s\tplayercharger",
				info.Channel, shard,
				info.Platform, log.AccountID, info.Avatar, info.Order,
				info.GameOrderId, HeroTime3, info.PayTime, info.Money, info.HcBuy,
				info.HCGive, info.GoodIdx, info.Name, info.CorpLvl, info.VIP,
				info.Moneysum, info.Ip)

			//$$游戏id $$渠道id $$区服id$$ 平台id $$cp账户id $$角色id$$ 付费金额 $$币种 $$原因$$子原因
			//$$免费获得的一级货币增量 $$付费赠送的一级货币增量 $$付费获得的一级货币增量
			//$$当前免费获得的一级货币数量 $$当前付费赠送的一级货币数量 $$当前付费获得的一级货币数量
			//$$角色名称 $$角色等级 $$角色vip等级 $$账户总付费金额$$IP$$事件名称
			resCurrency1 := fmt.Sprintf("102\t%s\t%s\t%s\t%s\t%d\t%d\tCNY\t3\t"+
				"%s\t%d\t%d\t%d\t%d\t%s\t%d\t%s\t%d\t%d\t%d\t%s\t%s\tgoldobtain",
				info.Channel, shard,
				info.Platform, log.AccountID, info.Avatar, info.Money,
				"0", info.HcBuy, info.HCGive, info.HasHcCp, info.HasHcGive,
				"0", info.HasHcBuy, info.Name, info.CorpLvl, info.VIP,
				info.Moneysum, info.Ip, info.PayTime)
			if len(resCurrency1) > 0 {
				if _, err := heroLog.loggerCurrency.Write([]byte(resCurrency1 + "\n")); err != nil {
					logs.Error("hero ReceiveMessage write %v", err)
					return err
				}
			}

		}

	}

	if len(resLine) > 0 {
		if _, err := heroLog.logger.Write([]byte(resLine + "\n")); err != nil {
			logs.Error("hero ReceiveMessage write %v", err)
			return err
		}
	}
	if len(resCurrency) > 0 {
		if _, err := heroLog.loggerCurrency.Write([]byte(resCurrency + "\n")); err != nil {
			logs.Error("hero ReceiveMessage write %v", err)
			return err
		}

	}

	return nil
}

func (heroLog *heroLog) Flush() {

}

func (heroLog *heroLog) Close() error {
	err := heroLog.logger.Close()
	if err != nil {
		logs.Error("heroLog.logger.Close() err %v", err)
	}
	err1 := heroLog.loggerCurrency.Close()
	if err1 != nil {
		logs.Error("heroLog.loggerCurrency.Close() err %v", err1)
	}
	if err != nil {
		return err
	}
	return err1
}

func IsVietnamese(channel string) bool {
	return channel == "5003" || channel == "5004"
}
