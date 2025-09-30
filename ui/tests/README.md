# E2E Test Suite

This directory contains end-to-end tests for the OLake frontend application using Playwright.

## Structure

```
tests/
├── fixtures/           # Test fixtures and setup
│   └── auth.fixture.ts  # Authentication and page object fixtures
├── pages/              # Page Object Models
│   ├── LoginPage.ts     # Login page interactions
│   ├── SourcesPage.ts   # Sources listing page
│   ├── CreateSourcePage.ts # Source creation flow
│   ├── EditSourcePage.ts   # Source editing flow
│   ├── DestinationsPage.ts # Destinations listing page
│   ├── CreateDestinationPage.ts # Destination creation flow
│   ├── EditDestinationPage.ts   # Destination editing flow
│   ├── JobsPage.ts      # Jobs listing and sync operations
│   └── CreateJobPage.ts # Job creation workflow
├── flows/              # Test scenarios organized by user flows
│   ├── login.spec.ts    # Login flow tests
│   ├── create-source.spec.ts # Source creation tests
│   ├── edit-source.spec.ts   # Source editing tests
│   ├── create-destination.spec.ts # Destination creation tests
│   ├── edit-destination.spec.ts   # Destination editing tests
│   ├── destination-end-to-end.spec.ts # Destination user journey tests
│   ├── create-job.spec.ts    # Job creation tests
│   ├── job-sync.spec.ts      # Job sync and execution tests
│   ├── job-end-to-end.spec.ts # Complete job workflow tests
│   └── end-to-end.spec.ts    # Complete user journey tests
├── auth/               # Authentication-specific tests
│   └── login.spec.ts    # Detailed login validation tests
├── setup/              # Test configuration and data
│   └── test-data.ts     # Shared test data and constants
├── utils/              # Test utilities and helpers
│   └── test-helpers.ts  # Common test helper functions
└── README.md           # This file
```

## Design Principles

### Page Object Model (POM)

- Each page has its own class with locators and actions
- Methods are named descriptively (e.g., `fillSourceName()`, `expectValidationError()`)
- Locators are defined once and reused throughout tests
- Actions are abstracted to focus on user intent, not implementation details

### Test Organization

- **flows/**: Tests organized by complete user workflows
- **auth/**: Detailed authentication and authorization tests
- Each test file focuses on a specific area of functionality
- Tests are independent and can run in any order

### Clean Code Practices

- Abstracted repetitive logic into helper functions
- Used shared test data to avoid duplication
- Clear, descriptive test names that explain the scenario
- Consistent patterns across all test files
- Comprehensive error handling and validation

### Fixtures

- Custom Playwright fixtures provide page objects automatically
- Fixtures handle page setup and teardown
- Shared authentication state and test isolation

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

## Test Data

Shared test data is defined in `setup/test-data.ts`:

- User credentials
- MongoDB connection configuration
- Amazon S3 connection configuration
- Job configuration (streams, frequency, sync modes)
- Validation messages
- URL constants
- CSS selectors

## Helper Functions

Common operations are abstracted in `utils/test-helpers.ts`:

- Admin login shortcut
- Test data generation (sources, destinations, jobs)
- Modal waiting utilities
- Form interaction helpers
- Dropdown selection utilities

## Best Practices

1. **Use Page Objects**: Always interact with pages through Page Object Models
2. **Wait for Elements**: Use `expect()` and `waitFor()` for reliable element interaction
3. **Unique Test Data**: Generate unique names/IDs to avoid test conflicts
4. **Independent Tests**: Each test should set up its own data and clean up
5. **Descriptive Names**: Test names should clearly describe the scenario
6. **Error Scenarios**: Include both happy path and error scenarios
7. **Accessibility**: Test keyboard navigation and screen reader compatibility

## Example Test

```typescript
test("should create MongoDB source successfully", async ({
  loginPage,
  sourcesPage,
  createSourcePage
}) => {
  // Login
  await loginPage.goto()
  await loginPage.login("admin", "password")
  await loginPage.waitForLogin()

  // Navigate to create source
  await sourcesPage.navigateToSources()
  await sourcesPage.clickCreateSource()

  // Fill form and create
  await createSourcePage.fillMongoDBForm({
    name: "test-source",
    host: "localhost:27017",
    database: "testdb",
    username: "admin",
    password: "password"
  })

  await createSourcePage.clickCreate()

  // Verify success
  await createSourcePage.expectSuccessModal()
  await sourcesPage.expectSourceExists("test-source")
})
```

This structure provides maintainable, reliable tests that focus on user workflows while keeping the code clean and reusable.
