# Copilot Agent — pemilo-golang

This document defines the architecture, rules, feature scope, and behavioral logic for the pemilo-golang backend.  
GitHub Copilot must use this file to generate consistent, secure, and correct backend code.

---

# 🧠 Copilot Agent Purpose

The agent assists development by:
- Generating backend code in Go
- Creating PostgreSQL queries and schema migrations
- Designing API endpoints
- Maintaining project consistency across services
- Helping build room creation logic, candidate management, voter flows, authentication tokens, and quota systems
- Always generating idiomatic, clean Golang code
- Including error handling and context.Context
- Using dependency injection
- Following hexagonal or clean architecture
- Providing SQL migrations and queries
- Avoiding placeholder logic; producing complete working code
- Offering improvements and alternatives when useful

---

# 🏗 Technology Stack

- **Language**: Golang (Go 1.21+)
- **Framework**: Gin (selected)
- **Database**: PostgreSQL 15+
- **Driver**: pgx (PostgreSQL driver)
- **Migrations**: Manual SQL migrations
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Password Security**: AES-256 encryption + bcrypt hashing
- **Live Reload**: Air (development)
- **Containerization**: Docker + Docker Compose

---

# 📌 System Overview

`pemilo-golang` is a backend API for an online election/voting system consisting of:

1. **Client Admin POV** — full dashboard to configure elections.
2. **Voters POV** — end-user voting experience that changes based on room configuration.

Strict Clean Architecture principles must be followed.

---

# 🟦 Client Admin POV — Features & Rules

## 0. Authentication & Admin Management

### 0.1 Admin Account System

Each admin has:
- `admin_id` (UUID)
- `username` (unique)
- `password` (bcrypt hashed in DB)
- `max_room` (quota limit, default: 10)
- `max_voters` (total voters across all rooms, default: 100)
- `is_active` (enabled/disabled flag)

### 0.2 Authentication Flow

**Login Process:**
1. Admin sends `POST /api/v1/auth/login` with:
   ```json
   {
     "username": "admin_user",
     "password": "AES-256-encrypted-password"
   }
   ```
2. Backend:
   - Decrypts password using `ENCRYPTION_KEY` (32 bytes for AES-256)
   - Verifies against bcrypt hash in database
   - Checks rate limiting (3 failed attempts → 5 minute lockout)
   - Records login attempt in `login_attempts` table
3. On success, returns:
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
4. On rate limit exceeded, returns 429:
   ```json
   {
     "error": "Too many failed login attempts. Please try again later."
   }
   ```

**Rate Limiting Rules:**
- Max 3 failed login attempts per username
- After 3 failures: 5 minute lockout
- Counter resets on successful login
- Tracked via `login_attempts` table with timestamps

**JWT Authentication:**
- All `/api/v1/admin/*` endpoints require JWT token
- JWT contains: `admin_id`, `username`, `exp` (expiration)
- Token sent via `Authorization: Bearer <token>` header
- Middleware extracts `admin_id` and stores in Gin context

### 0.3 Admin Creation (Owner Only)

**Endpoint:** `POST /api/v1/owner/create-admin`  
**Authentication:** HTTP Basic Auth only (not JWT)

Request:
```json
{
  "username": "new_admin",
  "password": "AES-256-encrypted-password",
  "maxRoom": 10,
  "maxVoters": 100
}
```

Authorization:
- Basic Auth credentials from env: `OWNER_USERNAME`, `OWNER_PASSWORD`
- This is owner-only access for creating admin accounts
- No JWT required

### 0.4 Admin Quota System

**Quota Enforcement:**
1. **Room Quota:**
   - Before creating room, backend checks: `current_room_count < max_room`
   - Query: `SELECT COUNT(*) FROM rooms WHERE admin_id = ?`
   - Returns 403 if quota exceeded

