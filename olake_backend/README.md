# Olake Backend

A Beego-based backend service for the Olake data Ingestion platform.

## Project Overview

Olake Backend is a RESTful API service built with the Beego framework that manages data sources, destinations, jobs, and users for data ingestion workflows.

## Prerequisites

- Go 1.23 or later
- PostgreSQL 12 or later
- Git

## Setup Instructions

### 1. Clone the Repository

```bash
git clone https://github.com/datazip-inc/olake-frontend.git
cd olake_backend
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure PostgreSQL

1. Install PostgreSQL if not already installed
2. Create a database named `olakedb`:

```bash
createdb olakedb
```

3. Update the database connection string in `utils/db.go` if needed:

```go
connStr := "host=localhost port=5432 user=postgres password=12345678 dbname=olakedb sslmode=disable"
```

### 4. Configure Application Settings

Review and update the configuration in `conf/app.conf` as needed:

```
appname = olake_backend
httpport = 8080
runmode = dev
copyrequestbody = true

# Session configuration
sessionon = true
sessionprovider = memory
sessionname = olake_session
sessiongcmaxlifetime = 3600
```

### 5. Install Bee Tool

The Bee tool is required to run the application. Install it using:

```bash
go install github.com/beego/bee/v2@latest
```

### 6. Run the Application

```bash
bee run
```

The server will start on port 8080 (or the port specified in your configuration).

## Project Structure

- **conf/** - Configuration files
- **controllers/** - Request handlers for API endpoints
- **middleware/** - HTTP middleware components
- **models/** - Data models and database schema
- **routers/** - URL routing definitions
- **utils/** - Utility functions and helpers
- **main.go** - Application entry point

## API Endpoints

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
