package invalid

import "sakiko.local/sakiko-core/interfaces"

type Matrix struct{}

func (m *Matrix) Type() interfaces.MatrixType {
	return interfaces.MatrixInvalid
}

func (m *Matrix) MacroJob() interfaces.MacroType {
	return interfaces.MacroInvalid
}

func (m *Matrix) Extract(entry interfaces.MatrixEntry, macro interfaces.Macro) {
	_ = entry
	_ = macro
}

func (m *Matrix) Payload() any {
	return map[string]any{}
}
