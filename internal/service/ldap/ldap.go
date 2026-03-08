package ldap

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	ldap "github.com/go-ldap/ldap/v3"
)

type RawUser struct {
	DN         string
	Attributes map[string][]string
}

type Config struct {
	URL        string
	BaseDN     string // Например, "dc=example,dc=com"
	UseTLS     bool
	Attributes []string // Список того, что хотим забирать (memberOf, mail и т.д.)
	ServerName string
}

func LoadConfig(url string, baseDN string, UseTLS bool, Attributes []string, serverName string) Config {
	cfg := Config{
		URL:        url,
		BaseDN:     baseDN,
		UseTLS:     UseTLS,
		Attributes: Attributes,
		ServerName: serverName,
	}

	return cfg
}

type Client struct {
	Conn   *ldap.Conn
	config Config
}

// connect with TLS LDAP
func NewLDAPConnection(cfg Config) (*Client, error) {
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
	//authentification
	err := c.Conn.Bind("tp\\"+login, password)
	if err != nil {
		ldapErr := parseLDAPError(err)
		return ldapErr
	}
	return nil
}

func (c *Client) Search(ctx context.Context, login string, filter string, attributes []string) (*RawUser, error) {

	searchRequest := ldap.NewSearchRequest(
		c.config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		filter, //передавать только готовый фильтр
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

func parseLDAPError(err error) error {
	ldapErr, ok := err.(*ldap.Error)
	if !ok {
		return err
	}

	if ldapErr.ResultCode == ldap.LDAPResultInvalidCredentials {
		errText := ldapErr.Error()

		switch {
		case strings.Contains(errText, "data 52e"):
			return errors.New("неверный логин или пароль")

		case strings.Contains(errText, "data 775"):
			return errors.New("аккаунт заблокирован")

		default:
			return errors.New("ошибка аутентификации")
		}
	}

	return err
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
