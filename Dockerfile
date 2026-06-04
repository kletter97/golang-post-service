# сборка приложения
FROM golang:1.25-alpine AS builder

WORKDIR /app

# копируем зависимости (для кэша)
COPY go.mod go.sum ./
RUN go mod download

# копируем остальное
COPY . .

# бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go


# финальный минимальный образ
FROM alpine:latest

WORKDIR /root/

# Копируем бинарник
COPY --from=builder /app/server .

EXPOSE 8090

CMD ["./server"]