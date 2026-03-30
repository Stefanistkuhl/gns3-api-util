package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type DocsConfig struct {
	Schema     string                 `json:"$schema"`
	Theme      string                 `json:"theme"`
	Name       string                 `json:"name"`
	PrimaryTab map[string]string      `json:"primaryTab"`
	ModeToggle map[string]interface{} `json:"modeToggle"`
	Colors     map[string]string      `json:"colors"`
	Favicon    string                 `json:"favicon"`
	Navigation Navigation             `json:"navigation"`
	Logo       map[string]string      `json:"logo"`
	Navbar     map[string]interface{} `json:"navbar"`
}

type Navigation struct {
	Tabs   []Tab                  `json:"tabs"`
	Global map[string]interface{} `json:"global,omitempty"`
}

type Tab struct {
	Tab     string  `json:"tab"`
	Groups  []Group `json:"groups,omitempty"`
	OpenAPI string  `json:"openapi,omitempty"`
	Href    string  `json:"href,omitempty"`
}

type Group struct {
	Group string   `json:"group"`
	Pages []string `json:"pages"`
}

type CommandInfo struct {
	Name     string   // e.g., "cluster"
	Path     []string // e.g., ["cluster", "config"]
	FilePath string   // e.g., "cli/gns3util_cluster_config.md"
	Category string   // "root" or "subcommand"
}

// GenerateDocsJSON scans CLI docs and updates docs.json with complete hierarchy
func GenerateDocsJSON(cliDir, docsJsonPath string) error {
	// Parse all CLI markdown files
	commands, err := parseCLIFiles(cliDir)
	if err != nil {
		return fmt.Errorf("failed to parse CLI files: %w", err)
	}

	// Read existing docs.json to preserve other tabs
	config, err := readDocsJSON(docsJsonPath)
	if err != nil {
		return fmt.Errorf("failed to read existing docs.json: %w", err)
	}

	// Generate hierarchical CLI reference groups
	cliGroups := buildHierarchy(commands)

	// Find and replace the "CLI Reference" tab
	for i, tab := range config.Navigation.Tabs {
		if tab.Tab == "CLI Reference" {
			config.Navigation.Tabs[i].Groups = cliGroups
			config.Navigation.Tabs[i].Href = "cli/gns3util" // Entry point when clicking the tab
			break
		}
	}

	// Write updated docs.json
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal docs.json: %w", err)
	}

	// #nosec G306
	return os.WriteFile(docsJsonPath, append(data, '\n'), 0o644)
}

// parseCLIFiles scans the CLI directory and extracts command structure
func parseCLIFiles(cliDir string) ([]CommandInfo, error) {
	var commands []CommandInfo

	entries, err := os.ReadDir(cliDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		// Extract command structure from filename
		// e.g., gns3util_cluster_config_apply.md -> ["cluster", "config", "apply"]
		// or gns3util.md -> ["gns3util"]
		parts := strings.Split(strings.TrimSuffix(entry.Name(), ".md"), "_")
		if len(parts) < 1 {
			continue // Skip invalid files
		}

		var cmdParts []string
		switch {
		case len(parts) == 1 && parts[0] == "gns3util":
			// Root command case
			cmdParts = []string{"gns3util"}
		case len(parts) >= 2:
			// Skip the "gns3util" prefix for subcommands
			cmdParts = parts[1:]

			// Normalize underscores to hyphens for consistency
			for i := range cmdParts {
				cmdParts[i] = strings.ReplaceAll(cmdParts[i], "_", "-")
			}
		default:
			continue // Skip invalid files
		}

		commands = append(commands, CommandInfo{
			Name:     cmdParts[0],
			Path:     cmdParts,
			FilePath: fmt.Sprintf("cli/%s", strings.TrimSuffix(entry.Name(), ".md")),
		})
	}

	return commands, nil
}

