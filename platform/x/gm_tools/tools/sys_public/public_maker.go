package sys_public

import (
	"encoding/json"
	"fmt"
	//"time"

	"strings"

	"strconv"

	"errors"
	"os"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/ftp"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/qiniu"
	"vcs.taiyouxi.net/platform/x/gm_tools/config"
)

var (
	s3 storehelper.IStore
)

type PublicToServer struct {
	Title     string `json:"Title"`
	Type      int64  `json:"Type"`
	Priority  int64  `json:"Priority"`
	Index     string `json:"Index"`
	SendTime  int64  `json:"SendTime"`
	Body      string `json:"Body"`
	Begin     int64  `json:"Begin"`
	End       int64  `json:"End"`
	Language  string `json:"Lang"`
	MultiLang string `json:"MultiLang"`
}

type PublicSetToServer struct {
	Publics      []PublicToServer  `json:"Publics"`
	Maintaince   []PublicToServer  `json:"Maintaince"`
	Forceupdate  []PublicToServer  `json:"Forceupdate"`
	Endpoint     string            `json:"Endpoint"`
	EndpointChat string            `json:"Chatserver"`
	DataVer      string            `json:"DataVer"`
	BundleVer    string            `json:"BundleVer"`
	Whitelistpwd string            `json:"Whitelistpwd"`
	BIOption     string            `json:"BIOption"`
	TokenTime    string            `json:"TokenTime"`
	TimeoutTime  string            `json:"TimeoutTime"`
	PrUrl        string            `json:"PrUrl"`
	PayUrls      map[string]string `json:"PayUrls"`
	GZIPSize     int64             `json:"gzip_size"`
	DataMin      string            `json:"DataMin"`
	BundleMin    string            `json:"BundleMin"`
}

func (s *sysPublicMng) FindAllVersion() map[string]map[int64]bool {
	ver := make(map[string]map[int64]bool, 64)
	s.lock.RLock()
	for _, v := range s.Pub {
		_, ok := ver[v.Version]
		if !ok {
			ver[v.Version] = make(map[int64]bool, 64)
		}
		gids := ver[v.Version]
		gids[v.Gid] = true
		ver[v.Version] = gids
	}
	s.lock.RUnlock()
	return ver
}

func (s *sysPublicMng) MkPublicForOneVersion(gid int64, ver string) ([]byte, error) {
	mkPublicForOneVersionTime := time.Now()
	res := PublicSetToServer{}
	s.lock.RLock()
	pubLen := len(s.Pub)
	s.lock.RUnlock()
	res.Publics = make([]PublicToServer, 0, pubLen)
	res.Maintaince = make([]PublicToServer, 0, pubLen)
	res.Forceupdate = make([]PublicToServer, 0, pubLen)
	res.Endpoint = "http://54.223.177.52:8081"

	logs.Debug("MkPublicForOneVersion Init using %v time",time.Since(mkPublicForOneVersionTime))
	mkPublicForOneVersionTime = time.Now()

	//now_t := time.Now().Unix()
	s.lock.RLock()
	var maintaince_start_time int64
	var maintaince_end_time int64
	for _, v := range s.Pub {
		if v.Gid == gid &&
			v.Version == ver &&
			//v.Begin <= now_t &&
			//now_t <= v.End &&
			v.State > 0 {

			t := PublicToServer{
				Title:     v.Title,
				Type:      v.Typ,
				Priority:  v.Priority,
				Index:     v.Id,
				SendTime:  v.Begin,
				Body:      v.Body,
				Begin:     v.Begin,
				End:       v.End,
				Language:  v.Language,
				MultiLang: v.MultiLang,
			}
			switch v.Class {
			case "Publics":
				res.Publics = append(res.Publics, t)
			case "Maintaince":
				res.Maintaince = append(res.Maintaince, t)
				maintaince_start_time = t.Begin
				maintaince_end_time = t.End
			case "Forceupdate":
				res.Forceupdate = append(res.Forceupdate, t)
			}
		}
	}
	s.lock.RUnlock()
	logs.Debug("MkPublicForOneVersion read data using %v time",time.Since(mkPublicForOneVersionTime))
	mkPublicForOneVersionTime = time.Now()

	ends := s.GetEndpoint(gid, ver)
	if ends == nil {
		// TODO 删除
		//j, _ := json.Marshal(res)
		//logs.Error(string(j))
		logs.Error("[sysPublicMng]No Endpoint Info by %d - %s", gid, ver)
		//return j[:], nil // TODO 删除
		return nil, errors.New("NoEndpointInfo")
	} else {
		res.Endpoint = ends.ServerEndpoint
		res.EndpointChat = ends.ChatEndpoint
		res.DataVer = ends.DataVer
		res.BundleVer = ends.BundleVer
		res.DataMin = ends.DataMin
		res.BundleMin = ends.BundleMin
		res.Whitelistpwd = DefaultEncode.Encode64ForNet([]byte(ends.Whitelistpwd))
		res.BIOption = ends.BIOption
		res.TokenTime = ends.TokenTime
		res.TimeoutTime = ends.TimeoutTime
		res.PrUrl = ends.PrUrl

		// pay url
		res.PayUrls = make(map[string]string, 5)
		if ends.PayUrls == "" {
			res.PayUrls["default"] = "http://autopay-874898615.cn-north-1.elb.amazonaws.com.cn:9080/quick/android/v1/billingcb"
			res.PayUrls["106"] = "http://autopay-874898615.cn-north-1.elb.amazonaws.com.cn:9080/quick/ios/v1/billingcb"
			res.PayUrls["32"] = "http://119.29.232.161:9080/quick/android/v1/billingcb"
		} else {
			lines := strings.Split(ends.PayUrls, "\n")
			for _, line := range lines {
				ws := strings.Split(line, ";")
				if len(ws) < 2 {
					logs.Error("[sysPublicMng] MkPublicForOneVersion payurl len err %s", ws)
					continue
				}
				res.PayUrls[ws[0]] = ws[1]
				logs.Debug("[sysPublicMng] MkPublicForOneVersion  payurl %s %s", ws[0], ws[1])
			}
		}
		size, err := strconv.ParseInt(ends.GZIPSize, 10, 0)
		if err != nil {
			logs.Error("strconv.ParseInt Err by %s", err.Error())
		} else {
			res.GZIPSize = size
		}
		// set endpoint to etcd
		key := fmt.Sprintf("%s/%s/%d/%s/%s", config.Cfg.GameSerEtcdRoot,
			etcd.KeyEndPoint, gid, ver, etcd.KeyEndPoint_Whitelistpwd)
		etcd.Set(key, res.Whitelistpwd, 0)
		key = fmt.Sprintf("%s/%s/%d/%s/%s", config.Cfg.GameSerEtcdRoot,
			etcd.KeyEndPoint, gid, ver, etcd.KeyEndPoint_maintain_starttime)
		etcd.Set(key, fmt.Sprintf("%d", maintaince_start_time), 0)
		key = fmt.Sprintf("%s/%s/%d/%s/%s", config.Cfg.GameSerEtcdRoot,
			etcd.KeyEndPoint, gid, ver, etcd.KeyEndPoint_maintain_endtime)
		etcd.Set(key, fmt.Sprintf("%d", maintaince_end_time), 0)
	}

	logs.Debug("MkPublicForOneVersion etcd options using %v time",time.Since(mkPublicForOneVersionTime))
	mkPublicForOneVersionTime = time.Now()

	j, err := json.Marshal(res)
	if err != nil {
		logs.Error("[sysPublicMng] MkPublicToServer Marshal Err by %s", err.Error())
		return nil, err
	}
	logs.Debug("MkPublicForOneVersion marshal data using %v time",time.Since(mkPublicForOneVersionTime))
	return j[:], nil
}

