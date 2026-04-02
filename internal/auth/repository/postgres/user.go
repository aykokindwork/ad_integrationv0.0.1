package postgres

import (
	"ad_integration/core/apperr"
	"ad_integration/internal/auth/model"
	"ad_integration/internal/auth/repository"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/pgxpool"

	"context"
)

type UserRepo struct {
	*DbConn
}

func NewUserRepo(pool *DbConn) repository.UserRepository {
	return &UserRepo{
		pool,
	}
}

func (u *UserRepo) getExecutor(ctx context.Context) Querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}

	return u.Pool
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

	executor := u.getExecutor(ctx)

	err := pgxscan.Get(ctx, executor, &user, sqlQuery, login, email)
	if err != nil {
		return model.User{}, apperr.ErrSyncUser.WithErr(err)
	}

	return user, nil

}

func (u UserRepo) SyncGroups(ctx context.Context, groups []string) error {
	sqlQuery := `
	INSERT into ldap_group_roles (ad_group_name)
	SELECT unnest($1::text[])
	ON CONFLICT (ad_group_name) WHERE role_id IS NULL DO NOTHING;
	`

	executor := u.getExecutor(ctx)

	_, err := executor.Exec(ctx, sqlQuery, groups)
	if err != nil {
		return err
	}

	return nil
}

func (u UserRepo) RefreshUserRoles(ctx context.Context, userID int, groups []string) error {

	executor := u.getExecutor(ctx)

	_, err := executor.Exec(ctx, "DELETE FROM users_roles WHERE user_id = $1", userID)
	if err != nil {
		return apperr.ErrRefreshUserRoles.WithErr(err)
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
	_, err = executor.Exec(ctx, query, userID, groups)
	if err != nil {
		return apperr.ErrRefreshUserRoles.WithErr(err)
	}

	return nil
}

func (u UserRepo) GetUserByID(ctx context.Context, id int) (model.User, error) {
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

	executor := u.getExecutor(ctx)

	var user model.User
	err := pgxscan.Get(ctx, executor, &user, sqlQuery, id)
	if err != nil {
		return model.User{}, apperr.ErrGetFullUserByID.WithErr(err)
	}

	return user, nil
}
