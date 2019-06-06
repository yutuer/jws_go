package account

import (
	"math"
	"testing"

	"vcs.taiyouxi.net/jws/gamex/models/account/events"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"

	"github.com/stretchr/testify/assert"
)

func DebugUnlockCorpAvatars(corp *Corp, amount int) {
	if amount > AVATAR_NUM_CURR {
		amount = AVATAR_NUM_CURR
	}

	for i := 0; i < amount; i++ {
		corp.UnlockAvatars[i] = 1
	}
}

func TestCorp_OnAccountInit(t *testing.T) {
	corp := Debuger.GetNewCorp()
	corp.OnAccountInit()

	assert.NotNil(t, corp.UnlockAvatars)
}

func TestCorp_OnAfterLogin(t *testing.T) {
	corp := Debuger.GetNewCorp()
	unlockAmount := 3

	DebugUnlockCorpAvatars(corp, unlockAmount)
	corp.OnAfterLogin()

	assert.NotNil(t, corp.UnlockAvatars)
	assert.Equal(t, len(corp.unlockAvatar2Client), unlockAmount)
}

func TestCorp_IsAvatarHasUnlock(t *testing.T) {
	corp := Debuger.GetNewCorp()
	DebugUnlockCorpAvatars(corp, 3)

	inputs := []int{-255, 0, 1, 3, AVATAR_NUM_CURR + 1, AVATAR_NUM_MAX + 1}
	expects := []bool{false, true, true, false, false, false}

	for i := range inputs {
		assert.Equal(t, corp.IsAvatarHasUnlock(inputs[i]), expects[i])
	}
}

func TestCorp_HasAvatarHasUnlok(t *testing.T) {
	corp := Debuger.GetNewCorp()
	assert.Equal(t, corp.HasAvatarHasUnlok(), 0)

	DebugUnlockCorpAvatars(corp, 5)
	assert.Equal(t, corp.HasAvatarHasUnlok(), 5)
}

func TestCorp_UnlockAvatar(t *testing.T) {
	account := Debuger.GetNewAccount()
	corp := account.Profile.CorpInf

	// 解锁
	corp.UnlockAvatar(account, 0)
	assert.Equal(t, corp.UnlockAvatars[0], byte(1))

	// 新英雄解锁
	corp.UnlockAvatar(account, AVATAR_NUM_CURR-1)
	assert.Equal(t, corp.UnlockAvatars[AVATAR_NUM_CURR-1], byte(1))
}

func TestCorp_GetUnlockedAvatar(t *testing.T) {
	account := Debuger.GetNewAccount()
	corp := account.Profile.CorpInf

	unlockAvatarIndex := []int{0, 2, 5, 11}

	for _, index := range unlockAvatarIndex {
		corp.UnlockAvatar(account, index)
	}

	assert.Equal(t, unlockAvatarIndex, corp.GetUnlockedAvatar())
}

func TestCorp_IsNeedSyncUnlocked(t *testing.T) {
	corp := Debuger.GetNewCorp()
	corp.isNeedSync = false

	assert.False(t, corp.isNeedSync)
}

func TestCorp_SetNoNeedSync(t *testing.T) {
	corp := Debuger.GetNewCorp()
	corp.isNeedSync = true
	corp.SetNoNeedSync()

	assert.False(t, corp.isNeedSync)
}

func TestCorp_SetHandler(t *testing.T) {
	corp := Debuger.GetNewCorp()
	handler := events.NewHandler()
	corp.SetHandler(handler)

	assert.NotNil(t, corp.handler)
}

func TestCorp_AddExp(t *testing.T) {
	corp := Debuger.GetNewProfile().CorpInf
	exp := uint32(1)

	corp.AddExp("", exp, "")

	assert.Equal(t, corp.Xp, exp)

	// 溢出
	corp.AddExp("", math.MaxUint32, "")
	corp.AddExp("", math.MaxUint32, "")

	// assert.Equal(t, corp.Xp, math.MaxUint32)   //居然Not Equal，我也是醉了
	assert.True(t, corp.Xp == math.MaxUint32)
}

func TestCorp_GetXpInfo(t *testing.T) {
	corp := Debuger.GetNewProfile().CorpInf
	corp.AddExp("", 5, "")

	level, exp := corp.GetXpInfo()

	assert.Equal(t, level, uint32(1))
	assert.Equal(t, exp, uint32(5))
}

func TestCorp_GetXpInfoNoUpdate(t *testing.T) {
	corp := Debuger.GetNewProfile().CorpInf
	corp.AddExp("", 5, "")

	level, exp := corp.GetXpInfo()

	assert.Equal(t, level, uint32(1))
	assert.Equal(t, exp, uint32(5))
}

func TestCorp_GetLvlInfo(t *testing.T) {
	corp := Debuger.GetNewProfile().CorpInf
	lLimit := gamedata.GetCommonCfg().GetCorpLevelUpperLimit()

	corp.AddExp("", 10, "")
	assert.Equal(t, corp.GetLvlInfo(), uint32(1))

	corp.AddExp("", math.MaxUint32, "")
	assert.Equal(t, corp.GetLvlInfo(), lLimit)
}

func TestCorp_OnLevelUp(t *testing.T) {
	corp := Debuger.GetNewProfile().CorpInf
	corp.OnLevelUp("", 1, "")
}

func TestCorpLevelUp(t *testing.T) {
	corp := Debuger.GetNewProfile().CorpInf

	lLimit := gamedata.GetCommonCfg().GetCorpLevelUpperLimit()
	lXPs := make([]uint32, lLimit+1)

	for i := uint32(0); i < lLimit; i++ {
		lXPs[i] = gamedata.GetCorpLvConfig(i).CorpXpNeed
	}

	xps := []uint32{
		lXPs[1] - 1, // 少于1级
		1,           // 正好到下一级
		lXPs[2],     // 完整升1级
		15,          // 不到1级
		lXPs[3] + lXPs[4] + lXPs[5] + lXPs[6], // 连续升4级
		math.MaxUint32,                        // 满级
		math.MaxUint32,                        // 溢出
	}

	expects := []uint32{
		1, 2, 3, 3, 7, lLimit, lLimit, // 对应的等级
	}

	for i, xp := range xps {
		corp.AddExp("", xp, "")
		assert.Equal(t, corp.Level, expects[i])
	}
}
