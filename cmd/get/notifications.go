package get

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api"
	"github.com/stefanistkuhl/gns3util/pkg/api/endpoints"
	"github.com/stefanistkuhl/gns3util/pkg/authentication"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/tidwall/pretty"
)

func NewGetNotificationsCmd() *cobra.Command {
	var timeout int = 5
	var cmd = &cobra.Command{
		Use:     "notifications",
		Short:   "Stream the notification of the controller",
		Long:    `Stream the notification of the controller`,
		Example: "gns3util -s https://controller:3080 get notifications",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to get global options: %v\n", err)
				return
			}

			token, err := authentication.GetKeyForServer(cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to get authentication token: %v\n", err)
				return
			}

			settings := api.NewSettings(
				api.WithBaseURL(cfg.Server),
				api.WithVerify(cfg.Insecure),
				api.WithToken(token),
				api.WithTimeout(time.Duration(timeout)*time.Second),
			)

			ep := endpoints.GetEndpoints{}
			client := api.NewGNS3Client(settings)
			reqOpts := api.NewRequestOptions(settings).
				WithURL(ep.Notifications()).
				WithMethod(api.GET).
				WithStream()

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "request failed: %v\n", err)
				return
			}
			if resp == nil {
				fmt.Fprintln(os.Stderr, "no response received")
				return
			}
			defer resp.Body.Close()

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
				if err == context.Canceled || err == context.DeadlineExceeded {
					return
				}
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				fmt.Fprintf(os.Stderr, "error reading stream: %v\n", err)
			}
		},
	}
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 5, "Timeout in seconds (0 for stream until cancellation)")
	return cmd
}

func NewGetProjectNotificationCmd() *cobra.Command {
	var timeout int = 5
	var cmd = &cobra.Command{
		Use:     "notifications [project-name/id]",
		Short:   "Stream the notification of a project by id or name",
		Long:    `Stream the notification of a project by id or name`,
		Example: "gns3util -s https://controller:3080 project notifications my-project",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to get global options: %v\n", err)
				return
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			token, err := authentication.GetKeyForServer(cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to get authentication token: %v\n", err)
				return
			}

			settings := api.NewSettings(
				api.WithBaseURL(cfg.Server),
				api.WithVerify(cfg.Insecure),
				api.WithToken(token),
				api.WithTimeout(time.Duration(timeout)*time.Second),
			)

			ep := endpoints.GetEndpoints{}
			client := api.NewGNS3Client(settings)
			reqOpts := api.NewRequestOptions(settings).
				WithURL(ep.ProjectNotifications(id)).
				WithMethod(api.GET).
				WithStream()

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "request failed: %v\n", err)
				return
			}
			if resp == nil {
				fmt.Fprintln(os.Stderr, "no response received")
				return
			}
			defer resp.Body.Close()

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
				if err == context.Canceled || err == context.DeadlineExceeded {
					return
				}
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				fmt.Fprintf(os.Stderr, "error reading stream: %v\n", err)
			}
		},
	}
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 5, "Timeout in seconds (0 for stream until cancellation)")
	return cmd
}
