# Authentication & User Management

Complete guide to LunaSentri's authentication system and user management features.

---

## Overview

LunaSentri implements a robust JWT-based authentication system with role-based access control, secure password management, and password reset functionality.

---

## Authentication System

### JWT-Based Sessions

**Features:**

- Stateless authentication with JSON Web Tokens
- Configurable token expiry (default: 15 minutes)
- HTTPOnly cookies for security
- Automatic token validation on protected endpoints

**Configuration:**

```bash
AUTH_JWT_SECRET="your-secret-key-min-32-chars"  # Required
ACCESS_TOKEN_TTL="15m"                          # Optional (default: 15m)
SECURE_COOKIE=true                              # Required in production
```

### Secure Cookies

**Production Settings:**

- `HTTPOnly` - Prevents JavaScript access
- `Secure` - HTTPS only (when `SECURE_COOKIE=true`)
- `SameSite=Lax` - CSRF protection
- Path-scoped to backend API

**Development Settings:**

- `SECURE_COOKIE=false` allows HTTP for local development
- Still uses HTTPOnly for security

---

## User Roles

### Admin Users

**Capabilities:**

- Full access to all features
- User management (create, delete users)
- Alert rule management
- Notification configuration
- System monitoring

**Admin Creation Methods:**

1. **Environment Variables** (Recommended for production)

   ```bash
   ADMIN_EMAIL="admin@example.com"
   ADMIN_PASSWORD="secure-password-here"
   ```

   - User created/updated on server start
   - Password updated if email exists

2. **First Registration** (Development)
   - First user to register automatically becomes admin
   - Navigate to `/register` and create account
   - Subsequent users are regular users

3. **Database Access** (Manual)

   ```bash
   # SSH into container
   docker exec -it <container> /bin/sh
   
   # Open database
   sqlite3 /app/data/lunasentri.db
   
   # Promote user to admin
   UPDATE users SET is_admin = 1 WHERE email = 'user@example.com';
   ```

### Regular Users

**Capabilities:**

- View dashboard and metrics
- Manage own notification settings (Telegram, Webhooks)
- View alert events
- Change own password
- No access to user management or alert rules

---

## Password Management

### Password Requirements

- Minimum length: 8 characters
- Hashed using bcrypt (cost factor: 10)
- Never stored in plain text
- Cannot be retrieved (only reset)

### Password Reset Flow

1. **Request Reset:**
   - User enters email at `/auth/forgot-password`
   - System generates unique reset token
   - Token expires after 1 hour (configurable)

2. **Reset Token:**

   ```bash
   PASSWORD_RESET_TTL="1h"  # Default: 1 hour
   ```

   - Cryptographically secure random token
   - Stored hashed in database
   - Single-use only

3. **Reset Password:**
   - User receives token (via notification or direct link)
   - Navigates to `/auth/reset-password?token=...`
   - Enters new password
   - Token invalidated after use

### Change Password

**Authenticated users can change their password:**

- Navigate to Settings
- Enter current password
- Enter new password (min 8 chars)
- Submit to update

---

## API Endpoints

### Public Endpoints (No Auth Required)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/auth/register` | Register new user |
| `POST` | `/auth/login` | Login and get session |
| `POST` | `/auth/logout` | Logout and clear session |
| `POST` | `/auth/forgot-password` | Request password reset |
| `POST` | `/auth/reset-password` | Reset password with token |

### Protected Endpoints (Auth Required)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/auth/me` | Get current user profile |
| `POST` | `/auth/change-password` | Change password |

### Admin Endpoints (Admin Role Required)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/auth/users` | List all users |
| `POST` | `/auth/users` | Create new user |
| `DELETE` | `/auth/users/:id` | Delete user |

---

## Authentication Flow

### Registration

```bash
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response (201 Created):**

```json
{
  "id": 1,
  "email": "user@example.com",
  "is_admin": true,  // true if first user
  "created_at": "2025-10-09T12:00:00Z"
}
```

**Cookie Set:**

```
Set-Cookie: access_token=<jwt_token>; HttpOnly; Secure; Path=/
```

### Login

```bash
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response (200 OK):**

```json
{
  "access_token": "<jwt_token>",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "is_admin": true
  }
}
```

**Cookie Set:**

```
Set-Cookie: access_token=<jwt_token>; HttpOnly; Secure; Path=/
```

### Accessing Protected Endpoints

