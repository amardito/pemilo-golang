# Copilot Agent — pemilo-golang

This document defines the architecture, rules, feature scope, and behavioral logic for the pemilo-golang backend.  
GitHub Copilot must use this file to generate consistent, secure, and correct backend code.

---

# 📌 System Overview

`pemilo-golang` is a backend API for an online election/voting system consisting of:

1. **Client Admin POV** — full dashboard to configure elections.
2. **Voters POV** — end-user voting experience that changes based on room configuration.

Strict Clean Architecture principles must be followed.

---

# 🟦 Client Admin POV — Features & Rules

## 1. Room Election Configuration

### 1.1 Create Room Session  
Base room fields:
- `room_id`
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

### Usecases:
- All business rules  
- Implements room types logic  
- Handles time-range validation  
- Handles ticket verification  

### Repository:
- Persistent storage only  
- Implements domain interfaces  

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
| `internal/middleware` | JWT/auth/logging |
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

---

# 🧩 Agent Commands Understanding

Example commands Copilot must support:

- “Create voter validator for all voters_types”
- “Add endpoint for custom-ticket verification”
- “Implement wild-limited vote counter”
- “Add session active time-range validation”
- “Generate migration for session range”
- “Create usecase for voting flow”
- “Add websocket for realtime graph updates”

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

---