2. **Voters Quota:**
   - Total voters = sum of all tickets + wild_limited voters across all admin's rooms
   - Query combines:
     - `SELECT COUNT(*) FROM tickets WHERE room_id IN (admin's rooms)`
     - `SELECT SUM(voters_limit) FROM rooms WHERE admin_id = ? AND voters_type = 'wild_limited'`
   - Checked before room creation/update
   - Returns 403 if total would exceed `max_voters`

**Get Quota Info:**  
`GET /api/v1/admin/quota` (JWT protected)

Response:
```json
{
  "maxRoom": 10,
  "currentRoom": 5,
  "maxVoters": 100,
  "currentVoters": 45
}
```

**Quota Exceeded Response:**
When creating a room that exceeds quota, returns 403:
```json
{
  "error": "Room quota exceeded. Maximum rooms allowed: 10"
}
```
or
```json
{
  "error": "Voters quota exceeded. Maximum voters allowed: 100"
}
```

### 0.5 Password Security

**Encryption Flow:**
1. Frontend encrypts plaintext password with AES-256 before sending
2. Backend decrypts using `ENCRYPTION_KEY` from config
3. Backend hashes decrypted password with bcrypt (cost 10)
4. Bcrypt hash stored in database

**Deterministic Salt Configuration:**
- System uses environment-based salt (not random) for consistent encryption
- `ENCRYPTION_SALT_FRONT` + `ENCRYPTION_SALT_BACK` combined to create 8-byte salt
- Same plaintext + same salt = same ciphertext (needed for admin creation)
- Final security layer: bcrypt with random salt on top of encrypted password

**Validation Flow:**
1. Login receives AES-encrypted password
2. Backend decrypts to plaintext using deterministic salt
3. Compares plaintext against bcrypt hash with `bcrypt.CompareHashAndPassword`

**Key Requirements:**
- `ENCRYPTION_KEY` must be exactly 32 bytes (256 bits)
- `ENCRYPTION_SALT_FRONT` and `ENCRYPTION_SALT_BACK` must match between frontend and backend
- Frontend and backend must share same encryption key and salt values
- JWT secret separate from encryption key

**Environment Variables Required:**
```env
ENCRYPTION_KEY=your-32-character-encryption-key  # Exactly 32 chars
ENCRYPTION_SALT_FRONT=frontSalt1234              # Must match frontend
ENCRYPTION_SALT_BACK=backSalt5678                # Must match frontend
JWT_SECRET=your-super-secret-jwt-key-change-this
OWNER_USERNAME=owner
OWNER_PASSWORD=change-this-password-immediately
```

---

## 1. Room Election Configuration

### 1.1 Create Room Session  
Base room fields:
- `room_id`
- `admin_id` (owner of this room, extracted from JWT)
- `room_name`
- `voters_type`:
  - `custom_tickets`
  - `wild_limited`
  - `wild_unlimited`
- `voters_limit` (required for wild_limited)
- `session_active_range`  
  (start_time, end_time — required for wild_unlimited)
- `status` (enabled/disabled)
- `publish_state` (draft/published)
- `session_state` (open/closed) - auto-managed by backend