// buildHierarchy creates hierarchical structure: New Orchestration, Legacy Orchestration, and GNS3 Operations
func buildHierarchy(commands []CommandInfo) []Group {
	newOrch := make(map[string][]string)
	legacyOrch := make(map[string][]string)
	gns3Ops := make(map[string][]string)

	// Categorize commands
	for _, cmd := range commands {
		filePath := cmd.FilePath
		category := getOrchestrationCategory(cmd.Path)

		switch category {
		case "new":
			groupName := getGroupName(cmd.Path)
			newOrch[groupName] = append(newOrch[groupName], filePath)
		case "legacy":
			groupName := getGroupName(cmd.Path)
			legacyOrch[groupName] = append(legacyOrch[groupName], filePath)
		default:
			groupName := getGroupName(cmd.Path)
			gns3Ops[groupName] = append(gns3Ops[groupName], filePath)
		}
	}

	// Sort pages within each group
	for k := range newOrch {
		sort.Strings(newOrch[k])
	}
	for k := range legacyOrch {
		sort.Strings(legacyOrch[k])
	}
	for k := range gns3Ops {
		sort.Strings(gns3Ops[k])
	}

	// Build the final groups list
	var groups []Group

	// Add Root Command at the top
	groups = append(groups, Group{
		Group: "CLI Overview",
		Pages: []string{"cli/gns3util"},
	})

	// Add New Orchestration section
	if len(newOrch) > 0 {
		orchGroups := buildGroupsFromMap(newOrch)
		for _, grp := range orchGroups {
			groups = append(groups, Group{
				Group: "Orchestration (New) / " + grp.Group,
				Pages: grp.Pages,
			})
		}
	}

	// Add Legacy Orchestration section
	if len(legacyOrch) > 0 {
		orchGroups := buildGroupsFromMap(legacyOrch)
		for _, grp := range orchGroups {
			groups = append(groups, Group{
				Group: "Orchestration (Legacy) / " + grp.Group,
				Pages: grp.Pages,
			})
		}
	}

	// Add GNS3 Operations section
	if len(gns3Ops) > 0 {
		gnsGroups := buildGroupsFromMap(gns3Ops)
		for _, grp := range gnsGroups {
			groups = append(groups, Group{
				Group: "GNS3 Operations / " + grp.Group,
				Pages: grp.Pages,
			})
		}
	}

	return groups
}

// isOrchestrationCommand checks if a command is orchestration-related and returns the category
// Returns: "new", "legacy", or ""
func getOrchestrationCategory(path []string) string {
	if len(path) == 0 {
		return ""
	}

	rootCmd := path[0]

	// New Orchestration: ctl and cluster-control
	if rootCmd == "ctl" || rootCmd == "cluster-control" {
		return "new"
	}

	// Legacy Orchestration: cluster with certain subcommands
	if rootCmd == "cluster" && len(path) > 1 {
		subCmd := path[1]
		if subCmd == "config" || subCmd == "create" || subCmd == "add-node" || subCmd == "add-nodes" {
			return "legacy"
		}
	}

	return ""
}

// getGroupName determines the group name for a command path
func getGroupName(path []string) string {
	if len(path) == 0 {
		return "Root"
	}

	rootCmd := path[0]

	switch rootCmd {
	case "gns3util":
		// Root command stays in its own section
		if len(path) == 1 {
			return "Root Command"
		}
		return "Root & Core"
	case "ctl", "cluster-control":
		if len(path) > 1 {
			switch path[1] {
			case "object-store":
				return "Object Storage"
			case "auth":
				return "Authentication"
			case "jobs":
				return "Jobs Management"
			case "create":
				return "Cluster Creation"
			case "discover":
				return "Cluster Discovery"
			default:
				return cases.Title(language.English).String(strings.ReplaceAll(path[1], "-", " "))
			}
		}
		return "Cluster Control"
	case "cluster":
		if len(path) > 1 {
			switch path[1] {
			case "config":
				return "Cluster Configuration"
			case "add-node", "add-nodes":
				return "Node Management"
			default:
				return "Cluster Operations"
			}
		}
		return "Cluster Management"
	case "project":
		return "Project Management"
	case "node":
		return "Node Operations"
	case "acl":
		return "Access Control"
	case "role":
		return "Role Management"
	case "group":
		return "Group Management"
	case "user":
		return "User Management"
	case "auth":
		return "Authentication"
	case "appliance":
		return "Appliances"
	case "class":
		return "Classes"
	case "image":
		return "Images"
	case "pool":
		return "Resource Pools"
	case "compute":
		return "Compute Resources"
	case "template":
		return "Templates"
	case "drawing":
		return "Drawings"
	case "link":
		return "Links"
	case "symbol":
		return "Symbols"
	case "exercise":
		return "Exercises"
	case "snapshot":
		return "Snapshots"
	case "share":
		return "Sharing"
	case "remote":
		return "Remote Operations"
	case "system":
		return "System"
	default:
		return cases.Title(language.English).String(strings.ReplaceAll(rootCmd, "-", " "))
	}
}

// buildGroupsFromMap converts a map of groupName -> pages into sorted Group slices
func buildGroupsFromMap(m map[string][]string) []Group {
	groups := make([]Group, 0, len(m))

	// Sort group names for consistent ordering
	groupNames := make([]string, 0, len(m))
	for name := range m {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	for _, name := range groupNames {
		groups = append(groups, Group{
			Group: name,
			Pages: m[name],
		})
	}

	return groups
}

// readDocsJSON reads and parses the existing docs.json
func readDocsJSON(path string) (*DocsConfig, error) {
	// #nosec G304
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config DocsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
