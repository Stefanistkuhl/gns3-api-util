package ctlcmd

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	clusteraccess "github.com/0xveya/gns3util/internal/shared/cluster_access"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type TokenRequest struct {
	UserID string `json:"user_id"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

func NewCreateTokenCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "token [username]",
		Short: "Mint a new JWT for a user (Admin only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetUser := args[0]
			configPath = viper.GetString("config")

			configData, readErr := os.ReadFile(configPath) //#nosec G304
			if readErr != nil {
				return fmt.Errorf("failed to read access config: %w", readErr)
			}

			var accessConfig clusteraccess.ClientConfig
			if unmarshallErr := toml.Unmarshal(configData, &accessConfig); unmarshallErr != nil {
				return fmt.Errorf("failed to parse access config: %w", unmarshallErr)
			}

			cert, parseCertSErr := tls.X509KeyPair([]byte(accessConfig.UserCert), []byte(accessConfig.UserKey))
			if parseCertSErr != nil {
				return fmt.Errorf("failed to load admin client cert/key: %w", parseCertSErr)
			}

			caPool := x509.NewCertPool()
			caPool.AppendCertsFromPEM([]byte(accessConfig.CACert))

			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						Certificates: []tls.Certificate{cert},
						RootCAs:      caPool,
					},
				},
			}

			reqBody := map[string]string{"user_id": targetUser}
			jsonBody, _ := json.Marshal(reqBody)

			url := fmt.Sprintf("%s/auth/token", accessConfig.ServerURL)

			req, reqErr := http.NewRequestWithContext(cmd.Context(), http.MethodPost, url, bytes.NewBuffer(jsonBody))
			if reqErr != nil {
				return fmt.Errorf("failed to create request: %w", reqErr)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, resqErr := client.Do(req)
			if resqErr != nil {
				return fmt.Errorf("failed to reach master node: %w", resqErr)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				cleanMsg := strings.TrimSpace(string(body))
				return fmt.Errorf("failed to generate token (status %d): %s", resp.StatusCode, cleanMsg)
			}

			var tokenResp TokenResponse
			if decodeErr := json.NewDecoder(resp.Body).Decode(&tokenResp); decodeErr != nil {
				return fmt.Errorf("failed to decode token response: %w", decodeErr)
			}

			fmt.Printf("Token for user '%s' generated successfully:\n\n%s\n", targetUser, tokenResp.Token)
			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "/data/master/tls/cluster_access.toml", "Path to cluster_access.toml")
	_ = viper.BindPFlag("config", cmd.Flags().Lookup("config"))

	return cmd
}
