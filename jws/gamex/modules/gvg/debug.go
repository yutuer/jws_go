package gvg

func (m *gvgModule) SetDebugOffSetTime(t int64) {
	m.world.DebugOffSetTime = t
}
func (m *gvgModule) GetDebugOffsetTime() int64 {
	return m.world.DebugOffSetTime
}
func (m *gvgModule) ResetDebugTime() {
	m.world.DebugOffSetTime = 0
	m.world.LastBalanceTime = 0
	m.world.LastDayBalanceTime = 0
	m.world.LastResetTime = 0
	m.world.LastGuildDayBalanceTime = 0
}
