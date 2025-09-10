package ip

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func IsValidIP(ip string) bool {
	// Pattern matches: xxx.xxx.xxx.xxx/xx format
	pattern := regexp.MustCompile(
		`^` +
			`(` +
			`(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\.` +
			`){3}` +
			`(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])` +
			`/` +
			`(3[0-2]|[12]?[0-9])` +
			`$`,
	)
	return pattern.MatchString(ip)
}

func IsValidDomain(domain string) bool {
	// Pattern matches: [A-Za-z0-9-]{1,63}\.[A-Za-z]{2,6}
	pattern := regexp.MustCompile(`^[A-Za-z0-9-]{1,63}(\.[A-Za-z0-9-]{1,63})*\.[A-Za-z]{2,6}$`)
	return pattern.MatchString(domain)
}

func IsValidPort(port string) bool {
	portNum, err := ParsePort(port)
	return err == nil && portNum >= 1 && portNum <= 65535
}

func ParsePort(s string) (int, error) {
	s = strings.Trim(s, `"`)

	if s == "" {
		return 0, fmt.Errorf("empty port")
	}

	port, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", s)
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port number out of range: %d", port)
	}

	return port, nil
}

func IsValidSubject(subject string) bool {
	// Pattern matches /key=value[/...] format with at least /CN=value
	pattern := regexp.MustCompile(
		`^/` +
			`(?:C=[A-Z]{2}` +
			`|ST=[^/=]+` +
			`|L=[^/=]+` +
			`|O=[^/=]+` +
			`|OU=[^/=]+` +
			`|CN=[^/=]+` +
			`|emailAddress=[^/=]+)` +
			`(?:/` +
			`(?:C=[A-Z]{2}` +
			`|ST=[^/=]+` +
			`|L=[^/=]+` +
			`|O=[^/=]+` +
			`|OU=[^/=]+` +
			`|CN=[^/=]+` +
			`|emailAddress=[^/=]+)` +
			`)*$`,
	)
	return pattern.MatchString(subject) && strings.Contains(subject, "/CN=")
}
