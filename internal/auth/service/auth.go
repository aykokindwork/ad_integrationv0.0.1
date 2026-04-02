package service

import (
	"ad_integration/core/apperr"
	"ad_integration/internal/auth/model"
	"context"
	"fmt"
	"regexp"
	"strings"

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
	provider       IdentityProvider
	userRepository UserRepository
	txManager      TxManager
}

func NewAuthService(provider IdentityProvider, repository UserRepository, txManager TxManager) *AuthService {
	return &AuthService{
		provider:       provider,
		userRepository: repository,
		txManager:      txManager,
	}
}

func (s *AuthService) Authenticate(
	ctx context.Context,
	login string,
	passwd string,
) (*model.LDAPUser, error) {

	if err := s.provider.BindUser(login, passwd); err != nil {
		return nil, apperr.ErrLdapBind.WithErr(err)
	}

	filter := fmt.Sprintf(userSearchFilterTmpl, ldap.EscapeFilter(login)) //go change ldap, in layer http this can be checked
	raw, err := s.provider.Search(ctx, filter, adUserAttributes)
	if err != nil {
		return nil, apperr.ErrLdapSearch.WithErr(err)
	}

	val, ok := raw.Attributes[attrUserPrincipalName]
	if !ok || len(val) == 0 || val[0] == "" {
		return nil, apperr.ErrLdapNoEmail
	}

	val, ok = raw.Attributes[attrSAMAccountName]
	if !ok || len(val) == 0 || val[0] == "" {
		return nil, apperr.ErrLdapNoName
	}

	username := strings.ToLower(raw.Attributes[attrSAMAccountName][0])
	email := strings.TrimSpace(strings.ToLower(raw.Attributes[attrUserPrincipalName][0]))

	var cleanGroups []string
	for _, dn := range raw.Attributes[attrMemberOf] {
		parts := strings.Split(dn, ",")
		if len(parts) > 0 {
			name := strings.TrimPrefix(parts[0], "CN=")
			cleanGroups = append(cleanGroups, name)
		}
	}
	groups := cleanGroups

	userLdap := &model.LDAPUser{
		Username: username,
		Email:    email,
		Groups:   groups,
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
		user, err := s.userRepository.SyncUser(txCtx, login, userLdap.Email)
		if err != nil {
			return err
		}

		if err := s.userRepository.SyncGroups(txCtx, userLdap.Groups); err != nil {
			return err
		}

		if err := s.userRepository.RefreshUserRoles(txCtx, user.ID, userLdap.Groups); err != nil {
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

	user, err := s.userRepository.GetUserByID(ctx, userID)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}
