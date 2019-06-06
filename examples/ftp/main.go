package main

import (
	//"crypto/tls"

	"vcs.taiyouxi.net/platform/planx/util/logs"

	"bytes"

	"github.com/dutchcoders/goftp"
)

func main() {
	var err error
	var ftp *goftp.FTP

	// For debug messages: goftp.ConnectDbg("ftp.server.com:21")
	if ftp, err = goftp.Connect("proxy.wsfdupload.lxdns.com:21"); err != nil {
		panic(err)
	}

	defer ftp.Close()

	if err = ftp.Login("ifsgcdn", "ifsgcdn#1118"); err != nil {
		panic(err)
	}

	if err = ftp.Cwd("/"); err != nil {
		panic(err)
	}

	var curpath string
	if curpath, err = ftp.Pwd(); err != nil {
		panic(err)
	}

	logs.Trace("Current path: %s", curpath)

	var files []string
	if files, err = ftp.List(""); err != nil {
		panic(err)
	}

	logs.Trace("files %v", files)

	data := "hjkl;adfsjkl;adfsjkl;adfshjkladfshjkladsfhjklasdfhjkadfshjkl000"

	r := bytes.NewReader([]byte(data))

	if err := ftp.Mkd("/public/"); err != nil {
		panic(err)
	}

	if err := ftp.Stor("/public/test.txt", r); err != nil {
		panic(err)
	}

}
