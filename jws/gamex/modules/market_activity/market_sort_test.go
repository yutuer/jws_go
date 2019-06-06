package market_activity

import (
	"math/rand"
	"testing"
	"time"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
)

func TestSort(t *testing.T) {
	a2s := make(map[string]float64)
	for i := 0; i < 500000; i++ {
		a2s[uuid.NewV4().String()] = rand.Float64()
	}

	t1 := time.Now()
	ret := sortFromAcid2score(a2s)
	t.Logf("sort cost %v", time.Now().Sub(t1).String())
	t.Logf("Top 10: %v", ret[:10])
}
