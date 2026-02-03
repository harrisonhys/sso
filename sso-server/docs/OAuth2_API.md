# OAuth2 Endpoint Documentation

## Base URL
`http://localhost:3000`

## Authentication Flow Endpoints

### 1. Authorization Endpoint
**Endpoint:** `GET /oauth2/authorize`  
**Authentication:** Required (Session Token)  
**Description:** Initiates OAuth2 authorization code flow

**Query Parameters:**
- `response_type`: `code` (required)
- `client_id`: OAuth2 client identifier (required)
- `redirect_uri`: Callback URL for client (required)
- `scope`: Space-separated scope list (required) - e.g., `openid profile email`
- `state`: CSRF protection token (recommended)
- `code_challenge`: PKCE challenge (required for public clients)
- `code_challenge_method`: `S256` or `plain` (required for public clients)

**Response:** 
- If authenticated and consented: Redirects to `redirect_uri?code=AUTH_CODE&state=STATE`
- If not consented: Renders consent screen UI
- If not authenticated: Redirects to login

**Example:**
```bash
GET /oauth2/authorize?response_type=code&client_id=abc123&redirect_uri=https://app.example.com/callback&scope=openid%20profile&code_challenge=xyz...&code_challenge_method=S256&state=random_state
```

---

### 2. Consent Submission
**Endpoint:** `POST /oauth2/authorize/consent`  
**Authentication:** Required (Session Token)  
**Content-Type:** `application/x-www-form-urlencoded`

**Form Parameters:**
- `client_id`: OAuth2 client ID
- `redirect_uri`: Callback URL
- `scope`: Requested scopes
- `state`: CSRF token
- `code_challenge`: PKCE challenge
- `code_challenge_method`: PKCE method
- `approved`: `true` or `false`

**Response:**
- If approved: Redirects with authorization code
- If denied: Redirects with error

---

### 3. Token Endpoint
**Endpoint:** `POST /oauth2/token`  
**Authentication:** Client credentials (HTTP Basic or form body)  
**Content-Type:** `application/x-www-form-urlencoded`

#### Grant Type: Authorization Code
**Parameters:**
- `grant_type`: `authorization_code`
- `code`: Authorization code from /authorize
- `redirect_uri`: Must match authorization request
- `code_verifier`: PKCE verifier (for PKCE flow)
- `client_id`: Client identifier
- `client_secret`: Client secret (for confidential clients)

**Example:**
```bash
curl -X POST http://localhost:3000/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=AUTH_CODE_HERE" \
  -d "redirect_uri=https://app.example.com/callback" \
  -d "code_verifier=VERIFIER_STRING" \
  -d "client_id=abc123" \
  -d "client_secret=secret"
```

**Response:**
```json
{
  "access_token": "eyJhbGciOi...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "def456...",
  "scope": "openid profile email"
}
```

#### Grant Type: Client Credentials
**Parameters:**
- `grant_type`: `client_credentials`
- `scope`: Requested scopes (optional)
- `client_id`: Client identifier
- `client_secret`: Client secret

**Example:**
```bash
curl -X POST http://localhost:3000/oauth2/token \
  -u "client_id:client_secret" \
  -d "grant_type=client_credentials&scope=api:read"
```

**Response:**
```json
{
  "access_token": "eyJhbGciOi...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "api:read"
}
```

#### Grant Type: Refresh Token
**Parameters:**
- `grant_type`: `refresh_token`
- `refresh_token`: Valid refresh token
- `client_id`: Client identifier
- `client_secret`: Client secret

**Example:**
```bash
curl -X POST http://localhost:3000/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=def456..." \
  -d "client_id=abc123" \
  -d "client_secret=secret"
```

**Response:** New access token + new refresh token (rotation)

---

### 4. Token Revocation
**Endpoint:** `POST /oauth2/revoke`  
**Authentication:** Client credentials  
**Content-Type:** `application/x-www-form-urlencoded`

**Parameters:**
- `token`: Token to revoke (required)
- `token_type_hint`: `access_token` or `refresh_token` (optional)
- `client_id`: Client identifier
- `client_secret`: Client secret

