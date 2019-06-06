package server

import (
	"github.com/gin-gonic/gin"
	"vcs.taiyouxi.net/jws/multiplayer/helper"
	"vcs.taiyouxi.net/jws/multiplayer/match_server/match"
	"vcs.taiyouxi.net/jws/multiplayer/match_server/notify"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type MultiplayServerValue struct {
	ServerID string `json:"sid"`
	GameNum  int    `json:"n"`
}

//取消排队匹配api/v2/match/{boss}?cancel=1
//开始排队匹配api/v2/match/{boss}
func MatchServer(g *gin.Engine) {
	logs.Info("reg match server service")
	//V2 url
	g.POST(helper.MatchPostUrlAddressV2+"/:boss", func(c *gin.Context) {
		s := helper.MatchValue{}
		err := c.Bind(&s)

		boss := c.Param("boss")
		cancelStr := c.Query("cancel")
		matchToken := c.Query("token")
		if matchToken == "" {
			matchToken = helper.MatchDefaultToken
		}

		cancel := false
		if cancelStr == "1" {
			cancel = true
		}

		if err != nil {
			logs.Error("MatchServer match boss, Bind err %s", err.Error())
			c.String(400, err.Error())
			return
		}

		logs.Info("matchv2 acid %s, hard %v, cancel %v, token:%s, boss:%s",
			s.AccountID, s.IsHard, cancel, matchToken, boss)
		notify.GetNotify(matchToken)
		err = match.GVEMatchV2_GetOrCreate(matchToken, boss).AddPlayer(s.AccountID, s.CorpLv, s.IsHard, cancel)

		if err == nil {
			c.String(200, string("ok"))
		} else {
			c.String(401, err.Error())
		}
	})

	//V1 url
	//g.POST(helper.MatchPostUrlAddress, func(c *gin.Context) {
	//	s := helper.MatchValue{}
	//	err := c.Bind(&s)
	//
	//	if err != nil {
	//		c.String(400, err.Error())
	//		return
	//	}
	//
	//	logs.Info("match %s, %s", s.AccountID, s.IsHard)
	//
	//	err = match.GetGVEMatch().AddPlayer(s.AccountID, s.CorpLv, s.IsHard, false)
	//
	//	if err == nil {
	//		c.String(200, string("ok"))
	//	} else {
	//		c.String(401, err.Error())
	//	}
	//})

	//g.POST(helper.MatchCancelPostUrlAddress, func(c *gin.Context) {
	//	s := helper.MatchValue{}
	//	err := c.Bind(&s)
	//
	//	if err != nil {
	//		c.String(400, err.Error())
	//		return
	//	}
	//
	//	logs.Info("matchCancel %s, %s", s.AccountID, s.IsHard)
	//	err = match.GetGVEMatch().AddPlayer(s.AccountID, s.CorpLv, s.IsHard, true)
	//
	//	if err == nil {
	//		c.String(200, string("ok"))
	//	} else {
	//		c.String(401, err.Error())
	//	}
	//})
}
