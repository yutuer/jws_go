package data_ver

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/timingwheel"

	"strconv"

	"strings"

	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func genDataVerModule(sid uint) *DataVerModule {
	return &DataVerModule{
		shardId: sid,
	}
}

type DataVerModule struct {
	shardId uint
	tWheel  *timingwheel.TimingWheel
}

func (m *DataVerModule) AfterStart(g *gin.Engine) {
}

func (m *DataVerModule) BeforeStop() {
}

func (m *DataVerModule) Start() {
	if game.Cfg.LocalDebug {
		return
	}
	logs.Info("DataVerModule start")
	m.tWheel = timingwheel.NewTimingWheel(time.Second, 10*60)
	t := 30 * time.Second
	if game.Cfg.IsRunModeProd() {
		t = 5 * time.Minute
	}
	m.getClientDataBundleVer()
	m.getActValid(m.shardId)
	m.updateCfgByEtcd(m.shardId)
	go func() {
		_timer := m.tWheel.After(t)
		for {
			select {
			case <-_timer:
				_timer = m.tWheel.After(t)
				m.getClientDataBundleVer()
				m.getActValid(m.shardId)
				m.updateCfgByEtcd(m.shardId)
			}
		}
	}()
}

func (m *DataVerModule) Stop() {
	if game.Cfg.LocalDebug {
		return
	}
	m.tWheel.Stop()
}

func (m *DataVerModule) getClientDataBundleVer() {
	gid := game.Cfg.Gid
	key := fmt.Sprintf("%s/%d/%s", game.Cfg.EtcdRoot, gid, etcd.KeyGlobalClientDataVer)
	ds, err := etcd.GetAllSubKeys(key)
	if err != nil {
		if !game.Cfg.LocalDebug {
			logs.Trace("globleclientinfo GetAllSubKeys from etcd err %v", err)
		}
	} else {
		for _, d := range ds {
			pg := strings.Split(d, "/")
			ver := pg[len(pg)-1]
			data_ver_str, err := etcd.Get(d)
			if err != nil {
				logs.Error("globleclientinfo get %s from etcd err %v", d, err)
				continue
			}
			data_ver, err := strconv.Atoi(data_ver_str)
			if err != nil {
				data_ver = 0
			}
			logs.Trace("globleclientinfo ver %s %d", ver, data_ver)
			game.Cfg.UpdateDataVer(ver, int32(data_ver))
		}
	}

	bundleKey := fmt.Sprintf("%s/%d/%s", game.Cfg.EtcdRoot, gid, etcd.KeyGlobalClientBundleVer)
	bs, err := etcd.GetAllSubKeys(bundleKey)
	if err != nil {
		if !game.Cfg.LocalDebug {
			logs.Debug("globleclientbundleinfo GetAllSubKeys from etcd err %v", err)
		}
	} else {
		for _, b := range bs {
			pg := strings.Split(b, "/")
			ver := pg[len(pg)-1]
			bundle_ver_str, err := etcd.Get(b)
			if err != nil {
				logs.Error("globleclientbundleinfo get %s from etcd err %v", b, err)
				continue
			}
			bundle_ver, err := strconv.Atoi(bundle_ver_str)
			if err != nil {
				bundle_ver = 0
			}
			//logs.Trace("globleclientbundleinfo ver %s %d", ver, bundle_ver)
			game.Cfg.UpdateBundleVer(ver, int32(bundle_ver))
		}
	}

	dataMinKey := fmt.Sprintf("%s/%d/%s", game.Cfg.EtcdRoot, gid, etcd.KeyGlobalClientDataVerMIn)
	dm, err := etcd.GetAllSubKeys(dataMinKey)
	if err != nil {
		if !game.Cfg.LocalDebug {
			logs.Trace("globleclientinfo GetAllSubKeys from etcd err %v", err)
		}
	} else {
		for _, d := range dm {
			pg := strings.Split(d, "/")
			ver := pg[len(pg)-1]
			data_ver_str, err := etcd.Get(d)
			if err != nil {
				logs.Error("globleclientinfo get %s from etcd err %v", d, err)
				continue
			}
			data_ver, err := strconv.Atoi(data_ver_str)
			if err != nil {
				data_ver = 0
			}
			logs.Trace("globleclientinfo ver %s %d", ver, data_ver)
			game.Cfg.UpdateDataMin(ver, int32(data_ver))
		}
	}

	boudleMinKey := fmt.Sprintf("%s/%d/%s", game.Cfg.EtcdRoot, gid, etcd.KeyGlobalClientBundleVerMin)
	bm, err := etcd.GetAllSubKeys(boudleMinKey)
	if err != nil {
		if !game.Cfg.LocalDebug {
			logs.Trace("globleclientinfo GetAllSubKeys from etcd err %v", err)
		}
	} else {
		for _, d := range bm {
			pg := strings.Split(d, "/")
			ver := pg[len(pg)-1]
			data_ver_str, err := etcd.Get(d)
			if err != nil {
				logs.Error("globleclientinfo get %s from etcd err %v", d, err)
				continue
			}
			data_ver, err := strconv.Atoi(data_ver_str)
			if err != nil {
				data_ver = 0
			}
			logs.Trace("globleclientinfo ver %s %d", ver, data_ver)
			game.Cfg.UpdateBundleMin(ver, int32(data_ver))
		}
	}
}

