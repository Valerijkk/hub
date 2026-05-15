# Dockerfile для Go бэкенда Hub
# Многоступенчатая сборка: builder + финальный минимальный образ

FROM golang:1.24-alpine AS builder

# Позволяем Go автоматически скачать нужную версию toolchain
ENV GOTOOLCHAIN=auto

WORKDIR /app

# Зависимости (кешируем слой)
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Генерируем Ent ORM код
RUN go generate ./ent

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o /hub-server ./cmd/server

# --- Финальный образ ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

# Для pg_restore и mongorestore (backup restore)
RUN apk add --no-cache postgresql16-client mongodb-tools

# LO-Inf: запускаем под непривилегированным пользователем (не root).
RUN addgroup -g 10001 -S hub && adduser -u 10001 -S -G hub hub

WORKDIR /app
COPY --from=builder /hub-server .
RUN chown -R hub:hub /app

USER hub:hub

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:3000/api/health || exit 1

ENTRYPOINT ["./hub-server"]
