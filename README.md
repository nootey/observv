# observv

A small showcase for the Grafana observability stack (LGTM: Loki, Grafana, Tempo, Mimir).

The app is kept small on purpose, so the focus stays on observability:

- A Go API (chi, Postgres, sqlc) that tracks a fleet of terminals.
- A simulator that acts as the fleet. It calls the API over HTTP, so every signal comes from real traffic.
- The stack is meant to be reusable in other projects.

> Work in progress. The observability stack is not added yet.

## Quick start

Run everything in Docker:

```sh
make up      # db, migrations, seed, api, simulator
make logs    # follow api and simulator logs
make down
```

Or run the Go code on the host, with only the database in Docker:

```sh
cp .env.example .env
make db
make migrate-up
make seed
make run       # terminal 1
make simulate  # terminal 2
```

The API listens on `http://localhost:8080`.

## API

| Method | Path | Body |
|---|---|---|
| `GET` | `/health` | |
| `GET` | `/api/terminals` | |
| `GET` | `/api/terminals/{id}` | |
| `POST` | `/api/terminals/{id}/heartbeat` | `{"status": "online"}` or `{"status": "offline"}` |
| `POST` | `/api/terminals/{id}/transfers` | `{"to_id": 2, "bytes": 1024}` |

A transfer needs both terminals online. If one is offline, the API answers `409 Conflict`. The simulator does not know the status of its peers, so part of its traffic gets this answer. This gives the demo a normal share of client errors.

## Development

```sh
make test     # unit and integration tests (integration tests need Docker)
make lint
make sqlc     # regenerate the database code from the SQL queries
make mock     # regenerate mocks
```

`make mock` and `make lint` need [mockery](https://github.com/vektra/mockery) v3.8+ and [golangci-lint](https://golangci-lint.run) v2 built with Go 1.27 or newer. Install them with `go install`. sqlc runs through `go run` at a pinned version, so you do not need to install it.
