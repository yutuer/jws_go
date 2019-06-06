package gamedata

import (
	"fmt"

	"strings"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/servers/game"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var (
	gdGameId        string
	gdQuick2Channel map[string]string
)

const CORRECT_CHANNEL_DIGIT = 6

func loadSerChannelConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.BRANCHANDROID13_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	items := dataList.GetItems()
	gdQuick2Channel = make(map[string]string, len(items))

	for _, item := range items {
		if strings.Contains(item.GetChannelID(), ".") {
			panic(fmt.Errorf("ChannelId %s has wrong format ", item.GetChannelID()))

		}
		if item.GetChannelID() != "0" {
			if strings.Count(item.GetChannelID(), "")-1 != CORRECT_CHANNEL_DIGIT {
				panic(fmt.Errorf("ChannelId %s has wrong digitally ", item.GetChannelID()))
			}
		}

		gdQuick2Channel[item.GetQuickSDKChannelID()] = item.GetChannelID()
	}
}

func loadChannelConstConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.CHANNELCONST_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	items := dataList.GetItems()
	gdGameId = items[0].GetGameID()
}

func GetChannelId(quickId string) string {
	if quickId == "" || quickId == "-1" {
		return quickId
	}
	q2c, ok := gdQuick2Channel[quickId]
	if !ok {
		return quickId
	}
	gameid := game.Gid2Channel[game.Cfg.Gid]
	if uutil.IsVNVer() {
		if quickId == "5003" {
			gameid = "13"
		} else if quickId == "5004" {
			gameid = "11"
		} else {
			logs.Error("VN channel err, quickId %s", quickId)
		}
	}
	return fmt.Sprintf("%s%s%s", gameid, gdGameId, q2c)
}
