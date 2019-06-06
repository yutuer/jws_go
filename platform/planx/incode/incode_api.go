package incode

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

const api_secret = "581c541ed5599aaefec0f41fad488ea4f7948f485f1e0e0e995ac6c770b40361"

// {"error_code":0,"code":"18C3F552","status":"0","sign":"c9d2bd15efb515f286c6429a44ff8c3e"}
type CheckIsCodeVerifyResp struct {
	Code   int    `json:"error_code"`
	Status string `json:"status"`
	Error  string `json:"error"`
}

func CheckIsCodeVerify(api_key, code string) (bool, error) {
	v := url.Values{}
	v.Set("sign", getSign(code))
	v.Set("api_key", api_key)
	v.Set("code", code)
	fmt.Println(v.Encode())
	resp, err := http.Get("http://incode.fir.im/api/verify?" + v.Encode())
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))

	var resp_info CheckIsCodeVerifyResp
	resp_info.Code = -1
	err = json.Unmarshal(body, &resp_info)
	if err != nil {
		return false, err
	}

	if resp_info.Code != 0 {
		return false, errors.New("APIErr:" + resp_info.Error)
	}

	return resp_info.Status == "0", nil
}

var CodeHasUsed = errors.New("CodeHasUsed")

func UseCode(api_key, code string) error {
	v := url.Values{}
	v.Set("sign", getSign(code))
	v.Set("api_key", api_key)
	v.Set("code", code)
	fmt.Println(v.Encode())
	resp, err := Put("http://incode.fir.im/api/occupy?" + v.Encode())

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))

	var resp_info CheckIsCodeVerifyResp
	resp_info.Code = -1
	err = json.Unmarshal(body, &resp_info)
	if err != nil {
		return err
	}

	if resp_info.Code != 0 {
		if resp_info.Code == 1004 {
			return CodeHasUsed
		} else {
			return errors.New("APIErr:" + resp_info.Error)
		}
	}

	return nil
}

func Put(url string) (*http.Response, error) {
	client := &http.Client{}
	reqest, _ := http.NewRequest("PUT", url, nil)

	return client.Do(reqest)
}

func getSign(code string) string {
	h := md5.New()
	io.WriteString(h, api_secret+code)
	keyMd5 := hex.EncodeToString(h.Sum(nil))
	return keyMd5
}
