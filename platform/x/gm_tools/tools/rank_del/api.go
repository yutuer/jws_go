package rank_del

import (
	"errors"
	"fmt"
	"time"

	"github.com/astaxie/beego/httplib"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	"vcs.taiyouxi.net/platform/x/gm_tools/config"
)

const (
	RankCorpGs      = "RankCorpGs"
	RankSimplePvp   = "RankSimplePvp"
	RankByCorpTrial = "RankByCorpTrial"
	RankByHeroStar  = "RankByHeroStar"
)

var args = []string{RankCorpGs,
	RankSimplePvp,
	RankByCorpTrial,
	RankByHeroStar}

func RegCommands() {
	gm_command.AddGmCommandHandle("delRank", delRank)
}

func delRank(c *gm_command.Context, server, accountid string, params []string) error {
	logs.Info("delRank %v %v", server, accountid)
	logs.Info("delRank Params %v", params)
	if server == "" || accountid == "" || len(params) < 1 {
		return errors.New("delRank params error")
	}
	serCfg := config.Cfg.GetServerCfgFromName(server)
	if serCfg.ServerHotDataUrl == "" {
		return fmt.Errorf("server not found")
	}

	resp := httplib.Post(serCfg.RankReloadUrl).SetTimeout(5*time.Second, 5*time.Second).
		Param("rank_name", params[0]).Param("param", params[1]).Param("acid", accountid)
	res, err := resp.String()

	if err != nil {
		return err
	}
	c.SetData(res)
	return nil
}

//伪造申请，调用delRank，删除希望删除的榜单
func DelPlayerRanks(c *gm_command.Context, server, accountid string) error {
	var flag bool
	var targeterr error
	for _, arg := range args {
		err := delRank(c, server, accountid, []string{arg, "", "System"})
		if err != nil {
			logs.Error("删除榜单%v时失败", arg)
			targeterr = err
			flag = true
		}
	}
	if flag {
		return targeterr
	}
	return nil
}
