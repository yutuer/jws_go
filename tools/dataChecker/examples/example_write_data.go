package examples

import (
	"github.com/golang/protobuf/proto"
	"os"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

// WriteBuff2Bin 将buff写入文件
func WriteBuff2Bin(binFilename string, buff []byte) (err error) {
	f, err := os.Create(binFilename)
	if err != nil {
		return
	}

	defer f.Close()

	_, err = f.Write(buff)
	if err != nil {
		return
	}

	err = f.Sync()

	return
}

// PutTestData2File 将生成的测试数据写入文件
func PutTestData2File(filename string) {
	h1 := newHotActivityTime()
	h2 := newHotActivityTime()

	hs := &ProtobufGen.HotActivityTime_ARRAY{
		Items: []*ProtobufGen.HotActivityTime{h1, h2},
	}

	buff, err := proto.Marshal(hs)
	if err != nil {
		panic(err)
	}

	WriteBuff2Bin(filename, buff)
}

func newHotActivityTime() (h *ProtobufGen.HotActivityTime) {
	h = &ProtobufGen.HotActivityTime{
		ActivityID:     new(uint32),
		ActivityType:   new(uint32),
		ActivityPID:    new(uint32),
		ActivityValid:  new(uint32),
		ActivityTitle:  new(string),
		TimeType:       new(uint32),
		StartTime:      new(string),
		EndTime:        new(string),
		Duration:       new(uint32),
		ServerGroupID:  new(uint32),
		ChannelGroupID: new(uint32),
		TeleID:         new(string),
		ConditionID:    new(uint32),
		HotActivity:    new(uint32),
		TabIDS:         new(string),
	}
	return
}
