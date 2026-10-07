# Local setup

Use Go 1.27.1+, Node.js 22+, PostgreSQL 18+, and npm. Production credentials
must never be used for development or automated tests.

## Complete development stack

From the repository root:

```sh
cp .env.example .env
docker compose up -d --build
docker compose ps
```

Compose starts PostgreSQL, applies migrations, and starts the API, background
worker, and web frontend. Open `http://localhost:3000`. The development account
is `admin` / `Admin@ERP123!`; both values can be replaced in `.env` before first
startup. Changing a bootstrap password does not reset an existing account.
Use `docker compose logs api worker` when startup fails.

To keep data, stop with `docker compose down` without `--volumes`. Set
`POSTGRES_PORT` if the host already runs PostgreSQL on port 5432. For a custom
web port, set `WEB_PORT`; for a custom API port, set both `API_PORT` and
`NEXT_PUBLIC_API_URL` and rebuild the web image.

## Native Go and Node processes

Go programs read process environment variables, not `.env` files. For an existing
disposable local PostgreSQL database in PowerShell:

```powershell
$env:DATABASE_URL = 'postgres://erp_user:erp_secret_password@localhost:5432/erp_dev?sslmode=disable'
$env:APP_ENV = 'development'
$env:MFA_ENCRYPTION_KEY = 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA='
go run ./cmd/erp-migrate up
go run ./cmd/erp-api
```

In a second terminal, set the same environment variables and run
`go run ./cmd/erp-worker`. In a third terminal:

```powershell
cd apps/web
npm ci
npm run dev
```

On Bash, use `export DATABASE_URL='...'`, `export APP_ENV=development`, and
`export MFA_ENCRYPTION_KEY='...'` before the same Go commands. To use just the
Compose PostgreSQL instance, start `docker compose up -d postgres` first.

The migration CLI supports `up`, `down` (one migration), `status`, `--path`, and
`--timeout`. It reads `MIGRATION_DATABASE_URL`, then `GOOSE_DBSTRING`, then
`DATABASE_URL`. It does not require SMTP or API secrets. `erp-cli migrate`
uses the same runner, while `erp-cli seed` applies schema/reference migrations.
Demo business records are initialized only when the API starts with
`BOOTSTRAP_DEMO_DATA_ENABLED=true`; production rejects that option.

## Verification

```powershell
./scripts/doctor.ps1 -RequireApi
```

On Bash: `REQUIRE_API=true ./scripts/doctor.sh`. Without the required-API option,
an offline API is reported as a warning. An API that is alive but not ready
always fails the check. Doctor does not replace the test suites.

For integration tests, set `TEST_DATABASE_URL` to a separate disposable database
such as `erp_integration_test`. The tests rebuild its application schemas. Never point
`TEST_DATABASE_URL` at the development database used by a running API.