**Quota Validation on Creation:**
- Check `admin.current_room_count < admin.max_room`
- Check `admin.projected_voters <= admin.max_voters` (considering new room's voters)
- Reject with 403 if quota exceeded

**Admin Ownership:**
- `admin_id` automatically extracted from JWT token
- Room creation endpoint uses `c.Get("admin_id")` from Gin context
- All rooms are linked to the admin who created them
- Quota enforcement happens in usecase layer before repository call

### 1.2 Candidate Management

#### Add Candidate
- photo upload
- name
- description (rich text)
- optional sub-candidates:
  - photo
  - name
  - description (optional)

#### Edit Candidate  
Modify all fields including images, name, description, sub-candidates.

### 1.3 Room Editing

Editable fields:
- room name  
- session active state  
- enable/disable room  
- voters configuration  
- publish / draft toggle  

#### Voter Configuration Modes

##### a. custom-tickets
- Voters must use a ticket code
- Admin uploads:
  - CSV bulk ticket list  
  - OR manual input for small groups  
- Each ticket is unique and single-use

##### b. wild-limited
- No ticket required  
- Room closes automatically when vote count reaches limit

##### c. wild-unlimited
- No ticket required  
- Unlimited voter participation  
- **Admin MUST set session active time range**  
  - `session_start_time`
  - `session_end_time`

### 1.4 Realtime Monitoring

Admin can view realtime vote graph for active rooms only.

Graph characteristics:
- x-axis: timestamp  
- y-axis: voters count  
- each candidate = separate line  

### 1.5 Rooms Listing

Filterable list:
- active  
- inactive  
- draft  
- published  
- disabled  

Actions:
- open realtime vote monitor (only if active)
- quick edit  
- quick delete  

---

# 🟩 Voters POV — Behavior by Room Type

Voters access the voting portal via a **shared link**:

/vote?room_id=<ROOM_ID>

The backend determines the voter validation path based on the room’s `voters_type`.

---

## 1. Voter Flow by Room Type

### 🅰️ custom-tickets

**Flow:**
1. Voter clicks shared link  
   `/vote?room_id=<id>`
2. Backend responds:
   - room details  
   - voting rules requiring ticket input  
3. Voter must enter a valid ticket code  
4. Backend verifies:  
   - ticket belongs to this room  
   - ticket not used before  
5. On success, voter can proceed to candidate list  
6. Voter votes → ticket becomes “used”

**Rules:**
- Ticket is single-use  
- Ticket must belong to room  
- All invalid attempts must be logged  

---

### 🅱️ wild-limited

**Flow:**
1. Voter clicks shared link  
2. No ticket needed  
3. Immediately allowed to see candidate list  
4. When total vote count reaches `voters_limit`:  
   - backend closes voting automatically  
   - room session transitions to “closed”

**Rules:**
- No voter identity validation  
- Vote until limit reached  
- Must handle race-condition for final few votes  

---

### 🅾️ wild-unlimited

**Flow:**
1. Voter clicks shared link  
2. No ticket needed  
3. Voter can proceed immediately  
4. Voting is allowed only if **current timestamp is within the required session active range**:
   - `session_start_time ≤ NOW ≤ session_end_time`

**Rules:**
- Session active range is **mandatory** for this mode  
- If outside time range:
  - show “session not active”  
  - deny votes  
- No vote limit  

---

## 2. Voting Flow (Shared Across All Types)

Regardless of voters_type:

1. Voter enters room portal  
2. Backend validates room + voter eligibility based on rules  
3. User sees candidate + sub-candidate list  
4. User selects a candidate (or sub-candidate)  
5. Backend stores:
   - room_id  
   - candidate_id  
   - voter_identifier (ticket or autogen if wild)  
   - timestamp  
6. Backend pushes update to realtime monitor (websocket, event stream, or polling)

---

# 🧱 Architecture Rules

### Clean Architecture Layers

handler → usecase → domain → repository → db

### Handlers:
- No logic  
- Only request parsing + response formatting
- Extract `admin_id` from JWT context for authenticated endpoints

### Usecases:
- All business rules  
- Implements room types logic  
- Handles time-range validation  
- Handles ticket verification
- Enforces admin quota limits
- Rate limiting logic for login attempts
- Password decryption and bcrypt validation

### Repository:
- Persistent storage only  
- Implements domain interfaces
- Quota counting queries (room count, voters count per admin)
- Login attempt tracking for rate limiting  

---

# 📂 File Structure Rules

| Directory | Role |
|----------|------|
| `cmd/server` | main entrypoint |
| `internal/domain` | entities/interfaces |
| `internal/usecase` | business logic |
| `internal/repository` | DB implementation |
| `internal/handler` | controllers |
| `internal/dto` | request/response models |
| `internal/middleware` | JWT/Basic auth/logging |
| `internal/config` | environment & config |
| `pkg` | utilities |
| `migrations` | SQL migrations |

---

# 🧪 Testing Requirements

Tests must include:
- ticket validation logic  
- voters_type branching logic  
- time-range enforcement  
- vote closing behavior for wild-limited  
- repository interaction  
- race-condition handling for last votes
- login rate limiting (3 attempts → 5 min lockout)
- password encryption/decryption flow
- quota enforcement (room and voters limits)
- JWT token generation and validation
- Basic Auth for owner endpoints

**Security Testing:**
- Verify AES-256 encryption/decryption with 32-byte key
- Test bcrypt hashing and verification
- Validate rate limiting resets after successful login
- Test quota counting accuracy (tickets + wild_limited voters)
- Verify JWT expiration and validation
- Test Basic Auth credentials validation

**Integration Testing:**
- Create admin → Login → Create room → Check quota
- Test quota exceeded scenarios (room and voters limits)
- Verify admin ownership of rooms
- Test rate limiting with multiple failed login attempts

---

# 🧩 Agent Commands Understanding

Example commands Copilot must support:

- "Create voter validator for all voters_types"
- "Add endpoint for custom-ticket verification"
- "Implement wild-limited vote counter"
- "Add session active time-range validation"
- "Generate migration for session range"
- "Create usecase for voting flow"
- "Add websocket for realtime graph updates"
- "Add admin authentication with rate limiting"
- "Implement quota enforcement for admin accounts"
- "Create login endpoint with password encryption"
- "Add Basic Auth for owner endpoints"
- "Update room creation to extract admin_id from JWT"
- "Add quota validation before creating room"
- "Implement login attempt tracking for rate limiting"

---

# ✔ Final Instructions for Copilot Agent

- Always respect room voters_type logic  
- Ensure correct branching in usecases  
- Generate clean, idiomatic Go  
- Follow file placement rules  
- Validate all inputs (ticket, time-range, room status)  
- Write secure vote-recording logic  
- Prevent double voting  
- Avoid business logic in handlers  
- Document complex flows as needed
- Enforce admin quota limits before room creation
- Extract admin_id from JWT context for ownership tracking
- Use AES-256 for password transmission, bcrypt for storage
- Implement 3-attempt/5-minute rate limiting for login
- Separate JWT auth (admin endpoints) from Basic Auth (owner endpoints)
- Never store plaintext passwords
- Always validate ENCRYPTION_KEY is 32 bytes

## Database Schema Requirements

### Admin Tables

**admins:**
```sql
CREATE TABLE admins (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,      -- bcrypt hash
    max_room INTEGER NOT NULL DEFAULT 10,
    max_voters INTEGER NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_admins_username ON admins(username);
CREATE INDEX idx_admins_is_active ON admins(is_active);
```

**login_attempts:**
```sql
CREATE TABLE login_attempts (
    id VARCHAR(36) PRIMARY KEY,
    identifier VARCHAR(255) NOT NULL,    -- username
    attempt_at TIMESTAMP NOT NULL,
    success BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX idx_login_attempts_identifier ON login_attempts(identifier);
CREATE INDEX idx_login_attempts_attempt_at ON login_attempts(attempt_at);
CREATE INDEX idx_login_attempts_identifier_success ON login_attempts(identifier, success);
```

**rooms (updated):**
```sql
CREATE TABLE rooms (
    id VARCHAR(36) PRIMARY KEY,
    admin_id VARCHAR(36),
    name VARCHAR(255) NOT NULL,
    voters_type VARCHAR(50) NOT NULL CHECK (voters_type IN ('custom_tickets', 'wild_limited', 'wild_unlimited')),
    voters_limit INTEGER,
    session_start_time TIMESTAMP,
    session_end_time TIMESTAMP,
    status VARCHAR(20) NOT NULL CHECK (status IN ('enabled', 'disabled')),
    publish_state VARCHAR(20) NOT NULL CHECK (publish_state IN ('draft', 'published')),
    session_state VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (session_state IN ('open', 'closed')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (admin_id) REFERENCES admins(id) ON DELETE CASCADE
);
CREATE INDEX idx_rooms_admin_id ON rooms(admin_id);
CREATE INDEX idx_rooms_voters_type ON rooms(voters_type);
CREATE INDEX idx_rooms_status ON rooms(status);
CREATE INDEX idx_rooms_publish_state ON rooms(publish_state);
CREATE INDEX idx_rooms_session_state ON rooms(session_state);
```

**Enum Types:**
- `voters_type`: custom_tickets | wild_limited | wild_unlimited
- `status`: enabled | disabled
- `publish_state`: draft | published
- `session_state`: open | closed (auto-managed, closes when voting ends)

---

# 📁 Recommended Project Structure

The current implementation follows this structure:

```
/cmd/server              # Application entrypoint
/internal/
    /domain              # Entities and interfaces
    /usecase             # Business logic
    /repository          # Database implementation
    /handler             # HTTP controllers
    /dto                 # Request/response models
    /middleware          # JWT/Basic auth/logging
    /config              # Environment & config
/pkg/utils               # Shared utilities (crypto, etc.)
/migrations              # SQL migration files
/tmp                     # Air build artifacts (gitignored)
.air.toml                # Air configuration
docker-compose.yml       # PostgreSQL orchestration
Makefile                 # Build and dev commands
```

**Key Principles:**
- Clean Architecture: handler → usecase → domain → repository → db
- No circular dependencies
- Domain layer has no external dependencies
- All business logic in usecases
- Handlers only parse requests and format responses

---

# ✅ Copilot Usage Instructions

When generating code for this project:
- Generate only production-ready code
- Avoid pseudocode
- Follow Golang best practices
- Always include proper error handling
- Use the existing project structure
- Respect Clean Architecture boundaries
- Include context.Context for database operations
- Add appropriate logging where needed
- Suggest improvements when applicable
- Ensure type safety and null handling

---

# 📡 API Endpoint Reference

## Authentication Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/auth/login` | None | Admin login with rate limiting |
| POST | `/api/v1/owner/create-admin` | Basic Auth | Create admin account (owner only) |
| GET | `/api/v1/admin/quota` | JWT | Get current admin quota info |

## Room Management Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/admin/rooms` | JWT | Create room (checks quota) |
| GET | `/api/v1/admin/rooms` | JWT | List all rooms for admin |
| GET | `/api/v1/admin/rooms/:id` | JWT | Get room details |
| PUT | `/api/v1/admin/rooms/:id` | JWT | Update room |
| DELETE | `/api/v1/admin/rooms/:id` | JWT | Delete room |
| GET | `/api/v1/admin/rooms/:roomId/realtime` | JWT | Realtime vote monitoring |

## Candidate Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/admin/candidates` | JWT | Create candidate |
| GET | `/api/v1/admin/candidates/:id` | JWT | Get candidate details |
| PUT | `/api/v1/admin/candidates/:id` | JWT | Update candidate |
| DELETE | `/api/v1/admin/candidates/:id` | JWT | Delete candidate |
| GET | `/api/v1/admin/candidates/room/:roomId` | JWT | List candidates for room |

## Ticket Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/admin/tickets` | JWT | Create single ticket |
| POST | `/api/v1/admin/tickets/bulk` | JWT | Bulk upload tickets (CSV) |
| GET | `/api/v1/admin/tickets/room/:roomId` | JWT | List tickets for room |
| DELETE | `/api/v1/admin/tickets/:id` | JWT | Delete ticket |

## Voter Endpoints (Public)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/vote?room_id=<id>` | None | Get room info for voting |
| POST | `/api/v1/vote` | None | Cast vote |
| POST | `/api/v1/vote/verify-ticket` | None | Verify ticket before voting |

---
