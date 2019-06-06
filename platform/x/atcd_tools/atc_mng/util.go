package atc_mng

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
)

func httpGet(url string) ([]byte, error) {
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

func httpPost(url, typ string, data []byte) ([]byte, error) {
	body := bytes.NewBuffer([]byte(data))
	resp, err := http.Post(url, typ, body)
	if err != nil {
		log.Fatal(err)
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

func httpDelete(url string) ([]byte, error) {
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
