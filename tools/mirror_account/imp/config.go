package imp

type Config struct {
	RankRedis     string `toml:"rank_redis"`
	RankRedisAuth string `toml:"rank_redis_auth"`
	RankRedisDB   int    `toml:"rank_redis_db"`
	RankTable     string `toml:"rank_table"`

	DynamoRegion string `toml:"dynamo_region"`
	DynamoGMInfo string `toml:"dynamodb_gm"`

	GamexLogicLog     string   `toml:"gamex_logic_log"`
	AuthLogicLog      string   `toml:"auth_logic_log"`
	MirrorAccountName []string `toml:"mirror_account_name"`
}

var (
	Cfg Config
)

func (c *Config) FindAccountName(n string) bool {
	for _, mn := range c.MirrorAccountName {
		if n == mn {
			return true
		}
	}
	return false
}
