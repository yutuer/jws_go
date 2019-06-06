package uutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/cloud_db"
	"vcs.taiyouxi.net/platform/planx/util/cloud_db/ossdb"
	"vcs.taiyouxi.net/platform/planx/util/cloud_db/s3db"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/version"
)

var (
	cloudDb cloud_db.CloudDb
)

func InitCloudDb() error {
	cloudDb = getCloudDb()
	if err := cloudDb.Open(); err != nil {
		logs.Error("DataHotUpdateModule s3.Open err %s", err.Error())
		return err
	}
	return nil
}

func getCloudDb() cloud_db.CloudDb {
	if game.Cfg.CloudDbDriver == "S3" {
		return s3db.NewStoreS3(game.Cfg.AWS_Region, "",
			game.Cfg.AWS_AccessKey, game.Cfg.AWS_SecretKey, "", "")
	} else if game.Cfg.CloudDbDriver == "OSS" {
		return ossdb.NewStoreOSS(game.Cfg.OSSEndpoint, game.Cfg.OSSAccessId,
			game.Cfg.OSSAccessKey, "", "", "")
	}
	return nil
}

func LoadHotData2LocalFromS3(bucket, build, hotAbsPath string) error {
	// 清理本地目录
	logs.Info("LoadHotData2LocalFromS3 newPath %s", hotAbsPath)
	os.RemoveAll(hotAbsPath)
	if err := os.MkdirAll(hotAbsPath, os.ModePerm); err != nil {
		logs.Error("LoadHotData2LocalFromS3 MkdirAll %s err %s", hotAbsPath, err.Error())
		return err
	}
	// s3上获取所有文件
	prefix := fmt.Sprintf("%d/%s/%s/", game.Cfg.Gid, version.Version, build)
	logs.Info("LoadHotData2LocalFromS3 prepare List files %s %s", bucket, prefix)
	files, err := cloudDb.ListObjectWithBucket(bucket, prefix, 1000)
	if err != nil {
		logs.Error("LoadHotData2LocalFromS3 ListObjectWithBucket err %s", err.Error())
		return err
	}
	logs.Info("LoadHotData2LocalFromS3 prepare update files %v", files)
	for _, f := range files {
		fc, err := cloudDb.GetWithBucket(bucket, f)
		if err != nil {
			logs.Error("LoadHotData2LocalFromS3 GetWithBucket err %s", err.Error())
			return err
		}
		fn := filepath.Join(hotAbsPath, strings.TrimLeft(f, prefix))
		logs.Info("DataHotUpdateModule write file %s", fn)
		fh, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			logs.Error("LoadHotData2LocalFromS3 OpenFile failed %s", err.Error())
			return err
		}
		err = func() error {
			defer fh.Close()
			if _, err := fh.Write(fc); err != nil {
				logs.Error("LoadHotData2LocalFromS3 filewrite failed %s", err.Error())
				return err
			}
			return nil
		}()
		if err != nil {
			logs.Error("LoadHotData2LocalFromS3 File Write failed %s", err.Error())
			return err
		}
	}
	return nil
}
