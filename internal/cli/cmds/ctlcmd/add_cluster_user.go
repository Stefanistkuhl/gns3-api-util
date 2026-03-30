package ctlcmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/api"
)

type addUserRow struct {
	Cluster string `json:"cluster"`
	User    string `json:"user"`
	Status  string `json:"status"`
}

func (r *addUserRow) GetHeaders() []string { return []string{"CLUSTER", "USER", "STATUS"} }
func (r *addUserRow) GetRow() []string     { return []string{r.Cluster, r.User, r.Status} }

func NewAddClusterUserCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add-user [token]",
		Short: "Store a user's JWT token for a cluster in the keyfile",
		Long: `Store a user credential in the keyfile for a cluster so you can switch
between multiple users with --user.

The token is sent to the cluster master's /auth/status endpoint to verify it
and retrieve the username - no local JWT decoding needed. If [token] is omitted
you will be prompted to paste it.

Use 'ctl auth set-default-user' to change which user is the default.`,
		Args: cobra.MaximumNArgs(1),
		Example: `  gns3util ctl auth add-user eyJhbGci... -c mycluster
  gns3util ctl auth add-user -c mycluster   # prompts for token`,
		Annotations: map[string]string{"auth-mode": "cluster-only"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			var rawToken string
			if len(args) == 1 {
				rawToken = strings.TrimSpace(args[0])
			} else {
				if _, writeErr := fmt.Fprint(cmd.OutOrStdout(), "Paste JWT token: "); writeErr != nil {
					return fmt.Errorf("failed to write prompt: %w", writeErr)
				}
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					rawToken = strings.TrimSpace(scanner.Text())
				}
				if scanErr := scanner.Err(); scanErr != nil {
					return fmt.Errorf("failed to read token: %w", scanErr)
				}
			}

			if rawToken == "" {
				return fmt.Errorf("token cannot be empty")
			}

			masterURL := cfg.ClusterEntry.Master.URL
			caCert := []byte(cfg.ClusterEntry.CaCert)

			settings := api.NewSettings(
				api.WithBaseURLV2(masterURL+"/api/v1"),
				api.WithToken(rawToken),
				api.WithVerify(!cfg.Insecure),
				api.WithCA(caCert),
			)
			client := api.NewClientV2(&settings)

			resp, err := client.GetAuthStatus(cmd.Context())
			if err != nil {
				return fmt.Errorf("token rejected by master: %w", err)
			}
			if !resp.Authenticated {
				return fmt.Errorf("token is not authenticated (master returned authenticated=false)")
			}

			userID := resp.User
			if userID == "" {
				return fmt.Errorf("master returned an empty username")
			}

			keyFilePath, err := pathutils.ResolveKeyFilePath(cfg.KeyFile)
			if err != nil {
				return fmt.Errorf("failed to resolve key file path: %w", err)
			}

			kf, err := pathutils.LoadGNS3KeysFile(keyFilePath)
			if err != nil {
				return fmt.Errorf("failed to load key file: %w", err)
			}

			kf.UpsertClusterUser(cfg.Cluster, pathutils.ClusterUser{
				User:        userID,
				AccessToken: rawToken,
			})

			if saveErr := pathutils.SaveKeysFile(keyFilePath, kf); saveErr != nil {
				return fmt.Errorf("failed to save key file: %w", saveErr)
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}
			return printer.PrintObj([]utils.TableRecord{&addUserRow{
				Cluster: cfg.Cluster,
				User:    userID,
				Status:  "Added",
			}}, cmd.OutOrStdout())
		},
	}
}
