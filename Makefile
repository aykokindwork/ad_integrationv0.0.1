include .env
export

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