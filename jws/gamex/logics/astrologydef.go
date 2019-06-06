package logics

import (
	"vcs.taiyouxi.net/jws/gamex/models/astrology"
)

//AstrologySoul ..
type AstrologySoul struct {
	ID    string `codec:"id"`
	Count uint32 `codec:"c"`
}

//AstrologySoulBag ..
type AstrologySoulBag struct {
	Souls [][]byte `codec:"ss"` //[]AstrologySoul
}

//AstrologyHeroHole ..
type AstrologyHeroHole struct {
	HoleID uint32 `codec:"id"`
	Rare   uint32 `codec:"r"`
	Level  uint32 `codec:"l"`
}

//AstrologyHero ..
type AstrologyHero struct {
	HeroID uint32   `codec:"id"`
	Holes  [][]byte `codec:"hs"` //[][]AstrologyHeroHole
}

//AstrologyAugur ..
type AstrologyAugur struct {
	AugurLevel uint32 `codec:"l"`
}

func buildNetAstrologySoul(src *astrology.Soul) *AstrologySoul {
	ret := &AstrologySoul{}

	ret.ID = src.SoulID
	ret.Count = src.Count

	return ret
}

func buildNetAstrologySoulBag(src *astrology.SoulBag) *AstrologySoulBag {
	ret := &AstrologySoulBag{}

	ret.Souls = make([][]byte, 0, len(src.Souls))
	for _, soul := range src.Souls {
		ret.Souls = append(ret.Souls, encode(buildNetAstrologySoul(soul)))
	}

	return ret
}

func buildNetAstrologyHeroHole(src *astrology.Hole) *AstrologyHeroHole {
	ret := &AstrologyHeroHole{}

	ret.HoleID = src.HoleID
	ret.Rare = src.Rare
	ret.Level = src.Upgrade

	return ret
}

func buildNetAstrologyHero(src *astrology.Hero) *AstrologyHero {
	ret := &AstrologyHero{}

	ret.HeroID = src.HeroID

	ret.Holes = make([][]byte, 0, len(src.Holes))
	for _, hole := range src.Holes {
		ret.Holes = append(ret.Holes, encode(buildNetAstrologyHeroHole(hole)))
	}

	return ret
}

func buildNetAstrologyAugur(src *astrology.Factory) *AstrologyAugur {
	ret := &AstrologyAugur{}

	ret.AugurLevel = src.CurrLevel

	return ret
}
