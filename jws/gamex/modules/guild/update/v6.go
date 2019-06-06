package update

import (
	"fmt"
	"vcs.taiyouxi.net/jws/gamex/modules/guild"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

func init() {
	guild.VerAdd(5, V5ToV6)
}

func V5ToV6(FromVersion int64, gi *guild.GuildInfo) error {
	if FromVersion != gi.Ver {
		return fmt.Errorf("v5Tov6 err FromVersion %d Profile.Ver %d",
			FromVersion, gi.Ver)
	}

	logs.Debug("v5Tov6 %s %d", gi.Base.GuildUUID, gi.Ver)
	return nil
}
