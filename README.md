# Fire_depart

Курсовий проєкт: система обліку викликів пожежної частини.

**Стек:** Go (REST API) · PostgreSQL (БД) · React + TypeScript + Vite (фронтенд) · SCSS + BEM (стилі).

## Структура репозиторію

- `cmd/firedepart/` — точка входу HTTP API (стандартна бібліотека `net/http`, патерни Go 1.22)
- `internal/api/` — REST-хендлери (`/api/health`, `/api/schema`)
- `internal/config/` — конфігурація через змінні оточення
- `internal/db/` — DDL-схема PostgreSQL та метадані таблиць для ERD-мапи
- `db/migrations/` — SQL-міграції (PK, FK, CHECK, UNIQUE, індекси, сід довідників)
- `frontend/` — React + TypeScript (Vite), SCSS/BEM, ERD-мапа на React Flow
- `docs/` — документація проєкту БД
- `.github/workflows/ci.yml` — CI: Go (vet/build/test) + frontend (npm build)

## Швидкий старт

### 1. База даних (PostgreSQL у Docker)

```bash
docker compose up -d
```

Міграція `db/migrations/001_initial_schema.sql` застосовується автоматично при першому старті тому.

### 2. Backend

```bash
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

## Дорожня карта

- [x] Схема PostgreSQL з PK/FK/CHECK/UNIQUE/індексами
- [x] ERD-мапа схеми (React Flow) з перетягуванням таблиць
- [ ] CRUD-ендпоінти для всіх сутностей + підключення pgx/v5
- [ ] Збереження позицій вузлів ERD-мапи (`schema_layouts` + API)
- [ ] Збереження/редагування таблиць і колонок з UI
- [ ] Дашборд: активні виклики, техніка, зміни
- [ ] PDF-звіти по викликах (час, адреса, бригада, збитки)
