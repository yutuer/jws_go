package events

type Handler interface {
	// 更换出站角色回调
	OnAvatarChg(avatarID int)
	// 战队等级提升回调
	OnCorpLvUp(toLv, toExp uint32, reason string)
	// 主将等级提升回调
	OnHeroLvUp(fromLv, toLv, toExp uint32, reason string)
	// 战队经验增加
	OnCorpExpAdd(oldV, chgV uint32, reason string)
	// VIP等级提升回调
	OnVIPLvUp(toLv uint32)
	// 首次通关副本
	OnFirstPassStage(stageID string)
	// 使用体力
	OnEnergyUsed(energy int64)
	// 增加、消耗sc
	OnScChg(isAdd bool, typ int, oldV, chgV int64, reason string)
	// 增加、消耗hc
	OnHcChg(isAdd bool, typ int, oldV, chgV int64, reason string)
	// 头顶称号变化
	OnTitleOnChg(oldTitle, newTitle string)
	// 充值回掉
}

type Handlers struct {
	handlers []Handler
}

func (h *Handlers) Add(handler Handler) {
	h.handlers = append(h.handlers, handler)
}

func (h *Handlers) OnAvatarChg(avatarID int) {
	for _, handler := range h.handlers {
		handler.OnAvatarChg(avatarID)
	}
}

func (h *Handlers) OnCorpLvUp(toLv, toExp uint32, reason string) {
	for _, handler := range h.handlers {
		handler.OnCorpLvUp(toLv, toExp, reason)
	}
}

func (h *Handlers) OnHeroLvUp(fromLv, toLv, toExp uint32, reason string) {
	for _, handler := range h.handlers {
		handler.OnHeroLvUp(fromLv, toLv, toExp, reason)
	}
}

func (h *Handlers) OnCorpExpAdd(oldV, chgV uint32, reason string) {
	for _, handler := range h.handlers {
		handler.OnCorpExpAdd(oldV, chgV, reason)
	}
}

func (h *Handlers) OnVIPLvUp(toLv uint32) {
	for _, handler := range h.handlers {
		handler.OnVIPLvUp(toLv)
	}
}

func (h *Handlers) OnFirstPassStage(stageID string) {
	for _, handler := range h.handlers {
		handler.OnFirstPassStage(stageID)
	}
}

func (h *Handlers) OnEnergyUsed(energy int64) {
	for _, handler := range h.handlers {
		handler.OnEnergyUsed(energy)
	}
}

func (h *Handlers) OnScChg(isAdd bool, typ int, oldV, chgV int64, reason string) {
	for _, handler := range h.handlers {
		handler.OnScChg(isAdd, typ, oldV, chgV, reason)
	}
}

func (h *Handlers) OnHcChg(isAdd bool, typ int, oldV, chgV int64, reason string) {
	for _, handler := range h.handlers {
		handler.OnHcChg(isAdd, typ, oldV, chgV, reason)
	}
}

func (h *Handlers) OnTitleOnChg(oldTitle, newTitle string) {
	for _, handler := range h.handlers {
		handler.OnTitleOnChg(oldTitle, newTitle)
	}
}

const handlersLen = 8

type onAvatarChgFunc func(avatarID int)
type onLvUpFunc func(toLv, toExp uint32, reason string)
type onHeroLvUpFunc func(fromLv, toLv, toExp uint32, reason string)
type onExpAddFunc func(oldV, chgV uint32, reason string)
type onVipLvFunc func(toLv uint32)
type onFirstPassStageFunc func(stageID string)
type onEnergyUsedFunc func(en int64)
type onScChgFunc func(isAdd bool, typ int, oldV, chgV int64, reason string)
type onHcChgFunc func(isAdd bool, typ int, oldV, chgV int64, reason string)
type onTitleOnChgFunc func(oldTitle, newTitle string)

type handlerByFunc struct {
	onAvatarChg      []onAvatarChgFunc
	onCorpLvUp       []onLvUpFunc
	onHeroLvUp       []onHeroLvUpFunc
	onCorpExpAdd     []onExpAddFunc
	onVIPLvUp        []onVipLvFunc
	onFirstPassStage []onFirstPassStageFunc
	onEnergyUsed     []onEnergyUsedFunc
	onScChg          []onScChgFunc
	onHcChg          []onHcChgFunc
	onTitleChg       []onTitleOnChgFunc
}

