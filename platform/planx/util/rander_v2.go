package util

import "math/rand"

//RandUintSetV2 ..
type RandUintSetV2 struct {
	set      []randIntElem
	powerMax uint32
}

type randIntElem struct {
	res   uint32
	power uint32
}

//NewRandUintSetV2 ..
func NewRandUintSetV2() *RandUintSetV2 {
	s := &RandUintSetV2{
		set:      []randIntElem{},
		powerMax: 0,
	}
	return s
}

//Add ..
func (s *RandUintSetV2) Add(res uint32, power uint32) {
	s.powerMax += power

	s.set = append(
		s.set,
		randIntElem{
			res:   res,
			power: power,
		},
	)
}

//Rand ..
func (s *RandUintSetV2) Rand(rd *rand.Rand) uint32 {
	rn := rd.Uint32() % s.powerMax
	for _, elem := range s.set {
		if rn < elem.power {
			return elem.res
		}
		rn -= elem.power
	}
	return 0
}
