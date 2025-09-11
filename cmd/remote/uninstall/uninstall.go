package uninstall

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/ssh"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/ssl"
)

// NewUninstallCmdGroup creates the uninstall command group
func NewUninstallCmdGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall services from remote GNS3 servers",
		Long:  "Uninstall various services and configurations from remote GNS3 servers",
	}

	// Add subcommands
	cmd.AddCommand(NewUninstallHTTPSCmd())

	return cmd
}

// NewUninstallHTTPSCmd creates the uninstall https command
func NewUninstallHTTPSCmd() *cobra.Command {
	var (
		port             int
		privateKeyPath   string
		reverseProxyPort int
		gns3Port         int
		domain           string
		subject          string
		firewallAllow    string
		firewallBlock    bool
		verbose          bool
		interactive      bool
	)

	cmd := &cobra.Command{
		Use:   "https [user]",
		Short: "Uninstall SSL reverse proxy setup from remote GNS3 server",
		Long: `Uninstall the SSL reverse proxy setup from a remote GNS3 server.

This command will:
- Stop and remove Caddy reverse proxy
- Remove SSL certificates
- Remove firewall rules
- Clean up systemd services
- Remove configuration files

The GNS3 server itself will remain running on its original port.

If a state file is found from a previous installation, all configuration values
will be automatically loaded and command line flags will be ignored.`,
		Args: cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			// Interactive mode
			if interactive {
				// Prompt user for each option
				fmt.Println("=== Interactive SSL Uninstallation Setup ===")
				fmt.Println("Press Enter to use default values (shown in brackets)")
				fmt.Println()

				// Prompt for reverse proxy port
				fmt.Printf("Reverse proxy port to uninstall [%d]: ", reverseProxyPort)
				var input string
				_, _ = fmt.Scanln(&input)
				if input != "" {
					if port, err := strconv.Atoi(input); err == nil {
						reverseProxyPort = port
					}
				}

				// Prompt for GNS3 port
				fmt.Printf("GNS3 server port [%d]: ", gns3Port)
				_, _ = fmt.Scanln(&input)
				if input != "" {
					if port, err := strconv.Atoi(input); err == nil {
						gns3Port = port
					}
				}

				// Prompt for domain
				fmt.Printf("Domain that was used (leave empty if none) [%s]: ", domain)
				_, _ = fmt.Scanln(&input)
				if input != "" {
					domain = input
				}

				// Prompt for subject
				fmt.Printf("SSL certificate subject that was used [%s]: ", subject)
				_, _ = fmt.Scanln(&input)
				if input != "" {
					subject = input
				}

				// Prompt for firewall allow
				fmt.Printf("Firewall allow subnet that was used (leave empty if none) [%s]: ", firewallAllow)
				_, _ = fmt.Scanln(&input)
				if input != "" {
					firewallAllow = input
				}

				// Prompt for firewall block
				fmt.Printf("Were firewall rules configured? (y/N): ")
				_, _ = fmt.Scanln(&input)
				firewallBlock = strings.ToLower(input) == "y" || strings.ToLower(input) == "yes"

				fmt.Println()
				fmt.Println("=== Uninstallation Configuration Summary ===")
				fmt.Printf("Reverse proxy port: %d\n", reverseProxyPort)
				fmt.Printf("GNS3 server port: %d\n", gns3Port)
				fmt.Printf("Domain: %s\n", domain)
				fmt.Printf("Subject: %s\n", subject)
				fmt.Printf("Firewall allow: %s\n", firewallAllow)
				fmt.Printf("Firewall block: %t\n", firewallBlock)
				fmt.Println()
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			user := args[0]

			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("%s Failed to get global options: %v\n", colorUtils.Error("✗"), err)
				return
			}

			// Parse server URL for SSH connection
			hostname, sshPort := ssl.ParseServerURLForSSH(cfg.Server, port)

			// Try to load state from local machine first
			stateManager, err := ssl.NewStateManager()
			var state *ssl.ServerState
			if err == nil {
				if localState, err := stateManager.LoadState(hostname); err == nil {
					state = localState
					fmt.Printf("%s Loaded state from local machine\n", colorUtils.Success("✓"))
				}
			}

			// Create SSL uninstall arguments (use state if available, otherwise use flags)
			sslArgs := ssl.InstallSSLArgs{
				FirewallAllow:    firewallAllow,
				FirewallBlock:    firewallBlock,
				ReverseProxyPort: reverseProxyPort,
				Domain:           domain,
				GNS3Port:         gns3Port,
				Subject:          subject,
				Verbose:          verbose,
			}

			// Override with state values if available
			if state != nil {
				sslArgs.FirewallAllow = state.FirewallAllow
				sslArgs.FirewallBlock = state.FirewallBlock
				sslArgs.ReverseProxyPort = state.ReverseProxyPort
				sslArgs.Domain = state.Domain
				sslArgs.GNS3Port = state.GNS3Port
				sslArgs.Subject = "/CN=localhost" // Default subject
				fmt.Printf("%s Using saved configuration: RP=%d, GNS3=%d, Firewall=%t\n",
					colorUtils.Info("ℹ"), state.ReverseProxyPort, state.GNS3Port, state.FirewallBlock)
			} else {
				// If no state found, show warning and use default values
				fmt.Printf("%s No state found, using command line flags or defaults\n", colorUtils.Warning("⚠"))
				if sslArgs.ReverseProxyPort == 443 && sslArgs.GNS3Port == 3080 && !sslArgs.FirewallBlock {
					fmt.Printf("%s Using default values: RP=443, GNS3=3080, Firewall=false\n", colorUtils.Info("ℹ"))
				}
			}

			// Validate arguments
			if err := ssl.ValidateInstallSSLInput(sslArgs); err != nil {
				fmt.Printf("%s Validation error: %v\n", colorUtils.Error("✗"), err)
				return
			}

			// Show uninstall header
			fmt.Printf("%s %s\n", colorUtils.Bold("🗑️"), colorUtils.Bold("GNS3 SSL Uninstallation"))
			fmt.Printf("%s\n", colorUtils.Seperator(strings.Repeat("─", 50)))
			fmt.Println()

			// Step 1: Connect via SSH
			fmt.Printf("%s Connecting to remote server...\n", colorUtils.Info("→"))
			sshClient, err := ssh.ConnectWithKeyOrPassword(hostname, user, sshPort, privateKeyPath, verbose)
			if err != nil {
				fmt.Printf("%s Failed to connect via SSH: %v\n", colorUtils.Error("✗"), err)
				return
			}
			defer sshClient.Close()
			fmt.Printf("%s Connected successfully\n", colorUtils.Success("✓"))

			// Step 2: Check privileges
			fmt.Printf("%s Checking user privileges...\n", colorUtils.Info("→"))
			if err := sshClient.CheckPrivileges(); err != nil {
				fmt.Printf("%s Privilege check failed: %v\n", colorUtils.Error("✗"), err)
				return
			}
			fmt.Printf("%s Privileges verified\n", colorUtils.Success("✓"))

			// Step 3: Prepare uninstall script
			fmt.Printf("%s Preparing SSL uninstall script...\n", colorUtils.Info("→"))
			script := ssl.GetUninstallScript()
			editedScript := ssl.EditUninstallScriptWithFlags(script, sslArgs)
			fmt.Printf("%s Script prepared\n", colorUtils.Success("✓"))

			// Step 4: Execute uninstall
			fmt.Printf("%s Uninstalling Caddy reverse proxy...\n", colorUtils.Info("→"))
			success, err := sshClient.ExecuteScript(editedScript, "/tmp/gns3_ssl_uninstall.sh")
			if err != nil {
				fmt.Printf("%s Failed to execute uninstall script: %v\n", colorUtils.Error("✗"), err)
				return
			}

			if !success {
				fmt.Printf("%s Uninstall script failed\n", colorUtils.Error("✗"))
				return
			}
			fmt.Printf("%s Uninstall completed\n", colorUtils.Success("✓"))

			// Clean up local state
			if stateManager != nil {
				if err := stateManager.DeleteState(hostname); err != nil {
					fmt.Printf("%s Warning: failed to delete local state: %v\n", colorUtils.Warning("⚠"), err)
				} else {
					fmt.Printf("%s Local state cleaned up\n", colorUtils.Success("✓"))
				}
			}

			// Show success message
			fmt.Printf("\n%s Successfully uninstalled Caddy reverse proxy\n", colorUtils.Success("✓"))
			fmt.Printf("%s GNS3 server is now accessible on port %d\n", colorUtils.Info("ℹ"), sslArgs.GNS3Port)
		},
	}

	// Add flags
	cmd.Flags().IntVarP(&port, "port", "p", 22, "SSH port")
	cmd.Flags().StringVarP(&privateKeyPath, "key", "", "", "Path to a custom SSH private key file")
	cmd.Flags().IntVarP(&reverseProxyPort, "reverse-proxy-port", "r", 443, "Port for the reverse proxy that was used")
	cmd.Flags().IntVarP(&gns3Port, "gns3-port", "g", 3080, "Port of the GNS3 Server")
	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Domain that was used for the reverse proxy")
	cmd.Flags().StringVarP(&subject, "subject", "", "/CN=localhost", "Subject that was used for the SSL certificate")
	cmd.Flags().StringVarP(&firewallAllow, "firewall-allow", "a", "", "Firewall allow subnet that was used. Example: 10.0.0.0/24")
	cmd.Flags().BoolVarP(&firewallBlock, "firewall-block", "b", false, "Whether firewall rules were configured")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVarP(&interactive, "interactive", "t", false, "Set the options for this command interactively")

	return cmd
}
