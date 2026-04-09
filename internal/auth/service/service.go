package service

import (
	"ad_integration/internal/auth/model"
	"context"
)

//go:generate mockgen -package=servicemocks -destination=./service/mocks.go -source=service.go *

type Author interface {
	Authenticate(ctx context.Context, login string, passwd string) (*model.LDAPUser, error)
	Authorization(ctx context.Context, login string, userLdap *model.LDAPUser) (int, error)
	GetUserByID(ctx context.Context, userID int) (model.User, error)
}

type Ldaper interface {
	Search(ctx context.Context, filter string, attributes []string) (*model.RawUser, error)
	BindUser(login string, password string) error
}

type MockLdaper interface {
	Search(ctx context.Context, filter string, attributes []string) (*model.RawUser, error)
	BindUser(login string, password string) error
}

type Userer interface {
	SyncUser(ctx context.Context, login, email string) (model.User, error)
	SyncGroups(ctx context.Context, groups []string) error
	RefreshUserRoles(ctx context.Context, userID int, groups []string) error
	GetUserByID(ctx context.Context, userID int) (model.User, error)
}

type EventProducer interface {
	SendMessage(ctx context.Context, topic string, key, value []byte) error
}

type AuthEventPublisher interface {
	PublishLDAPSuccess(ctx context.Context, email string) error
	PublishOTPSuccess(ctx context.Context, email string) error
	PublishAuthFailed(ctx context.Context, email string, reason string) error
}

type TxManager interface {
	WithInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
