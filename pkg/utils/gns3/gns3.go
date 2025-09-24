package gns3

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
)

//go:embed gns3_install.sh
var gns3InstallScript string

//go:embed gns3_uninstall.sh
var gns3UninstallScript string

const InteractiveOptionsText = `USERNAME="gns3"
HOME_DIR="/opt/gns3"
LISTEN_HOST="0.0.0.0"
GNS3_PORT=3080
DISABLE_KVM=False
INSTALL_DOCKER=False
INSTALL_VIRTUALBOX=False
INSTALL_VMWARE=False
USE_IOU=False
ENABLE_I386_FOR_IOU=False
VERBOSE=False
`

const UninstallInteractiveOptionsText = `HOME_DIR="/opt/gns3"
GNS3_PORT=3080
VERBOSE=False
PRESERVE_DATA=False
`

type InstallGNS3Args struct {
	Username          string
	HomeDir           string
	ListenHost        string
	GNS3Port          int
	DisableKVM        bool
	InstallDocker     bool
	InstallVirtualBox bool
	InstallVMware     bool
	UseIOU            bool
	EnableI386        bool
	Verbose           bool
	PreserveData      bool // For uninstall operations
}

type GNS3ServerState struct {
	ServerHost        string    `json:"server_host"`
	InstallTime       time.Time `json:"install_time"`
	Username          string    `json:"username"`
	HomeDir           string    `json:"home_dir"`
	ListenHost        string    `json:"listen_host"`
	GNS3Port          int       `json:"gns3_port"`
	DisableKVM        bool      `json:"disable_kvm"`
	InstallDocker     bool      `json:"install_docker"`
	InstallVirtualBox bool      `json:"install_virtualbox"`
	InstallVMware     bool      `json:"install_vmware"`
	UseIOU            bool      `json:"use_iou"`
	EnableI386        bool      `json:"enable_i386"`
	Distro            string    `json:"distro"`
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

func (sm *StateManager) SaveState(serverHost string, state GNS3ServerState) error {
	stateFile := filepath.Join(sm.StateDir, fmt.Sprintf("gns3_server_%s.json", serverHost))

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

func (sm *StateManager) LoadState(serverHost string) (*GNS3ServerState, error) {
	stateFile := filepath.Join(sm.StateDir, fmt.Sprintf("gns3_server_%s.json", serverHost))

	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no state found for server %s", serverHost)
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state GNS3ServerState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

func (sm *StateManager) DeleteState(serverHost string) error {
	stateFile := filepath.Join(sm.StateDir, fmt.Sprintf("gns3_server_%s.json", serverHost))

	if err := os.Remove(stateFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete state file: %w", err)
	}

	return nil
}

func (sm *StateManager) ListStates() ([]GNS3ServerState, error) {
	files, err := filepath.Glob(filepath.Join(sm.StateDir, "gns3_server_*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list state files: %w", err)
	}

	var states []GNS3ServerState
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var state GNS3ServerState
		if err := json.Unmarshal(data, &state); err != nil {
			continue
		}

		states = append(states, state)
	}

	return states, nil
}

func ValidateInstallGNS3Input(args InstallGNS3Args) error {
	if args.GNS3Port < 1 || args.GNS3Port > 65535 {
		return fmt.Errorf("invalid GNS3 port: %d (must be 1-65535)", args.GNS3Port)
	}

	if args.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if args.HomeDir == "" {
		return fmt.Errorf("home directory cannot be empty")
	}

	if args.ListenHost == "" {
		return fmt.Errorf("listen host cannot be empty")
	}

	return nil
}

func ValidateUninstallGNS3Input(args InstallGNS3Args) error {
	if args.GNS3Port < 1 || args.GNS3Port > 65535 {
		return fmt.Errorf("invalid GNS3 port: %d (must be 1-65535)", args.GNS3Port)
	}

	if args.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if args.HomeDir == "" {
		return fmt.Errorf("home directory cannot be empty")
	}

	// Note: ListenHost is not validated for uninstall as it's not critical

	return nil
}

func ParseServerURLForSSH(serverURL string, portOption int) (string, int) {
	cleanURL := serverURL
	if strings.HasPrefix(serverURL, "http://") {
		cleanURL = strings.TrimPrefix(serverURL, "http://")
	} else if strings.HasPrefix(serverURL, "https://") {
		cleanURL = strings.TrimPrefix(serverURL, "https://")
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

func EditScriptWithFlags(script string, args InstallGNS3Args) string {
	lines := strings.Split(script, "\n")

	for i, line := range lines {
		// Replace GNS3_USER="" with actual username
		if strings.Contains(line, `GNS3_USER="gns3"`) {
			lines[i] = strings.Replace(line, `GNS3_USER="gns3"`, fmt.Sprintf(`GNS3_USER="%s"`, args.Username), 1)
		}

		// Replace GNS3_HOME="" with actual home directory
		if strings.Contains(line, `GNS3_HOME="/opt/gns3"`) {
			lines[i] = strings.Replace(line, `GNS3_HOME="/opt/gns3"`, fmt.Sprintf(`GNS3_HOME="%s"`, args.HomeDir), 1)
		}

		// Replace GNS3_PORT with actual port
		if strings.Contains(line, `GNS3_PORT=3080`) {
			lines[i] = strings.Replace(line, `GNS3_PORT=3080`, fmt.Sprintf(`GNS3_PORT=%d`, args.GNS3Port), 1)
		}

		// Replace GNS3_LISTEN_HOST with actual listen host
		if strings.Contains(line, `GNS3_LISTEN_HOST="0.0.0.0"`) {
			lines[i] = strings.Replace(line, `GNS3_LISTEN_HOST="0.0.0.0"`, fmt.Sprintf(`GNS3_LISTEN_HOST="%s"`, args.ListenHost), 1)
		}

		// Replace DISABLE_KVM with actual value
		if strings.Contains(line, `DISABLE_KVM=0`) {
			if args.DisableKVM {
				lines[i] = strings.Replace(line, `DISABLE_KVM=0`, `DISABLE_KVM=1`, 1)
			}
		}

		// Replace INSTALL_DOCKER with actual value
		if strings.Contains(line, `INSTALL_DOCKER=0`) {
			if args.InstallDocker {
				lines[i] = strings.Replace(line, `INSTALL_DOCKER=0`, `INSTALL_DOCKER=1`, 1)
			}
		}

		// Replace INSTALL_VIRTUALBOX with actual value
		if strings.Contains(line, `INSTALL_VIRTUALBOX=0`) {
			if args.InstallVirtualBox {
				lines[i] = strings.Replace(line, `INSTALL_VIRTUALBOX=0`, `INSTALL_VIRTUALBOX=1`, 1)
			}
		}

		// Replace INSTALL_VMWARE with actual value
		if strings.Contains(line, `INSTALL_VMWARE=0`) {
			if args.InstallVMware {
				lines[i] = strings.Replace(line, `INSTALL_VMWARE=0`, `INSTALL_VMWARE=1`, 1)
			}
		}

		// Replace USE_IOU with actual value
		if strings.Contains(line, `USE_IOU=0`) {
			if args.UseIOU {
				lines[i] = strings.Replace(line, `USE_IOU=0`, `USE_IOU=1`, 1)
			}
		}

		// Replace ENABLE_I386_FOR_IOU with actual value
		if strings.Contains(line, `ENABLE_I386_FOR_IOU=0`) {
			if args.EnableI386 {
				lines[i] = strings.Replace(line, `ENABLE_I386_FOR_IOU=0`, `ENABLE_I386_FOR_IOU=1`, 1)
			}
		}
	}

	return strings.Join(lines, "\n")
}

func GetEmbeddedScript() string {
	return gns3InstallScript
}

func ParseInteractiveOptions(optionsText string) (InstallGNS3Args, error) {
	args := InstallGNS3Args{
		Username:          "gns3",
		HomeDir:           "/opt/gns3",
		ListenHost:        "0.0.0.0",
		GNS3Port:          3080,
		DisableKVM:        false,
		InstallDocker:     false,
		InstallVirtualBox: false,
		InstallVMware:     false,
		UseIOU:            false,
		EnableI386:        false,
		Verbose:           false,
	}

	lines := strings.Split(optionsText, "\n")

	for _, line := range lines {
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
		case "USERNAME":
			args.Username = value
		case "HOME_DIR":
			args.HomeDir = value
		case "LISTEN_HOST":
			args.ListenHost = value
		case "GNS3_PORT":
			if port, err := strconv.Atoi(value); err == nil {
				args.GNS3Port = port
			}
		case "DISABLE_KVM":
			args.DisableKVM = strings.ToLower(value) == "true"
		case "INSTALL_DOCKER":
			args.InstallDocker = strings.ToLower(value) == "true"
		case "INSTALL_VIRTUALBOX":
			args.InstallVirtualBox = strings.ToLower(value) == "true"
		case "INSTALL_VMWARE":
			args.InstallVMware = strings.ToLower(value) == "true"
		case "USE_IOU":
			args.UseIOU = strings.ToLower(value) == "true"
		case "ENABLE_I386_FOR_IOU":
			args.EnableI386 = strings.ToLower(value) == "true"
		case "VERBOSE":
			args.Verbose = strings.ToLower(value) == "true"
		case "PRESERVE_DATA":
			args.PreserveData = strings.ToLower(value) == "true"
		}
	}

	return args, nil
}

func GetUninstallScript() string {
	return gns3UninstallScript
}

func EditUninstallScriptWithFlags(script string, args InstallGNS3Args) string {
	lines := strings.Split(script, "\n")

	for i, line := range lines {
		if strings.Contains(line, "GNS3_USER=\"\"") {
			lines[i] = fmt.Sprintf("GNS3_USER=\"%s\"", args.Username)
		}

		if strings.Contains(line, "GNS3_HOME=\"\"") {
			lines[i] = fmt.Sprintf("GNS3_HOME=\"%s\"", args.HomeDir)
		}

		if strings.Contains(line, "GNS3_PORT=\"\"") {
			lines[i] = fmt.Sprintf("GNS3_PORT=\"%d\"", args.GNS3Port)
		}

		if strings.Contains(line, "VERBOSE=\"\"") {
			if args.Verbose {
				lines[i] = "VERBOSE=\"True\""
			} else {
				lines[i] = "VERBOSE=\"False\""
			}
		}

		if strings.Contains(line, "PRESERVE_DATA=\"\"") {
			if args.PreserveData {
				lines[i] = "PRESERVE_DATA=\"True\""
			} else {
				lines[i] = "PRESERVE_DATA=\"False\""
			}
		}
	}

	return strings.Join(lines, "\n")
}
