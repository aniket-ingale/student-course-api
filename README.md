# student-course-api

A small Go REST API for reading student records, backed by PostgreSQL via GORM.
Schema and seed data are managed with [golang-migrate](https://github.com/golang-migrate/migrate).

## Endpoints

| Method | Path             | Description                          |
| ------ | ---------------- | ------------------------------------ |
| GET    | `/students/{id}` | Fetch one student (`404` if missing) |
| GET    | `/students`      | List all students (`[]` when empty)  |
| GET    | `/healthz`       | Liveness probe (DB ping)             |

A student is `{ "studentId": <int>, "name": <string>, "address": <string>, "grade": <int> }`.

## Architecture

Layered: `handler` → `service` → `repository` → GORM/Postgres. Handlers never
touch GORM directly; the repository is an interface so the service and handlers
are unit-testable with fakes. Domain errors live in `internal/apperr` and are
mapped to HTTP status codes at the boundary.

```
cmd/server/main.go     entry point: config, db, mux, graceful shutdown
internal/config        env -> Config
internal/db            GORM Postgres connection + pool
internal/model         Student model
internal/repository    StudentRepository interface + GORM impl
internal/service       business logic + validation
internal/handler       HTTP handlers, router, middleware, JSON helpers
internal/apperr        domain errors -> HTTP status mapping
migrations             golang-migrate SQL (schema + seed)
```

## Configuration

Set via environment (see `.env.example`). Provide `DATABASE_URL`, or the
discrete `DB_*` vars from which a DSN is built:

| Var            | Default     | Notes                            |
| -------------- | ----------- | -------------------------------- |
| `DATABASE_URL` | —           | Full Postgres DSN (preferred)    |
| `DB_HOST`      | —           | Required if no `DATABASE_URL`    |
| `DB_PORT`      | `5432`      |                                  |
| `DB_USER`      | —           | Required if no `DATABASE_URL`    |
| `DB_PASSWORD`  | —           |                                  |
| `DB_NAME`      | —           | Required if no `DATABASE_URL`    |
| `DB_SSLMODE`   | `disable`   |                                  |
| `HTTP_PORT`    | `8080`      |                                  |
| `LOG_LEVEL`    | `info`      | `debug`/`info`/`warn`/`error`    |

## Run with Docker Compose (recommended)

Brings up Postgres, runs migrations as a one-shot job, then starts the API:

```bash
make compose-up        # docker compose up --build
```

The API is then on `http://localhost:8080`.

## Run on the host

Requires a local Postgres and the `migrate` CLI on `PATH`. With `DATABASE_URL`
exported (e.g. in `.env.local`):

```bash
make migrate-up        # apply schema + seed
make run               # go run ./cmd/server
```

## Migrations

```bash
make migrate-up                      # apply all
make migrate-down                    # roll back the latest
make migrate-create name=add_x       # new up/down pair
```

Every migration ships both an `.up.sql` and a `.down.sql`. The app does **not**
use `AutoMigrate`; the `migrations/` directory is the source of truth.

## Tests

```bash
make test              # unit tests (faked repository, httptest)
make test-integration  # repository tests against real Postgres via testcontainers (needs Docker)
make fmt vet           # format + static analysis
```

## Example requests

```bash
$ curl -s localhost:8080/students/1
{"studentId":1,"name":"Ada Lovelace","address":"12 Analytical Way, London","grade":10}

$ curl -s localhost:8080/students
[{"studentId":1,"name":"Ada Lovelace",...}, ...]

$ curl -s -o /dev/null -w '%{http_code}\n' localhost:8080/students/99999
404
```

## Out of scope / future work

- Write endpoints (POST/PUT/DELETE), auth, pagination/filtering.
- A `courses` domain (despite the repo name) — likely a follow-up.
