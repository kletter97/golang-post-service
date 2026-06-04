# Этап 1: Сборка приложения
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Сначала копируем зависимости (так Docker будет кэшировать слои и собирать быстрее)
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной исходный код
COPY . .

# Собираем статический бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go


# Этап 2: Финальный минимальный образ
FROM alpine:latest

WORKDIR /root/

# Копируем скомпилированный файл из предыдущего этапа
COPY --from=builder /app/server .

# Указываем порт (в твоем логе был 8090)
EXPOSE 8090

# Запускаем бинарник
CMD ["./server"]