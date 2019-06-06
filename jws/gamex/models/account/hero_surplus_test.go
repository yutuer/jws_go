package account

import (
	"testing"
	"time"

	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
	"vcs.taiyouxi.net/jws/gamex/models/helper"

	"github.com/stretchr/testify/assert"
)

func TestHeroSurplusInfo_AddUsedEn(t *testing.T) {
	hsi := &HeroSurplusInfo{}
	now := time.Now().Unix()
	openEn := gamedata.GetHeroCommonConfig().GetSPStoreAppear()
	openMax := gamedata.GetHeroCommonConfig().GetSPStoreTimes()
	openDur := gamedata.GetHeroCommonConfig().GetSPStoreDuration() * 60

	// 未达到
	hsi.AddUsedEn(int(openEn-100), now)
	assert.False(t, hsi.DailyFirstOpen2Client)

	// 达到并触发
	hsi.AddUsedEn(100, now)
	assert.True(t, hsi.DailyFirstOpen2Client)
	assert.Equal(t, 0, hsi.DailyUsedEN)
	assert.Equal(t, 1, hsi.OpenCount)
	assert.Equal(t, int64(openDur), hsi.EndTime-now)

	// 超过最大次数
	hsi.DailyFirstOpen2Client = false
	hsi.OpenCount = int(openMax)
	hsi.AddUsedEn(int(openEn), now)
	assert.False(t, hsi.DailyFirstOpen2Client)
	assert.EqualValues(t, openMax, hsi.OpenCount)
}

func TestHeroSurplusInfo_AddDailyDrawCount(t *testing.T) {
	hsi := &HeroSurplusInfo{}
	hsi.AddDailyDrawCount(0, 10)
	hsi.AddDailyDrawCount(1, 15)
	hsi.AddDailyDrawCount(2, 20)
	hsi.AddDailyDrawCount(helper.Hero_Surplus_Count, 25)

	assert.Equal(t, 3, len(hsi.DailyDrawCount))
}

func TestHeroSurplusInfo_TryDailyReset(t *testing.T) {
	now := time.Now().Unix()
	hsi := &HeroSurplusInfo{
		EndTime:        now + 3600,
		DailyUsedEN:    300,
		DailyResetTime: now,
	}
	hsi.AddDailyDrawCount(1, 15)

	// 不重置
	hsi.TryDailyReset(now)
	assert.True(t, hsi.DailyUsedEN == 300)
	assert.Contains(t, hsi.DailyDrawCount, 15)

	// 重置
	time5am := time.Date(2022, 7, 20, 5, 1, 0, 0, time.Local).Unix()
	hsi.TryDailyReset(time5am)
	assert.True(t, hsi.DailyUsedEN == 0)
	assert.NotContains(t, hsi.DailyDrawCount, 15)
}
