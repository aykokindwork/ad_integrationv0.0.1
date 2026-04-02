package postgres

import (
	"ad_integration/internal/auth/model"
	"ad_integration/internal/auth/repository"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"

	"context"
)

type UserRepo struct {
	*DbConn
}

func NewUserRepo(conn *DbConn) repository.UserRepository {
	return &UserRepo{
		conn,
	}
}

func (u UserRepo) SyncUser(ctx context.Context, login string, email string) (model.User, error) {
	sqlQuery := `
	INSERT INTO users (login, email)
	VALUES ($1, $2)
	ON CONFLICT (login) DO UPDATE SET
		email = EXCLUDED.email,
	    updated_at = NOW()
	RETURNING id, login, email, created_at, updated_at;
	`
	var user model.User
	err := pgxscan.Get(ctx, u.Conn, &user, sqlQuery, login, email)
	if err != nil {
		return model.User{}, fmt.Errorf("fail to sync user: %w", err)
	}

	return user, nil

}

func (u UserRepo) SyncGroups(ctx context.Context, groups []string) error {
	sqlQuery := `
	INSERT into ldap_group_roles (ad_group_name)
	SELECT unnest($1::text[])
	ON CONFLICT (ad_group_name) WHERE role_id IS NULL DO NOTHING;
	`

	_, err := u.DbConn.Conn.Exec(ctx, sqlQuery, groups)
	if err != nil {
		return err
	}

	return nil
}

func (u UserRepo) RefreshUserRoles(ctx context.Context, userID int, groups []string) error {
	tx, err := u.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "DELETE FROM users_roles WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

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

func (u UserRepo) GetFullUserByID(ctx context.Context, id int) (model.User, error) {
	sqlQuery := `
	SELECT 
    u.id, 
    u.login, 
    u.email,
    u.created_at,
    u.updated_at,
    -- Собираем массив объектов Ролей
    COALESCE(
        (SELECT json_agg(role_data) FROM (
            SELECT 
                r.id, r.code, r.name, r.created_at, r.updated_at,
                -- Внутри каждой роли собираем массив объектов Прав
                COALESCE(
                    (SELECT json_agg(perm_data) FROM (
                        SELECT p.id, p.code, p.name, p.created_at, p.updated_at
                        FROM auth.permissions p
                        JOIN auth.permissions_roles pr ON p.id = pr.permission_id
                        WHERE pr.role_id = r.id
                    ) perm_data), '[]'
                ) as permissions
            FROM auth.roles r
            JOIN auth.users_roles ur ON r.id = ur.role_id
            WHERE ur.user_id = u.id
        ) role_data), '[]'
    ) as roles
	FROM auth.users u
	WHERE u.id = $1;
	`
	var user model.User
	err := pgxscan.Get(ctx, u.Conn, &user, sqlQuery, id)
	if err != nil {
		return model.User{}, fmt.Errorf("fail to scan: %w", err)
	}

	return user, nil
}
