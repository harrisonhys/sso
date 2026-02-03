# SSO System - Concept Document

## 1. Project Overview

Sistem Single Sign-On (SSO) yang memungkinkan pengguna untuk mengakses multiple aplikasi dengan satu kali autentikasi. Sistem ini terdiri dari tiga komponen utama yang saling terintegrasi untuk menyediakan layanan autentikasi dan otorisasi yang aman dan terpusat.

### 1.1 Tujuan Sistem
- Menyediakan autentikasi terpusat untuk multiple aplikasi
- Meningkatkan keamanan dengan implementasi 2FA (Two-Factor Authentication)
- Memudahkan manajemen user, role, dan permission
- Menyediakan self-service untuk reset password dan recovery
- Memenuhi standar keamanan dengan password complexity dan expiry policy

### 1.2 Scope
Sistem ini mencakup:
- **SSO Server**: Server autentikasi utama dengan login page dan OAuth2 implementation
- **SSO Client**: Aplikasi client yang menggunakan SSO untuk autentikasi
- **SSO Management**: Dashboard untuk administrasi user, role, permission, dan konfigurasi

## 2. Komponen Sistem

### 2.1 SSO Server
**Teknologi Stack:**
- Backend: Go + Fiber Framework
- Database: MySQL
- Protocol: OAuth2

**Fitur Utama:**
- Login page dengan username/password
- Two-Factor Authentication (TOTP)
- OAuth2 Authorization Server
- Token management (Access Token & Refresh Token)
- Session management
- Password reset flow
- Forgot password flow
- Email notification service

**Tanggung Jawab:**
- Autentikasi user credentials
- Validasi 2FA code
- Generate dan validasi OAuth2 tokens
- Redirect ke aplikasi client setelah autentikasi berhasil
- Enforce password policy (complexity, expiry)

### 2.2 SSO Client
**Teknologi Stack:**
- Frontend: Nuxt.js
- Styling: Tailwind CSS

**Fitur Utama:**
- OAuth2 client implementation
- Protected routes
- Token refresh mechanism
- Logout functionality
- User profile display

**Tanggung Jawab:**
- Redirect ke SSO Server untuk autentikasi
- Menerima dan menyimpan OAuth2 tokens
- Protect application routes
- Handle token expiry dan refresh

### 2.3 SSO Management
**Teknologi Stack:**
- Backend: Go + Fiber Framework
- Frontend: Nuxt.js
- Database: MySQL (shared dengan SSO Server)
- Styling: Tailwind CSS

**Fitur Utama:**
- User Management (CRUD)
- Role Management (CRUD)
- Permission Management (CRUD)
- Client Application Management
- System Configuration:
  - Password expiry settings
  - Password complexity rules
  - 2FA enforcement policy
  - Session timeout settings
  - Audit log viewing

**Tanggung Jawab:**
- Administrasi data user, role, dan permission
- Konfigurasi security policy
- Monitoring aktivitas autentikasi
- Manajemen registered OAuth2 clients

## 3. Fitur Keamanan

### 3.1 Two-Factor Authentication (2FA)
- **Method**: TOTP (Time-based One-Time Password)
- **Compatible Apps**: Google Authenticator, Microsoft Authenticator, Authy
- **Setup Flow**:
  1. User enable 2FA dari profile settings
  2. System generate QR code untuk TOTP secret
  3. User scan QR code dengan authenticator app
  4. User input kode verifikasi untuk confirm setup
  5. System generate backup codes untuk recovery
- **Login Flow dengan 2FA**:
  1. User input username & password
  2. System validasi credentials
  3. Jika valid, tampilkan form input 2FA code
  4. User input code dari authenticator app
  5. System validasi TOTP code
  6. Jika valid, proceed dengan OAuth2 flow

### 3.2 Password Reset
**User-Initiated Reset (when logged in):**
1. User akses halaman change password
2. Input current password
3. Input new password (validate complexity)
4. Confirm new password
5. System update password dan send email notification

**Admin-Initiated Reset:**
1. Admin force reset password dari management dashboard
2. System generate temporary password
3. Send email ke user dengan temporary password
4. User login dengan temporary password
5. System force user untuk change password

### 3.3 Forgot Password
**Self-Service Recovery:**
1. User klik "Forgot Password" di login page
2. Input email address
3. System send reset link via email (valid 1 hour)
4. User klik link di email
5. Redirect ke reset password page
6. Input new password (validate complexity)
7. System update password
8. Redirect ke login page

**Security Measures:**
- Reset token adalah single-use
- Token expire setelah 1 jam
- Rate limiting untuk prevent abuse
- Email harus verified

