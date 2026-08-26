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

# --- Hardened QA runtime ---
FROM nginx:1.27-alpine AS qa-runtime

RUN echo 'worker_processes auto; \
error_log /dev/stderr notice; \
events { worker_connections 1024; } \
http { \
    include /etc/nginx/mime.types; \
    default_type application/octet-stream; \
    access_log /dev/stdout; \
    sendfile on; \
    keepalive_timeout 65; \
    client_body_temp_path /tmp/client_temp; \
    proxy_temp_path /tmp/proxy_temp; \
    fastcgi_temp_path /tmp/fastcgi_temp; \
    uwsgi_temp_path /tmp/uwsgi_temp; \
    scgi_temp_path /tmp/scgi_temp; \
    server { \
        listen 8080; \
        root /usr/share/nginx/html; \
        index index.html; \
        location /api/ { proxy_pass http://backend:3000; } \
        location /ws { proxy_pass http://backend:3000; proxy_http_version 1.1; proxy_set_header Upgrade $http_upgrade; proxy_set_header Connection "upgrade"; proxy_set_header Host $host; proxy_read_timeout 86400s; proxy_send_timeout 86400s; } \
        location / { try_files $uri $uri/ /index.html; } \
    } \
}' > /etc/nginx/nginx.conf

COPY --from=builder /app/dist /usr/share/nginx/html

USER 101:101

ENTRYPOINT []

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/api/health || exit 1

CMD ["nginx", "-g", "daemon off; pid /tmp/nginx.pid;"]

# --- Финальный образ ---
FROM nginx:1.27-alpine AS production

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
