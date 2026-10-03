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

Endpoints: `GET /healthz`, `GET /listings`, and `DELETE /listings/{id}`.

```sh
curl http://localhost:8080/healthz
curl http://localhost:8080/listings
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

The API is still being developed; the current listings handler needs its UUID import and parsing fixed before it will build.
