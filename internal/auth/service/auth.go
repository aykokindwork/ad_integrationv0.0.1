package service

import (
	"ad_integration/core/apperr"
	"ad_integration/internal/auth/model"
	"ad_integration/internal/infrasctructure/kafka"
	"context"
	"encoding/json"
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
	ldap      Ldaper
	userDB    Userer
	txManager TxManager
	kProducer *kafka.Producer
}

func NewAuthService(
	provider Ldaper,
	repository Userer,
	txManager TxManager,
	kProducer *kafka.Producer,
) *AuthService {
	return &AuthService{
		ldap:      provider,
		userDB:    repository,
		txManager: txManager,
		kProducer: kProducer,
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

	payload, _ := json.Marshal(map[string]string{
		"username": userLdap.Email,
		"event":    "LDAP_AUTH_SUCCESS",
		"status":   "AWAITING_OTP",
	})

	_ = s.kProducer.SendMessage(ctx, "auth-log", []byte(userLdap.Email), payload)

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
