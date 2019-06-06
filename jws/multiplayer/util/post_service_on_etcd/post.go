package postService

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func HttpPost(url, typ string, data []byte) ([]byte, error) {
	body := bytes.NewBuffer([]byte(data))
	c := &http.Client{Timeout: 65 * time.Second}
	resp, err := c.Post(url, typ, body)
	if err != nil {
		logs.Error("postService HttpPost err %s %s %d %s",
			url, typ, len(data), err.Error())
		return []byte{}, err
	}

	defer resp.Body.Close()
	body_res, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logs.Error("postService HttpPost readres err %s %s %s",
			url, typ, err.Error())
		return []byte{}, err
	}

	return body_res, nil
}
