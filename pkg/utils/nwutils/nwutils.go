package nwutils

import (
	"net"
	"net/url"
	"strconv"
	"strings"
)

func IsValidListenAddr(addr string) bool {
	if addr == "" {
		return false
	}
	if net.ParseIP(addr) != nil {
		return true
	}
	return strings.Contains(addr, ".") || strings.Contains(addr, ":")
}

func ParsePort(input string) (int, bool) {
	if input == "" {
		return 0, false
	}

	port, err := strconv.Atoi(input)
	if err != nil {
		return 0, false
	}
	if port < 0 || port > 65535 {
		return 0, false
	}
	return port, true
}

func ConvertMasterAPIURL(masterAPIURL string) string {
	elems := strings.Split(masterAPIURL, ":")
	elems[len(elems)-1] = "2379"
	return strings.Join(elems, ":")
}

func ParseURL(urlStr string) (*url.URL, bool) {
	if urlStr == "" {
		return nil, false
	}
	u, err := url.Parse(urlStr)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, false
	}
	return u, true
}
