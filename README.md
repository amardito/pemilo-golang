# Pemilo Golang - Online Voting System Backend

A robust backend API for online elections/voting built with Go, following Clean Architecture principles.

## 🌟 Features

### Admin Dashboard
- **Authentication System**: Secure login with rate limiting (3 attempts → 5 min lockout)
- **Quota Management**: Per-admin room and voter limits
- **Room Configuration**: Create and manage election rooms with 3 voter types
- **Candidate Management**: Add/edit candidates with photos, descriptions, and sub-candidates
- **Ticket Management**: CSV bulk upload or manual ticket creation
- **Real-time Monitoring**: Live vote tracking with WebSocket support
- **Room Filtering**: Filter by active/inactive, draft/published states

### Voter Experience
- **Dynamic Validation**: Automatic voter validation based on room configuration
- **Three Voter Types**:
  - `custom_tickets`: Secure ticket-based voting
  - `wild_limited`: First-come-first-served with vote limits
  - `wild_unlimited`: Time-range based unlimited voting
- **Secure Voting**: Single-vote enforcement, double-vote prevention

### Security Features
- **Password Security**: AES-256 encryption for transmission + Bcrypt hashing for storage
- **JWT Authentication**: Token-based auth for admin endpoints
- **Basic Auth**: HTTP Basic Auth for owner-only admin creation
- **Rate Limiting**: Login attempt tracking and lockout mechanism

## 🏗️ Architecture

Clean Architecture with strict layer separation:

```
cmd/server/          # Application entrypoint
internal/
  ├── domain/        # Entities & repository interfaces
  ├── usecase/       # Business logic
  ├── repository/    # Database implementation
  ├── handler/       # HTTP controllers
  ├── dto/           # Request/response models
  ├── middleware/    # Auth, CORS, logging
  └── config/        # Configuration management
pkg/                 # Shared utilities
migrations/          # SQL migration files
```

## 🚀 Getting Started

### Prerequisites
- Go 1.21+
- PostgreSQL 14+

### Installation

1. Clone the repository:
```bash
git clone https://github.com/amardito/pemilo-golang.git
cd pemilo-golang
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Run database migrations:
```bash
# Install migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations -database "postgres://user:password@localhost:5432/pemilo?sslmode=disable" up

# Or manually run each migration
psql -U postgres -d pemilo -f migrations/001_create_rooms_table.sql
psql -U postgres -d pemilo -f migrations/002_create_candidates_table.sql
psql -U postgres -d pemilo -f migrations/003_create_sub_candidates_table.sql
psql -U postgres -d pemilo -f migrations/004_create_tickets_table.sql
psql -U postgres -d pemilo -f migrations/005_create_votes_table.sql
psql -U postgres -d pemilo -f migrations/006_create_admins_table.sql
psql -U postgres -d pemilo -f migrations/007_create_login_attempts_table.sql
psql -U postgres -d pemilo -f migrations/008_add_admin_id_to_rooms.sql
```

5. Run the server:
```bash
go run cmd/server/main.go
```

Server will start on `http://localhost:8080`

## 📡 API Endpoints

### Authentication Endpoints

#### Admin Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin_user",
  "password": "encrypted-password"  // AES-256 encrypted
}
```

Response:
```json
{
  "token": "jwt-token",
  "expiresAt": "2024-01-01T00:00:00Z",
  "admin": {
    "id": "uuid",
    "username": "admin_user",
    "maxRoom": 10,
    "maxVoters": 100
  }
}
```

#### Create Admin (Owner Only)
```http
POST /api/v1/owner/create-admin
Authorization: Basic <base64(owner:password)>
Content-Type: application/json

