package main

import (
	auth "ad_integration/internal/auth"
	"ad_integration/internal/db"
	"context"
	"fmt"
	"os"

	"github.com/k0kubun/pp"
)

var login string
var password string

func main() {
	url := os.Getenv("URL")
	BaseDN := os.Getenv("BASEDN")
	Attributes := []string{
		"cn",
		"memberOf",       // Тут будет лежать твоя lab-test-admins
		"sAMAccountName", // Тут будет лежать lab-admin
	}
	UseTLS := true

	cfg := auth.LoadConfig(url, BaseDN, UseTLS, Attributes)

	ldapConnection, err := auth.NewLDAPConnection(cfg)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	/*fmt.Println("Введите логин, пароль")
	if _, err := fmt.Scanln(&login); err != nil {
		panic(err)
	}

	if _, err := fmt.Scanln(&password); err != nil {
		panic(err)
	}
	fmt.Println(login, password)*/

	login = os.Getenv("LOGIN")
	password = os.Getenv("PASS")
	fmt.Println(login, password)

	err = ldapConnection.AuthUser(login, password)
	if err != nil {
		fmt.Println("Доступ запрещен:", err)
		return
	}
	fmt.Println("Succesfull bind")

	/*testGroupsFromAD := []string{
		"CN=lab-test-admins,OU=Groups,DC=tp,DC=local",
		"CN=all-staff,OU=Global,DC=tp,DC=local",
		"CN=vpn-users,OU=Access,DC=tp,DC=local",
	}
	*/

	userDetails, err := ldapConnection.FetchUserDetails(login)
	if err != nil {
		fmt.Println("Fail to Fetch User's Details:", err)
		return
	}
	pp.Println(userDetails)

	fmt.Println("username:", userDetails.Username)
	fmt.Println("user_email:", userDetails.Email)
	fmt.Println("user_group:", userDetails.Groups)

	ctx := context.Background()
	addressDB := os.Getenv("CONN_STRING")

	conn, err := db.Connection(ctx, addressDB)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	fmt.Println("connection is succeded")

	userID, err := db.SyncUser(ctx, conn, userDetails)
	if err != nil {
		fmt.Println("Ошибка при сихронизации user:", err)
		return
	}

	if err := db.SyncGroups(ctx, conn, userDetails.Groups); err != nil {
		fmt.Println("Ошибка при синхронизации групп:", err)
		return
	}

	if err := db.RefreshUserRoles(ctx, conn, userID, userDetails.Groups); err != nil {
		fmt.Println("Ошибка при обновление ролей пользователя:", err)
		return
	}
}
