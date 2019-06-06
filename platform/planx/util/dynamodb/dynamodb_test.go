package dynamodb

import (
	"testing"
	"time"

	"github.com/cenk/backoff"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	Dynamo_DB     = "Mail"
	AWS_Region    = "cn-north-1"
	AWS_AccessKey = "AKIAPLNPSQYENX3LGB5A"
	AWS_SecretKey = "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"
)

func TestDynamoDBScan(t *testing.T) {
	t.Logf("Start")

	db := &DynamoDB{}

	err := db.Connect(
		AWS_Region,
		AWS_AccessKey,
		AWS_SecretKey,
		"")

	if err != nil {
		logs.Error("Connect Err %s", err.Error())
	}

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 500 * time.Millisecond
	b.MaxElapsedTime = 10 * time.Second

	//db.SyncMail(Dynamo_DB, "0:0:1001", []int64{2222, 3333}, []string{"ssss", "dsadasd"}, []int64{12, 232})

	//user_id string, time_min, time_max, limit int64
	ids, mails, err := db.QueryMail(Dynamo_DB, "profile:0:0:1331", 0, 2436413452, 100)
	logs.Info("ids %v", ids)
	logs.Info("mails %v", mails)

	logs.Trace("in %d", time.Now().Unix())
	logs.Trace("in %d", time.Now().UnixNano())
}

func TestConst(t *testing.T) {
	t.Logf("Mail_Send_By_Sys %v", Mail_Send_By_Sys)
	t.Logf("Mail_Send_By_GM %v", Mail_Send_By_GM)
	t.Logf("Mail_Send_By_Rank %v", Mail_Send_By_Rank_SimplePvp)
	t.Logf("Mail_Send_By_TMP %v", Mail_Send_By_TMP)
}
