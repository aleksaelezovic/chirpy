# Chirpy API Documentation

Chirpy is a Twitter-like social media platform API built in Go. It provides user management, authentication, chirp (message) creation/retrieval, and webhook integration.

## Table of Contents

- [Getting Started](#getting-started)
- [Authentication](#authentication)
- [API Endpoints](#api-endpoints)
  - [Health Check](#health-check)
  - [User Management](#user-management)
  - [Chirps](#chirps)
  - [Token Management](#token-management)
  - [Webhooks](#webhooks)
  - [Admin](#admin)
- [Data Models](#data-models)
- [Error Handling](#error-handling)

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL database
- Environment variables configured (see below)

### Environment Variables

Create a `.env` file with the following variables:

```bash
DB_URL=postgres://username:password@localhost:5432/chirpy
JWT_SECRET=your-secret-key-here
POLKA_KEY=your-polka-api-key
PLATFORM=dev  # Use "dev" for development, omit or set to "prod" for production
```

### Running the Server

```bash
go run .
```

Server starts on port `8080`.

## Authentication

Chirpy uses three authentication methods:

### 1. JWT Bearer Token
- **Used for**: User operations (create/delete chirps, update profile)
- **Format**: `Authorization: Bearer <jwt_token>`
- **Expiration**: 1 hour
- **Algorithm**: HS256

### 2. Refresh Token
- **Used for**: Obtaining new JWT tokens
- **Format**: `Authorization: Bearer <refresh_token>`
- **Expiration**: 60 days
- **Type**: 64-character hex string

### 3. API Key
- **Used for**: Webhook endpoints
- **Format**: `Authorization: ApiKey <api_key>`

## API Endpoints

Base URL: `http://localhost:8080`

### Health Check

#### GET /api/healthz

Check API health status.

**Response**
```
Status: 200 OK
Body: OK
```

---

### User Management

#### POST /api/users

Register a new user.

**Request Body**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response** (201 Created)
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2025-10-18T12:00:00Z",
  "updated_at": "2025-10-18T12:00:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

**Error Responses**
- `400`: Invalid email format or bad request
- `409`: Email already exists
- `500`: Internal server error

---

#### POST /api/login

Authenticate and receive tokens.

**Request Body**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response** (200 OK)
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2025-10-18T12:00:00Z",
  "updated_at": "2025-10-18T12:00:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "a1b2c3d4e5f6..."
}
```

**Error Responses**
- `400`: Bad request
- `401`: Incorrect email or password
- `500`: Internal server error

---

#### PUT /api/users

Update user credentials (email and/or password).

**Authentication**: Required (JWT)

**Headers**
```
Authorization: Bearer <jwt_token>
```

**Request Body**
```json
{
  "email": "newemail@example.com",
  "password": "newpassword123"
}
```

**Response** (200 OK)
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2025-10-18T12:00:00Z",
  "updated_at": "2025-10-18T12:00:00Z",
  "email": "newemail@example.com",
  "is_chirpy_red": false
}
```

**Error Responses**
- `400`: Bad request
- `401`: Unauthorized (missing or invalid token)
- `500`: Internal server error

---

### Chirps

#### POST /api/chirps

Create a new chirp.

**Authentication**: Required (JWT)

**Headers**
```
Authorization: Bearer <jwt_token>
```

**Request Body**
```json
{
  "body": "This is my chirp message"
}
```

**Constraints**
- Maximum 140 characters
- Profane words automatically censored: "kerfuffle", "sharbert", "fornax" → "****"

**Response** (201 Created)
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "body": "This is my chirp message",
  "created_at": "2025-10-18T12:00:00Z",
  "updated_at": "2025-10-18T12:00:00Z"
}
```

**Error Responses**
- `400`: Bad request or chirp too long
- `401`: Unauthorized
- `500`: Internal server error

---

#### GET /api/chirps

Retrieve all chirps with optional filtering and sorting.

**Query Parameters**
- `author_id` (optional): Filter by author UUID
- `sort` (optional): Set to `desc` for descending order (default: ascending by created_at)

**Examples**
```
GET /api/chirps
GET /api/chirps?sort=desc
GET /api/chirps?author_id=123e4567-e89b-12d3-a456-426614174000
GET /api/chirps?author_id=123e4567-e89b-12d3-a456-426614174000&sort=desc
```

**Response** (200 OK)
```json
[
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "body": "This is a chirp",
    "created_at": "2025-10-18T12:00:00Z",
    "updated_at": "2025-10-18T12:00:00Z"
  }
]
```

**Error Responses**
- `400`: Invalid author_id format
- `500`: Internal server error

---

#### GET /api/chirps/{id}

Retrieve a specific chirp by ID.

**Path Parameters**
- `id`: UUID of the chirp

**Response** (200 OK)
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "body": "This is a chirp",
  "created_at": "2025-10-18T12:00:00Z",
  "updated_at": "2025-10-18T12:00:00Z"
}
```

**Error Responses**
- `400`: Invalid UUID format
- `404`: Chirp not found
- `500`: Internal server error

---

#### DELETE /api/chirps/{id}

Delete a chirp (only the owner can delete).

**Authentication**: Required (JWT)

**Headers**
```
Authorization: Bearer <jwt_token>
```

**Path Parameters**
- `id`: UUID of the chirp to delete

**Response** (204 No Content)

**Error Responses**
- `400`: Invalid UUID
- `401`: Unauthorized (missing or invalid token)
- `403`: Forbidden (not the chirp owner)
- `404`: Chirp not found
- `500`: Internal server error

---

### Token Management

#### POST /api/refresh

Get a new JWT token using a refresh token.

**Headers**
```
Authorization: Bearer <refresh_token>
```

**Response** (200 OK)
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Error Responses**
- `401`: Invalid, expired, or revoked refresh token
- `500`: Internal server error

---

#### POST /api/revoke

Revoke (logout) a refresh token.

**Headers**
```
Authorization: Bearer <refresh_token>
```

**Response** (204 No Content)

**Error Responses**
- `401`: Unauthorized (missing or invalid token)
- `500`: Internal server error

---

### Webhooks

#### POST /api/polka/webhooks

Handle Polka payment events (e.g., user upgrades).

**Authentication**: Required (API Key)

**Headers**
```
Authorization: ApiKey <polka_api_key>
```

**Request Body**
```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000"
  }
}
```

**Supported Events**
- `user.upgraded`: Upgrades user to Chirpy Red status
- Other events: Ignored (returns 204)

**Response** (204 No Content)

**Error Responses**
- `400`: Bad request (invalid JSON)
- `401`: Unauthorized (missing or invalid API key)
- `404`: User not found
- `500`: Internal server error

---

### Admin

#### GET /admin/metrics

Display file server hit count (admin dashboard).

**Response** (200 OK)
```html
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited X times!</p>
  </body>
</html>
```

---

#### POST /admin/reset

Reset application state (development only).

**Restrictions**: Only available when `PLATFORM=dev`

**Response** (200 OK)
```
Reset successfully.
```

**Effects**
- Resets file server hit counter to 0
- Deletes all users and chirps

**Error Responses**
- `403`: Forbidden (not in dev environment)
- `500`: Internal server error

---

#### GET /app/*

Serve static files from the current directory.

---

## Data Models

### User

```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "email": "string",
  "is_chirpy_red": "boolean"
}
```

### Chirp

```json
{
  "id": "uuid",
  "user_id": "uuid",
  "body": "string (max 140 chars)",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Refresh Token

```json
{
  "token": "string (64 hex chars)",
  "user_id": "uuid",
  "expires_at": "timestamp",
  "revoked_at": "timestamp or null",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

---

## Error Handling

All error responses follow this format:

```json
{
  "error": "Error message describing what went wrong"
}
```

### Common HTTP Status Codes

- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `204 No Content`: Request successful, no content to return
- `400 Bad Request`: Invalid request format or parameters
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource already exists
- `500 Internal Server Error`: Server error

---

## Example Usage

### Complete User Flow

1. **Register**
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

2. **Login**
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

3. **Create Chirp**
```bash
curl -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt_token>" \
  -d '{"body":"Hello, Chirpy!"}'
```

4. **Get All Chirps**
```bash
curl http://localhost:8080/api/chirps
```

5. **Refresh Token**
```bash
curl -X POST http://localhost:8080/api/refresh \
  -H "Authorization: Bearer <refresh_token>"
```

6. **Logout (Revoke)**
```bash
curl -X POST http://localhost:8080/api/revoke \
  -H "Authorization: Bearer <refresh_token>"
```

---

## Security Features

- **Password Hashing**: Uses Argon2id algorithm
- **JWT Signing**: HS256 algorithm with configurable secret
- **Token Expiration**: JWT expires in 1 hour, refresh tokens in 60 days
- **Ownership Validation**: Users can only delete their own chirps
- **API Key Authentication**: Webhooks protected by API key
- **Profanity Filter**: Automatic censoring of inappropriate words

---

## Database

The application uses PostgreSQL with the following tables:
- `users`: User accounts
- `chirps`: User messages
- `refresh_tokens`: Authentication tokens

Database migrations are managed through SQL files in the `sql/schema/` directory.

---

## Development

### Project Structure

```
chirpy/
├── main.go                 # Server setup and routing
├── handlers.go             # Request handlers
├── helpers.go              # Helper functions
├── internal/
│   ├── auth/              # Authentication utilities
│   └── database/          # Database models and queries
└── sql/
    ├── schema/            # Database schema migrations
    └── queries/           # SQL queries
```

### Testing in Development

Use the reset endpoint to clear data between tests:

```bash
curl -X POST http://localhost:8080/admin/reset
```

Note: Only works when `PLATFORM=dev` is set.
