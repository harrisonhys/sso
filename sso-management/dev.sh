#!/bin/bash

# SSO Management Development Server Starter
# This script ensures the correct Node.js version is used

echo "🚀 Starting SSO Management Development Server..."
echo ""

# Check if nvm is available
if [ -s "$HOME/.nvm/nvm.sh" ]; then
    echo "✓ Loading nvm..."
    source "$HOME/.nvm/nvm.sh"
    
    # Use Node.js 22
    echo "✓ Switching to Node.js 22..."
    nvm use 22
    
    # Verify version
    NODE_VERSION=$(node --version)
    echo "✓ Using Node.js $NODE_VERSION"
    echo ""
    
    # Start dev server
    echo "🎨 Starting Nuxt development server on port 3002..."
    npm run dev
else
    echo "⚠️  nvm not found!"
    echo "Current Node.js version: $(node --version)"
    echo ""
    echo "If you see errors, please run:"
    echo "  nvm use 22"
    echo "  npm run dev"
    echo ""
    npm run dev
fi
