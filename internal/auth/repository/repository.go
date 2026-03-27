package repository

import (
	"ad_integration/internal/auth/model"
	"context"
)

//go:generate mockgen -package=repositorymocks -destination=./repositorymocks/mocks.go -source=repository.go *
type UserRepository interface {
	SyncUser(ctx context.Context, login, email string) (model.User, error)
	SyncGroups(ctx context.Context, groups []string) error
	RefreshUserRoles(ctx context.Context, userID int, groups []string) error
	GetFullUserByID(ctx context.Context, id int) (model.User, error)
}
