package sys_public

import (
	"encoding/json"
	"fmt"
	"time"

	"sync"

	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/store"
	"vcs.taiyouxi.net/platform/x/gm_tools/config"
)

type sysPublic struct {
	Id        string `json:"id"`
	Gid       int64  `json:"game_id"`
	Priority  int64  `json:"priority"`
	Typ       int64  `json:"typ"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Begin     int64  `json:"begin"`
	End       int64  `json:"end"`
	State     int64  `json:"state"`
	Version   string `json:"ver"`
	Class     string `json:"class"`
	Language  string `json:"lang"`
	MultiLang string `json:"multi_lang"`
}
type sysPrivate struct {
	Id        string `json:"id"`
	Gid       int64  `json:"game_id"`
	Priority  int64  `json:"priority"`
	Typ       int64  `json:"typ"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Begin     string `json:"begin"`
	End       string `json:"end"`
	State     int64  `json:"state"`
	Version   string `json:"ver"`
	Class     string `json:"class"`
	Language  string `json:"lang"`
	MultiLang string `json:"multi_lang"`
}
type sysEndPoints struct {
	ServerEndpoint string `json:"server"`
	ChatEndpoint   string `json:"chat"`
	DataVer        string `json:"ver"`
	DataVerEtcd    string `json:"ver_etcd"`
	BundleVer      string `json:"bundle_ver"`
	BundleVerEtcd  string `json:"bundle_ver_etcd"`
	Whitelistpwd   string `json:"whitelistpwd"`
	BIOption       string `json:"biOption"`
	TokenTime      string `json:"tokenTime"`
	TimeoutTime    string `json:"timeoutTime"`
	PrUrl          string `json:"pr_url"`
	PayUrls        string `json:"pay_urls"`
	GZIPSize       string `json:"gzip_size"`
	DataMin        string `json:"data_min"`
	BundleMin      string `json:"bundle_min"`
	DataMinEtcd    string `json:"etcd_data_min"`
	BundleMinEtcd  string `json:"etcd_bundle_min"`
}

func NewSysPublic(gid, priority, typ, begin, end, state int64, title, body, ver, class, language string, multiLang string) sysPublic {
	new_id := time.Now().UnixNano()
	return sysPublic{
		Id:        fmt.Sprintf("%d", new_id),
		Gid:       gid,
		Priority:  priority,
		Typ:       typ,
		Title:     title,
		Body:      body,
		Begin:     begin,
		End:       end,
		State:     state,
		Version:   ver,
		Class:     class,
		Language:  language,
		MultiLang: multiLang,
	}
}

var sysPublicManager sysPublicMng

func Init() {
	sysPublicManager.Load()

	if config.Cfg.StoreDriver == storehelper.S3 {
		s3 = storehelper.NewStore(storehelper.S3, storehelper.S3Cfg{
			Region:           config.Cfg.S3GongGao_Region,
			BucketsVerUpdate: config.Cfg.S3_Buckets_GongGao,
			AccessKeyId:      config.Cfg.AWS_AccessKey,
			SecretAccessKey:  config.Cfg.AWS_SecretKey,
		})
	} else if config.Cfg.StoreDriver == storehelper.OSS {
		s3 = storehelper.NewStore(storehelper.OSS, storehelper.OSSInitCfg{
			EndPoint:  config.Cfg.OSSPublicBucket,
			AccessId:  config.Cfg.OSSAccessId,
			AccessKey: config.Cfg.OSSAccessKey,
			Bucket:    config.Cfg.OSSChannelBucket,
		})
	}
	s3.Open()
}

func Save() {
	sysPublicManager.SaveAll()
}

type sysPublicMng struct {
	lock    sync.RWMutex
	Pub     map[string]sysPublic `json:"publics"`
	endLock sync.RWMutex
	End     map[string]sysEndPoints `json:"endpoints"`
}

