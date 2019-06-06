package serverinfo

import (
	"fmt"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/auth/models"
	CBConfig "vcs.taiyouxi.net/platform/x/cheat_beholder/config"
)

var (
	RSP Rsp
)

type RspSerInfo struct {
	Name  string `json:"name"`
	Serid string `json:"serverid"`
}

type Rsp struct {
	Result  int
	Servers []RspSerInfo
}

func RegisterAllServerInfo() {
	if err := UpdateServerInfo(); err != nil {
		logs.Error("get shard info err %v", err)
	}
	logs.Debug("get all shardInfo")
}

/*
更新服务器的列表信息
*/
func UpdateServerInfo() error {
	andgids := CBConfig.CommonConfig.AndGid
	data1 := make([]RspSerInfo, 0)
	logs.Debug("try to get info from etcd")

	for _, gid := range andgids {
		info, err := models.GetAllShard(gid)
		if err != nil {
			RSP.Result = 0
			return err
		}
		logs.Debug("get shard info: %v from andr %d", info, gid)
		for _, item := range info {
			data1 = append(data1, RspSerInfo{
				Name:  item.DName,
				Serid: fmt.Sprintf("%d:%d", item.Gid, item.Sid),
			})
		}
	}

	trsp := Rsp{
		Result:  1,
		Servers: data1,
	}
	RSP = trsp
	return nil
}
