# Apple OAuth Backend

Clean, production-ready Go backend service for Apple OAuth authentication.

## Features

- ✅ Apple Sign In with ID Token verification
- ✅ Clean architecture (handlers, services, models)
- ✅ JWT token validation with Apple's public keys
- ✅ Nonce verification for security
- ✅ RESTful API design
- ✅ CORS support
- ✅ Health check endpoint

## Project Structure

```
.
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── handler/         # HTTP handlers
│   ├── service/         # Business logic
│   └── model/          # Data models
├── pkg/
│   ├── config/         # Configuration management
│   └── apple/          # Apple OAuth verification
├── .env.example        # Environment variables template
├── Makefile           # Build and run commands
└── README.md
```

## Prerequisites

- Go 1.24.7 or higher
- Apple Developer account with configured Sign In with Apple

## Configuration

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Update `.env` with your Apple OAuth credentials:
```env
SERVER_PORT=8080
APPLE_CLIENT_ID=your.apple.client.id
APPLE_TEAM_ID=your_team_id
```

## Installation

Install dependencies:
```bash
make install-deps
```

## Running the Application

### Development
```bash
make run
```

### Build
```bash
make build
./bin/server
```

## API Endpoints

### Health Check
```
GET /health
```

Response:
```json
{
  "status": "healthy"
}
```

### Apple Sign In
```
POST /api/v1/auth/apple
Content-Type: application/json
```

Request body:
```json
{
  "id_token": "eyJraWQiOiJXNldjT0tCIiwiYWxnIjoiUlMyNTYifQ...",
  "nonce": "random-nonce-from-client"
}
```

Response (Success - 200):
```json
{
  "user_id": "001234.abcdef1234567890.1234",
  "email": "user@example.com"
}
```

Response (Error - 401):
```json
{
  "error": "authentication_failed",
  "message": "failed to verify token: token expired"
}
```

## How It Works

1. **Frontend** sends Apple ID token and nonce to backend
2. **Backend** fetches Apple's public keys from `https://appleid.apple.com/auth/keys`
3. **Verification** process:
   - Validates JWT signature using Apple's public key
   - Checks token expiration
   - Verifies issuer is Apple
   - Validates audience matches your client ID
   - Confirms nonce matches
4. **Response** returns user information from verified token

## Security Features

- ID token signature verification with Apple's RSA public keys
- Nonce validation to prevent replay attacks
- Token expiration checking
- Issuer and audience validation
- Public key caching with 24-hour refresh

## Testing

Run tests:
```bash
make test
```

## Clean Code Principles

This project follows clean code principles:

- **Separation of Concerns**: Handlers, services, and models are separated
- **Dependency Injection**: Services are injected into handlers
- **Single Responsibility**: Each package has a clear purpose
- **Error Handling**: Proper error propagation and handling
- **Configuration Management**: Centralized config loading
- **Validation**: Input validation at handler level

## License

MIT