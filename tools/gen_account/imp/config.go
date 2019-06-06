package imp

type Config struct {
	GidSid     string `toml:"gidsid"`
	Prefix     string `toml:"prefix"`
	Count      int    `toml:"count"`
	OutputPath string `toml:"output_path"`

	DynamoRegion          string `toml:"dynamo_region"`
	DynamoAccessKeyID     string `toml:"dynamo_accessKeyID"`
	DynamoSecretAccessKey string `toml:"dynamo_secretAccessKey"`
	DynamoDBName          string `toml:"dynamo_db_Name"`

	AuthUrl string `toml:"auth_url"`

	Redis     string `toml:"redis"`
	RedisDB   int    `toml:"redis_db"`
	RedisAuth string `toml:"redis_auth"`

	WriteAccount bool   `toml:"write_account"`
	AccountJson  string `toml:"account_json"`
}

var (
	Cfg Config
)
