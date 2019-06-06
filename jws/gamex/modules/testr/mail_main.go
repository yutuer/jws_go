package main

import (
	"fmt"
	"os"

	"time"

	"vcs.taiyouxi.net/jws/gamex/models/driver"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/mail/mailhelper"
	"vcs.taiyouxi.net/jws/gamex/modules"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

// 测试batch mail 发送
type GameConfig struct {
	GameConfig game.Config
}

const (
	cfgName   = "../../../conf/gate.toml"
	logConf   = "../../../conf/glog.xml"
	shard     = 10
	batch_cap = 2000
)

func main() {
	defer logs.Close()
	logs.LoadLogConfig(logConf)

	var gameConfig GameConfig
	cfgApp := config.NewConfigToml(cfgName, &gameConfig)
	game.Cfg = gameConfig.GameConfig
	if cfgApp == nil {
		logs.Critical("\n")
		os.Exit(1)
	}

	logs.Info("gamex config %v", game.Cfg)

	driver.SetupRedis(
		game.Cfg.Redis,
		game.Cfg.RedisDB,
		game.Cfg.RedisAuth,
		gamedata.IsGameDevMode(),
	)
	mailcfg := mailhelper.MailConfig{
		DBDriver:     game.Cfg.MailDBDriver,
		AWSRegion:    game.Cfg.AWS_Region,
		DBName:       game.Cfg.MailDBName,
		AWSAccessKey: game.Cfg.AWS_AccessKey,
		AWSSecretKey: game.Cfg.AWS_SecretKey,
		MongoDBUrl:   game.Cfg.MailMongoUrl,
	}
	mail_sender.SetConfig(mailcfg)

	shards := []uint{shard}
	modules.StartModule(shards)

	reason := "test"
	time_now := time.Now().Unix()
	mails := make([]timail.MailKey, 0, batch_cap)
	for i := 0; i < batch_cap; i++ {
		uid := fmt.Sprintf("0:0:%d", i)
		mail := timail.MailReward{
			IdsID:     mail_sender.IDS_MAIL_SIMPLEPPVP_RANKREWARD_TITLE,
			Param:     []string{fmt.Sprintf("%d", i)},
			TimeBegin: time_now,
			Reason:    reason,
			TimeEnd:   time_now + util.WeekSec,
			Idx:       timail.MkMailId(timail.Mail_Send_By_Rank_SimplePvp, int64(i*10)),
			ItemId:    []string{"VI_PVPC", "VI_DC"},
			Count:     []uint32{2500, 3750},
		}

		mails = append(mails, timail.MailKey{
			Uid: uid,
			Idx: mail.Idx,
		})

		mail_sender.AddMailBatchSend(shard, uid, mail, false, timail.Mail_send_By_Debug)
	}

	time.Sleep(time.Second * 10)

	/* 接口不存在了
	for _, k := range mails {
		if !mail_sender.MailExist(fmt.Sprintf("profile:%s", k.Uid), k.Idx) {
			logs.Error("%v not rec", k)
		}
	}
	*/

	logs.Info("finish test mail len %d", len(mails))
	modules.StopModule()
}
