package watcher

import (
	"vcs.taiyouxi.net/jws/crossservice/util/discover"
)

//Filter ..
type Filter interface {
	Check(discover.Service) bool
}

//FilterService ..
type FilterService struct {
	Service string
}

//Check ..
func (f FilterService) Check(s discover.Service) bool {
	return f.Service == s.Service
}

//FilterVersion ..
type FilterVersion struct {
	Version string
}

//Check ..
func (f FilterVersion) Check(s discover.Service) bool {
	return f.Version == s.Version
}

//FilterVersionMin ..
type FilterVersionMin struct {
	VersionMin string
}

//Check ..
func (f FilterVersionMin) Check(s discover.Service) bool {
	fMajor, fMinor, fFix, fDes := discover.Service{Version: f.VersionMin}.ParseVersion()
	sMajor, sMinor, sFix, sDes := s.ParseVersion()

	if sMajor > fMajor {
		return true
	} else if sMajor < fMajor {
		return false
	}

	if sMinor > fMinor {
		return true
	} else if sMinor < fMinor {
		return false
	}

	if sFix > fFix {
		return true
	} else if sFix < fFix {
		return false
	}

	return sDes >= fDes
}
