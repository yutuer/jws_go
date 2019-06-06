package examples

import (
	"github.com/golang/protobuf/proto"
	"io"
	"os"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

// LoadBin2Buff 将.data文件读取为 []byte
func LoadBin2Buff(binFilename string) ([]byte, error) {
	file, err := os.Open(binFilename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, fi.Size())
	_, err = io.ReadFull(file, buffer)
	if err != nil {
		return buffer, err
	}

	return buffer, nil
}

// GetHotActivityTime 从data文件中读取并生成HotActivityTime
func GetHotActivityTime() []*ProtobufGen.HotActivityTime {
	HATFilename := GetDataFileFullPath("hotactivitytime")
	buff, err := LoadBin2Buff(HATFilename)
	if err != nil {
		panic(err)
	}

	HATs := &ProtobufGen.HotActivityTime_ARRAY{}
	err = proto.Unmarshal(buff, HATs)
	if err != nil {
		panic(err)
	}

	return HATs.Items
}


