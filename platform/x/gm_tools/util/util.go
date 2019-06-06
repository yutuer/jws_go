package util

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func HttpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err.Error())
		return []byte{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println(err.Error())
		return []byte{}, err
	}

	return body, nil
}

func HttpPost(url, typ string, data []byte) ([]byte, error) {
	body := bytes.NewBuffer([]byte(data))
	resp, err := http.Post(url, typ, body)
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()
	body_res, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println(err.Error())
		return []byte{}, err
	}

	return body_res, nil
}

func HttpDelete(url string) ([]byte, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	resp, err := http.DefaultClient.Do(req)

	defer resp.Body.Close()
	body_res, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println(err.Error())
		return []byte{}, err
	}

	return body_res, nil
}

func DeleteBackslash(s string) string {
	res := s
	if res == "" {
		return ""
	}
	for res[0] == '/' {
		res = res[1:]
	}
	return res
}

func ToJson(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
