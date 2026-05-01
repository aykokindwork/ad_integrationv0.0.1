include .env
export

APP_NAME=auth-app
DOCKER_COMPOSE=docker-compose.yaml

service-run:
	go run cmd\app\main.go

script-run:
	go run cmd\seed\main.go

migrate-up:
	migrate -path migrations -database "${ADDRESS}" up

migrate-down:
	migrate -path migrations -database "${ADDRESS}" down

migrate-version:
	migrate -path migrations -database "${ADDRESS}" version

migrate-fix:
	migrate -path migrations -database "${ADDRESS}" force $(v)

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

.PHONY: build up down restart logs clean

deploy: build up

build:
	docker-compose build $(APP_NAME)

up:
	docker-compose up -d

down:
	docker-compose down

stop:
	docker-compose stop $(APP_NAME)

redeploy: stop deploy

logs:
	docker-compose logs -f $(APP_NAME)

status:
	docker ps
