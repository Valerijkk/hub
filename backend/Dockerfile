# Dockerfile для Go бэкенда Hub
# Многоступенчатая сборка: builder + финальный минимальный образ

FROM golang:1.22-alpine AS builder

WORKDIR /app

# Зависимости (кешируем слой)
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o /hub-server ./cmd/server

# --- Финальный образ ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

# Для pg_restore и mongorestore (backup restore)
RUN apk add --no-cache postgresql16-client mongodb-tools

WORKDIR /app
COPY --from=builder /hub-server .

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -qO- http://localhost:3000/api/health || exit 1

ENTRYPOINT ["./hub-server"]
