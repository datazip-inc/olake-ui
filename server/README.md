# Olake Server
Olake Server is a RESTful API service built with the Beego framework that manages data sources, destinations, jobs, and users for data ingestion workflows.

## Prerequisites
- Go 1.23 or later
- PostgreSQL 12 or later
- Git

## Setup Instructions

### 1. Clone the Repository
    ```bash
    git clone https://github.com/datazip-inc/olake-frontend.git
    ```
### 2. Configure Application Settings (Auth only works when session enabled)

Review and update the configuration in `conf/app.conf` as needed:

```bash
appname = olake-server
httpport = 8080
runmode = dev
copyrequestbody = true
postgresdb = postgres://postgres:testing@testing-postgres.postgres.database.azure.com:5432/olakedb

# Session configuration
sessionon = true
```
#### If session enabled, then manually run following command on your postgres db
```bash
CREATE TABLE session (
    session_key VARCHAR(64) PRIMARY KEY,
    session_data BYTEA,  -- Critical for binary storage
    session_expiry TIMESTAMP WITH TIME ZONE
);
```
### 3. Run the Application

```bash
go mod tidy
bee run 
# or
make run
```

The server will start on port 8080 (or the port specified in your configuration).

## Project Structure

- **conf/** - Configuration files
- **internal/handlers/** - Request handlers for API endpoints
- **internal/models/** - Data models and database schema
- **routes/** - URL routing definitions
- **utils/** - Utility functions and helpers
- **main.go** - Application entry point

## API Endpoints
All API Endpoints: [Postman](https://solar-capsule-662043.postman.co/workspace/Olake-Server~ad9c900c-0376-42e2-adf2-e3137b92b325/collection/24907154-6eaf11b3-4e36-4ec3-a05a-3fa3720125ee?action=share&creator=24907154&active-environment=24907154-dcc91e95-6699-48cb-bbe0-e0e92b9800bd)
### Authentication

- POST `/login` - User login
- POST `/signup` - User registration
- GET `/auth/check` - Check authentication status

### Sources

- GET `/sources` - Get all sources
- POST `/sources` - Create a new source
- PUT `/sources/:id` - Update a source
- DELETE `/sources/:id` - Delete a source

### Destinations

- GET `/destinations` - Get all destinations
- POST `/destinations` - Create a new destination
- PUT `/destinations/:id` - Update a destination
- DELETE `/destinations/:id` - Delete a destination

### Jobs

- GET `/jobs` - Get all jobs
- POST `/jobs` - Create a new job
- PUT `/jobs/:id` - Update a job
- DELETE `/jobs/:id` - Delete a job

### Users

- GET `/users` - Get all users
- POST `/users` - Create a new user
- PUT `/users/:id` - Update a user
- DELETE `/users/:id` - Delete a user

## Development

### Running in Development Mode

The application runs in development mode by default, with ORM debugging enabled.

## Deployment

For production deployment:

1. Set `runmode = prod` in `conf/app.conf`
2. Configure a secure database connection
3. Set up proper CORS settings in `routers/router.go`
4. Enable authentication middleware
