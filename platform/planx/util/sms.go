package util

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	serverURL = "http://sms.yingxiong.com/sms"
	group     = "ifsgsms"
	secret    = "0b316addef9771d7af94a04ddf0ceef5"
)

func SendHeroSMS(phone string, msg string) error {
	md5sum := md5.Sum([]byte(fmt.Sprintf("%s%s%s", phone, msg, secret)))

	data_send_sms := url.Values{
		"group":   {group},
		"phone":   {phone},
		"content": {msg},
		"type":    {"0"},
		"sign":    {fmt.Sprintf("%x", md5sum)}}
	return httpsPostForm(serverURL, data_send_sms)
}

func httpsPostForm(url string, data url.Values) error {
	logs.Trace("httpsPostForm url %s", url)
	resp, err := http.PostForm(url, data)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	logs.Trace("httpsPostForm Res %s", string(body))
	return nil
}
