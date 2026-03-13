package ctlcmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	clusteraccess "github.com/0xveya/gns3util/internal/shared/cluster_access"
	"github.com/0xveya/gns3util/pkg/state/pb"
	scopesPkg "github.com/0xveya/gns3util/pkg/web/scopes"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewCreateUserCmd() *cobra.Command {
	var (
		etcdEndpoint string
		scopesStr    string
	)

	cmd := &cobra.Command{
		Use:   "user [username]",
		Short: "Create or update a user's permissions in etcd",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			rawScopes := strings.Split(scopesStr, ",")
			var cleanScopes []string
			for _, s := range rawScopes {
				trimmed := strings.TrimSpace(s)
				if trimmed != "" {
					if !scopesPkg.IsValid(trimmed) {
						return fmt.Errorf("invalid scope: %s", trimmed)
					}
					cleanScopes = append(cleanScopes, trimmed)
				}
			}

			userPerms := &pb.UserPermissions{
				UserId:    targetUser,
				Scopes:    cleanScopes,
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

			fmt.Printf("Successfully created/updated user '%s' with scopes: %s\n", targetUser, strings.Join(cleanScopes, ","))
			return nil
		},
	}

	cmd.Flags().StringP("config", "", "", "Path to cluster_access.toml (Env: GNS3_CONFIG)")
	cmd.Flags().StringVar(&etcdEndpoint, "etcd", "localhost:2379", "Etcd endpoint")
	cmd.Flags().StringVar(&scopesStr, "scopes", "read,write", "Comma-separated list of scopes")

	_ = viper.BindPFlag("config", cmd.Flags().Lookup("config"))

	return cmd
}
