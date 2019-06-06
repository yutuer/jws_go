package config

type CommonConfig struct {
	Runmode string `toml:"runmode"`

	Dynamo_region          string `toml:"dynamo_region"`
	Dynamo_accessKeyID     string `toml:"dynamo_accessKeyID"`
	Dynamo_secretAccessKey string `toml:"dynamo_secretAccessKey"`

	Httpport  string `toml:"httpport"`
	SentryDSN string `toml:"SentryDSN"`
}

var (
	Cfg CommonConfig
)

const (
	Spec_Header         = "TYX-Request-Id"
	Spec_Header_Content = "0a2e427a115aa9a7130d00cb08316455"
)

func (c *CommonConfig) IsRunModeProdAndTest() bool {
	return c.Runmode == "prod" || c.Runmode == "test"
}

func (c *CommonConfig) IsRunModeLocal() bool {
	return c.Runmode == "local"
}