func (s *sysPublicMng) GetAll(ver string) []byte {
	getAllTime := time.Now()
	res := make(map[string]sysPrivate, 128)
	s.lock.RLock()
	for i, sp := range s.Pub {
		logs.Trace("GetAll:key:%v,value:%v", i, sp)
		//sysPublic的begin和end是int64,不能将int64直接交付给前端，应该先将其转换为对应时区的时间字符串格式再交付前端页面
		if ver != "" && sp.Version == ver {
			copy_sp := sysPrivate{
				sp.Id,
				sp.Gid,
				sp.Priority,
				sp.Typ,
				sp.Title,
				sp.Body,
				time.Unix(sp.Begin, 0).In(util.ServerTimeLocal).Format("2006/01/02 15:04:05"),
				time.Unix(sp.End, 0).In(util.ServerTimeLocal).Format("2006/01/02 15:04:05"),
				sp.State,
				sp.Version,
				sp.Class,
				sp.Language,
				sp.MultiLang,
			}
			res[i] = copy_sp
		}
	}
	s.lock.RUnlock()
	logs.Debug("func GetAll put copy to res using %v",time.Since(getAllTime))

	getAllTime = time.Now()
	j, err := json.Marshal(res)
	if err != nil {
		logs.Error("[sysPublicMng] GetAll Marshal Err by %s", err.Error())
		return nil
	}
	logs.Debug("func GetAll marshal res using %v",time.Since(getAllTime))
	return j[:]
}

func (s *sysPublicMng) Add(n sysPublic) {
	AddTime := time.Now()
	s.lock.Lock()
	s.Pub[n.Id] = n
	s.lock.Unlock()
	logs.Trace("Add: begin:%v,end:%v", n.Begin, n.End)
	s.Save(true, n.Gid, n.Version)
	logs.Debug("Add Using %s time",time.Since(AddTime))
}

func (s *sysPublicMng) Del(id string) {
	DelTime := time.Now()
	s.lock.Lock()
	p, ok := s.Pub[id]
	if ok {
		delete(s.Pub, id)
	}
	s.lock.Unlock()
	if ok {
		s.Save(true, p.Gid, p.Version)
	}
	logs.Debug("Del using %v time",time.Since(DelTime))
}

func (s *sysPublicMng) SetEndpoint(gid int64, ver string, end sysEndPoints) {
	k := fmt.Sprintf("%d-%s", gid, ver)
	s.endLock.Lock()
	s.End[k] = end
	s.endLock.Unlock()
	s.Save(true, gid, ver)
}

func (s *sysPublicMng) GetEndpoint(gid int64, ver string) *sysEndPoints {
	getEndpointTime := time.Now()
	k := fmt.Sprintf("%d-%s", gid, ver)

	s.endLock.RLock()
	e, ok := s.End[k]
	s.endLock.RUnlock()

	logs.Debug("GetEndpoint using %v time",time.Since(getEndpointTime))
	if !ok {
		return nil
	} else {
		return &e
	}
}

func (s *sysPublicMng) Load() {
	err := store.GetFromJson(store.NormalBucket, "public", s)
	logs.Trace("[PlanJobManager]Load %v", *s)
	s.lock.Lock()
	if s.Pub == nil {
		s.Pub = make(map[string]sysPublic, 128)
	}
	s.lock.Unlock()

	s.endLock.Lock()
	if s.End == nil {
		s.End = make(map[string]sysEndPoints, 128)
	}
	s.endLock.Unlock()

	if err == store.ErrNoKey {
		return
	}

	if err != nil {
		logs.Error("[PlanJobManager]Load Err by %s", err.Error())
	}

	return
}

func (s *sysPublicMng) SaveAll() {
	if err := store.SetIntoJson(store.NormalBucket, "public", *s); err != nil {
		logs.Error("sysPublicMng SaveAll SetIntoJson err %s", err.Error())
	}

	err := s.SendToHttp(true)
	if err != nil {
		logs.Error("send err by %s", err.Error())
	}
}

func (s *sysPublicMng) Save(is_debug bool, gid int64, ver string) {
	if err := store.SetIntoJson(store.NormalBucket, "public", *s); err != nil {
		logs.Error("sysPublicMng Save SetIntoJson err %s", err.Error())

	}
	for k, v := range s.Pub {
		logs.Trace("Save1: key:%v,value:%v", k, v)
	}
	err := s.SendOnePublic(is_debug, gid, ver)
	if err != nil {
		logs.Error("send err by %s", err.Error())
	}
	for k, v := range s.Pub {
		logs.Trace("Save2: key:%v,value:%v", k, v)
	}
}
