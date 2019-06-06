package mbot

import (
	"io"
	"net/http"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"

	"github.com/gin-gonic/gin"
)

var _bf *BotFactory

func api(bf *BotFactory) {
	_bf = bf
	r := gin.Default()
	// Global middlewares
	r.Use(gin.Logger())
	//r.Use(gin.Recovery())

	r.LoadHTMLGlob("resource/templ/*.templ.html")
	r.Static("/bower_components", "resource/bower_components")
	r.Static("/myjs", "resource/myjs")
	r.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"LiveBots": bf.GetNumberLiveBots(),
			"CCU":      bf.GetCCU(),
		})
	})
	r.GET("/", index)
	r.GET("/stream", stream)
	r.GET("/addbot", addbot)

	websocket_handler(r)

	// Listen and serve on 0.0.0.0:8080
	r.Run(":9080")
}

func addbot(c *gin.Context) {
	c.Request.ParseForm()
	method := c.Request.Form.Get("m")
	logs.Info("addbot %s", method)
	switch method {
	case "SimpleGenerator":
		type SimpleGeneratorForm struct {
			Number int `form:"botnumber"`
		}
		var form SimpleGeneratorForm
		c.Bind(&form)
		go _bf.SimpleGenerator(_bf.GetLogs()[0], form.Number)
	case "RandomGenerator":
		type RandomGeneratorForm struct {
			Number     int `form:"botnumber"`
			OnceNumber int `form:"oncenumber"`
			Sleep      int `form:"sleep"`
		}
		var form RandomGeneratorForm
		c.Bind(&form)
		go _bf.RandomGenerator(_bf.GetLogs(), form.Number, form.OnceNumber, form.Sleep)
	}
	c.JSON(http.StatusOK, gin.H{
		"method": method,
		"status": "ok",
	})
}

func index(c *gin.Context) {
	c.HTML(200, "stats.templ.html", gin.H{
		"timestamp": time.Now().Unix(),
	})
}

func stream(c *gin.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-ticker.C:
			c.SSEvent("stats", Stats())
		}
		return true
	})
}

func Stats() map[string]uint64 {
	//mutexStats.RLock()
	//defer mutexStats.RUnlock()
	savedStats := map[string]uint64{
		"timestamp": uint64(time.Now().Unix()),
		"LiveBots":  uint64(_bf.GetNumberLiveBots()),
		"CCU":       uint64(_bf.GetCCU()),
	}
	return savedStats
}
