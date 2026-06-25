# Olake Frontend UI

React + TypeScript + Ant Design + Tailwind CSS + Vite

## Requirements

- A latest LTS version of [Node.js](https://nodejs.org/en/download/).
- [pnpm](https://pnpm.io/installation), a fast, disk-space-efficient package manager for Node.js.

## Running the project locally

All commands below are run from the `ui/` directory.

1. Clone the repository using SSH or HTTPS.
2. Install dependencies:

```bash
pnpm install
```

3. Start the development server:

```bash
pnpm dev
```

### Build and preview

```bash
pnpm build
pnpm preview
```

### Format code

```bash
pnpm format
pnpm format:check
```

## Checks before commit and push

Check ESLint issues:

```bash
pnpm lint
```

Fix ESLint issues:

```bash
pnpm lint:fix
```

## Troubleshooting

To clean `node_modules` and reinstall dependencies:

```bash
pnpx npkill
pnpm install
```

## Architecture overview

The app is organized into **core** (cross-cutting concerns), **modules** (feature domains), and **common** (shared UI and utilities).

- **Routing** — `src/app/router.tsx` defines public and protected routes. Feature routes are registered through the module registry in `src/core/modules/registry.ts`.
- **Modules** — Each domain module (`ingestion`, `maintenance`) exports an `AppModule` descriptor with nav config, routes, and optional feature gates.
- **Data fetching** — API calls live in `services/`. TanStack Query hooks in `hooks/queries/` and `hooks/mutations/` handle server state, caching, and cache invalidation.
- **Client state** — Zustand stores remain for UI/auth state that does not belong in the server cache.
- **Path alias** — `@/` resolves to `src/` (configured in `vite.config.ts`).

To add a new module, create `src/modules/<name>/index.ts`, register it in `src/core/modules/registry.ts`, and add a corresponding entry in `useActiveModuleKeys`.

## Folder structure

```text
ui/
├── public/                          # Static assets (e.g. favicon)
├── tests/                           # Playwright E2E tests (see tests/README.md)
│
├── src/
│   ├── app/                         # App shell and router
│   │   ├── App.tsx
│   │   └── router.tsx
│   │
│   ├── assets/                      # Images, icons, and SVGs
│   │
│   ├── common/                      # Shared across all modules
│   │   ├── components/              # Reusable UI (DataTable, modals, form widgets)
│   │   ├── constants/
│   │   ├── hooks/
│   │   ├── types/
│   │   └── utils/
│   │
│   ├── config/                      # App configuration (e.g. API base URL)
│   │
│   ├── core/                        # Cross-cutting application infrastructure
│   │   ├── analytics/
│   │   ├── api/                     # Axios instance and interceptors
│   │   ├── auth/                    # Login, auth store, auth services
│   │   ├── layout/                  # Shell, sidebar, nav config
│   │   ├── modules/                 # Module registry and route builder
│   │   ├── notifications/
│   │   ├── platform/                # Platform-wide state and services
│   │   └── settings/                # System settings pages and hooks
│   │
│   ├── lib/                         # Shared library code (e.g. Ant Design theme)
│   │
│   ├── modules/                     # Feature domains
│   │   ├── ingestion/               # Jobs, sources, destinations
│   │   │   ├── common/              # Shared ingestion components and services
│   │   │   └── features/
│   │   │       ├── jobs/
│   │   │       ├── sources/
│   │   │       └── destinations/    # Each feature typically contains:
│   │   │                            #   components/, hooks/, pages/, services/,
│   │   │                            #   stores/, types/, utils/, constants/
│   │   └── maintenance/             # Catalogs, tables, compaction
│   │       ├── common/
│   │       └── features/
│   │           ├── catalogs/
│   │           └── tables/
│   │
│   ├── providers/                   # React context providers (QueryClient, Ant Design)
│   │
│   ├── main.tsx                     # Application entry point
│   └── index.css                    # Global styles
│
├── index.html
├── package.json
├── eslint.config.js
├── tsconfig.json
├── tsconfig.app.json
├── tsconfig.node.json
├── tailwind.config.js
├── postcss.config.js
├── vite.config.ts
└── playwright.config.ts
```

## Used packages

- **React** / **React DOM**: UI library.
- **TypeScript**: Static typing.
- **Vite**: Build tool and dev server.
- **Tailwind CSS**: Utility-first CSS framework.
- **Ant Design**: Component library.
- **TanStack Query** (`@tanstack/react-query`): Server-state management, caching, and data synchronization.
- **Axios**: HTTP client for API requests.
- **Zustand**: Client-side state management for auth and UI state.
- **React Router DOM**: Client-side routing.
- **Phosphor Icons** (`@phosphor-icons/react`): Icon library.
- **React JSON Schema Form** (`@rjsf/*`): Dynamic form rendering from JSON Schema.
- **clsx**: Conditional class name utility.
- **date-fns**, **croner**, **semver**, **uuid**, **react-virtuoso**, **react-markdown**: Supporting utilities.
- **Playwright** (`@playwright/test`): End-to-end testing (dev dependency).
- **ESLint** / **Prettier**: Linting and formatting (dev dependencies).

## UX tips

### Suspense wrapper

Use `Suspense` from `react` for loading states on lazily loaded routes and components. The app wraps `RouterProvider` in a `Suspense` boundary with a shared loading screen.
 
