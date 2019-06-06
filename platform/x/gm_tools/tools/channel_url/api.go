package channel_url

import (
	"crypto/md5"
	"fmt"

	"encoding/json"
	"strings"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	"vcs.taiyouxi.net/platform/x/gm_tools/config"
)

var (
	s3       storehelper.IStore
	_csv_bb  []byte
	_meta_bb []byte
)

func RegCommands() {
	gm_command.AddGmCommandHandle("getChannelUrl", getRemoteChannelUrl)
	gm_command.AddGmCommandHandle("updateChannelUrl", updateRemoteChannelUrl)
}

type res struct {
	Ret     string `json:"ret"`
	CsvInfo string `json:"csv"`
}

func Init() error {
	if config.Cfg.StoreDriver == storehelper.S3 {
		s3 = storehelper.NewStore(storehelper.S3, storehelper.S3Cfg{
			Region:           config.Cfg.AWS_Region,
			BucketsVerUpdate: config.Cfg.S3_Buckets_VerUpdate,
			AccessKeyId:      config.Cfg.AWS_AccessKey,
			SecretAccessKey:  config.Cfg.AWS_SecretKey,
		})
	} else if config.Cfg.StoreDriver == storehelper.OSS {
		s3 = storehelper.NewStore(storehelper.OSS, storehelper.OSSInitCfg{
			EndPoint:  config.Cfg.OSSEndpoint,
			AccessId:  config.Cfg.OSSAccessId,
			AccessKey: config.Cfg.OSSAccessKey,
			Bucket:    config.Cfg.OSSChannelBucket,
		})
	}
	s3.Open()
	return nil
}

func getRemoteChannelUrl(c *gm_command.Context, server, accountid string, params []string) error {
	csv_bb, err := s3.Get(config.Cfg.VerUpdate_FileName)
	if err != nil {
		return err
	}
	meta_bb, err := s3.Get(config.Cfg.VerUpdate_Meta_FileName)
	if err != nil {
		return err
	}
	if len(csv_bb) <= 0 {
		ret := res{"ok", ""}
		b, _ := json.Marshal(&ret)
		c.SetData(string(b))
		return nil
	}

	m := md5.Sum(csv_bb)
	if fmt.Sprintf("%x", m) != string(meta_bb) {
		return fmt.Errorf("get remote csv and meta diff !")
	}

	_csv_bb = csv_bb
	_meta_bb = meta_bb
	logs.Trace("channel csv : %v", string(csv_bb))
	if _csv_bb == nil {
		ret := res{"Please Refresh!", ""}
		b, _ := json.Marshal(&ret)
		c.SetData(string(b))
	} else {
		ret := res{"ok", strings.Replace(string(_csv_bb), ",", ";", -1)}
		b, _ := json.Marshal(&ret)
		c.SetData(string(b))
	}
	return nil
}

func updateRemoteChannelUrl(c *gm_command.Context, server, accountid string, params []string) error {
	if len(params) < 1 {
		logs.Error("updateRemoteChannelUrl param err: %v", params)
		return fmt.Errorf("param err")
	}
	csv := strings.Replace(params[0], ";", ",", -1)
	m := md5.Sum([]byte(csv))
	meta := fmt.Sprintf("%x", m)
	if err := s3.Put(config.Cfg.VerUpdate_FileName, []byte(csv), nil); err != nil {
		return err
	}
	if err := s3.Put(config.Cfg.VerUpdate_Meta_FileName, []byte(meta), nil); err != nil {
		return err
	}
	return nil
}
