package create

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewCreateACLCmd() *cobra.Command {
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
		Use:     utils.CreateSingleElementCmdName,
		Short:   "Create an ACE rule",
		Long:    "Create an Access Control Entry. If IDs are not UUIDv4, names will be resolved where possible.",
		Example: "gns3util -s https://controller:3080 acl create --ace-type allow --path /projects --role-id my-role",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			var payload map[string]any
			if useJSON == "" {
				if aceType == "" || path == "" || roleID == "" {
					return fmt.Errorf("for this command --ace-type, --path and --role-id are required or provide --use-json")
				}
				// resolve non-UUID ids
				if groupID != "" && !utils.IsValidUUIDv4(groupID) {
					id, err := utils.ResolveID(cfg, "group", groupID, nil)
					if err != nil {
						return err
					}
					groupID = id
				}
				if userID != "" && !utils.IsValidUUIDv4(userID) {
					id, err := utils.ResolveID(cfg, "user", userID, nil)
					if err != nil {
						return err
					}
					userID = id
				}
				if !utils.IsValidUUIDv4(roleID) {
					id, err := utils.ResolveID(cfg, "role", roleID, nil)
					if err != nil {
						return err
					}
					roleID = id
				}
				data := schemas.ACECreate{}
				if aceType != "" {
					v := aceType
					data.ACEType = &v
				}
				if path != "" {
					v := path
					data.Path = &v
				}
				if propagate {
					v := true
					data.Propagate = &v
				}
				if allow {
					v := true
					data.Allowed = &v
				}
				if userID != "" {
					v := userID
					data.UserID = &v
				}
				if groupID != "" {
					v := groupID
					data.GroupID = &v
				}
				if roleID != "" {
					// keep as string; backend expects UUID string; schemas.ACECreate uses uuid.UUID for role_id, but marshaling a string there requires conversion
					// Instead, pass raw payload map for role_id as string
					b, _ := json.Marshal(data)
					_ = json.Unmarshal(b, &payload)
					payload["role_id"] = roleID
				} else {
					b, _ := json.Marshal(data)
					_ = json.Unmarshal(b, &payload)
				}
			} else {
				if err := json.Unmarshal([]byte(useJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "createACL", nil, payload)
			return nil
		},
	}
	cmd.Flags().StringVar(&aceType, "ace-type", "", "ACE type: user or group")
	cmd.Flags().StringVarP(&path, "path", "p", "", "Path affected by the ACE (e.g., /pools/<id>)")
	cmd.Flags().BoolVar(&propagate, "propagate", true, "Apply ACE to nested resources")
	cmd.Flags().BoolVarP(&allow, "allow", "a", true, "Allow (true) or deny (false)")
	cmd.Flags().StringVarP(&userID, "user-id", "u", "", "User ID to apply the ACE to")
	cmd.Flags().StringVarP(&groupID, "group-id", "g", "", "Group ID to apply the ACE to")
	cmd.Flags().StringVarP(&roleID, "role-id", "r", "", "Role ID for the ACE")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")
	return cmd
}
