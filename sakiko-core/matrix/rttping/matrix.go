package rttping

import (
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/macro/ping"
)

type Matrix struct {
	Value uint16 `json:"value"`
}

func (m *Matrix) Type() interfaces.MatrixType {
	return interfaces.MatrixRTTPing
}

func (m *Matrix) MacroJob() interfaces.MacroType {
	return interfaces.MacroPing
}

func (m *Matrix) Extract(entry interfaces.MatrixEntry, macro interfaces.Macro) {
	_ = entry
	if p, ok := macro.(*ping.Macro); ok {
		m.Value = p.RTT
	}
}

func (m *Matrix) Payload() any {
	return m
}
