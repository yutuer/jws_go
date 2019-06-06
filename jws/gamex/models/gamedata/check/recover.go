package check

type AssertRecoverRetialEmpty func(recoverId uint32) bool

func RecoverCheck(dataName string, recoverIds []uint32, assertFunc AssertRecoverRetialEmpty) {
	if assertFunc == nil {
		panicf("RecoverCheck Err No Func By %s", dataName)
	}

	for _, id := range recoverIds {
		if !assertFunc(id) {
			panicf("RecoverCheck Err %d By %s", id, dataName)
		}
	}
}
