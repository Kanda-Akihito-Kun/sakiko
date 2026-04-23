package invalid

import (
	"context"

	"sakiko.local/sakiko-core/interfaces"
)

type Macro struct{}

func (m *Macro) Type() interfaces.MacroType {
	return interfaces.MacroInvalid
}

func (m *Macro) Run(ctx context.Context, proxy interfaces.Vendor, task *interfaces.Task) error {
	_ = ctx
	_ = proxy
	_ = task
	return nil
}
