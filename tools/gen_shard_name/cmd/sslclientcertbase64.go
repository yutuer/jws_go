package cmd

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"vcs.taiyouxi.net/platform/planx/util/secure"

	"github.com/codegangsta/cli"
)

const (
	v1dict = "ja-1qD40zXBO5kNHfcUiTYgVlMeQwLP7bmEsxo3RW_ZKFSnA86hGtIy9rdpC2uvJ"
	v1salt = "798hRbiyMOC" //这个不重要，可以乱写
)

func SslClientCertBase64(c *cli.Context) {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln(err.Error())
	}
	se := secure.New(v1dict, v1salt)

	if c.Bool("decode") {
		raw, err := se.Decode64FromNet(string(bytes))
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("raw md5:%x\n", md5.Sum(raw))

	} else {
		fmt.Printf("raw md5:%x\n", md5.Sum(bytes))
		es := se.Encode64ForNet(bytes)
		fmt.Printf("%s\n", es)
		raw, err := se.Decode64FromNet(es)
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("raw md5:%x\n", md5.Sum(raw))
	}

}

func init() {
	register(&cli.Command{
		Name:   "sslclientencrypt",
		Usage:  "类似cat client.pfx|base64中的base64",
		Action: SslClientCertBase64,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "decode, D",
				Usage: "用来在验证解密后md5",
			},
		},
	})
}
