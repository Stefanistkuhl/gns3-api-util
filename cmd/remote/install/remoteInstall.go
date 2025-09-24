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
	"github.com/stefanistkuhl/gns3util/pkg/utils/gns3"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

func NewRemoteInstallCmd() *cobra.Command {
	var (
		port              int
		privateKeyPath    string
		gns3Username      string
		homeDir           string
		listenHost        string
		gns3Port          int
		disableKVM        bool
		installDocker     bool
		installVirtualBox bool
		installVMware     bool
		useIOU            bool
		enableI386        bool
		verbose           bool
		interactive       bool
	)

	var cmd = &cobra.Command{
		Use:   "gns3 [user]",
		Short: "Install GNS3 server on a remote machine",
		Long: `Install GNS3 server on a remote machine via SSH.

This command will:
- Install GNS3 server and dependencies
- Configure GNS3 server with specified options
- Set up systemd service for GNS3
- Optionally install Docker, VirtualBox, VMware support
- Optionally configure IOU support

The installation supports Ubuntu LTS releases and requires Python 3.9+.`,
		Args: cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			// Handle interactive mode
			if interactive {
				editedText, err := utils.EditTextWithEditor(gns3.InteractiveOptionsText, "txt")
				if err != nil {
					log.Fatalf("failed to edit options: %v", err)
				}

				// Parse the edited options
				interactiveArgs, err := gns3.ParseInteractiveOptions(editedText)
				if err != nil {
					log.Fatalf("failed to parse interactive options: %v", err)
				}

				// Override the flag values with interactive values
				gns3Username = interactiveArgs.Username
				homeDir = interactiveArgs.HomeDir
				listenHost = interactiveArgs.ListenHost
				gns3Port = interactiveArgs.GNS3Port
				disableKVM = interactiveArgs.DisableKVM
				installDocker = interactiveArgs.InstallDocker
				installVirtualBox = interactiveArgs.InstallVirtualBox
				installVMware = interactiveArgs.InstallVMware
				useIOU = interactiveArgs.UseIOU
				enableI386 = interactiveArgs.EnableI386
				verbose = interactiveArgs.Verbose
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			sshUser := args[0]

			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				log.Fatalf("failed to get global options: %v", err)
			}

			gns3Args := gns3.InstallGNS3Args{
				Username:          gns3Username,
				HomeDir:           homeDir,
				ListenHost:        listenHost,
				GNS3Port:          gns3Port,
				DisableKVM:        disableKVM,
				InstallDocker:     installDocker,
				InstallVirtualBox: installVirtualBox,
				InstallVMware:     installVMware,
				UseIOU:            useIOU,
				EnableI386:        enableI386,
				Verbose:           verbose,
			}

			if err := gns3.ValidateInstallGNS3Input(gns3Args); err != nil {
				log.Fatalf("validation error: %v", err)
			}

			hostname, sshPort := gns3.ParseServerURLForSSH(cfg.Server, port)

			fmt.Printf("%s %s\n", messageUtils.Bold("🔧"), messageUtils.Bold("GNS3 Server Installation"))
			fmt.Printf("%s\n", messageUtils.Seperator(strings.Repeat("─", 50)))
			fmt.Println()

			fmt.Printf("%s Connecting to remote server...\n", messageUtils.InfoMsg("Connecting to remote server"))
			sshClient, err := ssh.ConnectWithKeyOrPassword(hostname, sshUser, sshPort, privateKeyPath, verbose)
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

			fmt.Printf("%s Preparing GNS3 installation script...\n", messageUtils.InfoMsg("Preparing GNS3 installation script"))
			scriptText := gns3.GetEmbeddedScript()
			editedScript := gns3.EditScriptWithFlags(scriptText, gns3Args)
			fmt.Printf("%s Script prepared\n", messageUtils.SuccessMsg("Script prepared"))

			fmt.Printf("%s Installing GNS3 server...\n", messageUtils.InfoMsg("Installing GNS3 server"))

			// Create the script
			createScriptCmd := fmt.Sprintf(`cat > /tmp/gns3_install.sh << 'SCRIPT_EOF'
%s
SCRIPT_EOF`, editedScript)

			result, err := sshClient.ExecuteCommand(createScriptCmd)
			if err != nil || !result.Success {
				fmt.Printf("%s Failed to create installation script: %v\n", messageUtils.ErrorMsg("Failed to create installation script"), err)
				return
			}

			// Make script executable
			chmodResult, err := sshClient.ExecuteCommand("chmod +x /tmp/gns3_install.sh")
			if err != nil || !chmodResult.Success {
				fmt.Printf("%s Failed to make script executable: %v\n", messageUtils.ErrorMsg("Failed to make script executable"), err)
				return
			}

			// Execute the script and capture output
			execResult, err := sshClient.ExecuteCommand("bash /tmp/gns3_install.sh")
			if err != nil {
				fmt.Printf("%s Failed to execute GNS3 installation script: %v\n", messageUtils.ErrorMsg("Failed to execute GNS3 installation script"), err)
				return
			}

			// Clean up script file
			_, _ = sshClient.ExecuteCommand("rm -f /tmp/gns3_install.sh")

			success := execResult.Success

			if success {
				fmt.Printf("%s GNS3 server installed successfully\n", messageUtils.SuccessMsg("GNS3 server installed successfully"))

				fmt.Printf("%s Saving installation state...\n", messageUtils.InfoMsg("Saving installation state"))
				stateManager, err := gns3.NewStateManager()
				if err != nil {
					fmt.Printf("%s failed to create state manager: %v\n", messageUtils.WarningMsg("failed to create state manager"), err)
				} else {
					state := gns3.GNS3ServerState{
						ServerHost:        hostname,
						InstallTime:       time.Now(),
						Username:          gns3Args.Username,
						HomeDir:           gns3Args.HomeDir,
						ListenHost:        gns3Args.ListenHost,
						GNS3Port:          gns3Args.GNS3Port,
						DisableKVM:        gns3Args.DisableKVM,
						InstallDocker:     gns3Args.InstallDocker,
						InstallVirtualBox: gns3Args.InstallVirtualBox,
						InstallVMware:     gns3Args.InstallVMware,
						UseIOU:            gns3Args.UseIOU,
						EnableI386:        gns3Args.EnableI386,
						Distro:            "unknown",
					}

					if err := stateManager.SaveState(hostname, state); err != nil {
						fmt.Printf("%s failed to save state: %v\n", messageUtils.WarningMsg("failed to save state"), err)
					} else {
						fmt.Printf("%s State saved for server %s\n", messageUtils.SuccessMsg("State saved for server"), hostname)
					}
				}

				fmt.Printf("\n%s Successfully installed GNS3 server\n", messageUtils.SuccessMsg("Successfully installed GNS3 server"))
				fmt.Printf("%s GNS3 server is accessible on %s:%d\n", messageUtils.InfoMsg("GNS3 server accessible"), gns3Args.ListenHost, gns3Args.GNS3Port)
				if gns3Args.InstallDocker {
					fmt.Printf("%s Docker support enabled\n", messageUtils.InfoMsg("Docker support enabled"))
				}
				if gns3Args.InstallVirtualBox {
					fmt.Printf("%s VirtualBox support enabled\n", messageUtils.InfoMsg("VirtualBox support enabled"))
				}
				if gns3Args.InstallVMware {
					fmt.Printf("%s VMware support enabled\n", messageUtils.InfoMsg("VMware support enabled"))
				}
				if gns3Args.UseIOU {
					fmt.Printf("%s IOU support enabled\n", messageUtils.InfoMsg("IOU support enabled"))
				}
			} else {
				if execResult.Stderr != "" {
					fmt.Printf("%s\n", messageUtils.ErrorMsg(execResult.Stderr))
				}
				if execResult.Stdout != "" {
					fmt.Printf("%s Script output:\n", messageUtils.InfoMsg("Script output"))
					fmt.Printf("%s\n", messageUtils.InfoMsg(execResult.Stdout))
				}
				return
			}
		},
	}

	cmd.Flags().IntVar(&port, "port", 22, "SSH port")
	cmd.Flags().StringVar(&privateKeyPath, "key", "", "Path to a custom SSH private key file")
	cmd.Flags().StringVar(&gns3Username, "username", "gns3", "Username for GNS3 service")
	cmd.Flags().StringVar(&homeDir, "home-dir", "/opt/gns3", "Home directory for GNS3 user")
	cmd.Flags().StringVar(&listenHost, "listen-host", "0.0.0.0", "Host address for GNS3 server to bind to")
	cmd.Flags().IntVar(&gns3Port, "gns3-port", 3080, "Port for GNS3 server to listen on")
	cmd.Flags().BoolVar(&disableKVM, "disable-kvm", false, "Disable KVM hardware acceleration")
	cmd.Flags().BoolVar(&installDocker, "install-docker", false, "Install Docker for GNS3 appliances")
	cmd.Flags().BoolVar(&installVirtualBox, "install-virtualbox", false, "Install VirtualBox support")
	cmd.Flags().BoolVar(&installVMware, "install-vmware", false, "Install VMware integration packages")
	cmd.Flags().BoolVar(&useIOU, "use-iou", false, "Install IOU support (requires valid license)")
	cmd.Flags().BoolVar(&enableI386, "enable-i386", false, "Enable i386 architecture for legacy IOU")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Set the options for this command interactively")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Run this command with extra logging")

	return cmd
}
