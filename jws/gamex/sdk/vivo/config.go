package vivo

type VivoConfig struct {
	Url          string `toml:"url"`
	SdkNotifyUrl string `toml:"sdk_notify_url"`
	AppID        string `toml:"app_id"`
	CPId         string `toml:"cp_id"`
	CPKey        string `toml:"cp_key"`
}

var Cfg VivoConfig
