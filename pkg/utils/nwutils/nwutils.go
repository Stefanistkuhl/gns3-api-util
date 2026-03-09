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

func GetLocalIPs() []net.IP {
	var ips []net.IP
	ips = append(ips, net.ParseIP("127.0.0.1"), net.ParseIP("::1"))

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP)
			}
		}
	}
	return ips
}

func GetActiveMulticastInterfaces() []net.Interface {
	var validIfaces []net.Interface
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	for _, i := range ifaces {
		if i.Flags&net.FlagUp == 0 {
			continue
		}
		if i.Flags&net.FlagLoopback != 0 {
			continue
		}
		if i.Flags&net.FlagMulticast == 0 {
			continue
		}

		validIfaces = append(validIfaces, i)
	}
	return validIfaces
}

func GetFirstNonLoopbackIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func NormalizeURL(urlStr string) string {
	if strings.HasPrefix(urlStr, "http://") {
		urlStr = urlStr[7:]
	} else if strings.HasPrefix(urlStr, "https://") {
		urlStr = urlStr[8:]
	}

	if colonIndex := strings.Index(urlStr, ":"); colonIndex != -1 {
		urlStr = urlStr[:colonIndex]
	}

	return urlStr
}
