package logics

func (p *Account) addHandle() {
	addRankHandle(p)
	addChgAvatarHandle(p)
	addCorpLvUpHandle(p)
	addHeroLvUpHandle(p)
	addCorpExpAddHandle(p)
	addEnergyUsedHandle(p)
	addScChgHandle(p)
	addHcChgHandle(p)
	addTrialHandle(p)
	addTitleHandle(p)
	addVipHandle(p)
}
