package invalid

import "sakiko.local/sakiko-core/interfaces"

type Macro struct{}

func (m *Macro) Type() interfaces.MacroType {
	return interfaces.MacroInvalid
}

func (m *Macro) Run(proxy interfaces.Vendor, task *interfaces.Task) error {
	_ = proxy
	_ = task
	return nil
}
