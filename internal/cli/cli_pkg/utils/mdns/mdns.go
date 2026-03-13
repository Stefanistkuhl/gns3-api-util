package mdns

import (
	"context"
	"fmt"
	"time"

	"github.com/grandcat/zeroconf"
)

type DiscoveredMaster struct {
	Instance string            `json:"instance"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	IPv4     string            `json:"ipv4,omitempty"`
	IPv6     string            `json:"ipv6,omitempty"`
	TXT      map[string]string `json:"txt,omitempty"`
}

func BrowseMasters(ctx context.Context, timeout time.Duration) ([]DiscoveredMaster, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create mDNS resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry)
	masters := make([]DiscoveredMaster, 0, 8)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for e := range entries {
			master := DiscoveredMaster{
				Instance: e.Instance,
				Port:     e.Port,
				TXT:      parseTXT(e.Text),
			}

			if len(e.AddrIPv4) > 0 && e.AddrIPv4[0] != nil {
				master.IPv4 = e.AddrIPv4[0].String()
			}

			if len(e.AddrIPv6) > 0 && e.AddrIPv6[0] != nil {
				master.IPv6 = e.AddrIPv6[0].String()
			}

			host := master.IPv4
			if host == "" && master.IPv6 != "" {
				host = fmt.Sprintf("[%s]", master.IPv6)
			}
			if host != "" {
				master.Address = fmt.Sprintf("%s:%d", host, e.Port)
			}

			masters = append(masters, master)
		}
	}()

	browseCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := resolver.Browse(browseCtx, "_gns3util_master_api._tcp", "local.", entries); err != nil {
		return nil, fmt.Errorf("mDNS browse failed: %w", err)
	}

	<-browseCtx.Done()
	<-done

	return masters, nil
}

func parseTXT(txt []string) map[string]string {
	result := make(map[string]string)
	for _, s := range txt {
		for i := 0; i < len(s); i++ {
			if s[i] == '=' {
				key := s[:i]
				value := s[i+1:]
				result[key] = value
				break
			}
		}
	}
	return result
}
