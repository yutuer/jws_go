package base

type Config struct {
	Host              string
	Url               string
	MaxListLen        int
	LegalInterval     int64
	MaxRetainMsgCount int
	SimilarityVal     float64
	SimilarityGagTime int64
	SmtpInfo          string
	SendFrom          string
	SendTo            []string
}

var Cfg Config
