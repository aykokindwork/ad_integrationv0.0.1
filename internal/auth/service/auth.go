package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

type RawUser struct {
	DN         string
	Attributes map[string][]string
}

const (
	attrSAMAccountName    = "sAMAccountName"
	attrUserPrincipalName = "userPrincipalName"
	attrMemberOf          = "memberOf"

	userSearchFilterTmpl = "(" + attrSAMAccountName + "=%s)"

	substrADInvalidCreds = "data 52e"
	substrADUserBlocked  = "data 775"
)

var adUserAttributes = []string{
	attrSAMAccountName,
	attrUserPrincipalName,
	attrMemberOf,
}

var loginRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

type LDAPUser struct {
	Username string
	Email    string
	Groups   []string // Сюда положим названия групп из memberOf
}

type UserData interface {
	GetDN() string
	GetAttr() map[string][]string
}

type IdentityProvider interface {
	Search(ctx context.Context, login string, filter string, attributes []string) (*RawUser, error)
	BindUser(login string, password string) error
}

type UserRepository interface {
	SyncUser(ctx context.Context, user *LDAPUser) (int, error)
	SyncGroups(ctx context.Context, groups []string) error
	RefreshUserRoles(ctx context.Context, userID int, groups []string) error
}
type AuthService struct {
	Provider       IdentityProvider
	UserRepository UserRepository
}

func NewAuthService(provider IdentityProvider, repository UserRepository) *AuthService {
	return &AuthService{
		Provider:       provider,
		UserRepository: repository,
	}
}

func (s *AuthService) Authenticate(
	ctx context.Context,
	login string,
	passwd string,
	attrs []string) (*LDAPUser, error) {

	if err := s.Provider.BindUser(login, passwd); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	filter := fmt.Sprintf(userSearchFilterTmpl, ldap.EscapeFilter(login))
	raw, err := s.Provider.Search(ctx, login, filter, adUserAttributes)
	if err != nil {
		return nil, fmt.Errorf("user search failed: %w", err)
	}

	if len(raw.DN) == 0 {
		return nil, fmt.Errorf("no saved raw.DN")
	}
	if len(raw.Attributes) == 0 {
		return nil, fmt.Errorf("no saved raw.Attributes")
	}

	if raw.Attributes[attrUserPrincipalName][0] == "" {
		return nil, fmt.Errorf("user has no email")
	}

	if raw.Attributes[attrSAMAccountName][0] == "" {
		return nil, fmt.Errorf("user has no name")
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

	user := &LDAPUser{
		Username: username,
		Email:    email,
		Groups:   groups,
	}

	return user, nil
}
