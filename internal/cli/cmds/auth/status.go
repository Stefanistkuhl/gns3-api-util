package auth

import (
	"encoding/json"
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/authentication"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathUtils"
	"github.com/0xveya/gns3util/pkg/api/schemas"
	"github.com/spf13/cobra"
)

func NewAuthStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check the current status of your Authentication",
		Long:  `Check the current status of your Authentication`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var user schemas.User

			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			keyFilePath, err := pathUtils.ResolveKeyFilePath(cfg.KeyFile)
			if err != nil {
				return fmt.Errorf("failed to resolve key file path: %w", err)
			}

			kf, err := pathUtils.LoadGNS3KeysFile(keyFilePath)
			if err != nil {
				return fmt.Errorf("failed to load keys: %w", err)
			}

			userData, err := authentication.TryKeys(kf, cfg)
			if err != nil {
				return err
			}

			err = json.Unmarshal([]byte(userData), &user)
			if err != nil {
				return fmt.Errorf("error unmarshaling JSON: %w", err)
			}
			fmt.Printf("%s logged in as user %s \n", messageUtils.SuccessMsg("Logged in as user"), messageUtils.Bold(*user.Username))
			return nil
		},
	}
	return cmd
}
