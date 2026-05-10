package outboundgeoip

import (
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/macro/geo"
)

type Matrix struct {
	interfaces.GeoIPInfo
}

func (m *Matrix) Type() interfaces.MatrixType {
	return interfaces.MatrixOutboundGeoIP
}

func (m *Matrix) MacroJob() interfaces.MacroType {
	return interfaces.MacroGeo
}

func (m *Matrix) Extract(entry interfaces.MatrixEntry, macro interfaces.Macro) {
	_ = entry
	if g, ok := macro.(*geo.Macro); ok {
		m.GeoIPInfo = g.Outbound
	}
}

func (m *Matrix) Payload() any {
	return m
}
