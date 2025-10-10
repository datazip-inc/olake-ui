# E2E Test Suite

This directory contains end-to-end tests for the OLake frontend application using Playwright.

## Structure

```
tests/
├── fixtures/           # Test fixtures and dependency injection
│   ├── base.fixture.ts         # Base fixtures with all page objects
│   ├── authenticated.fixture.ts # Authenticated fixture using saved state
│   └── index.ts                 # Central export point for fixtures
├── setup/              # Global setup scripts
│   └── auth.setup.ts            # Creates authentication state before tests
├── pages/              # Page Object Models (POM)
│   ├── BasePage.ts              # Abstract base class for all page objects
│   ├── LoginPage.ts             # Login page interactions
│   ├── SourcesPage.ts           # Sources listing page
│   ├── CreateSourcePage.ts      # Source creation flow
│   ├── EditSourcePage.ts        # Source editing flow
│   ├── DestinationsPage.ts      # Destinations listing page
│   ├── CreateDestinationPage.ts # Destination creation flow
│   ├── EditDestinationPage.ts   # Destination editing flow
│   ├── JobsPage.ts              # Jobs listing and sync operations
│   └── CreateJobPage.ts         # Job creation workflow
├── types/              # TypeScript type definitions
│   └── PageConfig.types.ts      # Form configs and test data interfaces
├── flows/              # Test scenarios organized by user flows
│   ├── login.spec.ts                # Login flow tests
│   ├── create-source.spec.ts        # Source creation tests
│   ├── edit-source.spec.ts          # Source editing tests
│   ├── create-destination.spec.ts   # Destination creation tests
│   ├── edit-destination.spec.ts     # Destination editing tests
│   ├── destination-end-to-end.spec.ts # Destination user journey
│   ├── create-job.spec.ts           # Job creation tests
│   ├── job-sync.spec.ts             # Job sync and execution tests
│   ├── job-end-to-end.spec.ts       # Complete job workflow tests
│   └── end-to-end.spec.ts           # Complete user journey tests
├── utils/              # Test utilities and helpers
│   ├── constants.ts                    # Shared constants and test data
│   ├── source-connector-configs.ts     # Source connector configurations
│   ├── destination-connector-configs.ts # Destination connector configs
│   ├── test-data-builder.ts            # Unique test data generator
│   ├── modal-utils.ts                  # Modal verification utilities
│   └── index.ts                        # Central export point for utils
└── README.md           # This file
```

## Design Principles

### Page Object Model (POM)

- Each page has its own class with locators and actions
- **All page objects extend `BasePage`** to inherit common functionality
- Methods are named descriptively (e.g., `fillSourceForm()`, `expectValidationError()`)
- Locators are defined once and reused throughout tests
- Actions are abstracted to focus on user intent, not implementation details
- **BasePage provides:** `goto()`, `expectVisible()`, `expectValidationError()`, `clickButton()`, etc.

### Test Organization

