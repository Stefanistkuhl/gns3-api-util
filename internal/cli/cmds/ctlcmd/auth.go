package ctlcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/0xveya/gns3util/pkg/api"
)

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication related operations",
		Long:  `Authentication related operations change this text later`,
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
		Short: "Check your authentication status",
		Long:  `Check your authentication status and store the token if correct`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			envErr := viper.BindEnv("token", "TOKEN")
			if envErr != nil {
				return fmt.Errorf("failed to bind token environment variable: %w", envErr)
			}
			token := viper.GetString("token")
			settings := api.NewSettings(
				api.WithBaseURLV2(cfg.Server+"/api/v1"),
				api.WithToken(token),
				api.WithVerify(!cfg.Insecure),
			)
			client := api.NewClientV2(settings)

			resp, err := client.GetAuthStatus(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get auth status: %w", err)
			}
			fmt.Printf("%s authenticated as user: %s with scopes: %s\n", messageUtils.SuccessMsg("Success:"), messageUtils.Bold(resp.User), messageUtils.Bold(resp.Scopes))
			return nil
		},
	}

	return cmd
}
