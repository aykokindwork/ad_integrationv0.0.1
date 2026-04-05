package repository

import (
	"ad_integration/internal/auth/model"
	"context"
)

//go:generate mockgen -package=repositorymocks -destination=./repositorymocks/mocks.go -source=repository.go *
type Userer interface {
	SyncUser(ctx context.Context, login, email string) (model.User, error)
	SyncGroups(ctx context.Context, groups []string) error
	RefreshUserRoles(ctx context.Context, userID int, groups []string) error
	GetUserByID(ctx context.Context, userID int) (model.User, error)
}

type TransactionManager interface {
	WithInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
