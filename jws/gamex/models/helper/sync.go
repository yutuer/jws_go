package helper

type ISyncObj interface {
	IsNeedSync() bool
	SetHadSync()
	SetNeedSync()
}

type SyncObj struct {
	isNeedSync bool
	I          int // 无用，只因为struct没有可导出的项的话，encode会报错，会导致一直dirty
}

func (p *SyncObj) IsNeedSync() bool {
	return p.isNeedSync
}

func (p *SyncObj) SetHadSync() {
	p.isNeedSync = false
}

func (p *SyncObj) SetNeedSync() {
	p.isNeedSync = true
}
