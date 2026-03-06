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
	scopesPkg "github.com/0xveya/gns3util/pkg/web/scopes"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func NewCreateUserCmd() *cobra.Command {
	var (
		configPath   string
		etcdEndpoint string
		scopesStr    string
	)

	cmd := &cobra.Command{
		Use:   "user [username]",
		Short: "Create or update a user's permissions in etcd",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath = viper.GetString("config")
			targetUser := args[0]

			configData, readErr := os.ReadFile(configPath) //#nosec G304
			if readErr != nil {
				return fmt.Errorf("failed to read access config: %w", readErr)
			}

			var accessConfig clusteraccess.ClientConfig
			if tomlUnmarshallErr := toml.Unmarshal(configData, &accessConfig); tomlUnmarshallErr != nil {
				return fmt.Errorf("failed to parse access config: %w", tomlUnmarshallErr)
			}

			cert, certErr := tls.X509KeyPair([]byte(accessConfig.UserCert), []byte(accessConfig.UserKey))
			if certErr != nil {
				return fmt.Errorf("failed to load admin client cert/key: %w", certErr)
			}

			caPool := x509.NewCertPool()
			if ok := caPool.AppendCertsFromPEM([]byte(accessConfig.CACert)); !ok {
				return fmt.Errorf("failed to append CA cert to pool")
			}

			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
				RootCAs:      caPool,
			}

			cli, cliErr := clientv3.New(clientv3.Config{
				Endpoints:   []string{etcdEndpoint},
				DialTimeout: 5 * time.Second,
				TLS:         tlsConfig,
			})
			if cliErr != nil {
				return fmt.Errorf("failed to connect to etcd: %w", cliErr)
			}
			defer cli.Close()

			scopes := strings.Split(scopesStr, ",")
			cleanScopes := []string{}
			for _, s := range scopes {
				trimmed := strings.TrimSpace(s)
				if trimmed != "" {
					if !scopesPkg.IsValid(trimmed) {
						return fmt.Errorf("invalid scope provided: %s", trimmed)
					}
					cleanScopes = append(cleanScopes, trimmed)
				}
			}
			val := strings.Join(cleanScopes, ",")

			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()

			key := fmt.Sprintf("/auth/scopes/%s", targetUser)
			_, putErr := cli.Put(ctx, key, val)
			if putErr != nil {
				return fmt.Errorf("failed to write permissions to etcd: %w", putErr)
			}

			fmt.Printf("Successfully set permissions for user '%s' with scopes: [%s]\n", targetUser, val)
			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "/data/master/tls/cluster_access.toml", "Path to cluster_access.toml containing root creds (env GNS3_CONFIG)")
	cmd.Flags().StringVar(&etcdEndpoint, "etcd", "localhost:2379", "Etcd endpoint")
	cmd.Flags().StringVar(&scopesStr, "scopes", "read,write", "Comma-separated list of scopes")

	_ = viper.BindPFlag("config", cmd.Flags().Lookup("config"))

	return cmd
}
