# API Documentation - Pemilo Golang

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
Admin endpoints require JWT authentication:
```
Authorization: Bearer <your_jwt_token>
```

---

## Public Endpoints (Voter)

### Get Voter Room Information
Retrieves room details and validates voter eligibility.

**Request:**
```http
GET /vote?room_id={uuid}
```

**Response:**
```json
{
  "room": {
    "id": "uuid",
    "name": "Presidential Election 2024",
    "voters_type": "custom_tickets",
    "status": "enabled",
    "publish_state": "published",
    "session_state": "open"
  },
  "candidates": [...],
  "requires_ticket": true,
  "is_active": true,
  "message": ""
}
```

### Cast Vote
Submit a vote for a candidate.

**Request:**
```http
POST /vote
Content-Type: application/json

{
  "room_id": "uuid",
  "candidate_id": "uuid",
  "sub_candidate_id": "uuid",  // optional
  "ticket_code": "TICKET123"   // required for custom_tickets
}
```

**Response:**
```json
{
  "id": "uuid",
  "room_id": "uuid",
  "candidate_id": "uuid",
  "voter_identifier": "TICKET123",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Verify Ticket
Pre-verify a ticket code before voting.

**Request:**
```http
POST /vote/verify-ticket
Content-Type: application/json

{
  "room_id": "uuid",
  "ticket_code": "TICKET123"
}
```

**Response:**
```json
{
  "valid": true,
  "message": "Proceed to vote"
}
```

---

## Admin Endpoints

### Room Management

#### Create Room
```http
POST /admin/rooms
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Presidential Election 2024",
  "voters_type": "custom_tickets",
  "voters_limit": 1000,              // required for wild_limited
  "session_start_time": "2024-01-15T08:00:00Z",  // required for wild_unlimited
  "session_end_time": "2024-01-15T20:00:00Z",    // required for wild_unlimited
  "status": "enabled",
  "publish_state": "draft"
}
```

#### List Rooms
```http
GET /admin/rooms?status=enabled&publish_state=published&session_state=open
Authorization: Bearer <token>
```

Query Parameters:
- `status`: `enabled` | `disabled`
- `publish_state`: `draft` | `published`
- `session_state`: `open` | `closed`

#### Get Room
```http
GET /admin/rooms/{room_id}
Authorization: Bearer <token>
```

#### Update Room
```http
PUT /admin/rooms/{room_id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Updated Name",
  "status": "disabled"
  // ... other fields
}
```

#### Delete Room
```http
DELETE /admin/rooms/{room_id}
Authorization: Bearer <token>
```

#### Get Real-time Vote Data
```http
GET /admin/rooms/{room_id}/realtime
Authorization: Bearer <token>
```

**Response:**
```json
{
  "room_id": "uuid",
  "room_name": "Presidential Election",
  "vote_data": [
    {
      "candidate_id": "uuid",
      "candidate_name": "John Doe",
      "vote_count": 150,
      "timestamp": "2024-01-15T10:30:00Z"
    }
  ],
  "total_votes": 300,
  "updated_at": "2024-01-15T10:30:00Z"
}
```

---

### Candidate Management

#### Create Candidate
```http
POST /admin/candidates
Authorization: Bearer <token>
Content-Type: application/json

{
  "room_id": "uuid",
  "name": "John Doe",
  "photo_url": "https://example.com/photo.jpg",
  "description": "<p>Rich text description</p>",
  "sub_candidates": [
    {
      "name": "Jane Smith",
      "photo_url": "https://example.com/photo2.jpg",
      "description": "Vice President"
    }
  ]
}
```

#### Get Candidate
```http
GET /admin/candidates/{candidate_id}
Authorization: Bearer <token>
```

#### List Candidates by Room
```http
GET /admin/candidates/room/{room_id}
Authorization: Bearer <token>
```

#### Update Candidate
```http
PUT /admin/candidates/{candidate_id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "John Doe Updated",
  "photo_url": "https://example.com/new-photo.jpg",
  "description": "<p>Updated description</p>"
}
```

#### Delete Candidate
```http
DELETE /admin/candidates/{candidate_id}
Authorization: Bearer <token>
```

---

### Ticket Management

#### Create Single Ticket
```http
POST /admin/tickets
Authorization: Bearer <token>
Content-Type: application/json

{
  "room_id": "uuid",
  "code": "TICKET001"
}
```

#### Create Tickets in Bulk
```http
POST /admin/tickets/bulk
Authorization: Bearer <token>
Content-Type: application/json

{
  "room_id": "uuid",
  "codes": [
    "TICKET001",
    "TICKET002",
    "TICKET003"
  ]
}
```

#### List Tickets by Room
```http
GET /admin/tickets/room/{room_id}
Authorization: Bearer <token>
```

**Response:**
```json
{
  "tickets": [
    {
      "id": "uuid",
      "room_id": "uuid",
      "code": "TICKET001",
      "is_used": false,
      "used_at": null,
      "created_at": "2024-01-15T08:00:00Z",
      "updated_at": "2024-01-15T08:00:00Z"
    }
  ]
}
```

#### Delete Ticket
```http
DELETE /admin/tickets/{ticket_id}
Authorization: Bearer <token>
```

---

## Error Responses

All endpoints may return error responses in this format:

```json
{
  "error": "Error message",
  "message": "Additional details"
}
```

Common HTTP Status Codes:
- `200 OK`: Success
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid input or business logic violation
- `401 Unauthorized`: Missing or invalid authentication
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

---

## Business Logic Rules

### Voter Type: custom_tickets
- Ticket code is **required**
- Each ticket can only be used **once**
- Ticket must belong to the room
- Invalid ticket attempts are logged

### Voter Type: wild_limited
- No ticket required
- First-come-first-served
- Vote limit **must** be set by admin
- Room automatically closes when limit reached
- Race-condition safe

### Voter Type: wild_unlimited
- No ticket required
- Session time range **must** be set
- Voting allowed only within active time range
- No vote limit

### General Voting Rules
- One vote per voter per room
- Room must be enabled
- Room must be published
- Session must be open
- Candidate must belong to room

---

## Example Workflows

### 1. Create Custom Tickets Election
```bash
# 1. Create room
POST /admin/rooms
{
  "name": "Class President Election",
  "voters_type": "custom_tickets",
  "status": "enabled",
  "publish_state": "draft"
}

# 2. Add candidates
POST /admin/candidates
{
  "room_id": "{room_id}",
  "name": "Alice Johnson",
  ...
}

# 3. Upload tickets (CSV or bulk)
POST /admin/tickets/bulk
{
  "room_id": "{room_id}",
  "codes": ["STUD001", "STUD002", ...]
}

# 4. Publish room
PUT /admin/rooms/{room_id}
{
  "publish_state": "published"
}

# 5. Voters vote
POST /vote
{
  "room_id": "{room_id}",
  "candidate_id": "{candidate_id}",
  "ticket_code": "STUD001"
}
```

### 2. Create Wild Limited Election
```bash
# 1. Create room with vote limit
POST /admin/rooms
{
  "name": "Quick Poll",
  "voters_type": "wild_limited",
  "voters_limit": 100,
  "status": "enabled",
  "publish_state": "published"
}

# 2. Share link: /vote?room_id={room_id}
# 3. First 100 voters can vote
# 4. Session closes automatically
```

### 3. Monitor Real-time Results
```bash
# While voting is active
GET /admin/rooms/{room_id}/realtime

# Returns live vote counts for all candidates
```
