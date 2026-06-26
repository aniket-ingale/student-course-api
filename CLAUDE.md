# CLAUDE.md

Guidance for Claude Code when working in this repository.

## Project overview

`student-course-api` is a Go-based REST API for managing student records. The
service exposes read endpoints backed by a **PostgreSQL** database.

### Tech stack

| Concern      | Choice                                            |
| ------------ | ------------------------------------------------- |
| Language     | Go (1.22+)                                         |
| HTTP router  | Standard library `net/http` (`ServeMux` patterns) |
| Database     | PostgreSQL                                         |
| ORM / DB     | [GORM](https://gorm.io) with the Postgres driver  |
| Migrations   | [golang-migrate](https://github.com/golang-migrate/migrate) |

> The structure and conventions below are implemented in the codebase (layered
> handler → service → repository, golang-migrate migrations, Docker Compose for
> local dev). See `README.md` for setup and run instructions.

## API endpoints

| Method | Path                  | Description                          |
| ------ | --------------------- | ------------------------------------ |
| GET    | `/students/{id}`      | Fetch a single student's details     |
| GET    | `/students`           | List all students                    |

Routes are registered on a `net/http` `ServeMux` using Go 1.22+ method+path
patterns, e.g. `mux.HandleFunc("GET /students/{id}", ...)`. Read path params
with `r.PathValue("id")`.

### Student model

A student has the following attributes:

| Field       | JSON key      | Type   | Notes                       |
| ----------- | ------------- | ------ | --------------------------- |
| Name        | `name`        | string |                             |
| Student ID  | `studentId`   | number | Unique identifier (integer) |
| Address     | `address`     | string |                             |
| Grade       | `grade`       | number | Integer                     |

GORM model (maps to the `students` table):

```go
type Student struct {
    StudentID int    `gorm:"column:student_id;primaryKey" json:"studentId"`
    Name      string `gorm:"column:name"                  json:"name"`
    Address   string `gorm:"column:address"               json:"address"`
    Grade     int    `gorm:"column:grade"                 json:"grade"`
}

// TableName overrides GORM's default pluralization rules if needed.
func (Student) TableName() string { return "students" }
```

### Response shapes

- `GET /students/{id}` → `200 OK` with a single `Student` JSON object, or
  `404 Not Found` if no student matches the ID.
- `GET /students` → `200 OK` with a JSON array of `Student` objects (empty array
  when there are no students).
- Errors return a JSON body, e.g. `{"error": "student not found"}`, with an
  appropriate HTTP status code.

## Suggested project layout

```
student-course-api/
├── cmd/server/main.go      # Entry point: wires config, db (GORM), mux, server
├── internal/
│   ├── handler/            # HTTP handlers (request/response, status codes)
│   ├── service/            # Business logic
│   ├── repository/         # DB access via GORM (queries -> models)
│   ├── model/              # Domain types (Student)
│   └── db/                 # GORM connection setup, config
├── migrations/             # golang-migrate SQL files (NNNN_*.up/down.sql)
├── go.mod
└── go.sum
```

Keep HTTP concerns in `handler`, data access in `repository`, and domain logic
in `service`. Handlers must not call GORM directly — go through the repository.

## Database

- Connect with GORM's Postgres driver (`gorm.io/driver/postgres`).
- Read connection settings from the environment; do not hardcode credentials.
  Expected vars: `DATABASE_URL` (or `DB_HOST`, `DB_PORT`, `DB_USER`,
  `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`).
- **Do not rely on `AutoMigrate` for schema changes** — schema is owned by the
  `migrations/` directory (golang-migrate). Treat the DB schema as the source of
  truth and keep the GORM model in sync with it.

### Migrations (golang-migrate)

```bash
# Create a new migration pair
migrate create -ext sql -dir migrations -seq <name>

# Apply / roll back (DATABASE_URL points at Postgres)
migrate -database "$DATABASE_URL" -path migrations up
migrate -database "$DATABASE_URL" -path migrations down 1
```

Every migration must ship both an `.up.sql` and a matching `.down.sql`.

## Common commands

```bash
go mod tidy            # Sync dependencies
go build ./...         # Build everything
go run ./cmd/server    # Run the API server locally
go test ./...          # Run all tests
go vet ./...           # Static analysis
gofmt -l -w .          # Format code (or: go fmt ./...)
```

A local Postgres (e.g. via Docker) must be running and migrations applied before
`go run ./cmd/server` will serve requests.

## Conventions

- **Formatting:** Always run `gofmt`/`go fmt` before committing. Code must be
  gofmt-clean.
- **Errors:** Return errors up the stack; wrap with `fmt.Errorf("...: %w", err)`
  to preserve the chain. Handle them at the HTTP boundary and map to status codes.
- **Context:** Pass `context.Context` as the first argument through service and
  repository calls; propagate the request context from handlers. Use
  `db.WithContext(ctx)` on GORM calls so queries respect request cancellation.
- **DB access:** Use GORM's parameterized query methods (`Where`, `First`,
  `Find`) — never string-concatenate user input into queries. Translate
  `gorm.ErrRecordNotFound` to a `404` at the boundary.
- **JSON:** Use struct tags for field names (see model above). Set
  `Content-Type: application/json` on responses.
- **Naming:** Exported identifiers use `CamelCase`; keep package names short and
  lowercase. DB columns use `snake_case`.

## Testing

- Use Go's standard `testing` package and table-driven tests.
- Test handlers with `net/http/httptest`.
- Mock or fake the repository layer when testing services/handlers so tests do
  not require a live database. Keep GORM behind a repository interface to make
  this straightforward.
