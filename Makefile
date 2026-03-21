include .env
export

service-run:
	go run cmd\app\main.go

script-run:
	go run cmd\seed\main.go

migrate-up:
	migrate -path migrations -database "${CONN_STRING}" up

migrate-down:
	migrate -path migrations -database "${CONN_STRING}" down

migrate-version:
	migrate -path migrations -database "${CONN_STRING}" version