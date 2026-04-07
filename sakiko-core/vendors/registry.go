package vendors

import (
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/vendors/invalid"
	"sakiko.local/sakiko-core/vendors/local"
	"sakiko.local/sakiko-core/vendors/mihomo"
)

var registered = map[interfaces.VendorType]func() interfaces.Vendor{
	interfaces.VendorMihomo: func() interfaces.Vendor { return &mihomo.Vendor{} },
	interfaces.VendorLocal:  func() interfaces.Vendor { return &local.Vendor{} },
}

func Find(t interfaces.VendorType) interfaces.Vendor {
	if f, ok := registered[t]; ok {
		return f()
	}
	return &invalid.Vendor{}
}
