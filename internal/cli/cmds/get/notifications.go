package get

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/authentication"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/pkg/api"
	"github.com/0xveya/gns3util/pkg/api/endpoints"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
)

func NewGetNotificationsCmd() *cobra.Command {
	timeout := 5
	cmd := &cobra.Command{
		Use:     "notifications",
		Aliases: []string{"notif", "n"},
		Short:   "Stream the notification of the controller",
		Long:    `Stream the notification of the controller`,
		Example: "gns3util -s https://controller:3080 get notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			token, err := authentication.GetKeyForServer(cfg)
			if err != nil {
				return fmt.Errorf("failed to get authentication token: %w", err)
			}

			settings := api.NewSettings(
				api.WithBaseURL(cfg.Server),
				api.WithVerify(cfg.Insecure),
				api.WithToken(token),
				api.WithTimeout(time.Duration(timeout)*time.Second),
			)

			ep := endpoints.GetEndpoints{}
			client := api.NewGNS3Client(&settings)
			reqOpts := api.NewRequestOptions(&settings).
				WithURL(ep.Notifications()).
				WithMethod(api.GET).
				WithStream()

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			if resp == nil {
				return fmt.Errorf("no response received")
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					// Keep silent on body close errors
					_ = err
				}
			}()

			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := scanner.Bytes()
				if len(line) == 0 {
					continue
				}

				var js json.RawMessage
				if err := json.Unmarshal(line, &js); err != nil {
					fmt.Fprintf(os.Stderr, "invalid JSON: %v\n", err)
					continue
				}

				formatted := pretty.Pretty(line)
				formatted = pretty.Color(formatted, nil)
				fmt.Println(string(formatted))
			}

			if err := scanner.Err(); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return nil
				}
				if strings.Contains(err.Error(), "use of closed network connection") {
					return nil
				}
				return fmt.Errorf("error reading stream: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 5, "Timeout in seconds (0 for stream until cancellation)")
	return cmd
}

func NewGetProjectNotificationCmd() *cobra.Command {
	timeout := 5
	cmd := &cobra.Command{
		Use:     "notifications [project-name/id]",
		Short:   "Stream the notification of a project by id or name",
		Long:    `Stream the notification of a project by id or name`,
		Example: "gns3util -s https://controller:3080 project notifications my-project",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					return err
				}
			}

			token, err := authentication.GetKeyForServer(cfg)
			if err != nil {
				return fmt.Errorf("failed to get authentication token: %w", err)
			}

			settings := api.NewSettings(
				api.WithBaseURL(cfg.Server),
				api.WithVerify(cfg.Insecure),
				api.WithToken(token),
				api.WithTimeout(time.Duration(timeout)*time.Second),
			)

			ep := endpoints.GetEndpoints{}
			client := api.NewGNS3Client(&settings)
			reqOpts := api.NewRequestOptions(&settings).
				WithURL(ep.ProjectNotifications(id)).
				WithMethod(api.GET).
				WithStream()

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			if resp == nil {
				return fmt.Errorf("no response received")
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					// Keep silent on body close errors
					_ = err
				}
			}()

			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := scanner.Bytes()
				if len(line) == 0 {
					continue
				}

				var js json.RawMessage
				if err := json.Unmarshal(line, &js); err != nil {
					fmt.Fprintf(os.Stderr, "invalid JSON: %v\n", err)
					continue
				}

				formatted := pretty.Pretty(line)
				formatted = pretty.Color(formatted, nil)
				fmt.Println(string(formatted))
			}

			if err := scanner.Err(); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return nil
				}
				if strings.Contains(err.Error(), "use of closed network connection") {
					return nil
				}
				return fmt.Errorf("error reading stream: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 5, "Timeout in seconds (0 for stream until cancellation)")
	return cmd
}
