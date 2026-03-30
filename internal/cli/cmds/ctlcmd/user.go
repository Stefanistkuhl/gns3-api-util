package ctlcmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	clusteraccess "github.com/0xveya/gns3util/internal/shared/cluster_access"
	"github.com/0xveya/gns3util/pkg/state/pb"
	"github.com/0xveya/gns3util/pkg/utils/globals"
)

type UserCreateResult struct {
	User    string   `json:"user"`
	Roles   []string `json:"roles"`
	EtcdKey string   `json:"etcd_key"`
	Status  string   `json:"status"`
}

func (u UserCreateResult) GetHeaders() []string {
	return []string{"USER", "ROLES", "ETCD KEY", "STATUS"}
}

func (u UserCreateResult) GetRow() []string {
	return []string{
		u.User,
		strings.Join(u.Roles, ", "),
		u.EtcdKey,
		u.Status,
	}
}

func NewCreateUserCmd() *cobra.Command {
	var (
		etcdEndpoint string
		rolesStr     string
		adminFlag    bool
	)

	cmd := &cobra.Command{
		Use:   "user [username]",
		Short: "Bootstrap a user's role assignments directly in etcd",
		Long: `Writes a UserPermissions record directly to etcd to bootstrap access for a user.

Use --admin to grant full admin access. This assigns the built-in "admin" role,
which is the only role that bypasses role-object lookups in the permission checker
and is required for commands like add-cluster (which calls /auth/status).

Use --roles for any other role names, but note those roles must already exist as
Role objects in etcd (created separately) or the permission checks will silently
fail at runtime.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			configPath, _ := cmd.Flags().GetString("config")
			if configPath == "" {
				configPath = viper.GetString("config")
			}

			if configPath == "" {
				return fmt.Errorf("config path is empty; please provide --config or set GNS3_CONFIG")
			}
			targetUser := args[0]

			configData, err := os.ReadFile(configPath) // #nosec G304
			if err != nil {
				return fmt.Errorf("failed to read access config: %w", err)
			}

			var accessConfig clusteraccess.ClientConfig
			if tomlErr := toml.Unmarshal(configData, &accessConfig); tomlErr != nil {
				return fmt.Errorf("failed to parse access config: %w", tomlErr)
			}

			cert, err := tls.X509KeyPair([]byte(accessConfig.UserCert), []byte(accessConfig.UserKey))
			if err != nil {
				return fmt.Errorf("failed to load client cert/key: %w", err)
			}

			caPool := x509.NewCertPool()
			if ok := caPool.AppendCertsFromPEM([]byte(accessConfig.CACert)); !ok {
				return fmt.Errorf("failed to append CA cert")
			}

			cli, err := clientv3.New(clientv3.Config{
				Endpoints:   []string{etcdEndpoint},
				DialTimeout: 5 * time.Second,
				TLS: &tls.Config{
					Certificates: []tls.Certificate{cert},
					RootCAs:      caPool,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to connect to etcd: %w", err)
			}
			defer cli.Close()

			var cleanRoles []string
			if adminFlag {
				cleanRoles = []string{globals.RoleAdmin}
			} else {
				if rolesStr == "" {
					return fmt.Errorf("specify --roles <role,...> or use --admin to grant the built-in admin role")
				}
				for s := range strings.SplitSeq(rolesStr, ",") {
					if trimmed := strings.TrimSpace(s); trimmed != "" {
						cleanRoles = append(cleanRoles, trimmed)
					}
				}
			}

			userPerms := &pb.UserPermissions{
				UserId:    targetUser,
				RoleNames: cleanRoles,
				UpdatedAt: timestamppb.Now(),
			}

			data, err := proto.Marshal(userPerms)
			if err != nil {
				return fmt.Errorf("failed to marshal proto: %w", err)
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()

			key := fmt.Sprintf("/gns3/v1/auth/scopes/%s", targetUser)
			_, err = cli.Put(ctx, key, string(data))
			if err != nil {
				return fmt.Errorf("failed to write to etcd: %w", err)
			}

			result := UserCreateResult{
				User:    targetUser,
				Roles:   cleanRoles,
				EtcdKey: key,
				Status:  "Created/Updated",
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}

			return printer.PrintObj(result, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringP("config", "", "", "Path to cluster_access.toml (Env: GNS3_CONFIG)")
	cmd.Flags().StringVar(&etcdEndpoint, "etcd", "localhost:2379", "Etcd endpoint")
	cmd.Flags().StringVar(&rolesStr, "roles", "", "Comma-separated list of role names to assign (roles must exist in etcd)")
	cmd.Flags().BoolVar(&adminFlag, "admin", false, `Grant the built-in "admin" role (bypasses role-object lookup; required for add-cluster and other admin-gated commands)`)

	_ = viper.BindPFlag("config", cmd.Flags().Lookup("config"))

	return cmd
}
