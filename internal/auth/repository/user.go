package repository

import (
	auth "ad_integration/internal/auth/service"
	"context"
)

func (db DbConn) SyncUser(ctx context.Context, user *auth.LDAPUser) (int, error) {
	sqlQuery := `
	INSERT INTO users (login, email)
	VALUES ($1, $2)
	ON CONFLICT (login) DO UPDATE SET
		email = EXCLUDED.email
	RETURNING id;
	`

	var userID int
	err := db.Conn.QueryRow(ctx, sqlQuery, user.Username, user.Email).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil

}

func (db DbConn) SyncGroups(ctx context.Context, groups []string) error {
	sqlQuery := `
	INSERT into ldap_group_roles (ad_group_name)
	SELECT unnest($1::text[])
	ON CONFLICT (ad_group_name) WHERE role_id IS NULL DO NOTHING;
	`

	_, err := db.Conn.Exec(ctx, sqlQuery, groups)
	if err != nil {
		return err
	}

	return nil
}

func (db DbConn) RefreshUserRoles(ctx context.Context, userID int, groups []string) error {
	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Удаляем все текущие роли пользователя.
	// Это проще, чем высчитывать, какую роль добавить, а какую забрать.
	_, err = tx.Exec(ctx, "DELETE FROM users_roles WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	// 2. Магия: берем группы из LDAP, находим их ID в нашей таблице маппинга
	// и вставляем результат в таблицу связей users_roles.
	// Если группа из AD не привязана ни к какой роли, она просто проигнорируется.
	query := `
    INSERT INTO users_roles (user_id, role_id)
    SELECT $1, role_id 
    FROM ldap_group_roles 
    WHERE ad_group_name = ANY($2)
		AND role_id IS NOT NULL
    ON CONFLICT DO NOTHING;
    `
	// $1 — ID юзера, $2 — слайс []string (Go-массив групп)
	_, err = tx.Exec(ctx, query, userID, groups)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
