package nwutils

import (
	"net"
	"strings"
	"testing"
)

func TestParsePort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		wantOK   bool
	}{
		{"valid port 8080", "8080", 8080, true},
		{"valid port 0", "0", 0, true},
		{"valid port 65535", "65535", 65535, true},
		{"invalid empty string", "", 0, false},
		{"invalid negative port", "-1", 0, false},
		{"invalid too high port", "65536", 0, false},
		{"invalid non-numeric", "abc", 0, false},
		{"invalid float", "80.5", 0, false},
		{"invalid with spaces", " 80 ", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParsePort(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ParsePort(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if ok && got != tt.expected {
				t.Errorf("ParsePort(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsValidListenAddr(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected bool
	}{
		{"valid IPv4", "192.168.1.1", true},
		{"valid IPv6", "::1", true},
		{"valid localhost", "127.0.0.1", true},
		{"valid hostname with dots", "example.com", true},
		{"valid hostname with colon", "localhost:8080", true},
		{"invalid empty string", "", false},
		{"invalid just dots", "...", true},
		{"invalid just colon", ":", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidListenAddr(tt.addr)
			if got != tt.expected {
				t.Errorf("IsValidListenAddr(%q) = %v, want %v", tt.addr, got, tt.expected)
			}
		})
	}
}

func TestConvertMasterAPIURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple URL", "http://example.com:8080", "http://example.com:2379"},
		{"https URL", "https://example.com:8443", "https://example.com:2379"},
		{"localhost", "http://localhost:8080", "http://localhost:2379"},
		{"IP address", "http://192.168.1.1:8080", "http://192.168.1.1:2379"},
		{"IPv6", "http://[::1]:8080", "http://[::1]:2379"},
		{"multiple colons", "http://example.com:8080/path", "http://example.com:2379"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertMasterAPIURL(tt.input)
			if got != tt.expected {
				t.Errorf("ConvertMasterAPIURL(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseURL(t *testing.T) {
	tests := []struct {
		name       string
		urlStr     string
		wantOK     bool
		wantHost   string
		wantScheme string
	}{
		{"valid HTTP URL", "http://example.com", true, "example.com", "http"},
		{"valid HTTPS URL", "https://example.com/path", true, "example.com", "https"},
		{"valid URL with port", "http://localhost:8080", true, "localhost:8080", "http"},
		{"invalid empty string", "", false, "", ""},
		{"invalid no scheme", "example.com", false, "", ""},
		{"invalid no host", "http://", false, "", ""},
		{"invalid malformed", "http://:", true, ":", "http"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseURL(tt.urlStr)
			if ok != tt.wantOK {
				t.Errorf("ParseURL(%q) ok = %v, want %v", tt.urlStr, ok, tt.wantOK)
				return
			}
			if ok {
				if got.Host != tt.wantHost {
					t.Errorf("ParseURL(%q) host = %v, want %v", tt.urlStr, got.Host, tt.wantHost)
				}
				if got.Scheme != tt.wantScheme {
					t.Errorf("ParseURL(%q) scheme = %v, want %v", tt.urlStr, got.Scheme, tt.wantScheme)
				}
			}
		})
	}
}

func TestGetLocalIPs(t *testing.T) {
	ips := GetLocalIPs()

	hasLocalhost := false
	hasIPv6Localhost := false

	for _, ip := range ips {
		if ip.Equal(net.ParseIP("127.0.0.1")) {
			hasLocalhost = true
		}
		if ip.Equal(net.ParseIP("::1")) {
			hasIPv6Localhost = true
		}
	}

	if !hasLocalhost {
		t.Error("GetLocalIPs() should contain 127.0.0.1")
	}
	if !hasIPv6Localhost {
		t.Error("GetLocalIPs() should contain ::1")
	}

	if len(ips) < 2 {
		t.Errorf("GetLocalIPs() returned %d IPs, expected at least 2", len(ips))
	}
}

func TestGetActiveMulticastInterfaces(t *testing.T) {
	ifaces := GetActiveMulticastInterfaces()

	if ifaces == nil {
		t.Error("GetActiveMulticastInterfaces() returned nil")
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			t.Errorf("Interface %v is not up", iface.Name)
		}
		if iface.Flags&net.FlagLoopback != 0 {
			t.Errorf("Interface %v is loopback", iface.Name)
		}
		if iface.Flags&net.FlagMulticast == 0 {
			t.Errorf("Interface %v does not support multicast", iface.Name)
		}
	}
}

func TestGetFirstNonLoopbackIP(t *testing.T) {
	ip := GetFirstNonLoopbackIP()

	if ip == "" {
		t.Error("GetFirstNonLoopbackIP() returned empty string")
	}

	if net.ParseIP(ip) == nil {
		t.Errorf("GetFirstNonLoopbackIP() returned invalid IP: %s", ip)
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"http://example.com", "example.com"},
		{"https://example.com", "example.com"},
		{"http://example.com:8080", "example.com"},
		{"https://example.com:8443", "example.com"},
		{"example.com", "example.com"},
		{"example.com:8080", "example.com"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeURL(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeURL(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPortBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		wantOK   bool
	}{
		{"min valid port", "0", 0, true},
		{"max valid port", "65535", 65535, true},
		{"just above max", "65536", 0, false},
		{"negative", "-1", 0, false},
		{"very large number", "999999", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParsePort(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ParsePort(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if ok && got != tt.expected {
				t.Errorf("ParsePort(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseURLPaths(t *testing.T) {
	tests := []struct {
		name       string
		urlStr     string
		wantOK     bool
		wantHost   string
		wantScheme string
	}{
		{"with path", "http://example.com/api/v1", true, "example.com", "http"},
		{"with query", "https://example.com/path?key=value", true, "example.com", "https"},
		{"with fragment", "http://example.com/path#section", true, "example.com", "http"},
		{"localhost", "http://localhost", true, "localhost", "http"},
		{"with userinfo", "http://user:pass@example.com", true, "example.com", "http"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseURL(tt.urlStr)
			if ok != tt.wantOK {
				t.Errorf("ParseURL(%q) ok = %v, want %v", tt.urlStr, ok, tt.wantOK)
				return
			}
			if ok {
				if got.Host != tt.wantHost {
					t.Errorf("ParseURL(%q) host = %v, want %v", tt.urlStr, got.Host, tt.wantHost)
				}
				if got.Scheme != tt.wantScheme {
					t.Errorf("ParseURL(%q) scheme = %v, want %v", tt.urlStr, got.Scheme, tt.wantScheme)
				}
			}
		})
	}
}

func TestConvertMasterAPIURLEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"no port specified", "http://example.com"},
		{"with path", "http://example.com:8080/some/path"},
		{"already etcd port", "http://example.com:2379"},
		{"multiple colons IPv6", "http://[::1]:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertMasterAPIURL(tt.input)
			if got == "" {
				t.Errorf("ConvertMasterAPIURL(%q) returned empty string", tt.input)
			}
			if !strings.Contains(got, "2379") {
				t.Errorf("ConvertMasterAPIURL(%q) should contain port 2379, got %q", tt.input, got)
			}
		})
	}
}

func TestIsValidListenAddrComplexCases(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected bool
	}{
		{"valid hostname with dot", "gns3.example.com", true},
		{"valid subdomain", "api.server.example.com", true},
		{"simple hostname no dot", "myserver", false},
		{"ipv4 with port", "192.168.1.1:8080", true},
		{"wildcard address", "0.0.0.0", true},
		{"localhost", "127.0.0.1", true},
		{"ipv6", "::1", true},
		{"hostname with port", "example.com:8080", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidListenAddr(tt.addr)
			if got != tt.expected {
				t.Errorf("IsValidListenAddr(%q) = %v, want %v", tt.addr, got, tt.expected)
			}
		})
	}
}
