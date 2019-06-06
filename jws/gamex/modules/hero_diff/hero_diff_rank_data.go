package hero_diff

type HeroDiffRankData struct {
	AcID       string `json:"acid" codec:"acid"`
	FreqAvatar []int  `json:"freq_avatar" codec:"freq_avatar"`
}
