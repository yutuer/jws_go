package error_code

import "vcs.taiyouxi.net/jws/gamex/models/gamedata/err_code"

const (
	_ = 10 * iota
	ErrCommonStart
	_
	ErrActBossStart
)

const (
	_ = ErrCommonStart + iota
	ErrCommonParam
)

const (
	_ = ErrActBossStart + iota
	ErrActBossNoFound
	ErrActBossCurrBossHasLocked
	ErrActBossCurrBossNoLocked
	ErrActBossCurrBossUnLocked // 大BOSS未解锁
)

func GetWarnCodeFromErr(err int) uint32 {
	switch err {
	case ErrCommonParam:
		return errCode.GuildBossParamErr
	case ErrActBossNoFound:
		return errCode.GuildBossNoFound
	case ErrActBossCurrBossHasLocked:
		return errCode.GuildBossCurrBossHasLocker
	case ErrActBossCurrBossNoLocked:
		return errCode.GuildBossCurrBossNoLocker
	case ErrActBossCurrBossUnLocked:
		return errCode.GuildBossBigUnlocked
	}

	return 0
}
