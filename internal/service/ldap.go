package service

import (
	"crypto/tls"
	"fmt"

	ldap "github.com/go-ldap/ldap/v3"
)

type Config struct {
	URL        string
	BaseDN     string // Например, "dc=example,dc=com"
	UseTLS     bool
	Attributes []string // Список того, что хотим забирать (memberOf, mail и т.д.)
}

func LoadConfig(url string, baseDN string, UseTLS bool, Attributes []string) Config {
	cfg := Config{
		URL:        url,
		BaseDN:     baseDN,
		UseTLS:     UseTLS,
		Attributes: Attributes,
	}

	return cfg
}

type LDAPAuth struct {
	Conn   *ldap.Conn
	config Config
}

// connect with TLS LDAP
func NewLDAPConnection(cfg Config) (*LDAPAuth, error) {
	conn, err := ldap.DialURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("LDAP connection error: ", err)
	}

	// TLS шифрование
	if cfg.UseTLS {
		err = conn.StartTLS(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return nil, fmt.Errorf("TLS connection failed:", err)
		}
		_, ok := conn.TLSConnectionState()
		if !ok {
			return nil, fmt.Errorf("TLS connection failed after succesful start:", err)
		}
	}

	fmt.Println("everything is good, you have LDAP connection")

	return &LDAPAuth{
		Conn:   conn,
		config: cfg,
	}, nil
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