- **flows/**: Tests organized by complete user workflows (including login, source, destination, and job tests)
- Each test file focuses on a specific area of functionality
- Tests are independent and can run in any order
- All tests requiring authentication use the `authenticated.fixture.ts` for auto-login

### Clean Code Practices

- Abstracted repetitive logic into helper functions
- Used shared test data to avoid duplication
- Clear, descriptive test names that explain the scenario
- Consistent patterns across all test files
- Comprehensive error handling and validation

### Fixtures & Authentication

1. **Setup Project (`setup/auth.setup.ts`)**:
   - Runs once before all tests
   - Logs in with admin credentials
   - Saves authentication state to `.auth/user.json`

2. **Base Fixtures (`base.fixture.ts`)**:
   - Provides all page object instances automatically
   - No authentication state applied

3. **Authenticated Fixture (`authenticated.fixture.ts`)**:
   - Extends base fixtures
   - Applies saved authentication state using `test.use({ storageState })`
   - No login needed - tests start already authenticated

4. **Central Export (`fixtures/index.ts`)**:
   - Import fixtures from one place

#### Usage

```typescript
// For tests requiring authentication (sources, destinations, jobs)
import { test, expect } from "../fixtures/authenticated.fixture"

// For tests NOT requiring authentication (login flow)
import { test, expect } from "../fixtures/base.fixture"
```

## Key Features

### Login Flow Tests

- Valid/invalid credential handling
- Form validation (empty fields, short inputs)
- Keyboard navigation support
- Error message verification
- Form state management

### Source Creation Tests

- Step-by-step form completion
- MongoDB connector configuration
- Host, database, and credential management
- SSL toggle functionality
- Validation error handling
- Test connection flow
- Success/failure modal handling

### Source Editing Tests

- Source name updates
- Associated jobs viewing
- Save/cancel operations
- Confirmation dialogs
- Navigation between pages

### Destination Creation Tests

- Step-by-step form completion
- Amazon S3 and Apache Iceberg connectors
- Bucket, region, and path configuration
- Catalog selection for Iceberg
- Validation error handling
- Test connection flow
- Success/failure modal handling

### Destination Editing Tests

- Destination name updates
- Associated jobs viewing
- Config section viewing
- Save/cancel operations
- Confirmation dialogs
- Navigation between pages

### Job Creation Tests

- Step-by-step job configuration
- Source and destination selection (existing)
- Stream configuration and sync modes
- Job naming and frequency settings
- Multi-step form validation
- Stream selection and sync options
- Full Refresh + Incremental sync mode

### Job Sync Tests

- Job execution and sync operations
- Log viewing and monitoring
- Job configuration inspection
- Multiple sync handling
- Navigation between job views
- Real-time sync status updates

### End-to-End Tests

- Complete user journeys from login to task completion
- Error scenario handling
- Keyboard accessibility testing
- Data persistence verification

## Running Tests

```bash
# Run all tests
npx playwright test

# Run specific test file
npx playwright test tests/flows/login.spec.ts

# Run tests in headed mode
npx playwright test --headed

# Run tests with UI mode
npx playwright test --ui

# Generate test report
npx playwright show-report
```

## Test Data Management

### Constants (`utils/constants.ts`)

Shared test data and configuration:

- **Login Credentials**: `LOGIN_CREDENTIALS` with admin/user roles
- **Source Configurations**: `POSTGRES_SOURCE_CONFIG`, `MONGODB_SOURCE_CONFIG`, etc.
- **Destination Configurations**: `ICEBERG_JDBC_CONFIG`, `S3_CONFIG`, etc.
- **Job Configuration**: `JOB_CONFIG` (streams, frequency, sync modes)
- **Validation Messages**: `VALIDATION_MESSAGES`
- **URL Constants**: `URLS`

### Test Data Builder (`utils/test-data-builder.ts`)

Generates unique test data to avoid conflicts:

```typescript
// Generate unique names with timestamps
TestDataBuilder.getUniqueSourceName("postgres")
// => "e2e_postgres_source_1234567890_123"

TestDataBuilder.getUniqueDestinationName("iceberg")
// => "e2e_iceberg_dest_1234567890_456"

TestDataBuilder.getUniqueJobName("postgres", "iceberg", "jdbc")
// => "postgres_iceberg_jdbc_job_1234567890_789"
```

### Connector Configurations

- **`source-connector-configs.ts`**: Helper functions to create source configs
  - `createPostgresSourceConfig()`, `createMongoDBSourceConfig()`, etc.
- **`destination-connector-configs.ts`**: Helper functions to create destination configs
  - `createIcebergJdbcConfig()`, `createS3Config()`, etc.

### Type Safety (`types/PageConfig.types.ts`)

TypeScript interfaces for type-safe test data:

- `SourceFormConfig`: Source creation form structure
- `DestinationFormConfig`: Destination creation form structure
- `JobFormConfig`: Job creation form structure
