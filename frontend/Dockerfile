# Dockerfile для React фронтенда Hub
# Многоступенчатая сборка: Node builder + Nginx для раздачи статики

FROM node:22-alpine AS builder

WORKDIR /app

# Зависимости (кешируем слой)
COPY package.json package-lock.json ./
RUN npm ci

# Копируем исходники и собираем
COPY . .
RUN npm run build

# --- Финальный образ ---
FROM nginx:1.27-alpine

# Конфигурация Nginx для SPA (все маршруты -> index.html)
RUN echo 'server { \
    listen 80; \
    root /usr/share/nginx/html; \
    index index.html; \
    location /api/ { proxy_pass http://backend:3000; } \
    location /ws { proxy_pass http://backend:3000; proxy_http_version 1.1; proxy_set_header Upgrade $http_upgrade; proxy_set_header Connection "upgrade"; proxy_set_header Host $host; proxy_read_timeout 86400s; proxy_send_timeout 86400s; } \
    location / { try_files $uri $uri/ /index.html; } \
}' > /etc/nginx/conf.d/default.conf

COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:80/ || exit 1
