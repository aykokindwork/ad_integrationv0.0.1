package service

import (
	ldap "ad_integration/internal/service/ldap"
	"context"
	"fmt"
	"regexp"
	"strings"
)

var loginRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

type LDAPUser struct {
	Username string
	Email    string
	Groups   []string // Сюда положим названия групп из memberOf
}

type IdentityProvider interface {
	Search(ctx context.Context, login string, filter string, attributes []string) (*ldap.RawUser, error)
	BindUser(login string, password string) error
}
type AuthService struct {
	provider IdentityProvider // Нам плевать, LDAP это или база
}

func NewAuthService(provider IdentityProvider) *AuthService {
	return &AuthService{
		provider: provider,
	}
}

func (s *AuthService) Authenticate(ctx context.Context, login string, passwd string, attrs []string) (*LDAPUser, error) {

	filter := fmt.Sprintf("(sAMAccountName=%s)", login)

	if err := s.provider.BindUser(login, passwd); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	raw, err := s.provider.Search(ctx, login, filter, attrs)
	if err != nil {
		return nil, fmt.Errorf("user search failed: %w", err)
	}

	if len(raw.DN) == 0 {
		return nil, fmt.Errorf("no saved raw.DN")
	}
	if len(raw.Attributes) == 0 {
		return nil, fmt.Errorf("no saved raw.Attributes")
	}

	if len(raw.Attributes["userPrincipalName"]) == 0 {
		fmt.Println("no email")
	}

	user := &LDAPUser{
		Username: raw.Attributes["sAMAccountName"][0],
		Email:    raw.Attributes["userPrincipalName"][0],
		Groups:   raw.Attributes["memberOf"],
	}

	user.cleanAttributes()

	return user, nil
}

func isLoginValid(login string) bool {
	return loginRegex.MatchString(login)
}

func (u *LDAPUser) cleanAttributes() {
	// 1. Чистим Логин (делаем маленькими буквами для единообразия в БД)
	u.Username = strings.ToLower(u.Username)

	// 2. Чистим Почту (убираем лишние пробелы, если они есть)
	u.Email = strings.TrimSpace(strings.ToLower(u.Email))

	// 3. Чистим Группы
	var cleanGroups []string
	for _, dn := range u.Groups {
		parts := strings.Split(dn, ",")
		if len(parts) > 0 {
			// Отрезаем "CN=" и получаем чистое имя
			name := strings.TrimPrefix(parts[0], "CN=")
			cleanGroups = append(cleanGroups, name)
		}
	}
	u.Groups = cleanGroups
}
