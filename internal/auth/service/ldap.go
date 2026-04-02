package service

import (
	"ad_integration/config"
	"ad_integration/core/apperr"
	"ad_integration/internal/auth/model"
	"context"
	"crypto/tls"
	"errors"
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
		return nil, apperr.ErrLdapUnexpected.WithErr(err)
	}

	// TLS шифрование
	if cfg.UseTLS {
		err = conn.StartTLS(&tls.Config{
			InsecureSkipVerify: false,
			ServerName:         cfg.ServerName,
		})
		if err != nil {
			return nil, apperr.ErrLdapTLS.WithErr(err)
		}
	}

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
				return apperr.ErrInvalidCredentials

			case strings.Contains(errText, substrADUserBlocked):
				return apperr.ErrUserIsBlocked

			default:
				return apperr.ErrLdapUnexpected
			}
		}
	}
	return nil
}

func (c *Client) Search(ctx context.Context, filter string, attributes []string) (*model.RawUser, error) {

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
		return nil, err
	}

	if len(searchResult.Entries) == 0 {
		return nil, apperr.ErrAdUserNotFound
	}

	entry := searchResult.Entries[0]
	rawUser := &model.RawUser{
		DN:         entry.DN,
		Attributes: make(map[string][]string),
	}

	if len(rawUser.DN) == 0 {
		return nil, apperr.ErrLdapNoDN
	}

	if len(rawUser.Attributes) == 0 {
		return nil, apperr.ErrLdapNoAttributes
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
