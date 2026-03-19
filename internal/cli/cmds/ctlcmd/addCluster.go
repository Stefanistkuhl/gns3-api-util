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

type ClusterAddResult struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	User        string `json:"user"`
	Fingerprint string `json:"fingerprint"`
	Status      string `json:"status"`
}

func (c *ClusterAddResult) GetHeaders() []string {
	return []string{"CLUSTER NAME", "URL", "USER", "FINGERPRINT", "STATUS"}
}

func (c *ClusterAddResult) GetRow() []string {
	return []string{
		c.Name,
		c.URL,
		c.User,
		c.Fingerprint,
		c.Status,
	}
}

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
			client := api.NewClientV2(&settings)

			fmt.Fprintf(cmd.ErrOrStderr(), "Bootstrapping trust with master...\n")

			fingerprint, caCertPEM, err := client.BootstrapConnect(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to bootstrap trust with master: %w", err)
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "Trusting new cluster: %s\n", clusterName)
			fmt.Fprintf(cmd.ErrOrStderr(), "Master Root CA Fingerprint: %s\n", fingerprint)

			if !cfg.Insecure {
				// Ensure utils.ConfirmPrompt writes to stderr natively if possible
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
			if saveErr := pathutils.SaveKeysFile(keyFilePath, kf); saveErr != nil {
				return fmt.Errorf("failed to save key file: %w", saveErr)
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "Discovering cluster nodes...\n")
			cfg.ClusterEntry = &newCluster
			if syncErr := ctlhelpers.DiscoverAndSyncNodes(cmd.Context(), cfg); syncErr != nil {
				return fmt.Errorf("cluster added, but node discovery failed: %w", syncErr)
			}

			result := ClusterAddResult{
				Name:        clusterName,
				URL:         cfg.Server,
				User:        resp.User,
				Fingerprint: fingerprint,
				Status:      "Discovered & Synced",
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}

			return printer.PrintObj(&result, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the cluster to add")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
