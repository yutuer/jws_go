package main

type FilterConfig struct {
	Name   []string
	Output []Output
}
type Output struct {
	Table     string
	Item      []string
	Ignore    []string
	Partition string
	KeyMap    []string
}
type FilterConfigs struct {
	FilterConfig []FilterConfig
	DynamoConfig DynamoConfig
}

type DynamoConfig struct {
	DynamoRegion          string `toml:"dynamo_region"`
	DynamoAccessKeyID     string `toml:"dynamo_accessKeyID"`
	DynamoSecretAccessKey string `toml:"dynamo_secretAccessKey"`
	DynamoSessionToken    string `toml:"dynamo_sessionToken"`
}
