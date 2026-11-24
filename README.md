# Cinema Ticket Booking

A small Go service that explores one of the classic concurrency problems in
ticketing systems: **the double-booking race condition**. Two users tap "Book"
on the same seat at the same instant — only one should win, even under high
contention.

Built as a learning project while going from a naive in-memory map to a
lock-protected store and finally to a distributed, TTL-based hold in Redis.

## The Problem

Two users click "Book" on seat A1 at the same instant. Only one should win.

```
User A ──► read seat A1 → "free" ──► write booking ──► success
User B ──► read seat A1 → "free" ──► write booking ──► ???
```

Without any protection, both reads see the seat as free and both writes succeed.
Now two people show up for the same seat.

## The Approaches

The project keeps three interchangeable implementations of `BookingStore`
(`internal/booking/domain.go`), each demonstrating one step of the fix:

| Store | File | Strategy |
|-------|------|----------|
| `MemoryStore` | `memory_store.go` | Naive map — **intentionally racy**, shows the bug |
| `ConcurrentStore` | `concurrent_store.go` | Check-and-set guarded by `sync.RWMutex` |
| `RedisStore` | `redis_store.go` | Atomic `SET NX` + TTL, works across processes |

### How the Redis hold works

```
seat:{movieID}:{seatID}   -> session JSON   (TTL = held, no TTL = confirmed)
session:{sessionID}       -> seat key       (reverse lookup for release/confirm)
```

A seat is claimed with a single atomic `SET ... NX EX`, so exactly one caller
can ever succeed. The key carries a short TTL (2 minutes), acting as a
**hold** that auto-expires if the user never completes checkout — the seat is
released without any cleanup job. The `session:*` key enables releasing or
confirming a hold from the session ID alone.

## Tech Stack

- **Go 1.26** — standard `net/http` routing with method + path patterns
- **Redis 7** (`go-redis/v9`) — atomic claim + TTL holds
- **Vanilla HTML/CSS/JS** — single-file seat-picker UI in `static/`
- **Docker Compose** — Redis + Redis Commander for inspection

## API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/movies` | List movies and their seat layout |
| `GET` | `/movies/{movieID}/seats` | List currently booked seats |
| `POST` | `/movies/{movieID}/seats/{seatID}/hold` | Hold a seat (body: `{"user_id": "..."}`) |
| `GET` | `/` | Seat-picker UI |

Example:

```bash
curl -X POST localhost:8080/movies/inception/seats/A1/hold \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"alice"}'
# 201 {"session_id":"A1","movie_id":"inception","seat_id":"<uuid>","expires_at":"..."}
```

## Running Locally

```bash
# 1. Start Redis (and the Redis Commander UI on :8081)
docker compose up -d

# 2. Run the server on :8080
go run ./cmd

# 3. Open the UI
open http://localhost:8080
```

## Testing

The core test fires **100,000 concurrent goroutines** at the same seat and
asserts that exactly one succeeds:

```bash
go test ./...
```

`TestConcurrentBooking_ExactlyOneWins` is parameterised by the store — swap the
`MemoryStore` back in to watch it fail, or point it at Redis to prove the
atomic claim holds.

## What I Took Away From This

- A map read-modify-write is three steps, not one — the race is between them.
- A mutex fixes correctness **within one process** only; it does nothing once
  you run more than one instance.
- Pushing the guarantee into the datastore (`SET NX`) makes the fix
  distributed and gives you TTL semantics almost for free.
- Modeling a hold as an expiring key turns "release abandoned carts" from a
  background job into a property of storage.
- Concurrency bugs are invisible to tests that run sequentially — the test suite
  has to generate contention on purpose.

## Project Layout

```
cmd/main.go                     HTTP wiring, movie list, static file server
internal/booking/
  domain.go                     Booking model + BookingStore interface
  service.go                    Thin service layer over the store
  handler.go                    HTTP handlers
  memory_store.go               Naive (racy) implementation
  concurrent_store.go           RWMutex-guarded implementation
  redis_store.go                Atomic SET NX + TTL implementation
  service_test.go               Contention test
internal/adapters/redis/        Redis client factory
internal/utils/                 JSON response helper
static/index.html               Seat-picker UI
docker-compose.yaml             Redis + Redis Commander
```

## Roadmap

- Implement `RedisStore.ListBookings` (scan `seat:{movieID}:*`)
- Confirm/release endpoints backed by `session:{sessionID}`
- Per-movie seat layouts driven from the store instead of hard-coded
- `-race` test target in CI
