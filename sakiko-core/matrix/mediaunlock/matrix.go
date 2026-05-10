package mediaunlock

import (
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/macro/media"
)

type Matrix struct {
	Items []interfaces.MediaUnlockPlatformResult `json:"items"`
}

func (m *Matrix) Type() interfaces.MatrixType {
	return interfaces.MatrixMediaUnlock
}

func (m *Matrix) MacroJob() interfaces.MacroType {
	return interfaces.MacroMedia
}

func (m *Matrix) Extract(entry interfaces.MatrixEntry, macro interfaces.Macro) {
	_ = entry
	if probe, ok := macro.(*media.Macro); ok {
		m.Items = append([]interfaces.MediaUnlockPlatformResult{}, probe.Result.Items...)
	}
}

func (m *Matrix) Payload() any {
	return m
}
