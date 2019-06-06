package merge

import (
	"fmt"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/modules/gvg"
	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func MergeGVG() error {
	ADB := GetADBConn()
	defer ADB.Close()
	BDB := GetBDBConn()
	defer BDB.Close()
	ResDB := GetResDBConn()
	defer ResDB.Close()

	ResGVG := &gvg.Gvg2DB{}
	ResGVG.InitPG()

	amerged, err := redis.Int(ADB.Do("EXISTS", gvg.TableGVGMerge(Cfg.AShardId)))
	if err != nil {
		logs.Error("MergeGVG EXISTS %s err %s", gvg.TableGVGMerge(Cfg.AShardId), err.Error())
		return err
	}
	bmerged, err := redis.Int(BDB.Do("EXISTS", gvg.TableGVGMerge(Cfg.BShardId)))
	if err != nil {
		logs.Error("MergeGVG EXISTS %s err %s", gvg.TableGVGMerge(Cfg.BShardId), err.Error())
		return err
	}
	if amerged > 0 && bmerged <= 0 || amerged <= 0 && bmerged > 0 {
		logs.Error("MergeGVG can't support 3 merge 1")
		return fmt.Errorf("MergeGVG can't support 3 merge 1")
	}
	if amerged > 0 {
		AGVGMerged := &gvg.Gvg2DB{}
		AGVGMerged.InitPG()
		err = driver.RestoreFromHashDB(ADB.RawConn(),
			gvg.TableGVGMerge(Cfg.AShardId), AGVGMerged, false, false)
		if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
			logs.Error("MergeGVG RestoreFromHashDB %s err %s", gvg.TableGVGMerge(Cfg.AShardId), err.Error())
			return err
		}

		BGVGMerged := &gvg.Gvg2DB{}
		BGVGMerged.InitPG()
		err = driver.RestoreFromHashDB(BDB.RawConn(),
			gvg.TableGVGMerge(Cfg.BShardId), BGVGMerged, false, false)
		if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
			logs.Error("MergeGVG RestoreFromHashDB %s err %s", gvg.TableGVGMerge(Cfg.BShardId), err.Error())
			return err
		}
		ResGVG.LastWorldInfo = AGVGMerged.LastWorldInfo
		ResGVG.CityInfo = make([]gvg.GvgCity2DB, 0, len(AGVGMerged.CityInfo)+len(BGVGMerged.CityInfo))
		ResGVG.CityInfo = append(ResGVG.CityInfo, AGVGMerged.CityInfo...)
		ResGVG.CityInfo = append(ResGVG.CityInfo, BGVGMerged.CityInfo...)
	} else {
		AGVG := &gvg.Gvg2DB{}
		err = driver.RestoreFromHashDB(ADB.RawConn(),
			gvg.TableGVG(Cfg.AShardId), AGVG, false, false)
		if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
			logs.Error("MergeGVG RestoreFromHashDB %s err %s", gvg.TableGVG(Cfg.AShardId), err.Error())
			return err
		}

		BGVG := &gvg.Gvg2DB{}
		err = driver.RestoreFromHashDB(BDB.RawConn(),
			gvg.TableGVG(Cfg.BShardId), BGVG, false, false)
		if err != nil && err != driver.RESTORE_ERR_Profile_No_Data {
			logs.Error("MergeGVG RestoreFromHashDB %s err %s", gvg.TableGVG(Cfg.BShardId), err.Error())
			return err
		}

		ResGVG.LastWorldInfo = AGVG.LastWorldInfo
		ResGVG.CityInfo = make([]gvg.GvgCity2DB, 0, len(AGVG.CityInfo)+len(BGVG.CityInfo))
		ResGVG.CityInfo = append(ResGVG.CityInfo, AGVG.CityInfo...)
		ResGVG.CityInfo = append(ResGVG.CityInfo, BGVG.CityInfo...)
	}

	cb := redis.NewCmdBuffer()
	if err := driver.DumpToHashDBCmcBuffer(cb, gvg.TableGVGMerge(Cfg.ResShardId), ResGVG); err != nil {
		return fmt.Errorf("MergeGVG DumpToHashDBCmcBuffer err %v", err)
	}

	if _, err := ResDB.DoCmdBuffer(cb, true); err != nil {
		logs.Error("MergeGVG save res err %s", err.Error())
		return err
	}
	return nil
}
