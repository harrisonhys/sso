# 1. Build Stage
FROM node:22-alpine as builder

WORKDIR /app

# Copy package files
COPY sso-management/package.json sso-management/package-lock.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY sso-management .

# Build static files
RUN npm run generate

# 2. Production Stage
FROM nginx:alpine

# Copy static files to nginx html root
COPY --from=builder /app/.output/public /usr/share/nginx/html

# Copy custom nginx config if needed (optional)
# COPY deploy/docker/nginx/default.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
