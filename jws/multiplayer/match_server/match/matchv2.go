package match

import (
	"fmt"
	"sync"
)

type GVEMatchV2 struct {
	mu         sync.RWMutex
	allMatches map[string]*GVEMatch
}

var _GVEMatches GVEMatchV2

func init() {
	_GVEMatches.allMatches = make(map[string]*GVEMatch, 10)
}

func GVEMatchV2_GetOrCreate(matchTokens, matchBoss string) *GVEMatch {
	_GVEMatches.mu.RLock()
	boss := fmt.Sprintf("%s.%s", matchBoss, matchTokens)
	m, ok := _GVEMatches.allMatches[boss]
	_GVEMatches.mu.RUnlock()

	if ok {
		return m
	}
	//Create
	_GVEMatches.mu.Lock()
	defer _GVEMatches.mu.Unlock()

	nm := &GVEMatch{}
	nm.Start(matchTokens)
	_GVEMatches.allMatches[boss] = nm
	return nm
}

func GVEMatchV2_Stop() {
	_GVEMatches.mu.Lock()
	defer _GVEMatches.mu.Unlock()

	for _, v := range _GVEMatches.allMatches {
		v.Stop()
	}
}
