package service

import "context"

type UserData interface {
	GetDN() string
	GetAttr() map[string][]string
}

type IdentityProvider interface {
	Search(ctx context.Context, filter string, attributes []string) (*RawUser, error)
	BindUser(login string, password string) error
}

type UserRepository interface {
	SyncUser(ctx context.Context, user *LDAPUser) (int, error)
	SyncGroups(ctx context.Context, groups []string) error
	RefreshUserRoles(ctx context.Context, userID int, groups []string) error
}
