package account

func (a *Account) GetProfileNowTime() int64 {
	return a.Profile.GetProfileNowTime()
}

func (a *Account) GetCorpLv() uint32 {
	return a.Profile.GetCorp().GetLvlInfo()
}
