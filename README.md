# Template

A slim Go HTTP service template: layered architecture, PostgreSQL, and a users CRUD as a worked example.

## Stack
- Go 1.26+
- PostgreSQL 14+
- `gorilla/mux`, `jmoiron/sqlx`, `lib/pq`

## Setup
Copy `.env.example` into `.env` and set your database URL.

Table schemas live in the `sql` folder. A migrate tool applies every `*.sql` file in order:
```
make run-migration
```

Run the HTTP server:
```
make run-http
```

Run tests:
```
make test
```

## Docker
Bring up Postgres, run migrations, then start the app:
```
docker compose up --build
```
The server listens on `:8080`.

## Project structure
```
cmd/
  httpserver/   - HTTP server entry point + DI container
  migrate/      - one-shot SQL bootstrap
internal/
  api/          - HTTP handlers, requests, responses
  services/     - business logic
  domain/       - models + repositories (SQL)
  database/     - connection + tx manager
  custerr/      - shared error values
sql/            - table schemas
```

## Endpoints

| Method | Path          | Description        |
| ------ | ------------- | ------------------ |
| GET    | `/health`     | Health check       |
| POST   | `/users`      | Create a user      |
| GET    | `/users`      | List users         |
| GET    | `/users/{id}` | Get a user by id   |
| PUT    | `/users/{id}` | Update a user      |
| DELETE | `/users/{id}` | Delete a user      |

## Sample requests

### Health
```
curl http://localhost:8080/health
```
```
{"data":{"status":"ok"}}
```

### Create user
```
curl --request POST \
  --url http://localhost:8080/users \
  --header 'content-type: application/json' \
  --data '{"name":"Ada Lovelace","email":"ada@example.com"}'
```
```
{
  "data": {
    "id": "564d5e8e-fd06-45b4-b3c9-69fbcf23093e",
    "name": "Ada Lovelace",
    "email": "ada@example.com",
    "created_at": "2026-07-06 09:00:00 +0000",
    "updated_at": "2026-07-06 09:00:00 +0000"
  },
  "message": "Successfully created user"
}
```
A duplicate email returns `409 Conflict`.

### List users
```
curl http://localhost:8080/users
```

### Get user
```
curl http://localhost:8080/users/564d5e8e-fd06-45b4-b3c9-69fbcf23093e
```

### Update user
```
curl --request PUT \
  --url http://localhost:8080/users/564d5e8e-fd06-45b4-b3c9-69fbcf23093e \
  --header 'content-type: application/json' \
  --data '{"name":"Ada L.","email":"ada@example.com"}'
```

### Delete user
```
curl --request DELETE \
  --url http://localhost:8080/users/564d5e8e-fd06-45b4-b3c9-69fbcf23093e
```

## Notes
- Requests and responses are decoupled from service inputs/outputs so each layer evolves independently.
- Repositories are transaction-aware via a context-scoped `TxManager`, so multi-write flows can share one transaction.
- Email is unique at the database level; the create path maps the constraint violation to a `409`.
