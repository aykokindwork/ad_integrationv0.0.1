package service

import (
	"ad_integration/internal/auth/model"
	"context"
)

//go:generate mockgen -package=servicemocks -destination=./service/mocks.go -source=service.go *
type UserData interface {
	GetDN() string
	GetAttr() map[string][]string
}

type IdentityProvider interface {
	Search(ctx context.Context, filter string, attributes []string) (*model.RawUser, error)
	BindUser(login string, password string) error
}

type UserRepository interface {
	SyncUser(ctx context.Context, login, email string) (model.User, error)
	SyncGroups(ctx context.Context, groups []string) error
	RefreshUserRoles(ctx context.Context, userID int, groups []string) error
	GetUserByID(ctx context.Context, userID int) (model.User, error)
}

type TxManager interface {
	WithInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
