package data_update

import (
	"errors"

	"vcs.taiyouxi.net/jws/gamex/models/account"
	"vcs.taiyouxi.net/jws/gamex/models/account/update"
	"vcs.taiyouxi.net/jws/gamex/models/helper"
)

// JustLoad表示仅仅是在加载离线玩家时的调用, 不会写回存档
type dataUpdateFunc func(FromVersion int64, justLoad bool, acc *account.Account) error
type dataUpdateFuncSet [][]dataUpdateFunc

func (d *dataUpdateFuncSet) Add(fromVer int64, f dataUpdateFunc) {
	dataUpdateFuncs[fromVer] = append(dataUpdateFuncs[fromVer], f)
}

var dataUpdateFuncs dataUpdateFuncSet

func toVersion(acc *account.Account, v int64) {
	acc.AntiCheat.Ver = v
	acc.BagProfile.Ver = v
	acc.GeneralProfile.Ver = v
	acc.GuildProfile.Ver = v
	acc.Profile.Ver = v
	acc.SimpleInfoProfile.Ver = v
	acc.StoreProfile.Ver = v
	acc.Tmp.Ver = v
}

// JustLoad表示仅仅是在加载离线玩家时的调用, 不会写回存档
func Update(fromVersion int64, justLoad bool, acc *account.Account) error {
	currVersion := fromVersion
	if fromVersion < 0 {
		currVersion = 0
	}

	if currVersion == 0 {
		toVersion(acc, helper.CurrDBVersion)
		return nil
	}

	for currVersion != helper.CurrDBVersion {
		if currVersion < 0 || currVersion >= helper.CurrDBVersion {
			return errors.New("account currVersion Err")
		}
		updateFuncs := dataUpdateFuncs[currVersion]
		if len(updateFuncs) == 0 {
			// 没什么更新需求
			currVersion = currVersion + 1
			continue
		}
		for _, upFunc := range updateFuncs {
			err := upFunc(currVersion, justLoad, acc)
			if err != nil {
				return err
			}
		}
		// 所有存档都是总体升级
		newVersion := acc.Profile.Ver
		if currVersion <= newVersion {
			currVersion += 1 // 正常的话都是升到下一级
		} else {
			currVersion = newVersion // 有时要跨等级升
		}
		acc.Profile.Ver = currVersion
	}
	toVersion(acc, helper.CurrDBVersion)
	return nil
}

func init() {
	// 实现更新在这里
	maxVer := helper.CurrDBVersion + 1 // 不可能有存档的Ver比现在最大的还要大 +1是为了预备下一次升级做
	dataUpdateFuncs = make([][]dataUpdateFunc, maxVer, maxVer)
	for i := 0; i < len(dataUpdateFuncs); i++ {
		dataUpdateFuncs[i] = make([]dataUpdateFunc, 0, 8)
	}

	dataUpdateFuncs.Add(0, func(FromVersion int64, justLoad bool, acc *account.Account) error {
		// 新账号直升到当前版本
		toVersion(acc, helper.CurrDBVersion)
		return nil
	})

	//
	// 严重警告 升级过程中不能向玩家发送邮件等会影响外部的操作
	// 由于LoadAccount也会调用数据更新逻辑但是LoadAccount
	// 的账号不会写回到玩家存档
	//

	dataUpdateFuncs.Add(5, update.V5toV6)
	dataUpdateFuncs.Add(6, update.V6toV7)
	dataUpdateFuncs.Add(7, update.V7toV8)
}
