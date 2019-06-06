package config

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"bytes"
	"crypto/md5"
	"io/ioutil"

	"github.com/siddontang/go/timingwheel"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

type verUpdateConfStruct map[string]map[string]string

var (
	VerUpdateConf verUpdateConfStruct
	_stop         chan struct{}

	_timingwheel *timingwheel.TimingWheel
	updateMutx   sync.RWMutex

	s3 storehelper.IStore
)

const (
	updateInterval = time.Minute * 5
)

func Start() {
	_stop = make(chan struct{})
	s3 = NewStore()
	_timingwheel = timingwheel.NewTimingWheel(time.Minute, 120)
	s3.Open()
	updateConf()
	updateGonggao()
	for {
		select {
		case <-_timingwheel.After(updateInterval):
			updateConf()
			updateGonggao()
		case <-_stop:
			return
		}
	}
}

func NewStore() storehelper.IStore {
	if Cfg.StoreDriver == storehelper.S3 {
		return storehelper.NewStore(storehelper.S3, storehelper.S3Cfg{
			Region:           Cfg.Dynamo_region,
			BucketsVerUpdate: Cfg.S3_Buckets_VerUpdate,
			AccessKeyId:      Cfg.Dynamo_accessKeyID,
			SecretAccessKey:  Cfg.Dynamo_secretAccessKey,
		})
	} else if Cfg.StoreDriver == storehelper.OSS {
		return storehelper.NewStore(storehelper.OSS, storehelper.OSSInitCfg{
			EndPoint:  Cfg.OSSEndPoint,
			AccessId:  Cfg.OSSAccessId,
			AccessKey: Cfg.OSSAccessKey,
			Bucket:    Cfg.OSSBucket,
		})
	}
	return nil
}

func GetChannelUrl(channel, subChannel string) string {
	updateMutx.RLock()
	defer updateMutx.RUnlock()
	if VerUpdateConf == nil {
		return ""
	}
	if sub, ok := VerUpdateConf[channel]; ok {
		if url, ok := sub[subChannel]; ok {
			return url
		}
	}
	return ""
}

func Stop() {
	close(_stop)
}

func updateConf() {
	if !Cfg.VerUpdate_Valid {
		return
	}
	if Cfg.VerUpdate_IsRemote {
		if err := updateByRemoteConf(); err != nil {
			logs.Error("[VerUpdate] updateByRemoteConf err %v", err)
		}
	} else {
		if err := updateByLocalConf(); err != nil {
			logs.Error("[VerUpdate] updateByLocalConf err %v", err)
		}
	}
}

func updateByLocalConf() error {
	f, err := os.OpenFile(config.NewConfigPath(Cfg.VerUpdate_FileName), os.O_RDONLY, os.FileMode(0644))
	if err != nil {
		return err
	}
	f_meta, err := os.OpenFile(config.NewConfigPath(Cfg.VerUpdate_Meta_FileName), os.O_RDONLY, os.FileMode(0644))
	if err != nil {
		return err
	}
	csv_bb, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	meta_bb, err := ioutil.ReadAll(f_meta)
	if err != nil {
		return err
	}
	conf, err := _read_csv(csv_bb, meta_bb)
	if err != nil {
		return err
	}

	updateMutx.Lock()
	defer updateMutx.Unlock()
	VerUpdateConf = conf
	logs.Debug("update ver conf by local: %v", VerUpdateConf)
	return nil
}

func updateByRemoteConf() error {
	csv_bb, err := s3.Get(Cfg.VerUpdate_FileName)
	if err != nil {
		return err
	}
	meta_bb, err := s3.Get(Cfg.VerUpdate_Meta_FileName)
	if err != nil {
		return err
	}
	conf, err := _read_csv(csv_bb, meta_bb)
	if err != nil {
		return err
	}
	updateMutx.Lock()
	defer updateMutx.Unlock()
	VerUpdateConf = conf
	logs.Debug("update ver conf by remote: %v", VerUpdateConf)
	return nil

}

func _read_csv(csv_bb, meta_bb []byte) (verUpdateConfStruct, error) {
	m := md5.Sum(csv_bb)
	if fmt.Sprintf("%x", m) != string(meta_bb) {
		return nil, fmt.Errorf("[VerUpdate] csv and meta diff !")
	}

	num := 0
	conf := make(verUpdateConfStruct, 10)
	r := csv.NewReader(bytes.NewReader(csv_bb))
	for {
		record, err := r.Read()
		num++
		// Stop at EOF.
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		// 去掉标题行
		if num == 1 {
			continue
		}
		if len(record) < 3 {
			return nil, fmt.Errorf("record len not enough, %v", record)
		}
		channel := record[0]
		subChannel := record[1]
		url := record[2]
		if _, ok := conf[channel]; !ok {
			conf[channel] = make(map[string]string, 64)
		}
		m := conf[channel]
		m[subChannel] = url
	}
	return conf, nil
}
