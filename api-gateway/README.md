# api-gateway

Единая точка входа для системы student-pm. Принимает запросы клиентов на `:8080` и проксирует на три backend-сервиса.

## Что делает

- **Reverse proxy** на `httputil.ReverseProxy` с маршрутизацией по префиксу пути:
  - `/auth/*`, `/users/*`        → auth-service:8081
  - `/groups/*`                  → groups-service:8082
  - `/projects/*`, `/tasks/*`    → projects-service:8083
- **JWT pre-validation** — проверяет подпись и срок токена до похода на upstream (defense in depth, отбрасывает мусор раньше).
- **Прокидывает заголовки**: `Authorization`, `X-Request-ID`, `X-User-ID`, `X-User-Role` — последние два полезны бэкенду в логах.
- **CORS** — настраивается через `CORS_ALLOW_ORIGINS`.
- **Rate limiting** — N запросов с IP за окно (по умолчанию 100/мин).
- **Swagger UI агрегатор** на `/swagger/` — переключатель между тремя сервисами; spec'и каждый сервис отдаёт сам у себя на `:8081/swagger/doc.json` и т.д.
- **Health-checks** `/health`, `/ready`.
- **Graceful shutdown** по `SIGINT/SIGTERM`.

## Маршрутизация

| Префикс            | Auth      | Upstream         |
| ------------------ | --------- | ---------------- |
| `POST /auth/register, /auth/login, /auth/refresh` | публично | auth-service     |
| `/auth/logout`, `/users/*`                        | Bearer   | auth-service     |
| `/groups/*`                                       | Bearer   | groups-service   |
| `/projects/*`, `/tasks/*`                         | Bearer   | projects-service |
| `/`, `/health`, `/ready`, `/swagger/*`            | публично | gateway сам      |

## Тесты

```bash
make test
```

Покрыто:
- публичные `/auth/*` без токена идут на upstream;
- защищённые без токена → 401, **upstream не вызывается**;
- с валидным токеном — gateway пробрасывает `X-User-ID`, `X-User-Role`, `X-Request-ID`;
- маршрутизация по префиксу действительно попадает в нужный upstream;
- `/`, `/health`, `/ready` отвечают локально без upstream.

## Замечания

- **Зачем JWT-проверка на gateway, если она есть на сервисах.** Затем, чтобы не нагружать backend заведомо мусорными запросами, и чтобы прокидывать `X-User-*` заголовки в логи бэкенда. Реальная авторизация и RBAC всё равно на сервисах — gateway не знает про роли в группах, owner проекта и т.п.
- **Почему один процесс на Go, а не nginx + lua.** Так проще для курсовой: один стек, одна команда `go run`. В проде тут был бы Envoy/nginx с дополнительной JWT-валидацией.
- **Swagger UI** — статичный HTML с `swagger-ui-dist` через CDN. JS внутри ходит напрямую на `:8081/8082/8083/swagger/doc.json`. Это работает только в dev (где у клиента есть прямой доступ к портам сервисов). В проде агрегатор бы тоже проксировал spec-файлы.
