package ctlcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/api"
)

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication related operations for clusters",
		Long:  `Authentication related operations for cluster management.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}
	cmd.AddCommand(NewAuthStatusCmd())
	return cmd
}

func NewAuthStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check authentication status against a cluster master",
		Long: `Check authentication status against a cluster master node.
Use --server to test a master URL before adding to a cluster.
Use --cluster to test against an existing cluster's master from the keyfile.
If GNS3_TOKEN environment variable isn't set, the token from the keyfile will be used.`,
		Annotations: map[string]string{
			"auth-mode": "flexible",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if envErr := viper.BindEnv("token", "GNS3_TOKEN"); envErr != nil {
				return fmt.Errorf("failed to bind token environment variable: %w", envErr)
			}

			keyFilePath, err := pathutils.ResolveKeyFilePath(cfg.KeyFile)
			if err != nil {
				return fmt.Errorf("failed to resolve key file path: %w", err)
			}

			kf, err := pathutils.LoadGNS3KeysFile(keyFilePath)
			if err != nil {
				return fmt.Errorf("failed to load keys: %w", err)
			}

			var serverURL string

			switch {
			case cfg.Cluster != "" && cfg.ClusterEntry != nil:
				serverURL = cfg.ClusterEntry.Master.URL
			case cfg.Server != "":
				serverURL = cfg.Server
			default:
				return fmt.Errorf("either --server or --cluster must be specified")
			}

			token := viper.GetString("token")
			if token == "" {
				if cfg.Cluster != "" && cfg.ClusterEntry != nil {
					token = cfg.ClusterEntry.Master.AccessToken
				} else {
					entry, ok := findMasterToken(kf, serverURL)
					if !ok {
						return fmt.Errorf("no token found for %q in key file (set GNS3_TOKEN or add cluster first)", serverURL)
					}
					token = entry.AccessToken
				}
			}

			settings := api.NewSettings(
				api.WithBaseURLV2(serverURL+"/api/v1"),
				api.WithToken(token),
				api.WithVerify(!cfg.Insecure),
				api.WithCA([]byte(cfg.ClusterEntry.CaCert)),
			)
			client := api.NewClientV2(settings)

			resp, err := client.GetAuthStatus(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get auth status: %w", err)
			}

			fmt.Printf("%s authenticated as user: %s with scopes: %s\n",
				messageUtils.SuccessMsg("Success:"),
				messageUtils.Bold(resp.User),
				messageUtils.Bold(resp.Scopes))

			return nil
		},
	}

	return cmd
}

func findMasterToken(kf *pathutils.KeyFileV2, masterURL string) (*pathutils.ServiceEntry, bool) {
	for i := range kf.Clusters {
		c := &kf.Clusters[i]
		if c.Master.URL == masterURL {
			return &c.Master, true
		}
	}
	return nil, false
}
