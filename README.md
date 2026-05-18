# student-pm

**Система управления проектами для студенческих групп.** Курсовая работа: микросервисная архитектура на Go, Clean Architecture, PostgreSQL, JWT, Docker.

---

## Быстрый старт

```bash
git clone <repo-url> && cd student-pm
cp .env.example .env                              

docker-compose up -d --build                                   
```

### Остановить и удалить

```bash
make down                 
make down-volumes            
```

---

## Тестирование

```bash
make test   
make cover
```
