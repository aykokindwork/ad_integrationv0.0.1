package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	LDAP LDAPConfig
	DB   DBConfig
}

type LDAPConfig struct {
	URL        string
	BaseDN     string
	UseTLS     bool
	ServerName string
	Attributes []string
}

type DBConfig struct {
	Address string
}

var defaultUse = map[string]string{
	"PROTOCOL":   "LDAP",
	"PORT":       "389",
	"BASEDN":     "DC=tp,DC=local",
	"USETLS":     "true",
	"SERVERNAME": "DC01.tp.local",
	"ATTRIBUTES": "userPrincipalName,memberOf,sAMAccountName",
}

func Load() (*Config, error) {
	protocol := getEnv("PROTOCOL")
	serverName := getEnv("SERVERNAME")
	port := getEnv("PORT")

	url := fmt.Sprintf("%s://%s:%s", protocol, serverName, port)

	baseDN := getEnv("BASEDN")

	useTLS, err := strconv.ParseBool(getEnv("USETLS"))
	if err != nil {
		return nil, fmt.Errorf("fail to convert string to int: %w", err)
	}

	attr := strings.Split(getEnv("ATTRIBUTES"), ",")

	address := os.Getenv("ADDRESS")

	return &Config{
		LDAP: LDAPConfig{
			URL:        url,
			BaseDN:     baseDN,
			UseTLS:     useTLS,
			ServerName: serverName,
			Attributes: attr,
		},
		DB: DBConfig{
			Address: address,
		},
	}, nil
}

func getEnv(s string) string {
	res := os.Getenv(s)

	if res == "" {
		res = defaultUse[s]
	}

	return res
}
