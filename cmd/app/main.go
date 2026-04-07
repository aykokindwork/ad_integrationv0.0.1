package main

import (
	"ad_integration/config"
	"ad_integration/internal/auth/delivery/http"
	"ad_integration/internal/auth/repository/postgres"
	"ad_integration/internal/auth/service"
	"context"
	"fmt"
	"os"
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
	defer Db.Pool.Close()

	userRepo := postgres.NewUserRepo(Db)

	var client service.Ldaper
	if os.Getenv("APP_ENV") == "local" {
		client = &service.MockClient{}
	} else {
		realClient, err := service.NewLDAPConnection(cfg.LDAP)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer realClient.Conn.Close()

		client = realClient
	}

	txManager := postgres.NewTranscationManager(Db.Pool)
	s := service.NewAuthService(client, userRepo, txManager)

	/*
		login = os.Getenv("LOGIN")
		password = os.Getenv("PASSWD")

		userLdap, err := s.Authenticate(ctx, login, password)
		if err != nil {
			fmt.Println("Fail to authenticate:", err)
			return
		}
		pp.Println(userLdap)

		userID, err := s.Authorization(ctx, login, userLdap)
		if err != nil {
			fmt.Println(err)
			return
		}

		user, err := s.GetUserByID(ctx, userID)
		if err != nil {
			fmt.Println(err)
			return
		}

		pp.Println(user)*/
	handler := http.NewHandler(s)

	srv := handler.Init()

	if err := srv.Run(":8080"); err != nil {
		panic(err)
	}
}
