package bag

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

func FromBagItem2Client(b *helper.BagItemToClient, bb *BagItem) {
	b.ID = bb.ID
	b.TableID = bb.TableID
	b.ItemID = bb.ItemID
	b.Count = bb.Count
	b.Data = bb.GetDataStr()
}

//BagItem 必须是能够拷贝的结构体, 请不要轻易使用指针
type BagItem struct {
	// 注意这三个ID的区别
	// ID是uint32类型，可以指定玩家背包中的某一个物品
	// TableID是这个物品原本的类型 对非固定物品类似 MAT_EEVO_45 对于固定物品 类似LG_WP_1_23
	// TableID发向客户端
	// ItemID和TableID一样，但是对于固定物品 id被替换成一个hash串 类似“9a5bbf4c-2082-4513-b12c-c351ce30b79c”
	ID      uint32 `codec:"id"`
	TableID string `codec:"tableid"`
	ItemID  string `codec:"itemid"`

	//物品当前数量
	Count int64 `codec:"count"`

	//StackSize uint16 `codec:"stacksize"` //堆叠上限
	data     string
	ItemData gamedata.BagItemData `codec:"datas"`
}

func (b *BagItem) ToClient() helper.BagItemToClient {
	b2c := helper.BagItemToClient{}
	FromBagItem2Client(&b2c, b)
	return b2c
}

func (b *BagItem) GetDataStr() string {
	if b.data == "" && !b.ItemData.IsNil() {
		b.data, _ = b.ItemData.ToData()
	}
	return b.data
}

// 得到Data具体信息的指针，供修改Data用，修改完成之后用SetItemData()将指针传回
func (b *BagItem) GetItemData() *gamedata.BagItemData {
	return &b.ItemData
}

// 根据GetItemData()获取的指针，修改Data信息
func (b *BagItem) SetItemData(n *gamedata.BagItemData) error {
	b.ItemData = *n
	return nil
}

func (b *BagItem) IsFixedID() bool {
	return IsFixedID(b.ID)
}

func (b *BagItem) CanDelete() bool {
	//if b.BagInfo {
	//return false
	//}
	if b.ID == 0 {
		return true
	}
	if b.Count == 0 {
		return true
	}
	return false
}
