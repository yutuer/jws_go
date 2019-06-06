package guild

import (
	"errors"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

//dataUpdateFunc的具体实现中, 不能修改Version

type dataUpdateFunc func(FromVersion int64, gi *GuildInfo) error
type dataUpdateFuncSet [][]dataUpdateFunc

var dataUpdateFuncs dataUpdateFuncSet

func VerAdd(fromVer int64, f dataUpdateFunc) {
	dataUpdateFuncs[fromVer] = append(dataUpdateFuncs[fromVer], f)
}

func toVersion(gi *GuildInfo, v int64) {
	gi.Ver = v
}

func VerUpdate(gi *GuildInfo) error {
	currVersion := gi.Ver

	if currVersion < 0 {
		currVersion = 0
	}
	if currVersion == 0 {
		toVersion(gi, helper.CurrDBVersion)
		return nil
	}

	for currVersion != helper.CurrDBVersion {
		if currVersion < 0 || currVersion >= helper.CurrDBVersion {
			return errors.New("account currVersion Err")
		}

		updateFuncs := dataUpdateFuncs[currVersion]
		if len(updateFuncs) != 0 {
			for _, upFunc := range updateFuncs {
				err := upFunc(currVersion, gi)
				if err != nil {
					return err
				}
			}
		}

		currVersion += 1
		gi.Ver = currVersion
	}

	toVersion(gi, helper.CurrDBVersion)
	return nil
}

func init() {
	// 实现更新在这里
	maxVer := helper.CurrDBVersion + 1 // 不可能有存档的Ver比现在最大的还要大 +1是为了预备下一次升级做
	dataUpdateFuncs = make([][]dataUpdateFunc, maxVer, maxVer)
	for i := 0; i < len(dataUpdateFuncs); i++ {
		dataUpdateFuncs[i] = make([]dataUpdateFunc, 0, 8)
	}

}
