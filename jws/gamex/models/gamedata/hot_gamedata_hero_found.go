package gamedata

const max_hero_found_iap = 16

const Android_Enjoy_Korea_GP_HeroFound_Iap  = "123"
const Android_Enjoy_Korea_OneStore_HeroFoud_Iap  = "223"
// 投资英雄
type HeroFoundData struct {
	ActivityIap map[int]bool // 投资英雄活动 相关联的充值表索引  当set使用
}

func isHeroFoundActivity(activityType int) bool {
	return ActHeroFund_Begin <= activityType && activityType <= ActHeroFund_End
}
