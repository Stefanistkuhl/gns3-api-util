package master

import (
	"context"
	"time"

	"github.com/0xveya/gns3util/pkg/state"
	"github.com/0xveya/gns3util/pkg/web/auth"
)

type Master struct {
	IDMgr *auth.IdentityManager
	Store *state.StateManager
}

func (m *Master) GrantAccess(userID string, role string, scopes []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.Store.PutUserPermissions(ctx, userID, scopes)
	if err != nil {
		return "", err
	}

	claims := auth.NewClaims(userID, role, scopes, 24*time.Hour)
	return m.IDMgr.Mint(claims)
}
