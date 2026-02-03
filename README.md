# SSO System

Complete, production-ready Single Sign-On system with Go backend and Nuxt 3 Frontends.

## Features
- **SSO Server**: Go + Fiber, OAuth2/OIDC, JWT, RBAC.
- **Login UI**: Custom Nuxt 3 embedded UI.
- **Admin Panel**: Separate Nuxt 3 dashboard for management.
- **Security**: 2FA (TOTP), Password History, Account Lockout, Audit Logs.
- **Architecture**: Headless Identity Provider.

## Development Setup
Check `sso-management/README.md` for specific UI dev instructions.

### Prerequisites
- Go 1.21+
- Node.js 18+ (22 Recommended)
- Docker & Docker Compose
- MySQL 8.0
- Redis 7.0

### Run Locally (Dev Mode)
1. **Database**: `docker-compose up -d` (in root)
2. **Server**: `cd sso-server && go run cmd/server/main.go`
3. **Login UI**: `cd sso-server/ui && npm run dev`
4. **Admin UI**: `cd sso-management && npm run dev` (Port 3002)
5. **Client**: `cd sso-client && npm run dev` (Port 3001)

## Production Deployment using Docker 🚀

We provide a specialized `docker-compose` setup for production.

### Architecture
- **sso-server**: Multi-stage build. Builds Login UI (Nuxt) -> Builds Go Binary -> Alpine Image serving both.
- **sso-management**: Nuxt generated static files served via Nginx.
- **sso-gateway**: Nginx reverse proxy.
- **Database**: MySQL & Redis containers.

### One-Command Deploy
```bash
./deploy-prod.sh
```
This script will:
1. Build all Docker images.
2. Start all services in detached mode.
3. Verify status.

### Access Points
- **Login/SSO**: http://localhost
- **Admin Panel**: http://localhost/admin-panel/
- **API**: http://localhost/api

### Directory Config
- `deploy/docker-compose.yml`: Main orchestration.
- `deploy/docker/`: Dockerfiles for each service.
- `deploy/nginx/`: Gateway configuration.
# sso
