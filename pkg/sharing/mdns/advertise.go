package mdns

import (
	"fmt"

	"github.com/grandcat/zeroconf"
)

type Advertiser struct {
	srv *zeroconf.Server
}

func (a *Advertiser) Start(instance string, port int, txt map[string]string) (func(), error) {
	var txtSlice []string
	for k, v := range txt {
		txtSlice = append(txtSlice, fmt.Sprintf("%s=%s", k, v))
	}
	srv, err := zeroconf.Register(
		instance,
		"_gns3util-share._udp",
		"local.",
		port,
		txtSlice,
		nil,
	)
	if err != nil {
		return nil, err
	}
	a.srv = srv
	return func() { srv.Shutdown() }, nil
}
