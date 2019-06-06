package qiniu

import (
	"testing"
	"time"
)

func TestUpLoad(t *testing.T) {
	SetCfg("_N4StMhwFM9oSkQ3SqDmpN1LGLfDljZ7K_-xmc-Q", "4AiW-FXGBn3YUWjNyD-IW40gRskx20XTfq1zmi49")
	err := UpLoad("jwscdn", "tttttest", []byte("dfsafadsfadsfasd555kfbwefjrhoCHsadkfhnaksfh000"))
	if err != nil {
		t.Logf("err : %s", err.Error())
	}
	time.Sleep(time.Second)
}
