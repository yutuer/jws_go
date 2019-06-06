package sysnotice

import (
	"encoding/json"
	"fmt"
	"time"

	"strings"
	"vcs.taiyouxi.net/jws/gamex/models/city_broadcast"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/etcd"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	connectTimeout   = 10 * time.Second
	readWriteTimeout = 10 * time.Second
)

const (
	ParamType_Context = iota
	ParamType_RollName
	ParamType_ItemId
	ParamType_Value
	ParamType_LevelId
	ParamType_Trial_LevelId
	ParamType_BossId
	ParamType_Hero
	ParamType_DGId
)

type NoticeParam struct {
	Key   string `json:"k"`
	Value string `json:"v"`
}

type SysRollNoticeInfo struct {
	MsgId int32         `json:"msgId"`
	KV    []NoticeParam `json:"info"`
	Acid  []string
}

type SysRollNotice struct {
	shard string
	info  *SysRollNoticeInfo
}

type Msg struct {
	Typ string `json:"type"`
	Shd string `json:"shd"`
	Msg string `json:"msg"`
}

func NewSysRollNotice(shard string, msgId int32) *SysRollNotice {
	instance := &SysRollNotice{shard, &SysRollNoticeInfo{msgId, []NoticeParam{}, []string{}}}
	return instance
}

func (srn *SysRollNotice) AddParam(key int, value string) *SysRollNotice {
	param := NoticeParam{fmt.Sprintf("%d", key), value}
	srn.info.KV = append(srn.info.KV, param)
	return srn
}

func (srn *SysRollNotice) AddSids(acid []string) *SysRollNotice {
	srn.info.Acid = append(srn.info.Acid, acid[:]...)
	return srn
}

func (srn *SysRollNotice) Send() {
	if game.Cfg.SysNoticeValid > 0 && (!uutil.IsJAVer() || srn.IsAvailable()) {
		logs.Trace("[cyt]send sysnotic shard :%v, MsgID : %v,Acid : %v, KV: %v", srn.shard, srn.info.MsgId, srn.info.Acid, srn.info.KV)
		content, _ := json.Marshal(*srn.info)
		realGidSid := GetRealSid(srn.shard)
		city_broadcast.Pool.UseRes2Send(
			city_broadcast.CBC_Typ_SysNotice,
			realGidSid,
			string(content),
			nil,
		)
	}
}

func GetRealSid(shardID string) string {
	info := strings.Split(shardID, ":")
	if len(info) < 2 {
		logs.Error("[cyt]wrong srn.shard need:gid:sid,but only gid or sid") //如果这里出错会导致info数组越界，必须return
		return ""
	}
	_realSid, err := etcd.Get(fmt.Sprintf("%s/%s/%s/gm/mergedshard", game.Cfg.EtcdRoot, info[0], info[1]))
	if err != nil {
		logs.Error("[cyt]get mergedshard from etcd err: %v", err)
		return ""
	}
	realGidSid := info[0] + ":" + _realSid
	return realGidSid
}

func (srn *SysRollNotice) SendGuild() {
	if game.Cfg.SysNoticeValid > 0 && (!uutil.IsJAVer() || srn.IsAvailable()) {
		content, _ := json.Marshal(*srn.info)
		city_broadcast.Pool.UseRes2Send(
			city_broadcast.CBC_Typ_GuildRoom,
			srn.shard,
			string(content),
			srn.info.Acid,
		)
	}
}

func (srn *SysRollNotice) String() string {
	return fmt.Sprintf("{%v %v}", srn.info.MsgId, srn.info.KV)
}

func (srn *SysRollNotice) IsAvailable() bool {
	return gamedata.GetIsAvailable(srn.info.MsgId)
}
