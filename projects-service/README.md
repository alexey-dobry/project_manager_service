# projects-service

Ядро системы: проекты, задачи, комментарии, статистика. Часть монорепозитория `student-pm`.

## Что внутри

- **Три связанные сущности**: `Project` (агрегат-корень) → `Task` → `Comment`.
- **State-machine** для статусов проекта и задачи. Невалидные переходы → 409.
- **Эндпоинт `/projects/:id/stats`**: одной SQL-агрегацией с `FILTER (WHERE …)` считаем разбивки по статусам и приоритетам, число просроченных задач и `done_percent`.
- **RBAC** в usecase: owner проекта или teacher/admin — менеджер; assignee может менять только статус и описание своей задачи.
- **IDOR-защита**: `GET/PATCH /projects/:id/tasks/:task_id` проверяет, что задача действительно принадлежит этому проекту, иначе `task_not_in_project`.
- Чистая архитектура, PostgreSQL + pgx/v5, миграции golang-migrate, Swagger, healthchecks, graceful shutdown.

## Эндпоинты

| Метод   | Путь                                   | Доступ                                  |
| ------- | -------------------------------------- | --------------------------------------- |
| POST    | `/projects`                            | любой авторизованный (становится owner) |
| GET     | `/projects`                            | любой авторизованный                    |
| GET     | `/projects/:id`                        | любой авторизованный                    |
| PATCH   | `/projects/:id`                        | owner / teacher / admin                 |
| DELETE  | `/projects/:id`                        | owner / teacher / admin                 |
| GET     | `/projects/:id/stats`                  | любой авторизованный                    |
| POST    | `/projects/:id/tasks`                  | менеджер проекта                        |
| GET     | `/projects/:id/tasks`                  | любой авторизованный                    |
| GET     | `/projects/:id/tasks/:task_id`         | любой авторизованный                    |
| PATCH   | `/projects/:id/tasks/:task_id`         | менеджер проекта (всё) / assignee (status, description) |
| DELETE  | `/projects/:id/tasks/:task_id`         | менеджер проекта                        |
| POST    | `/tasks/:task_id/comments`             | любой авторизованный                    |
| GET     | `/tasks/:task_id/comments`             | любой авторизованный                    |
| DELETE  | `/tasks/:task_id/comments/:comment_id` | автор / teacher / admin                 |
| GET     | `/health`, `/ready`, `/swagger/*`      | —                                       |

## State-machine

### Project

```
draft  ─────────────────► in_progress  ─► review  ─► completed
  │                          │              │            │
  │                          │              │            │
  └──► archived ◄────────────┴──────────────┴────────────┘
              │
              └─► draft  (расконсервация)
```

`review → in_progress` тоже разрешён (на доработку). Идемпотентные переходы (`status = текущему`) — норма.

### Task

```
todo  ◄──┐ ┌─► in_progress  ◄─► done   (done → in_progress = переоткрытие)
   │     │ │       │
   │     │ │       │
   └─►  blocked ◄──┘   (blocked возвращается в todo или in_progress)
```

`todo → done` напрямую запрещён — нужно пройти через `in_progress`. `blocked → done` тоже запрещён.

## Stats SQL

Один запрос с `COUNT(*) FILTER (WHERE …)`:

```sql
SELECT
  COUNT(*),
  COUNT(*) FILTER (WHERE status = 'todo'),
  COUNT(*) FILTER (WHERE status = 'in_progress'),
  COUNT(*) FILTER (WHERE status = 'done'),
  COUNT(*) FILTER (WHERE status = 'blocked'),
  COUNT(*) FILTER (WHERE priority = 'low'), …,
  COUNT(*) FILTER (WHERE due_date < $now AND status != 'done')
FROM tasks WHERE project_id = $1
```

`done_percent` считается в Go из `done / total * 100`.

## Структура

```
projects-service/
├── cmd/main.go
├── internal/
│   ├── domain/                       Project, Task, Comment, статусы, transitions, ProjectStats
│   ├── usecase/                      Service (3 use case-области в одном пакете)
│   ├── repository/postgres.go        Repos = {Projects, Tasks, Comments}
│   ├── delivery/http/
│   │   ├── handler.go                projects + system + mappers
│   │   ├── tasks_handler.go          tasks
│   │   ├── comments_handler.go       comments
│   │   ├── middleware.go
│   │   ├── routes.go
│   │   └── dto.go
│   ├── config/
│   └── pkg/{jwt,errors,validator,logger}
├── migrations/000001_init.{up,down}.sql
├── tests/{unit,integration}/
└── README · Dockerfile · Makefile · .air.toml · .env.example
```

## Запуск

```bash
cp .env.example .env
docker-compose up -d projects-postgres projects-service
```

Локально — как в auth/groups: поднять Postgres из compose, поправить `POSTGRES_HOST=localhost` в .env, `make run`.

JWT_SECRET должен совпадать с auth-service.

## Тесты

```bash
make test
make cover           # покрытие usecase + domain
```

Что покрыто:
- **state-machine** (`tests/unit/transitions_test.go`) — таблица переходов для проектов и задач, обе матрицы;
- **RBAC и сценарии**: создание/обновление/удаление проектов, IDOR-защита для задач, права assignee, удаление комментариев автором/админом;
- **stats**: агрегации, overdue, done_percent, 404 для несуществующего проекта;
- **integration**: реальный Fiber + JWT-handshake + полный сценарий «проект → задача → блок → комментарий → stats → удаление».

## Замечания для проверяющего

- **Почему один Service, а не три.** `Project`, `Task`, `Comment` — это один bounded context (проект — корневой агрегат). Разделение на три сервиса в usecase-слое привело бы к дублированию RBAC-проверок и циркулярным зависимостям. Структуры репозиториев — три, но фасадятся через `Repos` для удобной DI.
- **Почему `clear_*` флаги в DTO.** В JSON `null` и «поле не передано» неотличимы при `*time.Time` без custom Unmarshaler. Чтобы не городить unmarshal'ы, используется явный `clear_deadline: true` для очистки nullable-поля. Это однозначно и легко документируется.
- **Почему не event-driven с groups-service.** Для курсовой это избыточно. В реальном проде создание задачи проверяло бы через event/gRPC, что assignee — участник группы проекта. Здесь это TODO, явно отмечено в коде.
- **Почему статус `blocked` отдельный, а не флаг.** В ТЗ перечислено четырёхэлементное множество — реализация буквально следует ТЗ. State-machine это поддерживает: `blocked` доступен из `todo`/`in_progress`, выход — обратно в эти же состояния.
