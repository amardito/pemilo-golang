# Pemilo Golang Backend MVP

Backend MVP untuk Pemilo (secret ballot) menggunakan Go + PostgreSQL.

## Requirements

- Go 1.23+
- PostgreSQL 14+
- Migration tool (contoh: [golang-migrate](https://github.com/golang-migrate/migrate))

## Environment Variables

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/pemilo?sslmode=disable"
export PORT=8080
export JWT_SECRET="replace-me"
export REQUEST_TIMEOUT_SECONDS=5
```

## Run Migrations

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

## Run Server

```bash
go run ./cmd/server
```

## API MVP (`/api/v1`)

### 1) Create event
```bash
curl -X POST http://localhost:8080/api/v1/events \
  -H 'Content-Type: application/json' \
  -d '{"title":"Pemilihan Ketua BEM 2026","status":"OPEN"}'
```

### 2) Create slate + members
```bash
curl -X POST http://localhost:8080/api/v1/events/{eventId}/slates \
  -H 'Content-Type: application/json' \
  -d '{
    "number":1,
    "name":"Paslon 1",
    "vision":"Kampus inklusif",
    "mission":"Program kerja transparan",
    "members":[
      {"role":"Ketua","full_name":"Budi"},
      {"role":"Wakil","full_name":"Siti"}
    ]
  }'
```

### 3) Import voters CSV
Header wajib: `full_name,nim` (opsional: `class_name`)

```bash
curl -X POST http://localhost:8080/api/v1/events/{eventId}/voters/import \
  -F 'file=@voters.csv'
```

### 4) Generate tokens
```bash
curl -X POST http://localhost:8080/api/v1/events/{eventId}/voters/tokens/generate
```

### 5) Export tokens CSV
```bash
curl -L http://localhost:8080/api/v1/events/{eventId}/voters/tokens/export -o tokens.csv
```

### 6) Vote login (token + NIM)
- Token format: `^[A-Z0-9]{8}$`
- NIM dinormalisasi: trim + uppercase + hapus semua whitespace

```bash
curl -X POST http://localhost:8080/api/v1/events/{eventId}/vote/login \
  -H 'Content-Type: application/json' \
  -d '{"token":"A1B2C3D4","nim":"20 24 001"}'
```

### 7) Vote submit (atomic)
Gunakan `vote_session` dari endpoint login sebagai bearer token.

```bash
curl -X POST http://localhost:8080/api/v1/events/{eventId}/vote/submit \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer {vote_session}' \
  -d '{"slate_id":"{slateId}"}'
```

### 8) Realtime stats (polling endpoint)
```bash
curl http://localhost:8080/api/v1/events/{eventId}/stats
```

## Secret Ballot Design Notes

- Tidak ada `voter_id` di tabel `ballots`.
- Status partisipasi disimpan terpisah di `voters.has_voted` + `voters.voted_at`.
- Token dipakai sekali (`voter_tokens.status=USED`) saat vote submit.
- Submit vote berjalan dalam transaksi DB dengan row-lock (`FOR UPDATE`) untuk mencegah double vote.