const ActCount = 50

func (m *DataVerModule) getActValid(sid uint) {
	gid := game.Cfg.Gid
	key := fmt.Sprintf("%s/%d/%d/%s", game.Cfg.EtcdRoot,
		gid, sid, etcd.KeyActValid)
	v, err := etcd.Get(key)
	acts := strings.Split(v, ",")
	if v == "" {
		act_valueds := m.initState()
		etcd.Set(key, act_valueds, 0)
		v = act_valueds
	} else if len(acts) < ActCount {
		var act_valued []string
		for i := 0; i < ActCount; i++ {
			act_valued = append(act_valued, "1")
		}
		for i, act := range acts {
			_, err := strconv.Atoi(act)
			if err != nil {
				logs.Error("act_valid1 data err %s", v)
			}
			act_valued[i] = acts[i]
		}
		act_valueds := strings.Join(act_valued, ",")

		etcd.Set(key, act_valueds, 0)
	}
	if err == nil {
		logs.Trace("UpdateActValid get %s", v)
		game.Cfg.UpdateActValid(sid, v)
	} else {
		logs.Debug("UpdateActValid get err %v", err)
	}
}
func (m *DataVerModule) initState() string {
	var act_valued []string
	for i := 0; i < ActCount; i++ {
		// 确保某些活动是关闭的，硬编码
		switch i {
		case 30:
			fallthrough
		case 0:
			if uutil.IsVNVer() || uutil.IsKOVer() {
				act_valued = append(act_valued, "0")
			} else {
				act_valued = append(act_valued, "1")
			}
		case 31:
			if uutil.IsVNVer() {
				act_valued = append(act_valued, "0")
			} else {
				act_valued = append(act_valued, "1")
			}
		case 21:
			fallthrough
		case 22:
			if uutil.IsHMTVer() || uutil.IsVNVer() || uutil.IsKOVer() {
				act_valued = append(act_valued, "0")
			} else {
				act_valued = append(act_valued, "1")
			}
		default:
			act_valued = append(act_valued, "1")
		}
	}
	return strings.Join(act_valued, ",")
}

func (m *DataVerModule) updateCfgByEtcd(sid uint) {
	c := &game.Cfg
	c.ShardShowState, _ = etcd.Get(fmt.Sprintf("%s/%d/%d/%s",
		c.EtcdRoot, c.Gid, sid, etcd.KeyShowState))
}
