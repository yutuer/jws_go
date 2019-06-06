package accountWarm

import "vcs.taiyouxi.net/jws/gamex/models/driver"

type dbToWarm interface {
	DBName() string
	DBLoad(logInfo bool) error
}

func TryWarmData(isLog bool, shareID uint, dbprofile dbToWarm, isHasReg bool) error {
	err := dbprofile.DBLoad(isLog)
	if err == nil {
		return nil
	}

	//if isHasReg && err == driver.RESTORE_ERR_Profile_No_Data {
	//	// 尝试读取冷数据
	//	err := warm.Get(shareID).WarmKey(dbprofile.DBName())
	//	if err != nil {
	//		//TODO By Fanyang 改为有错就断开连接
	//		logs.Warn("WarmKey Err By %s", err.Error())
	//	}
	//	return dbprofile.DBLoad(isLog)
	//}

	return err
}

func PanicIfErr(err error) bool {
	if err == nil {
		return false
	}

	if err == driver.RESTORE_ERR_Profile_No_Data {
		return true
	}

	//NewAccount数据加载后如何处理玩家数据加载错误,
	//如果玩家数据不存在是不会引发错误的。这里的错误应该是数据库自身的错误。
	//此外此函数是通过GetMux|Player playerProcessor调用，是用户goroutine级别的错误，不会影响其他玩家
	panic(err)
	return false
}
