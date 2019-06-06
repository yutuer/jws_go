package chatserver

import (
	"crypto/md5"
	"sync"
	"time"

	"fmt"

	"github.com/siddontang/go/timingwheel"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/secure"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

type CommonConfig struct {
	RunMode         string   `toml:"run_mode"`
	EtcdEndPoint    []string `toml:"etcd_endpoint"`
	EtcdRoot        string   `toml:"etcd_root"`
	NetType         string   `toml:"net_type"`
	NetAddress      string   `toml:"net_address"`
	NetGamexAddress string   `toml:"net_gamex_address"`
	CityInitRoomNum int      `toml:"city_init_room_num"`
	RoomMaxNum      int      `toml:"room_player_max_num"`
	RoomMinNum      int      `toml:"room_player_min_num"`
	RoomAdjustNum   int      `toml:"room_player_adjust_num"`
}

var CommonCfg CommonConfig

var (
	chatAuthSource          [16]byte
	chatLastAuthSource      [16]byte
	chatLastAuthRefreshTime int64
	chatAuthSourceMutx      sync.RWMutex
	timingWheel             *timingwheel.TimingWheel
	gidsidmerge             map[string]SidMerge // key : 1:10
)

type SidMerge struct {
	GidSid   string // 1:10
	SidMerge string // 1:10
}

const (
	GameSource    = "87d5f092-7bb7-44a5-9869-b42fd9bf5899"
	ChgSourceTime = 24 * time.Hour // 5 * time.Minute TDB by zhangzhen
	PingWait      = 30 * time.Second
)

func init() {
	timingWheel = timingwheel.NewTimingWheel(time.Second, 120)
	gidsidmerge = make(map[string]SidMerge, 16)

	chatLastAuthRefreshTime = time.Now().Unix()
	chatAuthSource = [16]byte(uuid.NewV4())
	chatLastAuthSource = chatAuthSource
	fmt.Printf("chatAuth %v time %d \n", chatAuthSource, chatLastAuthRefreshTime)
}

func GetAuthSource(acid string) (curAuth, lastAuth string) {
	cur, last := _getAuthSource(acid)
	return secure.Encode64ForNet(cur), secure.Encode64ForNet(last)
}

func _getAuthSource(acid string) (curAuth, lastAuth []byte) {
	now := time.Now().Unix()
	var cur, last [16]byte
	chatAuthSourceMutx.RLock()
	if now-chatLastAuthRefreshTime > int64(ChgSourceTime) {
		chatAuthSourceMutx.RUnlock()

		chatAuthSourceMutx.Lock()
		if now-chatLastAuthRefreshTime > int64(ChgSourceTime) {
			chatLastAuthRefreshTime = now
			chatLastAuthSource = chatAuthSource
			chatAuthSource = [16]byte(uuid.NewV4())
			logs.Info("chatAuth %v lastAuth %v time %d", chatAuthSource, chatLastAuthSource, chatLastAuthRefreshTime)
		}
		cur = chatAuthSource
		last = chatLastAuthSource
		chatAuthSourceMutx.Unlock()
	} else {
		cur = chatAuthSource
		last = chatLastAuthSource
		chatAuthSourceMutx.RUnlock()
	}
	return genChatAuthSource(GameSource, acid, cur), genChatAuthSource(GameSource, acid, last)
}

// md5(md5(acid + gamexSource) + chatSource)
func genChatAuthSource(gameAuth, acid string, chatAuthSource [16]byte) []byte {
	res := make([]byte, 0, 32)
	v1 := md5.Sum([]byte(gameAuth + acid))
	res = append(res, v1[:]...)
	res = append(res, chatAuthSource[:]...)
	m := md5.Sum(res)
	return m[:]
}

func (cfg CommonConfig) SidMergeInfo() error {
	gid2sid, err := etcd.GetServerGidSid(cfg.EtcdRoot)
	if err != nil {
		return err
	}
	for gid, sids := range gid2sid {
		for _, sid := range sids {
			key := fmt.Sprintf("%s/%d/%d/gm/%s", cfg.EtcdRoot, gid, sid, etcd.KeyMergedShard)
			sidMerge, err := etcd.Get(key)
			if err != nil {
				logs.Info("SidMergeInfo get err %s", err.Error())
				continue
			}
			gidsid := fmt.Sprintf("%d:%d", gid, sid)
			_, ok := gidsidmerge[gidsid]
			if !ok {
				gidsidmerge[gidsid] = SidMerge{
					GidSid:   gidsid,
					SidMerge: fmt.Sprintf("%d:%s", gid, sidMerge),
				}
			}
		}
	}
	logs.Debug("SidMergeInfo %s", gidsidmerge)
	return nil
}

func GetGidSidMerge(gidsid string) string {
	info, ok := gidsidmerge[gidsid]
	if !ok {
		return gidsid
	}
	return info.SidMerge
}
