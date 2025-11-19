# Password Encryption Reference

## Overview
This document explains the password security implementation for pemilo-golang authentication.

## Two-Layer Security

### Layer 1: AES-256 Encryption (Transmission)
**Purpose**: Protect passwords during transmission from frontend to backend

**Frontend (Encryption)**:
```javascript
// Example using CryptoJS
import CryptoJS from 'crypto-js';

const encryptionKey = 'your-32-character-encryption-key'; // Same as backend ENCRYPTION_KEY

function encryptPassword(plainPassword) {
  return CryptoJS.AES.encrypt(plainPassword, encryptionKey).toString();
}

// Usage
const encryptedPassword = encryptPassword('user-typed-password');

// Send to backend
fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    username: 'admin_user',
    password: encryptedPassword  // Send encrypted password
  })
});
```

**Backend (Decryption)**:
```go
// pkg/utils/crypto.go
func DecryptPassword(encryptedPassword, key string) (string, error) {
    // 1. Base64 decode
    // 2. Extract IV (first 16 bytes)
    // 3. Decrypt using AES-256 CFB
    // 4. Return plaintext password
}
```

### Layer 2: Bcrypt Hashing (Storage)
**Purpose**: Secure storage in database (irreversible)

**Backend (Hashing)**:
```go
// pkg/utils/crypto.go
func HashPassword(password string) (string, error) {
    return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}
```

**Backend (Verification)**:
```go
func VerifyPassword(hashedPassword, plainPassword string) error {
    return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
```

## Complete Flow

### Registration/Admin Creation
```
1. Frontend: User enters password → "myPassword123"
2. Frontend: Encrypts with AES-256 → "U2FsdGVkX1+..." (encrypted)
3. Backend: Receives encrypted password
4. Backend: Decrypts to plaintext → "myPassword123"
5. Backend: Hashes with bcrypt → "$2a$10$..." (hash)
6. Database: Stores bcrypt hash only
```

### Login/Authentication
```
1. Frontend: User enters password → "myPassword123"
2. Frontend: Encrypts with AES-256 → "U2FsdGVkX1+..." (encrypted)
3. Backend: Receives encrypted password
4. Backend: Decrypts to plaintext → "myPassword123"
5. Backend: Retrieves bcrypt hash from database → "$2a$10$..."
6. Backend: Compares plaintext with hash using bcrypt.CompareHashAndPassword
7. Backend: Returns JWT token if match
```

## Configuration

### Environment Variables
```env
# MUST be exactly 32 characters (256 bits)
ENCRYPTION_KEY=your-32-character-encryption-key
```

### Validation
```go
// internal/config/config.go
if config.EncryptionKey == "" || len(config.EncryptionKey) != 32 {
    return nil, fmt.Errorf("ENCRYPTION_KEY must be exactly 32 characters for AES-256")
}
```

## Security Considerations

### ✅ DO
- Use HTTPS in production (prevent MITM attacks)
- Keep ENCRYPTION_KEY secret and identical on frontend/backend
- Use strong bcrypt cost (default: 10)
- Rotate encryption keys periodically
- Store bcrypt hashes only in database

### ❌ DON'T
- Don't send plaintext passwords over network
- Don't store plaintext passwords anywhere
- Don't use weak encryption keys
- Don't use the same key for JWT and AES encryption
- Don't expose encryption keys in client-side code

## Why Two Layers?

1. **AES-256 (Transmission)**:
   - Symmetric encryption
   - Fast and efficient
   - Reversible (decrypt on backend)
   - Protects password during network transit
   
2. **Bcrypt (Storage)**:
   - One-way hashing
   - Slow by design (prevents brute force)
   - Irreversible (can only compare)
   - Protects password if database is compromised

## Implementation Files

| File | Purpose |
|------|---------|
| `pkg/utils/crypto.go` | Encryption/decryption and hashing functions |
| `internal/usecase/auth_usecase.go` | Login flow: decrypt → verify |
| `internal/usecase/admin_usecase.go` | Admin creation: decrypt → hash |
| `internal/config/config.go` | Encryption key validation |

## Testing Password Security

### Test Encryption/Decryption
```go
func TestPasswordEncryption(t *testing.T) {
    key := "your-32-character-encryption-key"
    original := "myPassword123"
    
    encrypted, err := EncryptPassword(original, key)
    assert.NoError(t, err)
    
    decrypted, err := DecryptPassword(encrypted, key)
    assert.NoError(t, err)
    assert.Equal(t, original, decrypted)
}
```

### Test Hashing/Verification
```go
func TestPasswordHashing(t *testing.T) {
    password := "myPassword123"
    
    hashed, err := HashPassword(password)
    assert.NoError(t, err)
    
    err = VerifyPassword(hashed, password)
    assert.NoError(t, err)
    
    err = VerifyPassword(hashed, "wrongPassword")
    assert.Error(t, err)
}
```

## Frontend Example (Complete)

```javascript
// auth.js
import CryptoJS from 'crypto-js';

const API_BASE = 'http://localhost:8080/api/v1';
const ENCRYPTION_KEY = process.env.REACT_APP_ENCRYPTION_KEY; // 32 chars

export async function login(username, password) {
  // Encrypt password
  const encryptedPassword = CryptoJS.AES.encrypt(
    password, 
    ENCRYPTION_KEY
  ).toString();
  
  // Send to backend
  const response = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      username,
      password: encryptedPassword
    })
  });
  
  if (!response.ok) {
    if (response.status === 429) {
      throw new Error('Too many login attempts. Please wait 5 minutes.');
    }
    throw new Error('Invalid credentials');
  }
  
  const data = await response.json();
  // Store JWT token
  localStorage.setItem('token', data.token);
  return data;
}

export async function createAdmin(ownerUsername, ownerPassword, newAdmin) {
  // Encrypt new admin password
  const encryptedPassword = CryptoJS.AES.encrypt(
    newAdmin.password,
    ENCRYPTION_KEY
  ).toString();
  
  // Basic Auth header
  const basicAuth = btoa(`${ownerUsername}:${ownerPassword}`);
  
  const response = await fetch(`${API_BASE}/owner/create-admin`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Basic ${basicAuth}`
    },
    body: JSON.stringify({
      username: newAdmin.username,
      password: encryptedPassword,
      maxRoom: newAdmin.maxRoom || 10,
      maxVoters: newAdmin.maxVoters || 100
    })
  });
  
  if (!response.ok) {
    throw new Error('Failed to create admin');
  }
  
  return response.json();
}
```

## Common Issues & Solutions

### Issue: "Encryption key must be 32 characters"
**Solution**: Ensure `ENCRYPTION_KEY` environment variable is exactly 32 bytes.

```bash
# Generate a secure 32-character key
openssl rand -base64 24  # Generates 32-character string
```

### Issue: "Invalid credentials" on correct password
**Solution**: Verify encryption key matches between frontend and backend.

### Issue: Bcrypt error "invalid hash"
**Solution**: Ensure password is decrypted before passing to bcrypt comparison.

### Issue: Rate limiting not working
**Solution**: Check `login_attempts` table exists and has proper indexes.

## Security Checklist

- [x] AES-256 encryption for password transmission
- [x] Bcrypt hashing for password storage
- [x] Encryption key exactly 32 bytes
- [x] No plaintext passwords in database
- [x] HTTPS enforced in production
- [x] Rate limiting implemented (3 attempts/5 min)
- [x] JWT tokens expire
- [x] Basic Auth for owner endpoints only
- [x] Environment variables for secrets

---

**Note**: Always use HTTPS in production to prevent man-in-the-middle attacks, even with encrypted passwords.
