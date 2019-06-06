package csrob

//RankTeam ..
type RankTeam struct {
	Heros []HeroInfoForRank `json:"heros,omitempty"`
	Acid  string            `json:"acid,omitempty"`
	Name  string            `json:"-,omitempty"`
}

//HeroInfoForRank ..
type HeroInfoForRank struct {
	Idx       int   `json:"idx"`        // id
	StarLevel int   `json:"star_level"` // 星级
	Gs        int64 `json:"gs"`         // 战力
}
