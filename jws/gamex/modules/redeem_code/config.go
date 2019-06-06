package redeemCodeModule

type config struct {
	AWS_Region    string `toml:"mail_aws_region"`
	AWS_AccessKey string `toml:"mail_aws_accessKey"`
	AWS_SecretKey string `toml:"mail_aws_secretKey"`
	Db_Name       string `toml:"mail_db_name"`
	MongoURL      string `toml:"mongo_url"`
	DBDriver      string `toml:"db_driver"`
}

var cfg config

func SetConfig(region, access_key, secret_key, mongoUrl, db_driver, db_name string) {
	cfg.AWS_Region = region
	cfg.AWS_AccessKey = access_key
	cfg.AWS_SecretKey = secret_key
	cfg.Db_Name = db_name
	cfg.MongoURL = mongoUrl
	cfg.DBDriver = db_driver
}