{
  "username": "new_admin",
  "password": "encrypted-password",
  "maxRoom": 10,
  "maxVoters": 100
}
```

### Admin Endpoints (Requires JWT)

All admin endpoints require `Authorization: Bearer <token>` header.

#### Get Quota Info
```http
GET /api/v1/admin/quota
```

#### Room Management
```http
POST   /api/v1/admin/rooms           # Create room
GET    /api/v1/admin/rooms           # List rooms
GET    /api/v1/admin/rooms/:id       # Get room
PUT    /api/v1/admin/rooms/:id       # Update room
DELETE /api/v1/admin/rooms/:id       # Delete room
GET    /api/v1/admin/rooms/:id/realtime  # Real-time stats
```

#### Candidate Management
```http
POST   /api/v1/admin/candidates              # Create candidate
GET    /api/v1/admin/candidates/:id          # Get candidate
PUT    /api/v1/admin/candidates/:id          # Update candidate
DELETE /api/v1/admin/candidates/:id          # Delete candidate
GET    /api/v1/admin/candidates/room/:roomId # List by room
```

#### Ticket Management
```http
POST   /api/v1/admin/tickets              # Create ticket
POST   /api/v1/admin/tickets/bulk         # Bulk create tickets
GET    /api/v1/admin/tickets/room/:roomId # List by room
DELETE /api/v1/admin/tickets/:id          # Delete ticket
```

### Public Voter Endpoints

#### Get Voter Room Info
```http
GET /api/v1/vote?room_id={room_id}
```

#### Cast Vote
```http
POST /api/v1/vote
Content-Type: application/json

{
  "room_id": "uuid",
  "candidate_id": "uuid",
  "sub_candidate_id": "uuid",  // optional
  "ticket_code": "ABC123"      // required for custom_tickets
}
```

#### Verify Ticket
```http
POST /api/v1/vote/verify-ticket
Content-Type: application/json

{
  "room_id": "uuid",
  "ticket_code": "ABC123"
}
```

## 🔐 Voter Types Explained

### 1. custom_tickets
- Voters must provide a unique ticket code
- Each ticket is single-use
- Admin uploads tickets via CSV or manual entry
- Ideal for controlled, secure elections

**Flow:**
1. Voter receives ticket code
2. Voter enters room with ticket
3. System validates ticket (unused, belongs to room)
4. Voter casts vote
5. Ticket marked as used

### 2. wild_limited
- No ticket required
- Room automatically closes when vote limit reached
- Race-condition safe
- First-come-first-served basis

**Flow:**
1. Voter enters room
2. System checks if limit reached
3. Vote accepted if under limit
4. Session closes when limit reached

### 3. wild_unlimited
- No ticket required
- Time-range based access control
- Unlimited participants
- Admin MUST set start/end times

**Flow:**
1. Voter enters room
2. System checks current time in range
3. Vote accepted if within active period
4. Denied if outside time range

## 🧪 Testing

Run tests:
```bash
go test ./...
```

Run with coverage:
```bash
go test -cover ./...
```

## 📦 Database Schema

- **admins**: Admin accounts with quota limits
- **login_attempts**: Login attempt tracking for rate limiting
- **rooms**: Election room configuration (linked to admins)
- **candidates**: Main candidates
- **sub_candidates**: Optional sub-candidates (e.g., vice president)
- **tickets**: Single-use voting tickets
- **votes**: Recorded votes with timestamps

## 🔧 Configuration

Environment variables (see `.env.example`):

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `JWT_SECRET` | Secret for JWT signing | `change-this-secret` |
| `ENCRYPTION_KEY` | AES-256 key (must be 32 bytes) | Required |
| `OWNER_USERNAME` | Owner account username | `owner` |
| `OWNER_PASSWORD` | Owner account password | Required |
| `ENVIRONMENT` | Runtime environment | `development` |
| `ALLOWED_ORIGINS` | CORS allowed origins | `*` |

## 🛡️ Security Features

- **Authentication**: JWT for admin endpoints, Basic Auth for owner endpoints
- **Password Security**: AES-256 encryption + Bcrypt hashing
- **Rate Limiting**: 3 failed login attempts → 5 minute lockout
- **Quota Enforcement**: Per-admin room and voter limits
- **CORS Protection**: Configurable allowed origins
- **SQL Injection Prevention**: Prepared statements only
- **Double Voting Prevention**: Unique constraints on votes table
- **Ticket Security**: Single-use enforcement with database constraints

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details

## 👥 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📞 Support

For issues and questions, please open an issue on GitHub.

---

Built with ❤️ using Go and Clean Architecture principles
