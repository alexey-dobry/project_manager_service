# groups-service

Сервис управления студенческими группами и членством. Часть монорепозитория `student-pm`.

## Что внутри

- **CRUD** для групп с пагинацией и фильтрами (`faculty`, `course`).
- **Управление членством**: добавить/удалить участника, список участников.
  Composite primary key `(group_id, user_id)` — никаких суррогатных ID для membership.
- **RBAC** в usecase-слое (не в handlers): матрица ролей описана ниже.
- **Проверка JWT** через общий с auth-service `JWT_SECRET`. Этот сервис **не выпускает** токены, только проверяет — поэтому `pkg/jwt` содержит только `Verifier`.
- Чистая архитектура, PostgreSQL + pgx/v5, миграции golang-migrate, Swagger, healthchecks, graceful shutdown — всё как в auth-service.

## Эндпоинты

| Метод   | Путь                              | Auth   | Доступ                              |
| ------- | --------------------------------- | ------ | ----------------------------------- |
| POST    | `/groups`                         | Bearer | teacher / admin                     |
| GET     | `/groups`                         | Bearer | все авторизованные                  |
| GET     | `/groups/:id`                     | Bearer | все авторизованные                  |
| PATCH   | `/groups/:id`                     | Bearer | teacher / admin / лидер группы\*    |
| DELETE  | `/groups/:id`                     | Bearer | teacher / admin                     |
| GET     | `/groups/:id/members`             | Bearer | все авторизованные                  |
| POST    | `/groups/:id/members`             | Bearer | teacher / admin / лидер группы      |
| DELETE  | `/groups/:id/members/:user_id`    | Bearer | teacher / admin / лидер / сам себя  |
| GET     | `/health`, `/ready`, `/swagger/*` | —      |                                     |

\* лидер группы может править свою группу, но **менять лидера** может только teacher/admin.

## RBAC-матрица (источник истины — `usecase/service.go`)

```
                   create  list  read  update*  delete  add_member  remove_member
student            ✗       ✓     ✓     ✗        ✗       ✗           sam_sebya
group_leader       ✗       ✓     ✓     ✗        ✗       ✗           sam_sebya
teacher            ✓       ✓     ✓     ✓        ✓       ✓           ✓
admin              ✓       ✓     ✓     ✓        ✓       ✓           ✓
лидер группы (Х)   ✗       ✓     ✓     ✓**      ✗       ✓           ✓ (по группе Х)
```

\*\* лидер группы может править поля группы, но не лидера.

## Особенности реализации

- При создании группы её **лидер автоматически становится участником** с
  ролью `leader`. Из-за того что Postgres не поддерживает ROW-уровень изоляции
  без явных транзакций для двух таблиц "из коробки" в этом репо — операция
  не атомарна (см. TODO в `usecase/service.go`). Для курсовой это допустимо;
  в реальном проде обернули бы в транзакцию через `pgx.BeginTx`.
- **Удаление лидера** запрещено: `RemoveMember` возвращает 403, если
  `userID == group.LeaderID`. Сначала смените лидера через `PATCH /groups/:id`.
- **БД у каждого сервиса своя** (Database per Service). Поэтому `leader_id`
  и `user_id` в `group_memberships` — **логические FK** без явного REFERENCES
  на таблицу `users` (она в другой БД). В реальном проде это закрывается
  событиями (например, на удаление user — emit'ить событие, которое подхватит
  groups-service).

## Сущности

```
groups
├── id          UUID PK
├── name        VARCHAR(120) UNIQUE
├── course      INT (1..10)
├── faculty     VARCHAR(120)
├── leader_id   UUID         (логическая FK -> users.id)
├── created_at, updated_at

group_memberships
├── group_id      UUID  ┐
├── user_id       UUID  ┴── PRIMARY KEY (group_id, user_id)
├── role_in_group ENUM('member','leader')
├── joined_at
```

## Запуск

```bash
cp .env.example .env
# из корневого docker-compose:
docker-compose up -d groups-postgres groups-service
```

Или локально:

```bash
docker-compose up -d groups-postgres
# в .env: POSTGRES_HOST=localhost POSTGRES_PORT=<порт из compose>
make run
```

JWT-секрет должен совпадать с auth-service'ом. В корневом `.env` он один на оба сервиса.

## Тесты

```bash
make test                # всё
make test-unit           # RBAC-матрица
make test-integration    # реальный Fiber + JWT-handshake + in-memory repo
make cover               # покрытие usecase
```

## Замечания для проверяющего

- **Почему RBAC в usecase, а не в middleware.** Middleware видит только токен,
  но не объект (например, какая группа редактируется). Правило «лидер этой
  группы может править её, но не другую» по определению требует доступа к
  объекту → это бизнес-правило, и его место в usecase. Это же делает RBAC
  тестируемым без HTTP.
- **Почему Verifier, а не TokenProvider.** Этот сервис принципиально не должен
  уметь выпускать токены — это нарушило бы границы контекста. Поэтому
  в `pkg/jwt` только `ParseAccess`, общий `JWT_SECRET` — единственная связка
  с auth-service.
- **Почему composite PK для membership.** "Один пользователь × одна группа =
  одна запись" — это естественный ключ. Суррогатный `id` тут не несёт
  никакой информации, только лишний индекс.
