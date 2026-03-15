package ctlcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/ctlhelpers"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/api"
)

var clusterName string

func NewAddClusterCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-cluster",
		Short: "Add a cluster and discover its nodes",
		Long:  `Bootstraps trust with a GNS3 master, adds it to the keyfile, and automatically discovers all associated worker nodes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			keyFilePath, err := pathutils.ResolveKeyFilePath(cfg.KeyFile)
			if err != nil {
				return fmt.Errorf("failed to resolve key file path: %w", err)
			}

			token := viper.GetString("token")
			settings := api.NewSettings(
				api.WithBaseURLV2(cfg.Server+"/api/v1"),
				api.WithToken(token),
				api.WithVerify(false),
			)
			client := api.NewClientV2(settings)

			fingerprint, caCertPEM, err := client.BootstrapConnect(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to bootstrap trust with master: %w", err)
			}

			fmt.Printf("Trusting new cluster: %s\n", clusterName)
			fmt.Printf("Master Root CA Fingerprint: %s\n", fingerprint)
			if !cfg.Insecure {
				if !utils.ConfirmPrompt("Do you trust this fingerprint?", false) {
					return fmt.Errorf("connection aborted by user: untrusted fingerprint")
				}
			}

			resp, err := client.GetAuthStatus(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get auth status: %w", err)
			}

			kf, err := pathutils.LoadGNS3KeysFile(keyFilePath)
			if err != nil {
				return fmt.Errorf("failed to load keys: %w", err)
			}

			if kf.CheckIfClusterExists(clusterName) {
				return fmt.Errorf("cluster with name %q already exists in key file", clusterName)
			}

			newCluster := pathutils.ClusterEntry{
				Name:            clusterName,
				RootFingerprint: fingerprint,
				CaCert:          string(caCertPEM),
				Master: pathutils.ServiceEntry{
					Type:        pathutils.TypeClusterMaster,
					URL:         cfg.Server,
					User:        resp.User,
					AccessToken: token,
					TokenType:   "Bearer",
				},
			}

			kf.Clusters = append(kf.Clusters, newCluster)
			if err := pathutils.SaveKeysFile(keyFilePath, kf); err != nil {
				return fmt.Errorf("failed to save key file: %w", err)
			}
			fmt.Printf("Cluster %q master added successfully\n", clusterName)

			fmt.Println("Discovering cluster nodes...")
			cfg.ClusterEntry = &newCluster
			if err := ctlhelpers.DiscoverAndSyncNodes(cmd.Context(), cfg); err != nil {
				return fmt.Errorf("cluster added, but node discovery failed: %w", err)
			}

			fmt.Printf("Successfully initialized cluster %q with all nodes.\n", clusterName)
			return nil
		},
	}
	cmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster to add")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
