package crossservice

import (
	"vcs.taiyouxi.net/jws/crossservice/module"
)

//CallBackHandle ..
type CallBackHandle func(module.Param)

var listCallbackHandle = map[string]map[string]CallBackHandle{}

//RegCallbackHandle ..
func RegCallbackHandle(mo, me string, f CallBackHandle) {
	if nil == listCallbackHandle[mo] {
		listCallbackHandle[mo] = make(map[string]CallBackHandle)
	}
	listCallbackHandle[mo][me] = f
}

func getCallBackHandle(mo, me string) CallBackHandle {
	if nil == listCallbackHandle[mo] {
		return nil
	}
	return listCallbackHandle[mo][me]
}
