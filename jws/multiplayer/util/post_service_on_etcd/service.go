package postService

import (
	"encoding/json"
	"errors"
	"sync"

	"time"

	"strconv"

	"golang.org/x/net/context"
	mulutil "vcs.taiyouxi.net/jws/multiplayer/util"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/etcdClient"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timingwheel"
)

var (
	ErrSelectFuncRes = errors.New("ErrSelectFuncRes")
	ErrServiceNoFind = errors.New("ErrServiceNoFind")
	ErrTimeout       = errors.New("ErrTimeout")
)

type Service struct {
	Url string
	Id  string
	Num int
}

type selectServiceFunc func(services []Service) int

var NilFunc = func(s []Service) int {
	return -1
}

type postMsgRes struct {
	data []byte
	err  error
}

type postMsg struct {
	id       string
	typ      string
	postData []byte
	resChann chan postMsgRes
}

type ServiceMng struct {
	Services    []Service
	servicesMap map[string]Service

	selectFunc   selectServiceFunc
	postChann    chan postMsg
	wg           sync.WaitGroup
	quitChan     chan bool
	tWheel       *timingwheel.TimingWheel
	etcdRootPath string
	etcdEndPoint []string
}

func (s *ServiceMng) Init(etcdRootPath string, f selectServiceFunc, endPoint []string) {
	logs.Trace("ServiceMng Init with %s", etcdRootPath)
	s.etcdRootPath = etcdRootPath
	s.selectFunc = f
	s.etcdEndPoint = endPoint
	return
}

func (s *ServiceMng) Start() {
	s.quitChan = make(chan bool, 1)
	s.tWheel = mulutil.GetSlowTimeWheel()
	s.postChann = make(chan postMsg, 1024)
	s.updateFromEtcd()
	s.wg.Add(1)
	rspChan := make(chan *client.Response, 1024)
	etcd.NewWatcher(s.etcdEndPoint, s.etcdRootPath, rspChan)

	go func() {
		defer s.wg.Done()
		timerChan := s.tWheel.After(10 * time.Second)
		action := ""
		for {
			defer func() {
				if err := recover(); err != nil {
					logs.Error("Match Panic, Err %v", err)
				}
			}()
			select {
			case command, _ := <-s.postChann:
				s.wg.Add(1)
				func() {
					defer s.wg.Done()
					s.post(&command)

				}()
			case etcdRspInfo := <-rspChan:
				logs.Debug("receive change from etcd %v", etcdRspInfo)
				action = etcdRspInfo.Action
			case <-timerChan:
				s.wg.Add(1)
				func() {
					defer s.wg.Done()
					if action == etcd.WatchAction_Set ||
						action == etcd.WatchAction_Del {
						logs.Debug("%s has been changed Action:%s", s.etcdRootPath, action)
						s.updateFromEtcd()
					}
				}()
				action = ""
				timerChan = s.tWheel.After(10 * time.Second)

			case <-s.quitChan:
				return
			}
		}
	}()
}

func (s *ServiceMng) Stop() {
	s.quitChan <- true
	s.wg.Wait()
}

func (s *ServiceMng) updateFromEtcd() {
	time1 := time.Now().Unix()
	logs.Trace("ServiceMng updateFromEtcd with %s", s.etcdRootPath)
	nodes, err := etcd.GetAllSubKeysRcs(s.etcdRootPath)
	if err != nil {
		logs.Error("service updateFromEtcd err By %s", err.Error())
		return
	}
	newService := make([]Service, 0, len(nodes.Nodes))
	servicesMap := make(map[string]Service, len(nodes.Nodes))

	for _, node := range nodes.Nodes {

		var url, id, num string
		if urlNode := etcd.FindKeyFromNodes(node, node.Key+ServiceUrlKey); urlNode != nil {
			url = urlNode.Value
		}
		if idNode := etcd.FindKeyFromNodes(node, node.Key+ServiceIDKey); idNode != nil {
			id = idNode.Value
		}
		if numNode := etcd.FindKeyFromNodes(node, node.Key+ServiceNumKey); numNode != nil {
			num = numNode.Value
		}
		numn, err := strconv.Atoi(num)
		if url == "" || id == "" || num == "" || err != nil {
			continue
		}
		n := Service{
			Url: url,
			Id:  id,
			Num: numn,
		}
		newService = append(newService, n)
		servicesMap[id] = n
	}
	logs.Trace("ServiceMng updateFromEtcd get 2 %v", servicesMap)
	time2 := time.Now().Unix()
	if time2-time1 > 1 {
		logs.Warn("update from etcd over 1 s, %d", time2-time1)
	}
	s.Services = newService
	s.servicesMap = servicesMap
}

func (s *ServiceMng) post(m *postMsg) {
	var ser *Service

	if m.id == "" {
		idx := s.selectFunc(s.Services)
		if idx < 0 || idx >= len(s.Services) {
			logs.Error("selectFunc res Err By %d", idx)
			m.resChann <- postMsgRes{
				err: ErrSelectFuncRes,
			}
			return
		}
		ser = &s.Services[idx]
	} else {
		serv, ok := s.servicesMap[m.id]
		if !ok {
			logs.Error("service no find By %s", m.id)
			m.resChann <- postMsgRes{
				err: ErrServiceNoFind,
			}
			return
		}
		ser = &serv
	}
	//logs.Trace("multiplay service post info: %s, %s", ser.Url, m.typ)
	go func(url string, n *postMsg) {
		logs.Info("Post url: %v, %v", url, n.postData)
		resData, err := HttpPost(url, n.typ, n.postData)
		n.resChann <- postMsgRes{
			data: resData,
			err:  err,
		}
	}(ser.Url, m)
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (s *ServiceMng) PostBySelect(data interface{}) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	datas, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resChan := make(chan postMsgRes, 1)
	select {
	case s.postChann <- postMsg{
		postData: datas[:],
		typ:      util.JsonPostTyp,
		resChann: resChan,
	}:
	case <-ctx.Done():
		logs.Error("ServiceMng.PostBySelect postChann full")
		return nil, ErrTimeout
	}

	select {
	case r := <-resChan:
		return r.data[:], r.err
	case <-ctx.Done():
		logs.Error("ServiceMng.PostBySelect resChan timeout")
		return nil, ErrTimeout
	}
}

// 防Account主rountine锁死，有待重构成统一模式 byzhangzhenTDB
func (s *ServiceMng) PostById(serviceID string, data interface{}) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), util.ASyncCmdTimeOut)
	defer cancel()

	datas, err := json.Marshal(data)
	if err != nil {
		logs.Error("ServiceMng.PostById json.Marshal err %s", err.Error())
		return nil, err
	}

	resChan := make(chan postMsgRes, 1)
	select {
	case s.postChann <- postMsg{
		id:       serviceID,
		postData: datas[:],
		typ:      util.JsonPostTyp,
		resChann: resChan,
	}:
	case <-ctx.Done():
		logs.Error("ServiceMng.PostById postChann full")
		return nil, ErrTimeout
	}

	select {
	case r := <-resChan:
		return r.data[:], r.err
	case <-ctx.Done():
		logs.Error("ServiceMng.PostById resChan timeout")
		return nil, ErrTimeout
	}
}
