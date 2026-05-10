package udpnattype

import (
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/macro/udp"
)

type Matrix struct {
	interfaces.UDPNATInfo
}

func (m *Matrix) Type() interfaces.MatrixType {
	return interfaces.MatrixUDPNATType
}

func (m *Matrix) MacroJob() interfaces.MacroType {
	return interfaces.MacroUDP
}

func (m *Matrix) Extract(entry interfaces.MatrixEntry, macro interfaces.Macro) {
	_ = entry
	if probe, ok := macro.(*udp.Macro); ok {
		m.UDPNATInfo = probe.Info
	}
}

func (m *Matrix) Payload() any {
	return m
}
