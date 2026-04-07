package macro

import (
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/macro/geo"
	"sakiko.local/sakiko-core/macro/invalid"
	"sakiko.local/sakiko-core/macro/media"
	"sakiko.local/sakiko-core/macro/ping"
	"sakiko.local/sakiko-core/macro/speed"
)

var registered = map[interfaces.MacroType]func() interfaces.Macro{
	interfaces.MacroPing:  func() interfaces.Macro { return &ping.Macro{} },
	interfaces.MacroSpeed: func() interfaces.Macro { return &speed.Macro{} },
	interfaces.MacroGeo:   func() interfaces.Macro { return &geo.Macro{} },
	interfaces.MacroMedia: func() interfaces.Macro { return &media.Macro{} },
}

func Find(t interfaces.MacroType) interfaces.Macro {
	if f, ok := registered[t]; ok {
		return f()
	}
	return &invalid.Macro{}
}
