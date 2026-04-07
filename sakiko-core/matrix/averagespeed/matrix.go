package averagespeed

import (
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/macro/speed"
)

type Matrix struct {
	Value uint64 `json:"value"`
}

func (m *Matrix) Type() interfaces.MatrixType {
	return interfaces.MatrixAverageSpeed
}

func (m *Matrix) MacroJob() interfaces.MacroType {
	return interfaces.MacroSpeed
}

func (m *Matrix) Extract(entry interfaces.MatrixEntry, macro interfaces.Macro) {
	_ = entry
	if s, ok := macro.(*speed.Macro); ok {
		m.Value = s.AvgSpeed
	}
}

func (m *Matrix) Payload() any {
	return m
}