**Example:**
```bash
curl -X POST http://localhost:3000/oauth2/revoke \
  -u "client_id:client_secret" \
  -d "token=eyJhbGciOi..." \
  -d "token_type_hint=access_token"
```

**Response:** HTTP 200 (always, per RFC 7009)

---

## Admin Endpoints

### 5. Register OAuth2 Client
**Endpoint:** `POST /admin/oauth2/clients`  
**Authentication:** Required (Session Token)  
**Content-Type:** `application/json`

**Request Body:**
```json
{
  "name": "My Application",
  "description": "Application description",
  "redirect_uris": ["https://app.example.com/callback"],
  "allowed_scopes": ["openid", "profile", "email"],
  "grant_types": ["authorization_code", "refresh_token"],
  "is_public": false
}
```

**Response:**
```json
{
  "client": {
    "id": "uuid-here",
    "client_id": "abc123",
    "name": "My Application",
    "redirect_uris": ["https://app.example.com/callback"],
    "is_active": true
  },
  "client_secret": "secret_only_shown_once",
  "message": "Client created successfully. Save the client_secret - it will not be shown again!"
}
```

---

### 6. Get Client Details
**Endpoint:** `GET /admin/oauth2/clients/:client_id`  
**Authentication:** Required (Session Token)

**Response:**
```json
{
  "id": "uuid",
  "client_id": "abc123",
  "name": "My Application",
  "description": "...",
  "redirect_uris": ["..."],
  "allowed_scopes": ["openid", "profile"],
  "grant_types": ["authorization_code"],
  "is_public": false,
  "is_active": true
}
```

---

### 7. Regenerate Client Secret
**Endpoint:** `POST /admin/oauth2/clients/:client_id/regenerate-secret`  
**Authentication:** Required (Session Token)

**Response:**
```json
{
  "client_secret": "new_secret_here",
  "message": "Client secret regenerated. Save it now - it will not be shown again!"
}
```

---

### 8. Revoke Client
**Endpoint:** `DELETE /admin/oauth2/clients/:client_id`  
**Authentication:** Required (Session Token)

**Response:**
```json
{
  "message": "Client revoked successfully"
}
```

---

## User Consent Management

### 9. Get User Consents
**Endpoint:** `GET /user/oauth2/consents`  
**Authentication:** Required (Session Token)

**Response:**
```json
[
  {
    "id": "uuid",
    "user_id": "user-uuid",
    "client_id": "abc123",
    "scopes": ["openid", "profile", "email"],
    "granted_at": "2026-01-27T12:00:00Z"
  }
]
```

---

### 10. Revoke User Consent
**Endpoint:** `DELETE /user/oauth2/consents/:client_id`  
**Authentication:** Required (Session Token)

**Response:**
```json
{
  "message": "Consent revoked successfully"
}
```

---

## Error Responses

All OAuth2 errors follow RFC 6749 format:

```json
{
  "error": "invalid_request",
  "error_description": "Missing required parameter: redirect_uri"
}
```

**Common Error Codes:**
- `invalid_request` - Malformed request
- `unauthorized_client` - Client not authorized for grant type
- `access_denied` - User denied consent
- `unsupported_response_type` - Invalid response_type
- `invalid_grant` - Invalid authorization code or refresh token
- `invalid_client` - Client authentication failed

---

## Security Notes

1. **PKCE is enforced for public clients** - must provide code_challenge and code_verifier
2. **Redirect URIs are strictly validated** - must match registered URIs exactly
3. **Client secrets are bcrypt hashed** - never store in plaintext
4. **Access tokens are JWT-signed** - verify signature on resource servers
5. **Refresh tokens rotate** - old refresh token is revoked on use
6. **Authorization codes are single-use** - expire in 10 minutes by default
7. **State parameter required** - for CSRF protection

---

## Configuration

Environment variables in `.env`:

```env
OAUTH2_AUTH_CODE_EXPIRY=10m
OAUTH2_ACCESS_TOKEN_EXPIRY=1h
OAUTH2_REFRESH_TOKEN_EXPIRY=720h
OAUTH2_ENFORCE_PKCE=true
OAUTH2_ISSUER=http://localhost:3000
```
