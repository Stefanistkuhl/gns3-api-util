package ssl

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/stefanistkuhl/gns3util/pkg/utils/ip"
)

//go:embed setup_https.sh
var setupHTTPSScript string

//go:embed uninstall_https.sh
var uninstallHTTPSScript string

const InteractiveOptionsText = `REVERSE_PROXY_PORT=443
GNS3_PORT=3080
DOMAIN=""
SUBJECT=/CN=localhost
FIREWALL_ALLOW=""
FIREWALL_BLOCK=False
VERBOSE=False
`

type InstallSSLArgs struct {
	FirewallAllow    string
	FirewallBlock    bool
	ReverseProxyPort int
	Domain           string
	GNS3Port         int
	Subject          string
	Verbose          bool
}

type ServerState struct {
	ServerHost       string    `json:"server_host"`
	InstallTime      time.Time `json:"install_time"`
	ReverseProxyPort int       `json:"reverse_proxy_port"`
	GNS3Port         int       `json:"gns3_port"`
	Domain           string    `json:"domain"`
	FirewallBlock    bool      `json:"firewall_block"`
	FirewallAllow    string    `json:"firewall_allow"`
	Distro           string    `json:"distro"`
	UFWEnabled       bool      `json:"ufw_enabled"`
	UFWRules         []string  `json:"ufw_rules"`
}

type StateManager struct {
	StateDir string
}

func NewStateManager() (*StateManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	stateDir := filepath.Join(homeDir, ".gns3")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	return &StateManager{StateDir: stateDir}, nil
}

