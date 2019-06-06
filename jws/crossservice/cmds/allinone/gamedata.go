package allinone

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"vcs.taiyouxi.net/jws/crossservice/util/http_util"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

func regGinHandles() {
	http_util.GET("/ShowGamedataVer", ginShowGamedataVer)
}

func ginShowGamedataVer(c *gin.Context) {
	out := fmt.Sprintf("Gamedata Ver: Build: %s(%d), Data: (%s), Hash: %s\n", ProtobufGen.BuildBranch, ProtobufGen.Build, ProtobufGen.BuildDate, ProtobufGen.BuildHash)

	c.String(200, out)
}
