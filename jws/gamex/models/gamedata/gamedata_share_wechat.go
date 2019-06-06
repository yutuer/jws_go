package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

type share_key struct {
	Type  uint32
	Times uint32
}

var (
	shareWeChatData map[share_key]*ProtobufGen.SCLSHARE
)

func GetShareWeChatData() map[share_key]*ProtobufGen.SCLSHARE {
	return shareWeChatData
}

func GetShareWeChatDataByKey(Type uint32, Times uint32) *ProtobufGen.SCLSHARE {
	return shareWeChatData[share_key{Type: Type, Times: Times}]
}

func loadShareWeChatData(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.SCLSHARE_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	as := ar.GetItems()
	shareWeChatData = make(map[share_key]*ProtobufGen.SCLSHARE, len(as))
	for _, v := range as {
		shareWeChatData[share_key{Type: v.GetShareType(), Times: v.GetShareCount()}] = v
	}
}
