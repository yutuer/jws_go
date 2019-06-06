package hero_gag

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/olivere/elastic.v5"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/config"
)

const GagCommand = "banAccount"
const GetAccountData = "getInfoByNickName"
const GetAccountDataByNick = "getInfoByNickNameByNick"

func InitHeroGag(r *gin.Engine) {
	// init elastic client
	client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(config.Cfg.ElasticUrl),
		elastic.SetTraceLog(log.New(os.Stdout, "", 0)))
	if err != nil {
		// Handle error
		//panic(err)
		logs.Error("init elastic search err by %v", err)
	}
	elasticClient = client
	g := r.Group("/yingxiong", CommonIPLimit())

	// http://ip:port/yingxiong/ban/0:10/0:10:fdsafdsafdsafdsa/reason/10[seconds]
	g.POST("/ban/:server_name/:account_id/:reason/:time", handleGagAccount)
	// http://ip:port/yingxiong/get_account/0:10:fdsafdsafdsafdsa
	g.POST("/get_account/:account_id", handleGetAccountData)
	// http://ip:port/yingxiong/get_account/0/10/gdkgadsfad
	g.POST("/get_account_by_nick/:gid/:sid/:nick", handleGetAccountDataByNick)

}
