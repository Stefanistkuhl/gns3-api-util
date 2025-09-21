package mdns

import (
	"context"
	"fmt"
	"time"

	"github.com/grandcat/zeroconf"
)

type Peer struct {
	Instance string
	Addr     string
	Port     int
	IPv4     string
	IPv6     string
	TXT      map[string]string
}

func parseTXT(txt []string) map[string]string {
	out := make(map[string]string, len(txt))
	for _, kv := range txt {
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				out[kv[:i]] = kv[i+1:]
				break
			}
		}
	}
	return out
}

func Browse(ctx context.Context, timeout time.Duration) ([]Peer, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	entries := make(chan *zeroconf.ServiceEntry)
	peers := make([]Peer, 0, 8)

	// Collect goroutine. Do not close(entries); zeroconf owns it.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for e := range entries {
			ip4 := ""
			if len(e.AddrIPv4) > 0 {
				ip4 = e.AddrIPv4[0].String()
			}
			ip6 := ""
			if len(e.AddrIPv6) > 0 {
				ip6 = e.AddrIPv6[0].String()
			}
			m := parseTXT(e.Text)

			host := ip4
			if host == "" && ip6 != "" {
				host = fmt.Sprintf("[%s]", ip6)
			}
			addr := ""
			if host != "" {
				addr = fmt.Sprintf("%s:%d", host, e.Port)
			}

			peers = append(peers, Peer{
				Instance: e.Instance,
				Addr:     addr,
				Port:     e.Port,
				IPv4:     ip4,
				IPv6:     ip6,
				TXT:      m,
			})
		}
	}()

	// Time-bounded browse
	qctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := resolver.Browse(qctx, "_gns3util-share._udp", "local.", entries); err != nil {
		return nil, err
	}

	// Wait for timeout, then for collector to finish draining.
	<-qctx.Done()
	// Zeroconf closes entries after browse context is done; wait for our collector to exit.
	<-done

	return peers, nil
}
