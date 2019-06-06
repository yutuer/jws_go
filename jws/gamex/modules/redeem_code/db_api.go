package redeemCodeModule

import "vcs.taiyouxi.net/platform/planx/util/logs"

const (
	RedeemCodeStateNew = iota
	RedeemCodeStateDone
	RedeemCodeStateOutdatedDone
)

//
/*
	DoneBy string  用户gid:sid:uuid
	Bind   string  gid, gid:sid,
	State  int     New|Done|OutdatedDone
	Begin  int     开始时间
	End    int     结束时间
	Value  string  Json内容 {
		BatchID string   `json:"bid"`
		GroupID string   `json:"gid"`
		ItemIDs []string `json:"items"`
		Counts  []uint32 `json:"counts"`
	}
*/
type RedeemCodeExchange struct {
	DoneBy  string   `bson:"DoneBy,omitempty" json:"doneBy"`
	Bind    string   `bson:"Bind,omitempty" json:"bind"` // gid, gid:sid,
	State   int64    `bson:"State" json:"state"`         // New|Done|OutdatedDone
	Begin   int64    `bson:"Begin,omitempty" json:"begin"`
	End     int64    `bson:"End,omitempty" json:"end"`
	Code    string   `bson:"code" json:"code"`
	Value   string   `bson:"Value,omitempty"`
	BatchID string   `bson:"bid,omitempty" json:"bid,omitempty"`
	GroupID string   `bson:"gid,omitempty" json:"gid,omitempty"`
	ItemIDs []string `bson:"items,omitempty" json:"items,omitempty"`
	Counts  []uint32 `bson:"counts,omitempty" json:"counts,omitempty"`
	Title   string   `bson:"title,omitempty" json:"title,omitempty"`
}

type RedeemCodeDB interface {
	Open() error
	GetRedeemCodeData(code string) (*RedeemCodeExchange, error)
	SetRedeemCodeUsed(acID string, code string) error
}

var db RedeemCodeDB

func initDB() error {
	if cfg.DBDriver == "MongoDB" {
		db = NewRedeemCodeMongoDB(
			cfg.MongoURL,
			cfg.Db_Name)
	} else if cfg.DBDriver == "DynamoDB" {
		db = NewRedeemCodeDynamoDB(
			cfg.AWS_Region,
			cfg.Db_Name,
			cfg.AWS_AccessKey,
			cfg.AWS_SecretKey)
	}
	return db.Open()
}

func getRedeemCodeData(code string) *RedeemCodeExchange {
	data, err := db.GetRedeemCodeData(code)
	if err != nil {
		logs.Warn("isRedeemCodeHasUsed Err %s", err.Error())
		return nil
	}

	return data
}

func setRedeemCodeUsed(acID, code string) error {
	return db.SetRedeemCodeUsed(acID, code)
}
