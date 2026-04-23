package matrix

import (
	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/matrix/averagespeed"
	"sakiko.local/sakiko-core/matrix/httpping"
	"sakiko.local/sakiko-core/matrix/inboundgeoip"
	"sakiko.local/sakiko-core/matrix/invalid"
	"sakiko.local/sakiko-core/matrix/maxspeed"
	"sakiko.local/sakiko-core/matrix/mediaunlock"
	"sakiko.local/sakiko-core/matrix/outboundgeoip"
	"sakiko.local/sakiko-core/matrix/persecondspeed"
	"sakiko.local/sakiko-core/matrix/rttping"
	"sakiko.local/sakiko-core/matrix/trafficused"
	"sakiko.local/sakiko-core/matrix/udpnattype"
)

var registered = map[interfaces.MatrixType]func() interfaces.Matrix{
	interfaces.MatrixHTTPPing:      func() interfaces.Matrix { return &httpping.Matrix{} },
	interfaces.MatrixRTTPing:       func() interfaces.Matrix { return &rttping.Matrix{} },
	interfaces.MatrixUDPNATType:    func() interfaces.Matrix { return &udpnattype.Matrix{} },
	interfaces.MatrixAverageSpeed:  func() interfaces.Matrix { return &averagespeed.Matrix{} },
	interfaces.MatrixMaxSpeed:      func() interfaces.Matrix { return &maxspeed.Matrix{} },
	interfaces.MatrixPerSecSpeed:   func() interfaces.Matrix { return &persecondspeed.Matrix{} },
	interfaces.MatrixTrafficUsed:   func() interfaces.Matrix { return &trafficused.Matrix{} },
	interfaces.MatrixInboundGeoIP:  func() interfaces.Matrix { return &inboundgeoip.Matrix{} },
	interfaces.MatrixOutboundGeoIP: func() interfaces.Matrix { return &outboundgeoip.Matrix{} },
	interfaces.MatrixMediaUnlock:   func() interfaces.Matrix { return &mediaunlock.Matrix{} },
}

func Find(t interfaces.MatrixType) interfaces.Matrix {
	if f, ok := registered[t]; ok {
		return f()
	}
	return &invalid.Matrix{}
}
