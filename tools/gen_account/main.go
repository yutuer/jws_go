package main

import (
	"log"

	"crypto/md5"
	"encoding/base64"
	"fmt"

	"github.com/astaxie/beego/httplib"

	"math/rand"
	"time"

	"bufio"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"vcs.taiyouxi.net/platform/planx/util/secure"
	authconfig "vcs.taiyouxi.net/platform/x/auth/config"
	"vcs.taiyouxi.net/tools/gen_account/imp"
)

type AccountInfo struct {
	Name     string
	Password string
	Uid      string
}

var (
	account_infos []AccountInfo
)

func main() {
	rand.Seed(time.Now().UnixNano())

	if _, err := toml.DecodeFile("conf/config.toml", &imp.Cfg); err != nil {
		log.Fatalf("toml err %v", err)
		return
	}
	log.Println("conf ", imp.Cfg)

	err := imp.Init()
	if err != nil {
		log.Fatalf("imp.init err %v", err.Error())
		return
	}

	account_infos = make([]AccountInfo, 0, imp.Cfg.Count)
	for i := 1; i < imp.Cfg.Count+1; i++ {
		name := fmt.Sprintf("%s%03d", imp.Cfg.Prefix, i)
		_name := base64.StdEncoding.EncodeToString([]byte(name))
		pass := randPassword()

		md5Ctx := md5.New()
		md5Ctx.Write([]byte(pass))
		_pass := secure.DefaultEncode.Encode64ForNet(md5Ctx.Sum(nil))

		deviceId := fmt.Sprintf("%s@%d%03d@@apple.com", "gen_account", time.Now().Unix(), i)
		url := fmt.Sprintf("http://%s/auth/v1/user/reg/%s?name=%s&passwd=%s",
			imp.Cfg.AuthUrl, deviceId, _name, _pass)
		//log.Println(deviceId, _name, _pass)
		req := httplib.Get(url)
		req.Header(authconfig.Spec_Header, authconfig.Spec_Header_Content)

		_, err := req.String()
		if err != nil {
			log.Fatalf("err %s", err.Error())
			return
		}

		//log.Println(str)
		uid, err := imp.GetUid(name)
		if err != nil {
			log.Fatalf("imp.GetUid err %v", err.Error())
			return
		}
		log.Println(name, pass, uid, deviceId)
		account_infos = append(account_infos, AccountInfo{
			Name:     name,
			Password: pass,
			Uid:      uid,
		})

		if imp.Cfg.WriteAccount {
			if !imp.WriteAccount(fmt.Sprintf("%s:%s", imp.Cfg.GidSid, uid)) {
				log.Fatalf("imp.WriteAccount fail")
				return
			}
		}
	}

	fn := filepath.Join(imp.Cfg.OutputPath, "gen_account.csv")
	file, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("outPut OpenFile err %s", err.Error())
		return
	}
	defer file.Close()

	bufWriter := bufio.NewWriterSize(file, 10240)
	bufWriter.WriteString("name,password,acid\r\n")

	for _, v := range account_infos {
		bufWriter.WriteString(fmt.Sprintf("%s,%s,%s\r\n", v.Name, v.Password, v.Uid))
	}
	bufWriter.Flush()
}

func randPassword() string {
	return RandStringRunes(6)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
