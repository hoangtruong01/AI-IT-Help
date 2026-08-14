# EOMP — Web Application

Frontend for Enterprise Operations Management Platform built with Nuxt 4, Vue 3, TypeScript, and Nuxt UI.

## Tech Stack

- **Framework**: Nuxt 4 (v4.5.2) + Vue 3
- **Language**: TypeScript (strict mode)
- **Styling**: Tailwind CSS v4 + Nuxt UI (v4.10)
- **State Management**: Pinia (v4)
- **Server State / Data Fetching**: TanStack Vue Query (v5)
- **Utilities**: VueUse
- **Code Quality**: ESLint + Prettier

## Project Structure

```
apps/web/
├── app/
│   ├── assets/        # Styles and static assets
│   ├── components/    # Reusable UI components
│   ├── composables/   # Composition functions (useApi, etc.)
│   ├── layouts/       # Dashboard shell and page layouts
│   ├── middleware/    # Route and permission guards
│   ├── pages/         # Application views and routing
│   ├── plugins/       # Plugins (TanStack Query, etc.)
│   ├── stores/        # Pinia state stores
│   ├── types/         # TypeScript definitions
│   └── utils/         # Helper functions
├── public/            # Static public assets
├── nuxt.config.ts     # Nuxt configuration
├── package.json       # Dependencies and scripts
└── tsconfig.json      # TypeScript configuration
```

## Development

```bash
# Install dependencies
pnpm install

# Start development server (http://localhost:3000)
pnpm dev

# Type check
pnpm typecheck

# Lint and check formatting
pnpm lint

# Build for production
pnpm build

# Preview production build
pnpm preview
```
