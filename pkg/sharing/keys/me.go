package keys

import (
	"os"
	"strings"
)

func DeviceLabel() string {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	host, _ := os.Hostname()
	if host == "" {
		host = "device"
	}
	user = sanitize(user)
	host = sanitize(host)
	if user == "" {
		return host
	}
	return user + "@" + host
}

func sanitize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "-")
	if len(s) > 32 {
		s = s[:32]
	}
	return s
}
