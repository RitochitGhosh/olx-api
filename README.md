# Learning Go: OLX API

I'm learning Go by building a small OLX-style listings API. This project is a work in progress where I practice HTTP handlers, PostgreSQL, migrations, request contexts, and structured logging.

## Getting started

You need Go (the version specified in `olx-api/go.mod`), PostgreSQL, and Make. Create a local PostgreSQL database named `olx`, then:

```sh
cd olx-api
go mod download
```

Create or update `olx-api/.env` with your local settings:

```dotenv
PORT=8080
ENV=development
DATABASE_URL=postgres://postgres:password@localhost:5432/olx?sslmode=disable
```

Use your own database credentials. Keep secrets out of version control; `sslmode=disable` is for local development.

From `olx-api/`, apply migrations and start the API:

```sh
make migrate-up
make run
```

Endpoints: `GET /healthz`, `GET /listings`, `GET /listings/{id}`, `POST /listings`, `PUT /listings/{id}`, and `DELETE /listings/{id}`.

```sh
curl http://localhost:8080/healthz
curl http://localhost:8080/listings
```

Update a listing with `PUT /listings/{id}`. Send all four editable fields; title, description, and city must be non-empty, and price must be greater than zero. This updates `updated_at` and returns the saved listing with HTTP 200. A missing listing returns 404.

```sh
curl -X PUT http://localhost:8080/listings/YOUR_LISTING_UUID \
  -H 'Content-Type: application/json' \
  -d '{"title":"Bike","description":"Used bike","price":100,"city":"Pune"}'
```

## Creating migrations

Migrations live in `olx-api/migrations/` and use numbered SQL pairs. From `olx-api/`, create the next pair (use the next unused number):

```sh
touch migrations/000002_add_listing_category.up.sql
touch migrations/000002_add_listing_category.down.sql
```

Write the schema change in the **up** file:

```sql
ALTER TABLE listings ADD COLUMN category TEXT;
```

Write its reverse in the **down** file:

```sql
ALTER TABLE listings DROP COLUMN category;
```

Run `make migrate-up` to apply all pending migrations. **`make migrate-down` rolls back every applied migration**, which can delete tables and data.

Migration habits:

- Keep each migration focused on one schema change.
- Add a new migration instead of editing one already applied to a shared database.
- Test both directions on a disposable local database; rolling back can lose data.
- Back up important data before destructive changes.
- If a migration fails, inspect the SQL and database state before retrying.

Current caveats: the first down migration has invalid SQL (`DROP TABLE listings IF EXISTS;` should be `DROP TABLE IF EXISTS listings;`). The migration runner also reports an error when there are no pending changes.

## Go habits I'm practicing

- Keep startup code in `cmd/` and application code in `internal/`.
- Pass request contexts to database queries and close query results.
- Use parameterized SQL and validate input before querying.
- Log useful error details with request IDs; return safe errors to clients.
- Keep configuration in environment variables.
- Format and check code before committing:

```sh
gofmt -w cmd internal
go vet ./...
go test ./...
```

## Listing code structure

`internal/listing/` keeps related code together:

- **Handler:** reads HTTP requests and writes responses.
- **Service:** trims input and validates business rules.
- **Repository:** runs database queries using request contexts.
- **Request/response types:** keep the JSON contract separate from the database model.

New listings start with an `active` database status. Missing listings return 404; invalid IDs return 400; invalid listing fields return 422. A shared `ErrNotFound` lets the handler recognize missing records with `errors.Is`, even when an error is wrapped.

Tests cover input validation and HTTP responses without needing PostgreSQL. Run them from `olx-api/` with `go test ./...`.
