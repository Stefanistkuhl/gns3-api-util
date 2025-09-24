package install

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/ssh"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/ssl"
)

func NewInstallCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "install",
		Short: "Install something on the remote server via SSH",
		Long:  `Install something on the remote server via SSH`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	cmd.AddCommand(NewInstallHttpsCmd())
	cmd.AddCommand(NewRemoteInstallCmd())
	return cmd
}

func NewInstallHttpsCmd() *cobra.Command {
	var (
		user             string
		port             int
		privateKeyPath   string
		reverseProxyPort int
		gns3Port         int
		domain           string
		subject          string
		firewallAllow    string
		firewallBlock    bool
		interactive      bool
		verbose          bool
	)

	var cmd = &cobra.Command{
		Use:   "https [user]",
		Short: "Install a reverse proxy for HTTPS",
		Long:  `Install caddy for HTTPS and optionally set firewall rules`,
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			// Handle interactive mode
			if interactive {
				editedText, err := utils.EditTextWithEditor(ssl.InteractiveOptionsText, "txt")
				if err != nil {
					log.Fatalf("failed to edit options: %v", err)
				}

				// Parse the edited options
				interactiveArgs, err := ssl.ParseInteractiveOptions(editedText)
				if err != nil {
					log.Fatalf("failed to parse interactive options: %v", err)
				}

				// Override the flag values with interactive values
				reverseProxyPort = interactiveArgs.ReverseProxyPort
				gns3Port = interactiveArgs.GNS3Port
				domain = interactiveArgs.Domain
				subject = interactiveArgs.Subject
				firewallAllow = interactiveArgs.FirewallAllow
				firewallBlock = interactiveArgs.FirewallBlock
				verbose = interactiveArgs.Verbose
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			user = args[0]

			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				log.Fatalf("failed to get global options: %v", err)
			}

			sslArgs := ssl.InstallSSLArgs{
				FirewallAllow:    firewallAllow,
				FirewallBlock:    firewallBlock,
				ReverseProxyPort: reverseProxyPort,
				Domain:           domain,
				GNS3Port:         gns3Port,
				Subject:          subject,
				Verbose:          verbose,
			}

			if err := ssl.ValidateInstallSSLInput(sslArgs); err != nil {
				log.Fatalf("validation error: %v", err)
			}

			hostname, sshPort := ssl.ParseServerURLForSSH(cfg.Server, port)

			fmt.Printf("%s %s\n", messageUtils.Bold("🔧"), messageUtils.Bold("GNS3 SSL Installation"))
			fmt.Printf("%s\n", messageUtils.Seperator(strings.Repeat("─", 50)))
			fmt.Println()

			fmt.Printf("%s Connecting to remote server...\n", messageUtils.InfoMsg("Connecting to remote server"))
			sshClient, err := ssh.ConnectWithKeyOrPassword(hostname, user, sshPort, privateKeyPath, verbose)
			if err != nil {
				fmt.Printf("%s Failed to connect via SSH: %v\n", messageUtils.ErrorMsg("Failed to connect via SSH"), err)
				return
			}
			defer sshClient.Close()
			fmt.Printf("%s Connected successfully\n", messageUtils.SuccessMsg("Connected successfully"))

			fmt.Printf("%s Checking user privileges...\n", messageUtils.InfoMsg("Checking user privileges"))
			if err := sshClient.CheckPrivileges(); err != nil {
				fmt.Printf("%s Privilege check failed: %v\n", messageUtils.ErrorMsg("Privilege check failed"), err)
				return
			}
			fmt.Printf("%s Privileges verified\n", messageUtils.SuccessMsg("Privileges verified"))

			fmt.Printf("%s Preparing SSL installation script...\n", messageUtils.InfoMsg("Preparing SSL installation script"))
			scriptText := ssl.GetEmbeddedScript()
			modifiedScript := ssl.EditScriptWithFlags(scriptText, sslArgs)
			fmt.Printf("%s Script prepared\n", messageUtils.SuccessMsg("Script prepared"))

			fmt.Printf("%s Installing Caddy reverse proxy...\n", messageUtils.InfoMsg("Installing Caddy reverse proxy"))
			success, err := sshClient.ExecuteScript(modifiedScript, "/tmp/setup_https.sh")
			if err != nil {
				fmt.Printf("%s Failed to execute SSL installation script: %v\n", messageUtils.ErrorMsg("Failed to execute SSL installation script"), err)
				return
			}

			if success {
				fmt.Printf("%s Caddy reverse proxy installed successfully\n", messageUtils.SuccessMsg("Caddy reverse proxy installed successfully"))

				fmt.Printf("%s Saving installation state...\n", messageUtils.InfoMsg("Saving installation state"))
				stateManager, err := ssl.NewStateManager()
				if err != nil {
					fmt.Printf("%s failed to create state manager: %v\n", messageUtils.WarningMsg("failed to create state manager"), err)
				} else {
					state := ssl.ServerState{
						ServerHost:       hostname,
						InstallTime:      time.Now(),
						ReverseProxyPort: reverseProxyPort,
						GNS3Port:         gns3Port,
						Domain:           domain,
						FirewallBlock:    firewallBlock,
						FirewallAllow:    firewallAllow,
						Distro:           "unknown",
						UFWEnabled:       firewallBlock || firewallAllow != "",
						UFWRules:         []string{"allow ssh", "allow 22", fmt.Sprintf("deny %d", gns3Port)},
					}

					if err := stateManager.SaveState(hostname, state); err != nil {
						fmt.Printf("%s failed to save state: %v\n", messageUtils.WarningMsg("failed to save state"), err)
					} else {
						fmt.Printf("%s State saved for server %s\n", messageUtils.SuccessMsg("State saved for server"), hostname)
					}
				}

				fmt.Printf("\n%s Successfully installed Caddy reverse proxy on port %d\n", messageUtils.SuccessMsg("Successfully installed Caddy reverse proxy"), reverseProxyPort)
				if firewallBlock {
					fmt.Printf("%s Port %d is now blocked from all external access (including Tailscale/VPN)\n", messageUtils.InfoMsg("Port blocked from external access"), gns3Port)
				}
				fmt.Printf("%s GNS3 server is now accessible via HTTPS on port %d\n", messageUtils.InfoMsg("GNS3 server accessible via HTTPS"), reverseProxyPort)
			} else {
				fmt.Printf("%s SSL installation script failed\n", messageUtils.ErrorMsg("SSL installation script failed"))
				return
			}
		},
	}

	cmd.Flags().IntVar(&port, "port", 22, "SSH port")
	cmd.Flags().StringVar(&privateKeyPath, "key", "", "Path to a custom SSH private key file")
	cmd.Flags().IntVar(&reverseProxyPort, "reverse-proxy-port", 443, "Port for the reverse proxy to use")
	cmd.Flags().IntVar(&gns3Port, "gns3-port", 3080, "Port of the GNS3 Server")
	cmd.Flags().StringVar(&domain, "domain", "", "Domain to use for the reverse proxy")
	cmd.Flags().StringVar(&subject, "subject", "/CN=localhost", "Set the subject alternative name for the SSL certificate")
	cmd.Flags().StringVar(&firewallAllow, "firewall-allow", "", "Block all connections to the GNS3 server port and only allow a given subnet. Example: 10.0.0.0/24")
	cmd.Flags().BoolVar(&firewallBlock, "firewall-block", false, "Block all connections to the port of the GNS3 server")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Set the options for this command interactively")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Run this command with extra logging")

	return cmd
}
