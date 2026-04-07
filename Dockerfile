FROM golang:1.26-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o auth-app ./cmd/app/main.go

FROM alpine:latest

WORKDIR /root/

# Копируем только скомпилированный файл из первого этапа
COPY --from=builder /app/auth-app .


#COPY --from=builder /app/.env .

CMD ["./auth-app"]