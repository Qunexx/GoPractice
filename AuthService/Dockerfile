FROM ubuntu:latest
LABEL authors="Qunexx"

ENTRYPOINT ["top", "-b"]
# Используем официальный образ Go как базовый для сборки нашего приложения
FROM golang:1.21.6 as builder

WORKDIR /AuthService

# Копируем файлы go.mod и go.sum и скачиваем зависимости
COPY go.mod .
COPY go.sum .
RUN go mod download

# Копируем исходный код сервиса в контейнер
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o AuthService ./main/main.go

# Используем образ alpine для запуска собранного приложения
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем скомпилированный бинарник и файл config.yaml в рабочую директорию
COPY --from=builder /AuthService .
COPY --from=builder /AuthService/config.yaml .

# Задаём переменные окружения
ENV JWT_Token=fgrejegoiergjio
ENV LOCAL_SERVER_PORT=:8080

# Команда для запуска приложения
CMD ["./AuthService"]
