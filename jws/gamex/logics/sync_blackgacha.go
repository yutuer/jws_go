package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/protogen"
)

type BlackGachaSetting2Client struct {
	ActivityId      uint32 `codec:"bgs_id"`
	ActivitySubId   uint32 `codec:"bgs_subid"`
	GachaCoin1      string `codec:"bgs_coin"`
	OnePrice        uint32 `codec:"bgs_onep"`
	TenPrice        uint32 `codec:"bgs_tenp"`
	Ticket			string `codec:"bgs_ticket"`
	TAPrice			uint32 `codec:"bgs_tonep"`
	TTenPrice 		uint32 `codec:"bgs_ttenp"`
	ResetFreeTime   string `codec:"bgs_reset_time"`
	DailyLimit      uint32 `codec:"bgs_day_limit"`
	FreeTime        uint32 `codec:"bgs_free_time"`
	ActivitySubName string `codec:"bgs_sub_name"`
	Description     string `codec:"bgs_desc"`
	Icon            string `codec:"bgs_icon"`
	Discount        uint32 `codec:"bgs_dis"`
}

func (s *SyncResp) makeBlackGachaSettings() {
	blackGachaSettings := gamedata.GetHotDatas().HotBlackGachaData.BlackGachaSettings
	s.BlackGachaSettings = make([][]byte, 0)
	for _, gacha := range blackGachaSettings {
		s.BlackGachaSettings = append(s.BlackGachaSettings, encode(convertSettings2Client(gacha)))
	}
}

func convertSettings2Client(data *ProtobufGen.BOXSETTINGS) BlackGachaSetting2Client {
	return BlackGachaSetting2Client{
		ActivityId:      data.GetActivityID(),
		ActivitySubId:   data.GetActivitySubID(),
		GachaCoin1:      data.GetGachaCoin1(),
		OnePrice:        data.GetAPrice1(),
		TenPrice:        data.GetTenPrice1(),
		Ticket:			 data.GetGachaTicket(),
		TAPrice:		 data.GetTAPrice(),
		TTenPrice:		 data.GetTTenPrice(),
		ResetFreeTime:   data.GetResetFreeTime(),
		DailyLimit:      data.GetDailyLimit(),
		FreeTime:        data.GetFreeTime(),
		ActivitySubName: data.GetActivitySubName(),
		Description:     data.GetDescriptionIDS(),
		Icon:            data.GetIcon(),
		Discount:        data.GetDiscount(),
	}
}

type BlackGachaShow2Client struct {
	ActivitySubId uint32   `codec:"bgsw_subid"`
	ShowType      uint32   `codec:"bgsw_st"`
	ShowItem      []string `codec:"bgsw_item"`
}

func (s *SyncResp) makeBlackGachaShow() {
	blackGachaShows := gamedata.GetHotDatas().HotBlackGachaData.BlackGachaShow
	s.BlackGachaShows = make([][]byte, 0)
	for _, data := range blackGachaShows {
		s.BlackGachaShows = append(s.BlackGachaShows, encode(convertShow2Client(data)))
	}
}

func convertShow2Client(data *ProtobufGen.BOXSHOW) BlackGachaShow2Client {
	showItemArray := make([]string, 0)
	for _, item := range data.GetGachaShowItem_Template() {
		showItemArray = append(showItemArray, item.GetItemID())
	}

	return BlackGachaShow2Client{
		ActivitySubId: data.GetActivitySubID(),
		ShowType:      data.GetGachaShowType(),
		ShowItem:      showItemArray,
	}
}

type BlackGachaLowest2Client struct {
	ActivitySubId uint32   `codec:"bgl_subid"`
	ListId        uint32   `codec:"bgl_list"`
	LowestTimes   uint32   `codec:"bgl_times"`
	ItemId        []string `codec:"bgl_item_id"`
	ItemCount     []uint32 `codec:"bgl_item_c"`
}

func (s *SyncResp) makeBlackGachaLowest() {
	blackGachaLowest := gamedata.GetHotDatas().HotBlackGachaData.BlackGachaLowest
	s.BlackGachaLowest = make([][]byte, 0)
	for _, data := range blackGachaLowest {
		s.BlackGachaLowest = append(s.BlackGachaLowest, encode(convertLowest2Client(data)))
	}
}

func convertLowest2Client(data *ProtobufGen.BOXLOWEST) BlackGachaLowest2Client {
	itemIds := make([]string, 0)
	itemCounts := make([]uint32, 0)
	for _, cfg := range data.Fixed_Loot {
		itemIds = append(itemIds, cfg.GetItemID())
		itemCounts = append(itemCounts, cfg.GetItemCount())
	}
	return BlackGachaLowest2Client{
		ActivitySubId: data.GetActivitySubID(),
		ListId:        data.GetListID(),
		LowestTimes:   data.GetLowestTimes(),
		ItemId:        itemIds,
		ItemCount:     itemCounts,
	}
}
