package model

type RawUser struct {
	DN         string
	Attributes map[string][]string
}

type LDAPUser struct {
	Username string
	Email    string
	Groups   []string
}
