package update

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewUpdateACECmd() *cobra.Command {
	var (
		aceType   string
		path      string
		propagate bool
		allow     bool
		userID    string
		groupID   string
		roleID    string
		useJSON   string
	)

	cmd := &cobra.Command{
		Use:     utils.UpdateSingleElementCmdName + " [ace-id]",
		Short:   "Update an ACE",
		Long:    "Update an ACE. User, group and role IDs will be resolved from names if UUIDs are not provided.",
		Example: "gns3util -s https://controller:3080 acl update ace-id --ace-type user --path /some/endpoint --role-id some-role",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			aceID := args[0]
			if !utils.IsValidUUIDv4(aceID) {
				return fmt.Errorf("Please use a valid UUIDv4 for the ACE-ID")
			}

			if err := validateChoice(aceType, []string{"user", "group"}, "--ace-type"); err != nil {
				return err
			}

			if userID != "" && !utils.IsValidUUIDv4(userID) {
				resolvedID, err := utils.ResolveID(cfg, "user", userID, nil)
				if err != nil {
					return fmt.Errorf("failed to resolve user ID: %w", err)
				}
				userID = resolvedID
			}

			if groupID != "" && !utils.IsValidUUIDv4(groupID) {
				resolvedID, err := utils.ResolveID(cfg, "group", groupID, nil)
				if err != nil {
					return fmt.Errorf("failed to resolve group ID: %w", err)
				}
				groupID = resolvedID
			}

			if roleID != "" && !utils.IsValidUUIDv4(roleID) {
				resolvedID, err := utils.ResolveID(cfg, "role", roleID, nil)
				if err != nil {
					return fmt.Errorf("failed to resolve role ID: %w", err)
				}
				roleID = resolvedID
			}

			var payload map[string]any
			if useJSON == "" {
				if aceType == "" || path == "" {
					return fmt.Errorf("for this command --ace-type and --path are required or provide --use-json")
				}

				if aceType == "user" && userID == "" {
					return fmt.Errorf("when ace-type is 'user', --user-id is required")
				}
				if aceType == "group" && groupID == "" {
					return fmt.Errorf("when ace-type is 'group', --group-id is required")
				}
				if roleID == "" {
					return fmt.Errorf("--role-id is required for all ACE types")
				}

				data := schemas.ACEUpdate{}
				if aceType != "" {
					data.ACEType = &aceType
				}
				if path != "" {
					data.Path = &path
				}
				if propagate {
					data.Propagate = &propagate
				}
				if allow {
					data.Allowed = &allow
				}
				if userID != "" {
					data.UserID = &userID
				}
				if groupID != "" {
					data.GroupID = &groupID
				}
				if roleID != "" {
					data.RoleID = &roleID
				}
				b, err := json.Marshal(data)
				if err != nil {
					return fmt.Errorf("failed to encode request: %w", err)
				}
				if err := json.Unmarshal(b, &payload); err != nil {
					return fmt.Errorf("failed to prepare payload: %w", err)
				}
			} else {
				if err := json.Unmarshal([]byte(useJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "updateACE", []string{aceID}, payload)
			return nil
		},
	}

	cmd.Flags().StringVar(&aceType, "ace-type", "", "Desired type for the ACE (user/group)")
	cmd.Flags().StringVar(&path, "path", "", "Desired path for the ace to affect")
	cmd.Flags().BoolVar(&propagate, "propagate", true, "Apply ACE rules to all nested endpoints in the path")
	cmd.Flags().BoolVar(&allow, "allow", true, "Whether to allow or deny access to the set path")
	cmd.Flags().StringVar(&userID, "user-id", "", "Desired user ID to use for this ACE (name or UUID)")
	cmd.Flags().StringVar(&groupID, "group-id", "", "Desired group ID to use for this ACE (name or UUID)")
	cmd.Flags().StringVar(&roleID, "role-id", "", "Desired role ID to use for this ACE (name or UUID)")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")

	return cmd
}

func validateChoice(val string, allowed []string, flag string) error {
	if val == "" {
		return nil
	}
	if slices.Contains(allowed, val) {
		return nil
	}
	return fmt.Errorf("invalid value %q for %s (allowed: %v)", val, flag, allowed)
}
