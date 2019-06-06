package imp

type CommonConfig struct {
	AWS_Region       string
	AWS_AccessKey    string
	AWS_SecretKey    string
	PayTable         string   `toml:"pay_table"`
	ResPath          string   `toml:"res_path"`
	GidShard         []string `toml:"gidshard"`
	TimeBegin        string   `toml:"time_begin"`
	TimeDivision     string   `toml:"time_division"`
	CorrectFromRedis bool
	Redis            string
	RedisAuth        string
	RedisDB          int
}

var (
	Cfg CommonConfig
)
