package main

import (
	"ad_integration/config"
	repository "ad_integration/internal/auth/repository"
	"ad_integration/internal/auth/service"
	"ad_integration/internal/auth/service/ldap"
	"context"
	"fmt"

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
	Db, err := repository.NewConnection(ctx, cfg.DB)
	if err != nil {
		panic(err)
	}
	defer Db.Conn.Close(ctx)

	client, err := ldap.NewLDAPConnection(cfg.LDAP)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("connection is succeded")

	s := service.NewAuthService(client, Db)

	fmt.Println(login, password)

	LdapUser, err := s.Authenticate(context.Background(), login, password, cfg.LDAP.Attributes)
	if err != nil {
		fmt.Println("Fail to Fetch User's Details:", err)
		return
	}
	pp.Println(LdapUser)

	userID, err := s.UserRepository.SyncUser(ctx, LdapUser)
	if err != nil {
		fmt.Println("Ошибка при сихронизации user:", err)
		return
	}

	if err := s.UserRepository.SyncGroups(ctx, LdapUser.Groups); err != nil {
		fmt.Println("Ошибка при синхронизации групп:", err)
		return
	}

	if err := s.UserRepository.RefreshUserRoles(ctx, userID, LdapUser.Groups); err != nil {
		fmt.Println("Ошибка при обновление ролей пользователя:", err)
		return
	}
}