func NewHandler() *handlerByFunc {
	return &handlerByFunc{
		onAvatarChg:      make([]onAvatarChgFunc, 0, handlersLen),
		onCorpLvUp:       make([]onLvUpFunc, 0, handlersLen),
		onHeroLvUp:       make([]onHeroLvUpFunc, 0, handlersLen),
		onCorpExpAdd:     make([]onExpAddFunc, 0, handlersLen),
		onVIPLvUp:        make([]onVipLvFunc, 0, handlersLen),
		onFirstPassStage: make([]onFirstPassStageFunc, 0, handlersLen),
		onEnergyUsed:     make([]onEnergyUsedFunc, 0, handlersLen),
		onScChg:          make([]onScChgFunc, 0, handlersLen),
		onHcChg:          make([]onHcChgFunc, 0, handlersLen),
		onTitleChg:       make([]onTitleOnChgFunc, 0, handlersLen),
	}
}

func (h *handlerByFunc) OnAvatarChg(avatarID int) {
	for _, handler := range h.onAvatarChg {
		handler(avatarID)
	}
}

func (h *handlerByFunc) WithOnAvatarChg(nh onAvatarChgFunc) *handlerByFunc {
	h.onAvatarChg = append(h.onAvatarChg, nh)
	return h
}

func (h *handlerByFunc) OnCorpLvUp(toLv, toExp uint32, reason string) {
	for _, handler := range h.onCorpLvUp {
		handler(toLv, toExp, reason)
	}
}

func (h *handlerByFunc) WithOnCorpLvUp(nh onLvUpFunc) *handlerByFunc {
	h.onCorpLvUp = append(h.onCorpLvUp, nh)
	return h
}

func (h *handlerByFunc) OnHeroLvUp(fromLv, toLv, toExp uint32, reason string) {
	for _, handler := range h.onHeroLvUp {
		handler(fromLv, toLv, toExp, reason)
	}
}

func (h *handlerByFunc) WithOnHeroLvUp(nh onHeroLvUpFunc) *handlerByFunc {
	h.onHeroLvUp = append(h.onHeroLvUp, nh)
	return h
}

func (h *handlerByFunc) OnCorpExpAdd(oldV, chgV uint32, reason string) {
	for _, handler := range h.onCorpExpAdd {
		handler(oldV, chgV, reason)
	}
}

func (h *handlerByFunc) WithOnCorpExpAdd(nh onExpAddFunc) *handlerByFunc {
	h.onCorpExpAdd = append(h.onCorpExpAdd, nh)
	return h
}

func (h *handlerByFunc) OnVIPLvUp(toLv uint32) {
	for _, handler := range h.onVIPLvUp {
		handler(toLv)
	}
}

func (h *handlerByFunc) WithOnVIPLvUp(nh onVipLvFunc) *handlerByFunc {
	h.onVIPLvUp = append(h.onVIPLvUp, nh)
	return h
}

func (h *handlerByFunc) OnFirstPassStage(stage string) {
	for _, handler := range h.onFirstPassStage {
		handler(stage)
	}
}

func (h *handlerByFunc) WithFirstPassStage(nh onFirstPassStageFunc) *handlerByFunc {
	h.onFirstPassStage = append(h.onFirstPassStage, nh)
	return h
}

func (h *handlerByFunc) OnEnergyUsed(en int64) {
	for _, handler := range h.onEnergyUsed {
		handler(en)
	}
}

func (h *handlerByFunc) WithEnergyUsed(nh onEnergyUsedFunc) *handlerByFunc {
	h.onEnergyUsed = append(h.onEnergyUsed, nh)
	return h
}

func (h *handlerByFunc) OnScChg(isAdd bool, typ int, oldV, chgV int64, reason string) {
	for _, handler := range h.onScChg {
		handler(isAdd, typ, oldV, chgV, reason)
	}
}

func (h *handlerByFunc) WithScChg(sch onScChgFunc) *handlerByFunc {
	h.onScChg = append(h.onScChg, sch)
	return h
}

func (h *handlerByFunc) OnHcChg(isAdd bool, typ int, oldV, chgV int64, reason string) {
	for _, handler := range h.onHcChg {
		handler(isAdd, typ, oldV, chgV, reason)
	}
}

func (h *handlerByFunc) WithHcChg(hch onHcChgFunc) *handlerByFunc {
	h.onHcChg = append(h.onHcChg, hch)
	return h
}

func (h *handlerByFunc) OnTitleOnChg(oldTitle, newTitle string) {
	for _, handler := range h.onTitleChg {
		handler(oldTitle, newTitle)
	}
}

func (h *handlerByFunc) WithTitleOnChg(tlh onTitleOnChgFunc) *handlerByFunc {
	h.onTitleChg = append(h.onTitleChg, tlh)
	return h
}
