package hero_gag

import (
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
	"time"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	elasticClient *elastic.Client
)

type CommandInfo struct {
	CommandName string `json:"command_name"`
	IsOK        bool   `json:"is_ok"`
	Time        string `json:"@timestamp"`
	AccountID   string `json:"account_id"`
	ServerName  string `json:"server_id"`
	Reason      string `json:"reason,omitempty"`
}

const YingXiongGMIndex = "yingxiong_gm" // Warn: must lowercase
const ElasticType = "GM_Manual"

func logElasticInfo(info CommandInfo) {
	index := YingXiongGMIndex + "." + time.Now().Format("01.02.2006")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	exist, err := elasticClient.IndexExists(index).Do(ctx)
	if err != nil {
		logs.Error("HeroGag elasticSearch [IndexExists] err: %v", err)
		return
	}
	if !exist {
		// Create an index
		_, err = elasticClient.CreateIndex(index).Do(ctx)
		if err != nil {
			logs.Error("HeroGag elasticSearch [CreateIndex] err: %v", err)
			return
		}
	}
	_, err = elasticClient.Index().
		Index(index).
		Type(ElasticType).
		BodyJson(info).
		Refresh("true").
		Do(ctx)
	if err != nil {
		logs.Error("HeroGag elasticSearch [Index] err: %v", err)
		return
	}
}

func getTime() string {
	return time.Now().Format("2006-01-02T15:04:05.000Z07:00")
}
