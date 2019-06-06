package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/bag"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/platform/planx/servers/db"
	//"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

type PlayerBag struct {
	bag.StackBag
}

func NewPlayerBag(account db.Account) PlayerBag {
	AccountID := account

	mybag := PlayerBag{
		StackBag: *bag.NewStackBag(AccountID),
	}

	return mybag
}

// 向Account的背包发放物品
// TBD 确定data是否还需要由外层决定
func (b *PlayerBag) AddToBag(
	account *Account,
	data gamedata.BagItemData, ItemID string, count uint32) (
	errCode int, item_inner_type int, idx2OldCount map[uint32]int64) {

	aid := account.AccountID.String()

	// 加一个保护,如果是关卡中Buff类道具,就不加了 XXX 可以去掉了
	is_buffer := gamedata.IsItemToBuffWhenAdd(ItemID)
	if is_buffer {
		return helper.RES_AddToBag_NoItemADD, -1, nil
	}

	if ok, jadeCfg := gamedata.IsJade(ItemID); ok {
		return account.Profile.GetJadeBag().AddJadeByTableId(ItemID, int64(count), 0, jadeCfg)
	} else if ok, fashionCfg := gamedata.IsFashion(ItemID); ok {
		return account.Profile.GetFashionBag().AddFashionByTableId(ItemID, fashionCfg,
			account.Profile.GetProfileNowTime())
	} else {
		errCode, idx2OldCount = b.Add(data, ItemID, count, aid, account.GetRand(), account.Profile.GetProfileNowTime())
		return errCode, helper.Item_Inner_Type_Basic, idx2OldCount
	}
}
