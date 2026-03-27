package service

import (
	"ad_integration/config"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	ldap "github.com/go-ldap/ldap/v3"
)

const (
	substrADInvalidCreds = "data 52e"
	substrADUserBlocked  = "data 775"
)

type Client struct {
	Conn   *ldap.Conn
	config config.LDAPConfig
}

// connect with TLS LDAP
func NewLDAPConnection(cfg config.LDAPConfig) (*Client, error) {
	conn, err := ldap.DialURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("LDAP connection error: %w", err)
	}

	// TLS шифрование
	if cfg.UseTLS {
		err = conn.StartTLS(&tls.Config{
			InsecureSkipVerify: false,
			ServerName:         cfg.ServerName,
		})
		if err != nil {
			return nil, fmt.Errorf("TLS connection failed: %w", err)
		}
		_, ok := conn.TLSConnectionState()
		if !ok {
			return nil, fmt.Errorf("TLS connection failed after succesful start: %w", err)
		}
	}

	fmt.Println("everything is good, you have LDAP connection")

	return &Client{
		Conn:   conn,
		config: cfg,
	}, nil
}

func (c *Client) BindUser(login string, password string) error {
	err := c.Conn.Bind("tp\\"+login, password)
	if err != nil {
		var ldapErr *ldap.Error
		ok := errors.As(err, &ldapErr)
		if !ok {
			return err
		}

		if ldapErr.ResultCode == ldap.LDAPResultInvalidCredentials {
			errText := ldapErr.Error()

			switch {
			case strings.Contains(errText, substrADInvalidCreds):
				return errors.New("неверный логин или пароль")

			case strings.Contains(errText, substrADUserBlocked):
				return errors.New("аккаунт заблокирован")

			default:
				return errors.New("ошибка аутентификации")
			}
		}
	}
	return nil
}

func (c *Client) Search(ctx context.Context, filter string, attributes []string) (*RawUser, error) {

	searchRequest := ldap.NewSearchRequest(
		c.config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		filter,
		attributes,
		nil,
	)

	searchResult, err := c.Conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("fail to search: %w", err)
	}

	if len(searchResult.Entries) == 0 {
		return nil, fmt.Errorf("fail to find user in AD")
	}

	entry := searchResult.Entries[0]

	rawUser := &RawUser{
		DN:         entry.DN,
		Attributes: make(map[string][]string),
	}

	for _, attr := range attributes {
		rawUser.Attributes[attr] = entry.GetAttributeValues(attr)
	}

	return rawUser, nil

}

/* BASEDN for getting groups
func getBaseDN(l *ldap.Conn) (string, error) {
	req := ldap.NewSearchRequest(
		"",
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0, 0, false,
		"(objectClass=*)",
		[]string{"defaultNamingContext"},
		nil,
	)

	res, err := l.Search(req)
	if err != nil {
		return "", err
	}

	if len(res.Entries) == 0 {
		return "", fmt.Errorf("no RootDSE entries")
	}

	return res.Entries[0].GetAttributeValue("defaultNamingContext"), nil
}
*/
