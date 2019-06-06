package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/modules/mail_sender"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/timail"
)

func (p *Account) AutoExchangeShopProp(activityID uint32) {
	logs.Debug("auto exchange shop prop for activityID: %d", activityID)
	exchangeData := gamedata.GetHotDatas().HotExchangeShopData.GetCanAutoExchangePropData(activityID)
	reward := make(map[string]uint32, 0)
	for _, item := range exchangeData {
		costData := &gamedata.CostData{}
		for _, cost := range item.GetNeed_Table() {
			costData.AddItem(cost.GetItemID(), cost.GetItemNum())
		}
		for CostBySync(p, costData, nil, "auto exchange shop prop") {
			for _, loot := range item.GetLoot_Table() {
				for i := 0; i < int(loot.GetLootTime()); i++ {
					randReward := gamedata.GetRandWeightProp(loot.GetLootGroupID())
					if v, ok := reward[randReward.GetItemID()]; ok {
						reward[randReward.GetItemID()] = v + randReward.GetItemCount()
					} else {
						reward[randReward.GetItemID()] = randReward.GetItemCount()
					}
				}
			}
		}
	}
	if len(reward) > 0 {
		hotData := gamedata.GetHotDatas().Activity
		act := hotData.GetActivitySimpleInfoById(activityID)
		if act.Cfg.GetTabIDS() != "" {
			mail_sender.BatchSendMail2Account(p.AccountID.String(), timail.Mail_Send_By_Market_Activity,
				mail_sender.IDS_MAIL_AUTO_EXCHANGE_SHOP_PROP_CUSTOM_TITLE, []string{act.Cfg.GetTabIDS()}, reward,
				"AutoExchangeShopProp", true)

		} else {
			mail_sender.BatchSendMail2Account(p.AccountID.String(), timail.Mail_Send_By_Market_Activity,
				mail_sender.IDS_MAIL_AUTO_EXCHANGE_SHOP_PROP_TITLE, []string{}, reward,
				"AutoExchangeShopProp", true)
		}

	}
	p.Profile.GetMarketActivitys().ClearExchangePropInfo(p.AccountID.String(),
		p.Profile.GetProfileNowTime(), uint32(activityID))
}
