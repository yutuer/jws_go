package gamedata

import (
	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

const (
	GuildInventoryHeapUpLimit = 999
)

var (
	gdGuildInventoryItems map[string]*ProtobufGen.GUILDBAG
	gdLostGoodShopItems   map[string]*ProtobufGen.LOSTGOODSHOP
)

func loadGuildInventory(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.GUILDBAG_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdGuildInventoryItems = make(map[string]*ProtobufGen.GUILDBAG, len(data))
	for _, d := range data {
		gdGuildInventoryItems[d.GetGoodID()] = d
	}
}

func loadGuildLostInventory(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ar := &ProtobufGen.LOSTGOODSHOP_ARRAY{}
	err = proto.Unmarshal(buffer, ar)
	errcheck(err)

	data := ar.GetItems()
	gdLostGoodShopItems = make(map[string]*ProtobufGen.LOSTGOODSHOP, len(data))
	for _, d := range data {
		gdLostGoodShopItems[d.GetGoodID()] = d
	}
}

func GetGuildInventoryCfg(lootId string) *ProtobufGen.GUILDBAG {
	return gdGuildInventoryItems[lootId]
}

func GetGuildLostInventoryCfg(lootId string) *ProtobufGen.LOSTGOODSHOP {
	return gdLostGoodShopItems[lootId]
}
