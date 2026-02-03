# SSO Management - UI/UX Redesign

## Quick Start

### Running the Development Server

**Important**: This project requires **Node.js 18+**. Your system currently uses Node.js v16 by default (from Herd).

#### Option 1: Use the Helper Script (Recommended)
```bash
./dev.sh
```

#### Option 2: Manual Setup
```bash
# Switch to Node.js 22
nvm use 22

# Start dev server
npm run dev
```

#### Option 3: Set Node.js 22 as Default
```bash
# Set default Node version
nvm alias default 22

# Restart terminal, then:
npm run dev
```

### Accessing the Application

Once running, open your browser to:
- **URL**: http://localhost:3002
- **Port**: 3002 (configured to avoid conflicts)

## What's New in the UI

### 🎨 Design System
- Custom CSS with design tokens (colors, gradients, shadows, transitions)
- Enhanced Tailwind configuration
- Professional color palette and typography
- Smooth animations and micro-interactions

### 🧩 Component Library
7 reusable components created:
1. **Card** - Variants: default, gradient, bordered
2. **Button** - 5 variants, 3 sizes, loading state
3. **Badge** - 6 variants with pulse animation
4. **Modal** - 4 sizes, backdrop blur, ESC support
5. **Icon** - 30+ SVG icons (Heroicons style)
6. **StatCard** - Animated count-up numbers, trend indicators
7. **Layout** - Modern sidebar with collapsible navigation

### 📊 Dashboard Enhancements
- **Welcome Banner**: Gradient card with time-based greeting
- **Stat Cards**: Animated numbers with trend indicators
- **Quick Actions**: Modern card-based layout with hover effects
- **Recent Activity**: Enhanced table with user avatars and status badges

### 🎯 Key Features
- ✅ No more emoji icons - professional SVG icons
- ✅ Gradient backgrounds and modern color scheme
- ✅ Smooth animations (count-up, slide-in, fade-in)
- ✅ Better visual hierarchy and typography
- ✅ Responsive design for mobile
- ✅ Loading states and empty states
- ✅ User profile section in sidebar

## Project Structure

```
sso-management/
├── assets/
│   └── css/
│       └── main.css              # Custom CSS with design tokens
├── components/
│   ├── ui/
│   │   ├── Card.vue
│   │   ├── Button.vue
│   │   ├── Badge.vue
│   │   ├── Modal.vue
│   │   └── Icon.vue
│   └── stats/
│       └── StatCard.vue
├── layouts/
│   └── default.vue               # Redesigned layout
├── pages/
│   └── index.vue                 # Enhanced dashboard
├── dev.sh                        # Helper script to start dev server
└── package.json
```

## Troubleshooting

### Error: "Cannot read properties of undefined (reading 'prototype')"
This error occurs when using Node.js v16. Solution:
```bash
nvm use 22
npm run dev
```

### Port 3002 Already in Use
```bash
# Kill the process using port 3002
lsof -ti:3002 | xargs kill -9

# Or use a different port
npm run dev -- --port 3003
```

### Clear Cache and Reinstall
```bash
rm -rf node_modules .nuxt package-lock.json
npm install
npm run dev
```

## Development

### Available Scripts
- `npm run dev` - Start development server on port 3002
- `npm run build` - Build for production
- `npm run preview` - Preview production build

### Environment Variables
Create a `.env` file:
```env
API_BASE_URL=http://localhost:8080
```

## Documentation

For detailed information about the UI redesign, see:
- [UI Redesign Walkthrough](/.gemini/antigravity/brain/6361ebd4-a48d-4bbb-a43b-d851c5951dd0/ui-redesign-walkthrough.md)

## Next Steps

### Pages to Redesign (Optional)
- [ ] Users page - Enhanced table with filters
- [ ] Roles page - Card-based layout
- [ ] Permissions page - Grouped display
- [ ] OAuth2 Clients - Card layout with copy buttons
- [ ] Audit Logs - Timeline view option
- [ ] Settings - Tabbed interface

### Additional Features
- [ ] Dark mode toggle
- [ ] Global search functionality
- [ ] Advanced filtering
- [ ] Export to CSV/JSON
- [ ] Notification system

---

**Note**: Make sure to use Node.js 22 before running the development server!
