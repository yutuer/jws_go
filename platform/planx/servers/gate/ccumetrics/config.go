package ccumetrics

type LoginStatus struct {
	LoginToken string
	AccountID  string
	LogInOff   bool
}

type Config struct {
	EtcdEndpoint []string `toml:"etcd_endpoint"`
	EtcdRoot     string   `toml:"etcd_root"`
	LoginUrl     string   `toml:"loginurl"`

	URINotifyLogin  string `toml:"URI_login"`
	URINotifyLogout string `toml:"URI_logout"`
	URICcu          string `toml:"URI_ccu"`

	LoginConnectorTickTime int64 `toml:"LoginConnector_tick_time"`
}
