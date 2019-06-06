package metrics

type Config struct {
	GraphiteValid         bool   `toml:"graphite_valid"`
	GraphiteHost          string `toml:"graphite_host"`
	GraphiteMemStats      bool   `toml:"graphite_memstats"`
	GraphiteGCStats       bool   `toml:"graphite_gcstats"`
	GraphiteFlushInterval int64  `toml:"graphite_flush_interval"`
	GraphitePrefix        string `toml:"graphite_prefix"`
	SimpleGraphite        bool   `toml:"graphite_simple"`
	//GraphiteOutPutStdErr  bool   `toml:"graphite_output_stderr"`
}
