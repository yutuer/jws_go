package simple_update

type simpleUpdateFunc func(FromVersion int64, from []byte, to []byte) error

// 结构升级(小) TODO By Fanyang
