package auth

import (
	"fmt"
	"strings"

	ldap "github.com/go-ldap/ldap/v3"
)

type LDAPUser struct {
	Username string
	Email    string
	Groups   []string // Сюда положим названия групп из memberOf
}

// authorization
// ldap.bind
func (a *LDAPAuth) AuthUser(login string, password string) error {
	//authentification
	err := a.Conn.Bind("tp\\"+login, password)
	if err != nil {
		return fmt.Errorf("Fail to Bind", err)
	}
	return nil
}

// Достает атрибуты, которые нам нужны
func (a *LDAPAuth) FetchUserDetails(login string) (*LDAPUser, error) {
	// 1. Создаем поисковый запрос

	cleanLogin := login
	if strings.Contains(login, "\\") {
		parts := strings.Split(login, "\\")
		cleanLogin = parts[len(parts)-1] // Берем последнюю часть (самого юзера)
	}

	escapedLogin := ldap.EscapeFilter(cleanLogin)

	searchRequest := ldap.NewSearchRequest(
		a.config.BaseDN,        // "dc=tp,dc=local" — где ищем
		ldap.ScopeWholeSubtree, // Ищем во всех вложенных папках (OU)
		ldap.NeverDerefAliases,
		0, 0, false,
		fmt.Sprintf("(sAMAccountName=%s)", escapedLogin), // Наш фильтр
		[]string{"sAMAccountName", "mail", "memberOf"},   // Атрибуты, которые хотим забрать
		nil,
	)

	// 2. Выполняем поиск через сохраненное соединение a.conn
	sr, err := a.Conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("ldap search fatal: %w", err)
	}

	// 3. Проверяем результат
	if len(sr.Entries) == 0 {
		return nil, fmt.Errorf("user not found in AD")
	}

	// Берем первую найденную запись (т.к. логин уникален)
	entry := sr.Entries[0]

	// 4. Мапим данные в нашу структуру из User.go
	userDetails := &LDAPUser{
		Username: entry.GetAttributeValue("sAMAccountName"),
		Email:    entry.GetAttributeValue("mail"),
		Groups:   entry.GetAttributeValues("memberOf"), // Возвращает []string
	}

	userDetails.cleanAttributes()

	return userDetails, nil
}

// CleanAttributes приводит все данные пользователя в порядок
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
