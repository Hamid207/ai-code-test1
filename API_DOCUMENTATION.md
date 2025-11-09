# API Documentation

## Base URL
```
http://localhost:8080
```

---

## Endpoints

### 1. Health Check

**Endpoint:** `GET /health`

**Description:** Check the health status of the service and database connectivity

**Request:**
- No request body required
- No authentication required

**Response (200 OK - Healthy):**
```json
{
  "status": "healthy",
  "database": "connected"
}
```

**Response (503 Service Unavailable - Unhealthy):**
```json
{
  "status": "unhealthy",
  "database": "disconnected"
}
```

**Example cURL:**
```bash
curl -X GET http://localhost:8080/health
```

---

### 2. Sign In with Apple

**Endpoint:** `POST /api/v1/auth/apple`

**Description:** Authenticate user using Apple ID token and create/update user in database

**Rate Limit:** 10 requests per minute per IP

**Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "id_token": "eyJraWQiOiJXNldjT0tCIiwiYWxnIjoiUlMyNTYifQ.eyJpc3MiOiJodHRwczovL2FwcGxlaWQuYXBwbGUuY29tIiwiYXVkIjoiY29tLnlvdXJhcHAuaWQiLCJleHAiOjE3MzEyNzAwMDAsImlhdCI6MTczMTE4MzYwMCwic3ViIjoiMDAxMjM0LmFiY2RlZjEyMzQ1Njc4OTAuMTIzNCIsImVtYWlsIjoidXNlckBleGFtcGxlLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjoidHJ1ZSJ9.signature",
  "nonce": "random-nonce-from-client-abc123xyz"
}
```

**Request Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id_token` | string | ✅ Yes | JWT token received from Apple Sign In |
| `nonce` | string | ✅ Yes | Random nonce string used in Apple Sign In flow |

**Success Response (200 OK):**
```json
{
  "user_id": "001234.abcdef1234567890.1234",
  "email": "user@example.com",
  "token": null
}
```

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `user_id` | string | Apple's unique identifier for the user |
| `email` | string | User's email address (verified by Apple) |
| `token` | string\|null | JWT token for subsequent API requests (not implemented yet) |

**Error Response (400 Bad Request - Invalid Request):**
```json
{
  "error": "invalid_request",
  "message": "Key: 'AppleSignInRequest.IDToken' Error:Field validation for 'IDToken' failed on the 'required' tag"
}
```

**Error Response (401 Unauthorized - Invalid Token):**
```json
{
  "error": "authentication_failed",
  "message": "Invalid or expired token"
}
```

**Error Response (429 Too Many Requests - Rate Limit):**
```json
{
  "error": "rate_limit_exceeded",
  "message": "Too many requests"
}
```

**Example cURL:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/apple \
  -H "Content-Type: application/json" \
  -d '{
    "id_token": "eyJraWQiOiJXNldjT0tCIiwiYWxnIjoiUlMyNTYifQ...",
    "nonce": "your-random-nonce"
  }'
```

**Example Response Flow:**

1. **New User (First Login):**
   - User logs in with Apple
   - Database creates new user record
   - Returns user information

2. **Existing User:**
   - User logs in with Apple
   - Database finds existing user by `apple_id`
   - Updates email if changed
   - Returns user information

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "error_code",
  "message": "Human readable error message"
}
```

### Common Error Codes:

| Status Code | Error Code | Description |
|-------------|------------|-------------|
| 400 | `invalid_request` | Request validation failed |
| 401 | `authentication_failed` | Token verification failed |
| 429 | `rate_limit_exceeded` | Too many requests |
| 500 | `internal_server_error` | Server error (not exposed to client) |
| 503 | `service_unavailable` | Database or service unavailable |

---

## Database Operations

### Apple Sign In Flow:

1. **Validate Input:**
   - Apple ID: min 10 chars, max 255 chars
   - Email: valid email format, max 255 chars

2. **Database Operation (Atomic UPSERT):**
   ```sql
   INSERT INTO users (apple_id, email)
   VALUES ($1, $2)
   ON CONFLICT (apple_id)
   DO UPDATE SET
       email = EXCLUDED.email,
       updated_at = CURRENT_TIMESTAMP
   RETURNING id, apple_id, email, created_at, updated_at
   ```

3. **Scenarios:**
   - **New User:** Creates new record
   - **Existing User (Same Email):** Returns existing record
   - **Existing User (New Email):** Updates email and returns

---

## Security Features

### Request Validation:
- ✅ Required field validation
- ✅ Email format validation
- ✅ Apple ID format validation
- ✅ Request body size limit (1MB)

### Token Verification:
- ✅ JWT signature verification with Apple's public keys
- ✅ Token expiration check
- ✅ Issuer verification (must be Apple)
- ✅ Audience verification (must match client ID)
- ✅ Nonce verification (prevents replay attacks)
- ✅ Email verification status check

### Rate Limiting:
- ✅ 10 requests per minute per IP address
- ✅ Applies to all `/api/v1/auth/*` endpoints

### Database Security:
- ✅ Context timeouts (5 seconds per query)
- ✅ Connection pool limits
- ✅ Prepared statements (prevents SQL injection)
- ✅ Input validation before database operations

---

## Authentication Flow

```
┌─────────┐                 ┌──────────┐                ┌──────────┐
│  Client │                 │  Backend │                │  Database│
└────┬────┘                 └─────┬────┘                └─────┬────┘
     │                            │                           │
     │  1. Sign In with Apple     │                           │
     │───────────────────────────>│                           │
     │                            │                           │
     │                            │  2. Verify JWT with       │
     │                            │     Apple's public keys   │
     │                            │                           │
     │                            │  3. Validate nonce        │
     │                            │                           │
     │                            │  4. Check email verified  │
     │                            │                           │
     │                            │  5. UPSERT user           │
     │                            │──────────────────────────>│
     │                            │                           │
     │                            │  6. Return user data      │
     │                            │<──────────────────────────│
     │                            │                           │
     │  7. Return user info       │                           │
     │<───────────────────────────│                           │
     │                            │                           │
```

---

## Response Times

- Health Check: < 50ms (with DB ping < 100ms)
- Apple Sign In: 200-500ms (depends on Apple's token verification)
- Database Operations: < 50ms (timeout at 5000ms)

---

## CORS Configuration

Configured via `ALLOWED_ORIGINS` environment variable.

**Example:**
```env
ALLOWED_ORIGINS=http://localhost:3000,https://app.example.com
```

**Allowed Methods:**
- GET
- POST
- PUT
- DELETE
- PATCH
- OPTIONS

**Allowed Headers:**
- Content-Type
- Authorization
- Accept
- Origin
- X-Requested-With

---

## Notes

1. **Token Field**: Currently returns `null`. Implement your own JWT generation for session management.

2. **Email Updates**: If a user's email changes in their Apple account, it will be automatically updated in the database on next login.

3. **Concurrent Requests**: The UPSERT pattern handles concurrent login requests safely without race conditions.

4. **Error Messages**: Production environment returns generic error messages to prevent information disclosure.
