# Fire_depart

Курсовий проєкт: система обліку викликів пожежної частини.

**Стек:** Go (REST API) · PostgreSQL (БД) · React + TypeScript + Vite (фронтенд) · SCSS + BEM (стилі).

## Структура репозиторію

- `cmd/firedepart/` — точка входу HTTP API (стандартна бібліотека `net/http`, патерни Go 1.22)
- `internal/api/` — REST-хендлери (`/api/health`, `/api/schema`, CRUD, layout, дашборд, звіти)
- `internal/config/` — конфігурація через змінні оточення
- `internal/db/` — DDL-схема PostgreSQL, метадані, CRUD-шар на pgx/v5, дашборд, звіти
- `db/migrations/` — SQL-міграції (PK, FK, CHECK, UNIQUE, індекси, сід довідників)
- `frontend/` — React + TypeScript (Vite), SCSS/BEM, ERD-мапа, CRUD-браузер, дашборд
- `docs/` — документація проєкту БД
- `.github/workflows/ci.yml` — CI: Go (tidy/vet/build/test) + frontend (npm build)

## Швидкий старт

### 1. База даних (PostgreSQL у Docker)

```bash
docker compose up -d
```

Міграція `db/migrations/001_initial_schema.sql` застосовується автоматично при першому старті тому.

### 2. Backend

```bash
go mod tidy   # генерує go.sum (після pull — обов'язково)
go run ./cmd/firedepart
```

API слухає на `:8080` (змінна `HTTP_ADDR`). Дефолтний DSN: `postgres://postgres:postgres@localhost:5432/fire_depart?sslmode=disable` (змінна `DATABASE_DSN`).

### 3. Frontend

```bash
cd frontend
npm install
npm run dev
```

Dev-сервер на `http://localhost:5173`, запити `/api/*` проксіються на `:8080`.

## API

- `GET /api/health` — статус сервера
- `GET /api/schema` — метадані таблиць (колонки, PK, FK) для ERD-мапи і форм
- `GET /api/dashboard` — агреговані показники дашборда
- `GET /api/tables/{table}` — список рядків (до 500)
- `POST /api/tables/{table}` — створити запис
- `PATCH|PUT /api/tables/{table}/{id}` — оновити запис
- `DELETE /api/tables/{table}/{id}` — видалити запис
- `GET|PUT /api/layout/{view}` — позиції вузлів ERD-мапи (автозбереження при перетягуванні)
- `GET /api/report/incidents` — зведений звіт по викликах (HTML, друк → PDF)
- `GET /api/report/incident/{id}` — детальний звіт по виклику (HTML, друк → PDF)

## Дорожня карта

- [x] Схема PostgreSQL з PK/FK/CHECK/UNIQUE/індексами
- [x] ERD-мапа схеми (React Flow) з перетягуванням і збереженням позицій
- [x] CRUD-ендпоінти (pgx/v5, whitelist + параметризація) та CRUD-браузер у UI
- [x] Дашборд: активні виклики, техніка, склад, останні виклики
- [x] Звіти по викликах (друк/PDF: зведений + детальний)
- [ ] Редактор схеми: створення/редагування таблиць і колонок з UI
- [ ] Автентифікація користувачів
