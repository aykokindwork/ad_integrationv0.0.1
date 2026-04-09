package service

import (
	"ad_integration/core/apperr"
	"ad_integration/internal/auth/model"
	"strings"
)

type ldapRawAttrs map[string][]string

func mapLDAPUser(attrs ldapRawAttrs) (*model.LDAPUser, error) {
	upn, ok := attrs[attrUserPrincipalName]
	if !ok || len(upn) == 0 || upn[0] == "" {
		return nil, apperr.ErrLdapNoEmail
	}

	sam, ok := attrs[attrSAMAccountName]
	if !ok || len(sam) == 0 || sam[0] == "" {
		return nil, apperr.ErrLdapNoName
	}

	username := strings.ToLower(sam[0])
	email := strings.TrimSpace(strings.ToLower(upn[0]))

	groups := exctractGroups(attrs[attrMemberOf])

	return &model.LDAPUser{
		Username: username,
		Email:    email,
		Groups:   groups,
	}, nil
}

func exctractGroups(dns []string) []string {

	var cleanGroups []string
	for _, dn := range dns {
		parts := strings.Split(dn, ",")
		if len(parts) > 0 {
			name := strings.TrimPrefix(parts[0], "CN=")
			cleanGroups = append(cleanGroups, name)
		}
	}
	groups := cleanGroups

	return groups
}
