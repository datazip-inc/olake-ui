# Olake Backend API Contract

## Base URL

```
http://localhost:8080
```

## Authentication

### Login

- **Endpoint**: `/login`
- **Method**: POST
- **Description**: Authenticate user and get access token
- **Request Body**:
  ```json
  {
    "email": "string",
    "password": "string"
  }
  ```
- **Response**:
  ```json
  {
    "token": "string",
    "user": {
      "id": "integer",
      "email": "string",
      "name": "string"
    }
  }
  ```

### Signup

- **Endpoint**: `/signup`
- **Method**: POST
- **Description**: Register a new user
- **Request Body**:
  ```json
  {
    "name": "string",
    "email": "string",
    "password": "string"
  }
  ```
- **Response**:
  ```json
  {
    "id": "integer",
    "name": "string",
    "email": "string"
  }
  ```

### Check Authentication

- **Endpoint**: `/auth/check`
- **Method**: GET
- **Description**: Verify if user is authenticated
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
  ```json
  {
    "authenticated": "boolean",
    "user": {
      "id": "integer",
      "email": "string",
      "name": "string"
    }
  }
  ```

## Sources

### Create Source

- **Endpoint**: `/sources`
- **Method**: POST
- **Description**: Create a new source
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
  ```json
  {
    "name": "string",
    "project_id": "integer",
    "source_type": "string",
    "config": "object",
    "created_by": "string",
    "updated_by": "string"
  }
  ```
- **Response**:
  ```json
  {
    "source_id": "integer",
    "name": "string",
    "source_type": "string",
    "project_id": "integer",
    "config": "object",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "created_by": "string",
    "updated_by": "string",
    "deleted_at": "null"
  }
  ```

### Get All Sources

- **Endpoint**: `/sources`
- **Method**: GET
- **Description**: Retrieve all sources
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
  ```json
  [
    {
      "source_id": "integer",
      "name": "string",
      "project_id": "integer",
      "source_type": "string",
      "config": "object",
      "created_at": "timestamp",
      "updated_at": "timestamp",
      "created_by": "string",
      "updated_by": "string",
      "deleted_at": "null"
    }
  ]
  ```

### Update Source

- **Endpoint**: `/sources/:id`
- **Method**: PUT
- **Description**: Update an existing source
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
  ```json
  {
    "source_id": "integer",
    "name": "string",
    "source_type": "string",
    "config": "object",
    "project_id": "integer",
    "updated_by": "string"
  }
  ```
- **Response**:
  ```json
  {
    "source_id": "integer",
    "name": "string",
    "project_id": "integer",
    "source_type": "string",
    "config": "object",
    "updated_at": "timestamp",
    "created_at": "timestamp",
    "created_by": "string",
    "updated_by": "string",
    "deleted_at": "null"
  }
  ```

### Delete Source

- **Endpoint**: `/sources/:id`
- **Method**: DELETE
- **Description**: Delete a source
- **Headers**: `Authorization: Bearer <token>`
- **Response**: HTTP 204 No Content

## Destinations

### Create Destination

- **Endpoint**: `/destinations`
- **Method**: POST
- **Description**: Create a new destination
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
  ```json
  {
    "name": "string",
    "project_id": "integer",
    "dest_type": "string",
    "config": "object",
    "created_by": "string",
    "updated_by": "string"
  }
  ```
- **Response**:
  ```json
  {
    "dest_id": "integer",
    "name": "string",
    "project_id": "integer",
    "dest_type": "string",
    "config": "object",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "created_by": "string",
    "updated_by": "string",
    "deleted_at": "null"
  }
  ```

### Get All Destinations

- **Endpoint**: `/destinations`
- **Method**: GET
- **Description**: Retrieve all destinations
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
  ```json
  [
    {
      "dest_id": "integer",
      "name": "string",
      "project_id": "integer",
      "dest_type": "string",
      "config": "object",
      "created_at": "timestamp",
      "updated_at": "timestamp",
      "created_by": "string",
      "updated_by": "string",
      "deleted_at": "null"
    }
  ]
  ```

### Update Destination

