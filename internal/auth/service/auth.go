package service

import (
	"ad_integration/core/apperr"
	"ad_integration/internal/auth/model"
	"context"
	"fmt"
	"regexp"

	"github.com/go-ldap/ldap/v3"
)

const (
	attrSAMAccountName    = "sAMAccountName"
	attrUserPrincipalName = "userPrincipalName"
	attrMemberOf          = "memberOf"

	userSearchFilterTmpl = "(" + attrSAMAccountName + "=%s)"
)

var adUserAttributes = []string{
	attrSAMAccountName,
	attrUserPrincipalName,
	attrMemberOf,
}

var loginRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

type AuthService struct {
	ldap      Ldaper
	userDB    Userer
	txManager TxManager
	events    AuthEventPublisher
}

func NewAuthService(
	provider Ldaper,
	repository Userer,
	txManager TxManager,
	events AuthEventPublisher,
) *AuthService {
	return &AuthService{
		ldap:      provider,
		userDB:    repository,
		txManager: txManager,
		events:    events,
	}
}

func (s *AuthService) Authenticate(
	ctx context.Context,
	login string,
	passwd string,
) (*model.LDAPUser, error) {

	if err := s.ldap.BindUser(login, passwd); err != nil {
		return nil, apperr.ErrLdapBind.WithErr(err)
	}

	filter := fmt.Sprintf(userSearchFilterTmpl, ldap.EscapeFilter(login))
	raw, err := s.ldap.Search(ctx, filter, adUserAttributes)
	if err != nil {
		return nil, apperr.ErrLdapSearch.WithErr(err)
	}

	userLdap, err := mapLDAPUser(raw.Attributes)
	if err != nil {
		return nil, apperr.ErrLdapMap.WithErr(err)
	}

	if err := s.events.PublishLDAPSuccess(ctx, userLdap.Email); err != nil {
		_ = err
	}

	return userLdap, nil
}

func (s *AuthService) Authorization(
	ctx context.Context,
	login string,
	userLdap *model.LDAPUser,
) (int, error) {
	var userID int

	err := s.txManager.WithInTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.userDB.SyncUser(txCtx, login, userLdap.Email)
		if err != nil {
			return err
		}

		if err := s.userDB.SyncGroups(txCtx, userLdap.Groups); err != nil {
			return err
		}

		if err := s.userDB.RefreshUserRoles(txCtx, user.ID, userLdap.Groups); err != nil {
			return err
		}
		userID = user.ID

		return nil
	})
	if err != nil {
		return 0, err
	}

	return userID, nil

}

func (s *AuthService) GetUserByID(ctx context.Context, userID int) (model.User, error) {

	user, err := s.userDB.GetUserByID(ctx, userID)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}
