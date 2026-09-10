# Fire_depart

Система управління пожежною частиною: виклики, зміни, особовий склад, збитки.

Чистий старт на Go + PostgreSQL — без MS Access, ODBC і cgo.

## Структура

- `cmd/firedepart` — точка входу (HTTP-сервер, `/healthz`)
- `internal/config` — конфігурація через змінні оточення (`HTTP_ADDR`, `DATABASE_DSN`)
- `internal/db` — канонічна схема PostgreSQL (DDL)

## Запуск

```bash
go build ./...
go test ./...
go run ./cmd/firedepart
```

Дефолтний DSN: `postgres://postgres:postgres@localhost:5432/fire_depart?sslmode=disable`.

## CI

GitHub Actions (`.github/workflows/ci.yml`): `go vet` + `go build` + `go test` на кожен push і PR.

## Дорожня карта

- [ ] Підключити `pgx/v5` (пул з'єднань, `INSERT ... RETURNING`)
- [ ] CRUD-ендпоінти
- [ ] Міграції схеми
