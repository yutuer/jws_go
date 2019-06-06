package samsung

type SamsungConfig struct {
	Url          string `toml:"url"`
	SdkNotifyUrl string `toml:"sdk_notify_url"`
	AppID        string `toml:"app_id"`
	PrivateKey   string `toml:"private_key"`
	PublicKey    string `toml:"public_key"`
}

var Cfg SamsungConfig
