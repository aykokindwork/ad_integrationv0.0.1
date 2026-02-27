package main

import (
	"ad_integration/internal/repository"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func fullcreate(
	ctx context.Context,
	tx pgx.Tx,
	permCode string,
	permName string,
	roleCode string,
	roleName string,
	group string) error {
	if err := createPermission(ctx, tx, permCode, permName); err != nil {
		fmt.Println("ошибка при создании permission", err)
		return err
	}

	if err := createRole(ctx, tx, roleCode, roleName); err != nil {
		fmt.Println("ошибка при создании роли", err)
		return err
	}

	if err := createRolePermission(ctx, tx, roleCode, permCode); err != nil {
		fmt.Println("ошибка при связе роль-permission:", err)
		return err
	}

	if err := syncGroups(ctx, tx, group, roleCode); err != nil {
		fmt.Println("fail to sync group", err)
		return err
	}

	err := tx.Commit(ctx)
	if err != nil {
		fmt.Println(" fail to commit:", err)
		return err
	}

	return nil

}

func createPermission(ctx context.Context, tx pgx.Tx, permCode string, permName string) error {
	//создание прав
	sqlQueryForPermission := `
	INSERT INTO permissions (code, name)
	VALUES ($1, $2)
	ON CONFLICT (code) DO NOTHING
	`
	if _, err := tx.Exec(ctx, sqlQueryForPermission, permCode, permName); err != nil {
		return err
	}

	return nil
}

func createRole(ctx context.Context, tx pgx.Tx, roleCode string, roleName string) error {

	sqlQueryForRole := `
	INSERT INTO roles (code, name)
	VALUES ($1, $2)
	ON CONFLICT (code) DO NOTHING
	`

	if _, err := tx.Exec(ctx, sqlQueryForRole, roleCode, roleName); err != nil {
		return err
	}

	return nil
}

func createRolePermission(ctx context.Context, tx pgx.Tx, roleCode string, permCode string) error {
	sqlQueryForPermRoles := `
	INSERT INTO permissions_roles (permission_id, role_id)
	SELECT r.id, p.id
	FROM roles r, permissions p
	WHERE r.code = $1 AND p.code= $2
	ON CONFLICT DO NOTHING
	`

	if _, err := tx.Exec(ctx, sqlQueryForPermRoles, roleCode, permCode); err != nil {
		return err
	}

	return nil
}

func syncGroups(ctx context.Context, tx pgx.Tx, group string, roleCode string) error {
	updateQuery := `
    DELETE from ldap_group_roles 
    WHERE ad_group_name = $1 AND role_id IS NULL;
	`

	if _, err := tx.Exec(ctx, updateQuery, group); err != nil {
		fmt.Println("ошибка именно при удаление группы")
		return err
	}

	insertQuery := `
    INSERT INTO ldap_group_roles (ad_group_name, role_id)
    SELECT $1, id FROM roles WHERE code = $2
    ON CONFLICT (ad_group_name, role_id) DO NOTHING;
`
	if _, err := tx.Exec(ctx, insertQuery, group, roleCode); err != nil {
		fmt.Println("Ошибка при связи группы с ролью", err)
		return err
	}

	return nil
}

func main() {
	ctx, cancelCTX := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCTX()

	addressDB := os.Getenv("CONN_STRING")
	conn, err := repository.Connection(ctx, addressDB)
	if err != nil {
		fmt.Println("Fail to connect to DB: ", err)
		return
	}
	defer conn.Close(ctx)

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		fmt.Println("Fail to make tx:", err)
		return
	}
	defer tx.Rollback(ctx)

	permName := "Разрабатывать"
	permCode := "io.develope"
	roleCode := "developer"
	roleName := "Разработчик"
	group := "Developer"

	if err := fullcreate(ctx, tx, permCode, permName, roleCode, roleName, group); err != nil {
		fmt.Println("Fail:", err)
		return
	}

	fmt.Println("all good")

}