func (sm *StateManager) SaveState(serverHost string, state ServerState) error {
	stateFile := filepath.Join(sm.StateDir, fmt.Sprintf("server_%s.json", serverHost))

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

func (sm *StateManager) LoadState(serverHost string) (*ServerState, error) {
	stateFile := filepath.Join(sm.StateDir, fmt.Sprintf("server_%s.json", serverHost))

	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no state found for server %s", serverHost)
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state ServerState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

func (sm *StateManager) DeleteState(serverHost string) error {
	stateFile := filepath.Join(sm.StateDir, fmt.Sprintf("server_%s.json", serverHost))

	if err := os.Remove(stateFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete state file: %w", err)
	}

	return nil
}

func (sm *StateManager) ListStates() ([]ServerState, error) {
	files, err := filepath.Glob(filepath.Join(sm.StateDir, "server_*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list state files: %w", err)
	}

	var states []ServerState
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var state ServerState
		if err := json.Unmarshal(data, &state); err != nil {
			continue
		}

		states = append(states, state)
	}

	return states, nil
}

func ValidateInstallSSLInput(args InstallSSLArgs) error {
	if args.FirewallBlock && args.FirewallAllow != "" {
		return fmt.Errorf("cannot both block all connections and allow specific subnet")
	}

	if args.FirewallAllow != "" {
		if !ip.IsValidIP(args.FirewallAllow) {
			return fmt.Errorf("invalid firewall allow IP format: %s (expected format: x.x.x.x/xx)", args.FirewallAllow)
		}
	}

	if args.ReverseProxyPort < 1 || args.ReverseProxyPort > 65535 {
		return fmt.Errorf("invalid reverse proxy port: %d (must be 1-65535)", args.ReverseProxyPort)
	}
	if args.GNS3Port < 1 || args.GNS3Port > 65535 {
		return fmt.Errorf("invalid GNS3 port: %d (must be 1-65535)", args.GNS3Port)
	}
	if !ip.IsValidSubject(args.Subject) {
		return fmt.Errorf("invalid subject format: %s (expected format: /CN=value[/...])", args.Subject)
	}

	if args.Domain != "" {
		if !ip.IsValidDomain(args.Domain) {
			return fmt.Errorf("invalid domain format: %s (expected format: example.com)", args.Domain)
		}
	}

	return nil
}

func ParseServerURLForSSH(serverURL string, portOption int) (string, int) {
	cleanURL := serverURL
	if result, found := strings.CutPrefix(serverURL, "http://"); found {
		cleanURL = result
	} else if result, found := strings.CutPrefix(serverURL, "https://"); found {
		cleanURL = result
	}
	// Parse the URL to extract hostname and port
	parsedURL, err := url.Parse("http://" + cleanURL)
	if err != nil {
		hostname, port := splitHostPort(cleanURL)
		if portOption > 0 {
			return hostname, portOption
		}
		if port > 0 {
			return hostname, port
		}
		return hostname, 22
	}

	hostname := parsedURL.Hostname()
	port := 22

	if portOption > 0 {
		port = portOption
	} else if parsedURL.Port() != "" {
		if parsedPort, err := strconv.Atoi(parsedURL.Port()); err == nil {
			port = parsedPort
		}
	}

	return hostname, port
}

func splitHostPort(hostPort string) (string, int) {
	parts := strings.Split(hostPort, ":")
	if len(parts) == 2 {
		if port, err := strconv.Atoi(parts[1]); err == nil {
			return parts[0], port
		}
	}
	return hostPort, 0
}

func EditScriptWithFlags(script string, args InstallSSLArgs) string {
	lines := strings.Split(script, "\n")

	for i, line := range lines {
		// Replace UFW="" with UFW="ufw" if firewall block is enabled
		if strings.Contains(line, `UFW=""`) {
			if args.FirewallBlock || args.FirewallAllow != "" {
				lines[i] = strings.Replace(line, `UFW=""`, `UFW="ufw"`, 1)
			}
		}

		// Replace RP_PORT="" with actual reverse proxy port
		if strings.Contains(line, `RP_PORT=""`) {
			lines[i] = strings.Replace(line, `RP_PORT=""`, fmt.Sprintf(`RP_PORT="%d"`, args.ReverseProxyPort), 1)
		}

		// Replace GNS3_PORT="" with actual GNS3 port
		if strings.Contains(line, `GNS3_PORT=""`) {
			lines[i] = strings.Replace(line, `GNS3_PORT=""`, fmt.Sprintf(`GNS3_PORT="%d"`, args.GNS3Port), 1)
		}

		// Replace DOMAIN="" with actual domain if provided
		if strings.Contains(line, `DOMAIN=""`) {
			if args.Domain != "" {
				lines[i] = strings.Replace(line, `DOMAIN=""`, fmt.Sprintf(`DOMAIN="%s"`, args.Domain), 1)
			}
		}

		// Replace SUBJ="" with actual subject
		if strings.Contains(line, `SUBJ=""`) {
			lines[i] = strings.Replace(line, `SUBJ=""`, fmt.Sprintf(`SUBJ="%s"`, args.Subject), 1)
		}

		// Handle UFW_ENABLE replacement based on firewall settings
		if strings.Contains(line, "UFW_ENABLE") {
			if args.FirewallBlock {
				// Block all external access to GNS3 port, only allow localhost for reverse proxy
				firewallRules := []string{
					"# Ensure SSH access is preserved",
					"$SUDO ufw allow ssh",
					"$SUDO ufw allow 22",
					"# Block all external access to GNS3 port (including Tailscale/VPN)",
					fmt.Sprintf("$SUDO ufw deny from any to any port %d", args.GNS3Port),
					"# Allow only localhost access for the reverse proxy (Caddy)",
					fmt.Sprintf("$SUDO ufw allow from 127.0.0.1 to any port %d", args.GNS3Port),
					"# Allow only localhost access for the reverse proxy (Caddy) - IPv6",
					fmt.Sprintf("$SUDO ufw allow from ::1 to any port %d", args.GNS3Port),
					fmt.Sprintf("echo \"Port %d blocked from all external access (including Tailscale/VPN)\"", args.GNS3Port),
					fmt.Sprintf("echo \"Only localhost can access port %d (for reverse proxy)\"", args.GNS3Port),
				}
				lines[i] = strings.Join(firewallRules, "\n")
			} else if args.FirewallAllow != "" {
				// Allow specific subnet to GNS3 port and block others from GNS3 port only
				// First ensure SSH is allowed, then configure GNS3 port rules
				firewallRules := []string{
					"# Ensure SSH access is preserved",
					"$SUDO ufw allow ssh",
					"$SUDO ufw allow 22",
					"# Allow specific subnet to GNS3 port",
					fmt.Sprintf("$SUDO ufw allow from %s to any port %d", args.FirewallAllow, args.GNS3Port),
					"# Block all other LAN access to GNS3 port",
					fmt.Sprintf("$SUDO ufw deny %d", args.GNS3Port),
					fmt.Sprintf("echo \"Port %d allowed for subnet %s, blocked for others\"", args.GNS3Port, args.FirewallAllow),
				}
				lines[i] = strings.Join(firewallRules, "\n")
			} else {
				// No firewall rules
				lines[i] = "# No firewall rules configured"
			}
		}
	}

	return strings.Join(lines, "\n")
}

func GetEmbeddedScript() string {
	return setupHTTPSScript
}

func ParseInteractiveOptions(optionsText string) (InstallSSLArgs, error) {
	args := InstallSSLArgs{
		ReverseProxyPort: 443,
		GNS3Port:         3080,
		Subject:          "/CN=localhost",
		FirewallBlock:    false,
		Verbose:          false,
	}

	for line := range strings.SplitSeq(optionsText, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		value = strings.Trim(value, `"`)

		switch key {
		case "REVERSE_PROXY_PORT":
			if port, err := strconv.Atoi(value); err == nil {
				args.ReverseProxyPort = port
			}
		case "GNS3_PORT":
			if port, err := strconv.Atoi(value); err == nil {
				args.GNS3Port = port
			}
		case "DOMAIN":
			args.Domain = value
		case "SUBJECT":
			args.Subject = value
		case "FIREWALL_ALLOW":
			args.FirewallAllow = value
		case "FIREWALL_BLOCK":
			args.FirewallBlock = strings.ToLower(value) == "true"
		case "VERBOSE":
			args.Verbose = strings.ToLower(value) == "true"
		}
	}

	return args, nil
}

func GetUninstallScript() string {
	return uninstallHTTPSScript
}

func EditUninstallScriptWithFlags(script string, args InstallSSLArgs) string {
	lines := strings.Split(script, "\n")

	for i, line := range lines {
		if strings.Contains(line, "VERBOSE=\"\"") {
			if args.Verbose {
				lines[i] = "VERBOSE=\"True\""
			} else {
				lines[i] = "VERBOSE=\"False\""
			}
		}

		if strings.Contains(line, "REVERSE_PROXY_PORT=\"\"") {
			lines[i] = fmt.Sprintf("REVERSE_PROXY_PORT=\"%d\"", args.ReverseProxyPort)
		}

		if strings.Contains(line, "GNS3_PORT=\"\"") {
			lines[i] = fmt.Sprintf("GNS3_PORT=\"%d\"", args.GNS3Port)
		}

		if strings.Contains(line, "DOMAIN=\"\"") {
			if args.Domain != "" {
				lines[i] = fmt.Sprintf("DOMAIN=\"%s\"", args.Domain)
			} else {
				lines[i] = "DOMAIN=\"\""
			}
		}

		if strings.Contains(line, "FIREWALL_ALLOW=\"\"") {
			if args.FirewallAllow != "" {
				lines[i] = fmt.Sprintf("FIREWALL_ALLOW=\"%s\"", args.FirewallAllow)
			} else {
				lines[i] = "FIREWALL_ALLOW=\"\""
			}
		}

		if strings.Contains(line, "FIREWALL_BLOCK=\"\"") {
			if args.FirewallBlock {
				lines[i] = "FIREWALL_BLOCK=\"True\""
			} else {
				lines[i] = "FIREWALL_BLOCK=\"False\""
			}
		}
	}

	return strings.Join(lines, "\n")
}
