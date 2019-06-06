package gamedata

import "vcs.taiyouxi.net/jws/gamex/protogen"

var gdFriendConfig *ProtobufGen.FRIENDCHEAT

func loadFriendConfig(filepath string) {
	ar := &ProtobufGen.FRIENDCHEAT_ARRAY{}
	_common_load(filepath, ar)
	data := ar.GetItems()
	if data[0] == nil {
		panic("error data")
	}
	gdFriendConfig = data[0]
}

func GetFriendConfig() *ProtobufGen.FRIENDCHEAT {
	return gdFriendConfig
}
