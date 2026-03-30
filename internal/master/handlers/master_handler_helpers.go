package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	sharedpb "github.com/0xveya/gns3util/internal/shared/pb"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/state/pb"
)

func roleToInfo(role *pb.Role) models.RoleInfo {
	info := models.RoleInfo{
		Name:        role.Name,
		Description: role.Description,
		Scopes:      make([]models.ScopeInfo, 0, len(role.Scopes)),
	}
	if role.CreatedAt != nil {
		info.CreatedAt = role.CreatedAt.AsTime().UTC().Format(time.RFC3339)
	}
	if role.UpdatedAt != nil {
		info.UpdatedAt = role.UpdatedAt.AsTime().UTC().Format(time.RFC3339)
	}
	for _, s := range role.Scopes {
		info.Scopes = append(info.Scopes, models.ScopeInfo{
			Action:   actionToDisplay(s.Action),
			Resource: resourceToDisplay(s.Resource),
		})
	}
	return info
}

func scopesToInfo(scopes []*pb.Scope) []models.ScopeInfo {
	infos := make([]models.ScopeInfo, 0, len(scopes))
	for _, s := range scopes {
		infos = append(infos, models.ScopeInfo{
			Action:   actionToDisplay(s.Action),
			Resource: resourceToDisplay(s.Resource),
		})
	}
	return infos
}

func actionToDisplay(a sharedpb.Action) string {
	s := strings.TrimPrefix(a.String(), "ACTION_")
	return strings.ToLower(s)
}

func resourceToDisplay(r sharedpb.Resource) string {
	if r == sharedpb.Resource_RESOURCE_UNSPECIFIED {
		return "*"
	}
	s := strings.TrimPrefix(r.String(), "RESOURCE_")
	return strings.ToLower(s)
}

func scopeInfosToProto(infos []models.ScopeInfo) ([]*pb.Scope, error) {
	scopes := make([]*pb.Scope, 0, len(infos))
	for _, si := range infos {
		lower := strings.ToLower(si.Action)
		if lower == "superuser" || (lower == "admin" && si.Resource == "") {
			scopes = append(scopes, &pb.Scope{
				Action:   sharedpb.Action_ACTION_ADMIN,
				Resource: sharedpb.Resource_RESOURCE_UNSPECIFIED,
			})
			continue
		}

		action := "ACTION_" + strings.ToUpper(si.Action)
		actionVal, ok := sharedpb.Action_value[action]
		if !ok {
			return nil, fmt.Errorf("unknown action %q", si.Action)
		}
		if si.Resource == "" || si.Resource == "*" {
			scopes = append(scopes, &pb.Scope{
				Action:   sharedpb.Action(actionVal),
				Resource: sharedpb.Resource_RESOURCE_UNSPECIFIED,
			})
			continue
		}
		resource := "RESOURCE_" + strings.ToUpper(si.Resource)
		resourceVal, ok := sharedpb.Resource_value[resource]
		if !ok {
			return nil, fmt.Errorf("unknown resource %q", si.Resource)
		}
		scopes = append(scopes, &pb.Scope{
			Action:   sharedpb.Action(actionVal),
			Resource: sharedpb.Resource(resourceVal),
		})
	}
	return scopes, nil
}

func (m *Master) HasEffectivePermission(
	ctx context.Context,
	userID string,
	action sharedpb.Action,
	resource sharedpb.Resource,
) (bool, error) {
	return m.Store.HasEffectivePermission(ctx, userID, action, resource)
}

func (m *Master) IsTokenRevoked(
	ctx context.Context,
	jti string,
) (bool, error) {
	return m.Store.IsTokenRevoked(ctx, jti)
}

func (m *Master) userInfoFromPermissions(ctx context.Context, perms *pb.UserPermissions) (models.UserInfo, error) {
	info := models.UserInfo{
		UserID:      perms.UserId,
		RoleNames:   perms.RoleNames,
		Permissions: []models.ScopeInfo{},
		DenyScopes:  scopesToInfo(perms.DenyScopes),
	}
	if perms.UpdatedAt != nil {
		info.UpdatedAt = perms.UpdatedAt.AsTime().UTC().Format(time.RFC3339)
	}

	scopes, err := m.Store.GetUserScopes(ctx, perms.UserId)
	if err != nil {
		return models.UserInfo{}, err
	}
	info.Permissions = scopesToInfo(scopes)
	return info, nil
}
