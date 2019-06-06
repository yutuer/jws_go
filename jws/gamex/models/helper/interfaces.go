package helper

const (
	MemSyncReceiverIDGateEnemyNull = iota
	MemSyncReceiverIDGateEnemyAct
)

// 公会成员信息同步接受者, 注意由于公会是单gorountine驱动的这些接口实现必须不能有阻塞
type IGuildMemSyncReceiver interface {
	GetMemSyncReceiverID() int
	OnGuildChange(accounts []AccountSimpleInfo)
}

//  Interface
type IAccount interface {
	GetProfileNowTime() int64
	GetCorpLv() uint32
}