- **Endpoint**: `/destinations/:id`
- **Method**: PUT
- **Description**: Update an existing destination
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
  ```json
  {
    "dest_id": "integer",
    "name": "string",
    "project_id": "integer",
    "created_by": "string",
    "updated_by": "string",
    "dest_type": "string",
    "config": "object"
  }
  ```
- **Response**:
  ```json
  {
    "dest_id": "integer",
    "name": "string",
    "project_id": "integer",
    "dest_type": "string",
    "config": "object",
    "updated_at": "timestamp",
    "created_at": "timestamp",
    "created_by": "string",
    "updated_by": "string",
    "deleted_at": "null"
  }
  ```

### Delete Destination

- **Endpoint**: `/destinations/:id`
- **Method**: DELETE
- **Description**: Delete a destination
- **Headers**: `Authorization: Bearer <token>`
- **Response**: HTTP 204 No Content

## Jobs

### Create Job

- **Endpoint**: `/jobs`
- **Method**: POST
- **Description**: Create a new job
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:

  ```json
  {
    "name": "string",
    "source_id": "integer",
    "dest_id": "integer",
    "frequency": "string",
    "config": "object",
    "project_id": "integer",
    "connected": "boolean",
    "created_by": "string",
    "updated_by": "string",
    "last_sync_state": {
      "last_run_time": "timestamp",
      "last_run_state": "string"
    }
  }
  ```

- **Response**:
  ```json
  {
    "job_id": "integer",
    "name": "string",
    "source_id": "integer",
    "dest_id": "integer",
    "schedule": "string",
    "config": "object",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "created_by": "string",
    "updated_by": "string",
    "connected": "boolean",
    "deleted_at": "null",
    "frequency": "string",
    "last_sync_state": {
      "last_run_time": "timestamp",
      "last_run_state": "string"
    }
  }
  ```

### Get All Jobs

- **Endpoint**: `/jobs`
- **Method**: GET
- **Description**: Retrieve all jobs
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
  ```json
  [
    {
      "job_id": "integer",
      "name": "string",
      "source_id": "integer",
      "dest_id": "integer",
      "schedule": "string",
      "config": "object",
      "connected": "boolean",
      "frequency": "string",
      "updated_at": "timestamp",
      "created_at": "timestamp",
      "created_by": "string",
      "updated_by": "string",
      "last_sync_state": {
        "last_run_time": "timestamp",
        "last_run_state": "string"
      },
      "deleted_at": "null"
    }
  ]
  ```

### Update Job

- **Endpoint**: `/jobs/:id`
- **Method**: PUT
- **Description**: Update an existing job
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
  ```json
  {
    "name": "string",
    "connected": "boolean",
    "config": "object",
    "dest_id": "integer",
    "source_id": "integer",
    "frequency": "string",
    "updated_by": "string",
    "last_sync_state": {
      "last_run_time": "timestamp",
      "last_run_state": "string"
    }
  }
  ```
- **Response**:
  ```json
  {
    "job_id": "integer",
    "name": "string",
    "source_id": "integer",
    "dest_id": "integer",
    "frequency": "string",
    "config": "object",
    "status": "string",
    "updated_at": "timestamp",
    "created_at": "timestamp",
    "connected": "boolean",
    "last_sync_state": {
      "last_run_time": "timestamp",
      "last_run_state": "string"
    },
    "updated_by": "string",
    "created_by": "string",
    "deleted_at": "null"
  }
  ```

### Delete Job

- **Endpoint**: `/jobs/:id`
- **Method**: DELETE
- **Description**: Delete a job
- **Headers**: `Authorization: Bearer <token>`
- **Response**: HTTP 204 No Content

### Test Connection

- **Endpoint**: `/sources/test`
- **Method**: POST
- **Description**: Test configured source configuration
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:

  ```json
  {
    "source_id": "integer",
    "config": "object"
  }
  ```

- **Response**:

  ```json
  {
    "status": "string"
  }
  ```

- **Endpoint**: `/destinations/test`
- **Method**: POST
- **Description**: Test configured destination configuration
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:

  ```json
  {
    "dest_id": "integer",
    "config": "object"
  }
  ```

