# SSO Client Demo

Demo client application for SSO Server integration using OAuth2 with PKCE.

## Features

- OAuth2 Authorization Code Flow with PKCE
- Modern UI with Tailwind CSS
- Token management (access & refresh tokens)
- User profile display
- Secure authentication

## Setup

1. Install dependencies:
```bash
npm install
```

2. Copy environment variables:
```bash
cp .env.example .env
```

3. Update `.env` with your SSO server configuration

4. Run development server:
```bash
npm run dev
```

The application will be available at `http://localhost:3001`

## Environment Variables

- `SSO_SERVER_URL`: URL of the SSO server (default: http://localhost:3000)
- `CLIENT_ID`: OAuth2 client ID registered on SSO server
- `CLIENT_SECRET`: OAuth2 client secret (optional for PKCE)
- `REDIRECT_URI`: Callback URL after authentication (default: http://localhost:3001/callback)

## Usage

1. Click "Sign in with SSO" button
2. You will be redirected to SSO server login page
3. Enter your credentials
4. Complete 2FA if enabled
5. Grant consent for the application
6. You will be redirected back to the client app with authentication complete

## Build for Production

```bash
npm run build
npm run preview
```
