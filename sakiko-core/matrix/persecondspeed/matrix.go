package persecondspeed

import (
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/macro/speed"
)

type Matrix struct {
	Values []uint64 `json:"values"`
}

func (m *Matrix) Type() interfaces.MatrixType {
	return interfaces.MatrixPerSecSpeed
}

func (m *Matrix) MacroJob() interfaces.MacroType {
	return interfaces.MacroSpeed
}

func (m *Matrix) Extract(entry interfaces.MatrixEntry, macro interfaces.Macro) {
	_ = entry
	if s, ok := macro.(*speed.Macro); ok {
		m.Values = append([]uint64{}, s.Speeds...)
	}
}

func (m *Matrix) Payload() any {
	return m
}