- **Response**:

  ```json
  {
    "status": "string"
  }
  ```

- **Endpoint**: `/sources/source_type/spec`
- **Method**: GET
- **Description**: Give spec based on source type
- **Headers**: `Authorization: Bearer <token>`

- **Response**:

  ```json
  {
    "spec": "object"
  }
  ```

- **Endpoint**: `/sources/dest_type/spec`
- **Method**: GET
- **Description**: Give spec based on destination type
- **Headers**: `Authorization: Bearer <token>`

- **Response**:

  ```json
  {
    "spec": "object"
  }
  ```

- **Endpoint**: `/sources/:id/catalog`
- **Method**: GET
- **Description**: Give the streams details
- **Headers**: `Authorization: Bearer <token>`

- **Response**:

  ```json
  {
    "catalog": {
      "streams": "object"
    }
  }
  ```

- **Endpoint**: `/jobs/:id/history`
- **Method**: GET
- **Description**: Give the History of jobs
- **Headers**: `Authorization: Bearer <token>`

- **Response**:

  ```json
  {
    "data": [
      {
        "start_time": "timestamp",
        "runtime": "integer",
        "status": "string"
      }
    ]
  }
  ```

- **Endpoint**: `/jobs/:id/tasks`
- **Method**: GET
- **Description**: Give the History of jobs
- **Headers**: `Authorization: Bearer <token>`

- **Response**:

  ```json
  {
    "data": [
      {
        "start_time": "timestamp",
        "runtime": "integer",
        "status": "string"
      }
    ]
  }
  ```

- **Endpoint**: `/jobs/:id/task/:taskid`
- **Method**: GET
- **Description**: Give the Logs of that particular Job
- **Headers**: `Authorization: Bearer <token>`

- **Response**:

  ```json
  {
    "data": [
      {
        "created_at": "timestamp",
        "message": "integer",
        "state": "string"
      }
    ]
  }
  ```

- **Endpoint**: `/sources/:id/jobs`
- **Method**: GET
- **Description**: Give the associated jobs of source
- **Headers**: `Authorization: Bearer <token>`

- **Response**:

  ```json
  {
    "jobs": [
      {
        "job_id": "integer",
        "name": "string",
        "source_id": "integer",
        "dest_id": "integer",
        "schedule": "string",
        "config": "object",
        "connected": "boolean",
        "frequency": "string",
        "updated_at": "timestamp",
        "created_at": "timestamp",
        "created_by": "string",
        "updated_by": "string",
        "last_sync_state": {
          "last_run_time": "timestamp",
          "last_run_state": "string"
        },
        "deleted_at": "null"
      }
    ]
  }
  ```

- **Endpoint**: `/destinations/:id/jobs`
- **Method**: GET
- **Description**: Give the associated jobs of a destination
- **Headers**: `Authorization: Bearer <token>`

- **Response**:
  ```json
  {
    "jobs": [
      {
        "job_id": "integer",
        "name": "string",
        "source_id": "integer",
        "dest_id": "integer",
        "schedule": "string",
        "config": "object",
        "connected": "boolean",
        "frequency": "string",
        "updated_at": "timestamp",
        "created_at": "timestamp",
        "created_by": "string",
        "updated_by": "string",
        "last_sync_state": {
          "last_run_time": "timestamp",
          "last_run_state": "string"
        },
        "deleted_at": "null"
      }
    ]
  }
  ```

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request

```json
{
  "error": "string",
  "message": "string"
}
```

### 401 Unauthorized

```json
{
  "error": "string",
  "message": "Authentication required"
}
```

### 403 Forbidden

```json
{
  "error": "string",
  "message": "Insufficient permissions"
}
```

### 404 Not Found

```json
{
  "error": "string",
  "message": "Resource not found"
}
```

### 500 Internal Server Error

```json
{
  "error": "string",
  "message": "Internal server error"
}
```

## CORS Configuration

The API allows requests from:

- Origin: `http://localhost:5173`
- Methods: GET, POST, PUT, DELETE, OPTIONS
- Headers: Origin, Content-Type, Accept, Authorization
- Credentials: true