func (s *sysPublicMng) SendToHttp(is_debug bool) error {
	vers := s.FindAllVersion()

	for ver, gids := range vers {
		for gid, _ := range gids {
			err := s.SendOnePublic(is_debug, gid, ver)
			if err != nil {
				logs.Error("SendOnePublic Err by %s at %d %s", err.Error(), gid, ver)
			}
		}
	}

	return nil
}

func (s *sysPublicMng) SendOnePublic(is_debug bool, gid int64, ver string) error {
	sendOnePublicTime := time.Now()
	dstr := "d"
	if !is_debug {
		dstr = "r"
	}

	data, err := s.MkPublicForOneVersion(gid, ver)
	if err != nil {
		logs.Error("ReleasePublic Err by %d %s", gid, ver)
		return err
	} else {
		logs.Info("ReleasePublic %d %s %v", gid, ver, string(data))
	}
	logs.Debug("SendOnePublic get data using %v time",time.Since(sendOnePublicTime))
	sendOnePublicTime = time.Now()

	path := fmt.Sprintf("public/%d/", gid)
	fileName := fmt.Sprintf("gonggao_%s_%s.json", ver, dstr)

	if err = s.saveLocal(fmt.Sprintf("./gm_gonggao/%d/", gid), fileName, data); err != nil {
		logs.Error("save public sys notice to local err", err)
		return err
	}
	logs.Debug("SendOnePublic save local using %v time",time.Since(sendOnePublicTime))
	sendOnePublicTime = time.Now()

	if ftp.FtpAddress != "" {
		err = ftp.UploadToFtp(path, fileName, data)
		if err != nil {
			logs.Error("ReleasePublic Ftp UpLoad Err by %s %v", ver, err)
			return err
		}
	}
	logs.Debug("SendOnePublic ftp using %v time",time.Since(sendOnePublicTime))
	sendOnePublicTime = time.Now()

	if qiniu.IsActivity() {
		err = qiniu.UpLoad("jwscdn", path+fileName, data)
		if err != nil {
			logs.Error("ReleasePublic qiniu UpLoad Err by %s %v", ver, err)
			return err
		}
	}
	logs.Debug("SendOnePublic upload data using %v time",time.Since(sendOnePublicTime))
	sendOnePublicTime = time.Now()

	if config.Cfg.S3_Buckets_GongGao != "" {
		if err := s3.Put(path+fileName, data, nil); err != nil {
			logs.Error("ReleasePublic s3 UpLoad Err by %s %v", ver, err)
			return err
		}
		if config.Cfg.NoticeWatchKey != "" {
			etcd.Set(fmt.Sprintf("%s/%s", config.Cfg.GameSerEtcdRoot, config.Cfg.NoticeWatchKey), fmt.Sprintf("%d", time.Now().Unix()), 0)
		}
		logs.Debug("notice save s3 ok ")
	}
	logs.Debug("SendOnePublic save data using %v time",time.Since(sendOnePublicTime))
	return nil
}

func (s *sysPublicMng) saveLocal(path, fileName string, data []byte) error {
	if err := os.MkdirAll(path, 0777); err != nil {
		return err
	}
	outputFile, err := os.OpenFile(path+fileName, os.O_RDWR|os.O_CREATE, 0644)
	defer outputFile.Close()
	if err != nil {
		return err
	} else {
		realLen := len(data)
		wlen, err := outputFile.Write(data)
		if err != nil {
			return err
		}
		if wlen != realLen {
			return fmt.Errorf("not write completely")
		}
		logs.Info("save sys public to local ok %s", path+fileName)
		return nil
	}
	return nil
}
