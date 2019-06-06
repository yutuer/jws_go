package pay

type PayDBConfig struct {
	AWSRegion    string
	DBName       string
	AWSAccessKey string
	AWSSecretKey string

	MongoDBUrl string
	DBDriver   string
}

type PayDB interface {
	SetByHashM(hash interface{}, values map[string]interface{}) error
	SetByHashM_IfNoExist(hash interface{}, values map[string]interface{}) error
	GetByHashM(hash interface{}) (map[string]interface{}, error)
	UpdateByHash(hash interface{}, values map[string]interface{}) error
	//IsExist 被IsIOSOrderRepeat调用,用来判断是不是有重复的订单, 此处并不在意返回值的内容
	IsExist(hash interface{}) (bool, error)
	// QueryByUid 只应该在Gmtools中调用,因为如果支付内容很多,开销会很大
	QueryByUid(ucid string) ([]GateWay_Pay_Info, error)
}

type GateWay_Pay_Info struct {
	Order            string `json:"order_no" bson:"order_no"`
	SN               int64  `json:"sn" bson:"sn"`
	Uid              string `json:"uid" bson:"uid"`
	AccountName      string `json:"account_name" bson:"account_name"`
	RoleName         string `json:"role_name" bson:"role_name"`
	GoodIdx          string `json:"good_idx" bson:"good_idx"`
	GoodName         string `json:"good_name" bson:"good_name"`
	MoneyAmount      string `json:"money_amount" bson:"money_amount"`
	Platform         string `json:"platform" bson:"platform"`
	Channel          string `json:"channel" bson:"channel"`
	Status           string `json:"status" bson:"status"`
	TiStatus         string `json:"tistatus" bson:"tistatus"`
	HcBuy            uint32 `json:"hc_buy" bson:"hc_buy"`
	HcGive           uint32 `json:"hc_give" bson:"hc_give"`
	PayTimeS         string `json:"pay_time_s" bson:"pay_time_s"`
	PayTime          string `json:"pay_time" bson:"pay_time"`
	ReceiveTimeStamp int64  `json:"receiveTimestamp" bson:"receiveTimestamp"`
	IsTest           string `json:"is_test" bson:"is_test"`
}
