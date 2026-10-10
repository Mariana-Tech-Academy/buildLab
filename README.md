# FoodHub API

Backend for FoodHub, a food ordering platform where customers browse restaurants and order food,
restaurant managers run their restaurant and menu, and admins oversee the platform.

## Tech stack

Go, PostgreSQL, GORM, gorilla/mux, JWT (golang-jwt v5), bcrypt, GitHub Actions.

## Project structure

```
main.go          entrypoint: config, database, migrations, server
config/          environment variable loading and validation
auth/            JWT generation/verification and request-context claims
middleware/      authentication, role checks, CORS
routes/          route registration and dependency wiring per feature
handlers/        HTTP layer: decode/validate requests, write JSON responses
service/         business rules
repository/      database access with GORM
models/          GORM models and status/role constants
utils/           JSON response helpers, password hashing
docs/            ERD and design notes
```

Requests flow `routes -> middleware -> handlers -> service -> repository -> database`.
Handlers must not talk to the database directly, and repositories must not contain business rules.

## Conventions

- Every response uses `utils.Success` / `utils.Error`:
  `{ "success": true, "message": "...", "data": {} }` or `{ "success": false, "message": "..." }`.
- All routes live under `/api`. Protected routes go on the `protected(...)` subrouter;
  role-restricted routes add `middleware.RequireRole(...)`.
- Get the caller inside a handler with `auth.ClaimsFromContext(r.Context())`.
- Money is stored as integers in minor units (kobo/cents), never `float64`.

## Getting started

```bash
cp .env.example .env   # fill in DB credentials and JWT_SECRET
go run .
curl localhost:8080/api/health
```

## Environment variables

See [`.env.example`](.env.example). `JWT_SECRET` is required; the server refuses to start without it.

## CI

CI runs `gofmt`, `go vet` and `go build` on every pull request.

## Git workflow

- Branch from `main` as `feature/<name>`, `fix/<name>` or `test/<name>`. Never commit to `main` directly.
- Use conventional commit messages, e.g. `feat: add customer registration endpoint`.
- A PR can only merge when CI is green and a teammate has approved it.

## Documentation

- ERD: [`docs/erd.md`](docs/erd.md)
- API docs: _TODO (Swagger or Postman collection)_

## Deployment

_TODO: backend URL, frontend URL._

## Team

_TODO_


Logout behaviour: The frontend discards the JWT when the user logs out. The backend does not invalidate the token; it remains valid until it expires. Protected endpoints continue to require a valid JWT.
