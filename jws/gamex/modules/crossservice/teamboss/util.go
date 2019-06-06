package teamboss

import (
	"vcs.taiyouxi.net/jws/gamex/models/codec"
	"vcs.taiyouxi.net/jws/helper"
)

type RoomPushInfo struct {
	RoomID         string   `codec:"room_id"`
	LeadID         string   `codec:"lead_id"`
	PlayerPushInfo [][]byte `codec:"player_info"`
	IsPublic       bool     `codec:"is_public"`
	IsAdvance      bool     `codec:"is_advance"`
	SceneID        string   `codec:"scene_id"`
	BossID         string   `codec:"boss_id"`
	TeamTypID      int      `codec:"team_typ_id"`
}

type PlayerPushInfo struct {
	AcID            string   `codec:"acid"`
	Sid             uint     `codec:"sid"`
	GS              int      `codec:"gs"`
	Avatar          int      `codec:"avatar"`
	BattleAvatar    int      `codec:"battle_avatar"`
	Name            string   `codec:"name"`
	Wing            int      `codec:"wing"`
	Fashion         []string `codec:"fashion"`
	Status          int      `codec:"status"`
	Level           int      `codec:"level"`
	VIP             int      `codec:"vip"`
	StarLevel       int      `codec:"star_level"`
	MagicPet        int      `codec:"magic_pet"`
	ExclusiveWeapon string   `codec:"ex_wp"`
	Position        int      `codec:"position"`
	CompressGS      int      `codec:"compress_gs"`
}

func GenRoomPushInfo(info helper.RoomDetailInfo) []byte {
	psi := make([][]byte, len(info.SimpleInfo))
	for i, item := range info.SimpleInfo {
		position := -1
		for j, jtem := range info.PositionAcID {
			if jtem == item.AcID {
				position = j
				break
			}
		}
		psi[i] = codec.Encode(PlayerPushInfo{
			AcID:            item.AcID,
			Sid:             item.Sid,
			GS:              item.GS,
			Avatar:          item.Avatar,
			BattleAvatar:    item.BattleAvatar,
			Name:            item.Name,
			Wing:            item.Wing,
			Fashion:         item.Fashion,
			Status:          item.Status,
			Level:           item.Level,
			VIP:             item.VIP,
			StarLevel:       item.StarLevel,
			MagicPet:        item.MagicPet,
			ExclusiveWeapon: item.ExclusiveWeapon,
			Position:        position,
			CompressGS:      item.CompressGS,
		})
	}
	return codec.Encode(RoomPushInfo{
		RoomID:         info.RoomID,
		LeadID:         info.LeadID,
		PlayerPushInfo: psi,
		IsPublic:       info.RoomStatus == helper.TBRoomStateOpen,
		IsAdvance:      info.BoxStatus == helper.TBBoxStateAdvance,
		SceneID:        info.SceneID,
		BossID:         info.BossID,
		TeamTypID:      int(info.TeamTypID),
	})
}