**Option 1: Cookie (Automatic)**

```bash
# Cookie sent automatically by browser
GET /auth/me
```

**Option 2: Authorization Header**

```bash
GET /auth/me
Authorization: Bearer <jwt_token>
```

### Logout

```bash
POST /auth/logout
```

**Response (200 OK):**

```
Cookie cleared, session invalidated
```

---

## Security Features

### Password Security

- ✅ Bcrypt hashing with cost factor 10
- ✅ Minimum 8 character requirement
- ✅ Never stored in plain text
- ✅ Password reset with time-limited tokens
- ✅ Current password required for password change

### Session Security

- ✅ JWT with configurable expiry
- ✅ HTTPOnly cookies prevent XSS
- ✅ Secure flag for HTTPS
- ✅ SameSite protection against CSRF
- ✅ Token validation on every request

### Access Control

- ✅ Role-based permissions (admin vs user)
- ✅ Middleware enforces authentication
- ✅ Admin endpoints protected by role check
- ✅ Users isolated (can only access own data)

---

## User Management

### Admin User Management

**List All Users:**

```bash
GET /auth/users
```

**Response:**

```json
[
  {
    "id": 1,
    "email": "admin@example.com",
    "is_admin": true,
    "created_at": "2025-10-09T12:00:00Z"
  },
  {
    "id": 2,
    "email": "user@example.com",
    "is_admin": false,
    "created_at": "2025-10-09T13:00:00Z"
  }
]
```

**Create User:**

```bash
POST /auth/users
Content-Type: application/json

{
  "email": "newuser@example.com",
  "password": "password123",
  "is_admin": false
}
```

**Delete User:**

```bash
DELETE /auth/users/2
```

**Response (204 No Content)**

---

## Best Practices

### Production Deployment

1. **Set Strong JWT Secret:**

   ```bash
   openssl rand -base64 32
   ```

2. **Enable Secure Cookies:**

   ```bash
   SECURE_COOKIE=true
   ```

3. **Use HTTPS:**
   - Required for secure cookies
   - Protects credentials in transit

4. **Configure Admin Account:**

   ```bash
   ADMIN_EMAIL="admin@yourdomain.com"
   ADMIN_PASSWORD="<strong-password>"
   ```

5. **Rotate Passwords Regularly:**
   - Change admin password periodically
   - Regenerate JWT secret if compromised

### Development

1. **Use Test Credentials:**

   ```bash
   ADMIN_EMAIL="admin@test.com"
   ADMIN_PASSWORD="admin123"
   SECURE_COOKIE=false
   ```

2. **Reset Database for Fresh Start:**

   ```bash
   ./scripts/dev-reset.sh --reset-db
   ```

3. **Test Authentication Flow:**
   - Register first user (becomes admin)
   - Login and verify cookie set
   - Access protected endpoints
   - Test password reset flow

---

## Troubleshooting

### "401 Unauthorized" Errors

**Possible Causes:**

- Token expired (default: 15 minutes)
- Invalid JWT secret
- Cookie not sent (CORS issue)
- Token not in cookie or Authorization header

**Solutions:**

- Check `ACCESS_TOKEN_TTL` configuration
- Verify `AUTH_JWT_SECRET` is set and matches
- Check CORS configuration allows credentials
- Ensure cookie or header contains valid token

### "First User Not Admin"

**Possible Causes:**

- Database already has users
- Environment variables not loaded

**Solutions:**

- Reset database: `./scripts/dev-reset.sh --reset-db`
- Verify `ADMIN_EMAIL` and `ADMIN_PASSWORD` are set
- Manually promote user in database

### "Password Reset Token Invalid"

**Possible Causes:**

- Token expired (1 hour TTL)
- Token already used
- Token not found in database

**Solutions:**

- Request new reset token
- Check `PASSWORD_RESET_TTL` configuration
- Ensure token matches exactly

---

## Summary

LunaSentri's authentication system provides:

- ✅ Secure JWT-based sessions with HTTPOnly cookies
- ✅ Role-based access control (admin vs regular users)
- ✅ Bcrypt password hashing
- ✅ Password reset with time-limited tokens
- ✅ First-user admin promotion
- ✅ Environment variable admin bootstrap
- ✅ User management API for admins
- ✅ Production-ready security features
- ✅ CORS support for authenticated requests
- ✅ Session expiry with configurable TTL

**Status: Production Ready** 🔐
