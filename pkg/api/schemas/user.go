package schemas

import (
	"github.com/google/uuid"
)

type User struct {
	Username     *string   `json:"username"`
	IsActive     bool      `json:"is_active"`
	Email        *string   `json:"email"`
	FullName     *string   `json:"full_name"`
	CreatedAt    *string   `json:"created_at"`
	UpdatedAt    *string   `json:"updated_at"`
	UserID       uuid.UUID `json:"user_id"`
	LastLogin    *string   `json:"last_login"`
	IsSuperadmin bool      `json:"is_superadmin"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Token struct {
	AccessToken *string `json:"access_token"`
	TokenType   *string `json:"token_type"`
}

type UserCreate struct {
	Username *string `json:"username"`
	IsActive bool    `json:"is_active"`
	Email    *string `json:"email"`
	FullName *string `json:"full_name"`
	Password *string `json:"password"`
}

type UserUpdate struct {
	Username *string `json:"username,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
	Email    *string `json:"email,omitempty"`
	FullName *string `json:"full_name,omitempty"`
	Password *string `json:"password,omitempty"`
}

type LoggedInUserUpdate struct {
	Password *string `json:"password,omitempty"`
	Email    *string `json:"email,omitempty"`
	FullName *string `json:"full_name,omitempty"`
}
