package ftp

var (
	FtpAddress string
	User       string
	Passwd     string
	Path       string
)

func SetCfg(address, user, pass, path string) {
	FtpAddress = address
	User = user
	Passwd = pass
	Path = path
}
