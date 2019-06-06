package logics

import (
	"testing"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/platform/planx/servers"

	"github.com/stretchr/testify/assert"
)

// 验证线上Cheat修复：
// 修改请求中物品购买数量为特定值，导致HC花费溢出为一个很小的数字
func TestAccount_BuyInShop(t *testing.T) {
	req := servers.Request{}

	// 根据原结构生成request
	data := &struct {
		Req
		ShopId uint32 `codec:"s"`
		GoodId string `codec:"g"`
		Count  int    `codec:"c"`
	}{}

	data.ShopId = 0
	data.GoodId = "JD_1_6"
	data.Count = 95443718

	/*
		math.MaxUint32 % 45 = 30
		因此14是能够使其溢出的余数，验证如下：
		var num int
		num = (math.MaxUint32 + 1 + 14) / 45
		num -= math.MaxUint32 + 1
		fmt.Println("num: ", num)
		cost := uint32(45) * uint32(num)
		fmt.Println("cost: ", v)
		所以  num = 95443718, -4199523578 时 cost = 14
	*/

	req.RawBytes = encode(*data)

	p := new(Account)
	p.Account = account.Debuger.GetNewAccount()

	// 增加正好足够的硬通
	p.Profile.GetHC().AddHC(p.AccountID.String(), 2, 15, time.Now().Unix(), "Test Buy")

	// 验证数量为负数时无法获得宝石
	resp := p.BuyInShop(req)
	assert.Empty(t, p.Account.Profile.GetJadeBag().JadesMap)
	assert.NotEmpty(t, resp.RawBytes)

	// 验证数量为95443718时无法获得宝石
	data.Count = -4199523578
	req.RawBytes = encode(*data)

	resp = p.BuyInShop(req)
	assert.Empty(t, p.Account.Profile.GetJadeBag().JadesMap)
	assert.NotEmpty(t, resp.RawBytes)
}