### 3.4 Password Policy Configuration
**Configurable Parameters:**
- **Password Complexity**:
  - Minimum length (default: 8 characters)
  - Require uppercase (default: yes)
  - Require lowercase (default: yes)
  - Require numbers (default: yes)
  - Require special characters (default: yes)
  - Prevent common passwords
  - Prevent password reuse (last N passwords)

- **Password Expiry**:
  - Password lifetime (default: 90 days)
  - Grace period untuk reset (default: 7 days)
  - Warning notification (default: 7 days before expiry)

## 4. OAuth2 Flow

### 4.1 Authorization Code Flow
```
Client App -> SSO Server: Authorization Request
SSO Server -> User: Login Page
User -> SSO Server: Credentials + 2FA
SSO Server -> Client App: Authorization Code (redirect)
Client App -> SSO Server: Exchange Code for Token
SSO Server -> Client App: Access Token + Refresh Token
Client App -> Resource: Access Protected Resources
```

### 4.2 Token Types
- **Access Token**: JWT dengan expiry 15 menit
- **Refresh Token**: Opaque token dengan expiry 7 hari
- **ID Token**: JWT berisi user info

### 4.3 Token Claims
```json
{
  "sub": "user_id",
  "email": "user@example.com",
  "name": "User Name",
  "roles": ["user", "admin"],
  "permissions": ["read:profile", "write:profile"],
  "client_id": "client_app_id",
  "exp": 1234567890,
  "iat": 1234567890
}
```

## 5. Data Model Overview

### 5.1 Core Entities
- **Users**: Data pengguna dengan credentials
- **Roles**: Peran pengguna dalam sistem
- **Permissions**: Hak akses spesifik
- **Clients**: Aplikasi yang terdaftar untuk OAuth2
- **Sessions**: Active user sessions
- **AuditLogs**: Log aktivitas untuk security tracking

### 5.2 Relationships
- User ↔ Role: Many-to-Many
- Role ↔ Permission: Many-to-Many
- User ↔ Session: One-to-Many
- User ↔ TwoFactorAuth: One-to-One
- User ↔ PasswordResetToken: One-to-Many

## 6. Deployment Architecture

### 6.1 Docker Containers
```
- sso-server (Go + Fiber)
- sso-client (Nuxt.js)
- sso-management-api (Go + Fiber)
- sso-management-ui (Nuxt.js)
- mysql-db
- nginx (reverse proxy)
```

### 6.2 Network Flow
```
Internet -> Nginx -> SSO Server (Port 3000)
                  -> SSO Client (Port 3001)
                  -> SSO Management UI (Port 3002)
                  -> SSO Management API (Port 3003)
                  
SSO Server -> MySQL (Port 3306)
SSO Management API -> MySQL (Port 3306)
```

## 7. Security Considerations

### 7.1 Data Protection
- Password hashing: bcrypt dengan cost 12
- Encryption di transit: TLS/SSL
- Secure cookie flags: HttpOnly, Secure, SameSite
- TOTP secret encryption di database

### 7.2 Attack Prevention
- Rate limiting untuk login attempts
- CSRF protection
- XSS prevention
- SQL injection prevention (prepared statements)
- Brute force protection dengan account lockout

### 7.3 Session Management
- Session timeout: Configurable (default 30 menit)
- Secure session storage
- Single logout across applications
- Force logout on password change

## 8. Integration Points

### 8.1 Email Service
- Password reset emails
- Password expiry warnings
- Account lockout notifications
- 2FA setup confirmations
- Login alerts (optional)

### 8.2 Audit Logging
- Login attempts (success/failure)
- Password changes
- 2FA enable/disable
- Permission changes
- Configuration changes

## 9. User Experience

### 9.1 Login Flow (Normal User)
1. User mengakses client application
2. Redirect ke SSO Server login page
3. Input username & password
4. Jika 2FA enabled: Input TOTP code
5. Redirect kembali ke client application
6. Access granted

### 9.2 First-Time Setup (New User)
1. Admin create user di management dashboard
2. User receive email dengan temporary password
3. User login dengan temporary password
4. Force change password
5. Option untuk enable 2FA
6. Setup complete

### 9.3 Admin Workflow
1. Login ke management dashboard
2. Manage users, roles, permissions
3. Configure system settings
4. Monitor audit logs
5. Handle support requests

## 10. Future Enhancements (Out of Scope - Phase 1)

- Social login (Google, GitHub, etc.)
- Biometric authentication
- Push notification untuk 2FA
- Advanced analytics dashboard
- Multi-tenancy support
- API rate limiting dashboard
- Passwordless authentication (WebAuthn)

---

**Document Version**: 1.0  
**Last Updated**: 2026-01-26  
**Author**: Development Team
