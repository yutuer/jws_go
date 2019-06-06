package imp

type Config struct {
	GidSid    string `toml:"gidsid"`

	Redis     string `toml:"redis"`
	RedisDB   int    `toml:"redis_db"`
	RedisAuth string `toml:"redis_auth"`

	AccountNum      int      `toml:"account_num"`
	AccountUidCsv   string   `toml:"account_uid_csv"`
	AccountJsons    []string `toml:"account_jsons"`
	AccountSplitNum []int    `toml:"account_split_num"`
}

var (
	Cfg Config
)
