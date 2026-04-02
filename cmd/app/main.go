package main

import (
	"ad_integration/config"
	"ad_integration/internal/auth/repository/postgres"
	"ad_integration/internal/auth/service"
	"context"
	"fmt"
	"os"

	"github.com/k0kubun/pp"
)

var login string
var password string

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("fail to load config")
	}

	ctx := context.Background()
	Db, err := postgres.NewConnection(ctx, cfg.DB)
	if err != nil {
		panic(err)
	}
	defer Db.Conn.Close(ctx)

	userRepo := postgres.NewUserRepo(Db)

	client, err := service.NewLDAPConnection(cfg.LDAP)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer client.Conn.Close()

	s := service.NewAuthService(client, userRepo)

	login = os.Getenv("LOGIN")
	password = os.Getenv("PASSWD")
	LdapUser, err := s.Authenticate(ctx, login, password)
	if err != nil {
		fmt.Println("Fail to authenticate:", err)
		return
	}
	pp.Println(LdapUser)

	user, err := s.UserRepository.SyncUser(ctx, login, LdapUser.Email)
	if err != nil {
		fmt.Println("Ошибка при сихронизации user:", err)
		return
	}

	if err := s.UserRepository.SyncGroups(ctx, LdapUser.Groups); err != nil {
		fmt.Println("Ошибка при синхронизации групп:", err)
		return
	}

	if err := s.UserRepository.RefreshUserRoles(ctx, user.ID, LdapUser.Groups); err != nil {
		fmt.Println("Ошибка при обновление ролей пользователя:", err)
		return
	}
}
