# ============================================================
# Stage 1: Build
# ============================================================
FROM node:20-alpine AS builder

WORKDIR /app

# Сначала копируем только манифесты для кеширования слоя зависимостей
COPY package.json package-lock.json* ./

# npm ci быстрее и предсказуемее, чем npm install
RUN npm ci --no-audit --prefer-offline

# Теперь копируем исходники
COPY . .

# Production-сборка через Vite
RUN npm run build

# ============================================================
# Stage 2: Serve via Nginx
# ============================================================
FROM nginx:alpine AS runner

# Удаляем дефолтную страницу
RUN rm -rf /usr/share/nginx/html/*

# Шаблон конфига: официальный nginx-образ сам прогонит его через envsubst
# при старте контейнера и положит результат в /etc/nginx/conf.d/default.conf.
# NGINX_ENVSUBST_FILTER ограничивает подстановку только переменной
# GATEWAY_HOST, чтобы envsubst не затронул встроенные $host, $remote_addr
# и другие nginx-переменные внутри конфига.
ENV NGINX_ENVSUBST_FILTER=^GATEWAY_HOST$
COPY nginx.conf.template /etc/nginx/templates/default.conf.template

# Копируем артефакты сборки
COPY --from=builder /app/dist /usr/share/nginx/html

# Healthcheck для оркестратора
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
