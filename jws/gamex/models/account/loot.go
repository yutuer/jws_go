package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

// 随机掉落计算接口 通过掉落Template随机
func (p *Account) GetGivesByTemplate(template_id string) (gamedata.PriceDatas, error) {
	res, err := gamedata.GetGivesByTemplate(template_id, p.AccountID.String(), p.GetRand())
	return res, err
}

// 随机掉落计算接口 通过掉落ItemGroup随机
func (p *Account) GetGivesByItemGroup(item_group_id string) (gamedata.PriceDatas, error) {
	res, err := gamedata.GetGivesByItemGroup(item_group_id, p.AccountID.String(), p.GetRand())
	return res, err
}
