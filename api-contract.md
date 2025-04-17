# Olake Server API Contract

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
    "username": "string",
    "password": "string"
  }
  ```
- **Response**:
  ```json
  {
    "username": "string",
    "message": "string",
    "success": "boolean"
  }
  ```

### Signup

- **Endpoint**: `/signup`
- **Method**: POST
- **Description**: Register a new user
- **Request Body**:
  ```json
  {
    "email": "string",
    "username": "string",
    "password": "string"
  }
  ```
- **Response**:

  ```json
  {
    "email": "string",
    "message": "string",
    "success": "boolean"
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
    "email": "string"
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
    "type": "string",
    "config": "string"
  }
  ```
- **Response**:
  ```json
  {
    "success": "boolean",
    "data": {
      "name": "string",
      "type": "string",
      "project_id": "integer",
      "config": "string"
    }
  }
  ```

### Get All Sources

- **Endpoint**: `/sources`
- **Method**: GET
- **Description**: Retrieve all sources
- **Headers**: `Authorization: Bearer <token>`
- **Response**:

  ```json
  {
    "success": "boolean",
    "data": [
      {
        "id": "integer",
        "name": "string",
        "project_id": "integer",
        "type": "string",
        "config": "string",
        "created_at": "timestamp",
        "updated_at": "timestamp",
        "created_by": {
          "id": "integer",
          "username": "string"
        },
        "updated_by": {
          "id": "integer",
          "username": "string"
        }
      }
    ]
  }
  ```

### Update Source

- **Endpoint**: `/sources/:id`
- **Method**: PUT
- **Description**: Update an existing source
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
  ```json
  {
    "name": "string",
    "type": "string",
    "config": "string",
    "project_id": "integer"
  }
  ```
- **Response**:
  ```json
  {
    "success": "boolean",
    "data": {
      "name": "string",
      "type": "string",
      "project_id": "integer",
      "config": "string"
    }
  }
  ```

### Delete Source

- **Endpoint**: `/sources/:id`
- **Method**: DELETE
- **Description**: Delete a source
- **Headers**: `Authorization: Bearer <token>`
- **Response**:

```json
{
  "success": "boolean"
}
```

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
    "type": "string",
    "config": "string"
  }
  ```
- **Response**:

  ```json
  {
    "success": "boolean",
    "data": {
      "name": "string",
      "type": "string",
      "project_id": "integer",
      "config": "string"
    }
  }
  ```

### Get All Destinations

- **Endpoint**: `/destinations`
- **Method**: GET
- **Description**: Retrieve all destinations
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
  ```json
  {
    "success": "boolean",
    "data": [
      {
        "id": "integer",
        "name": "string",
        "project_id": "integer",
        "type": "string",
        "config": "string",
        "created_at": "timestamp",
        "updated_at": "timestamp",
        "created_by": {
          "id": "integer",
          "username": "string"
        },
        "updated_by": {
          "id": "integer",
          "username": "string"
        }
      }
    ]
  }
  ```

### Update Destination

- **Endpoint**: `/destinations/:id`
- **Method**: PUT
- **Description**: Update an existing destination
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
  ```json
  {
    "name": "string",
    "type": "string",
    "config": "string",
    "project_id": "integer"
  }
  ```
- **Response**:
  ```json
  {
    "success": "boolean",
    "data": {
      "name": "string",
      "type": "string",
      "project_id": "integer",
      "config": "string"
    }
  }
  ```

### Delete Destination

- **Endpoint**: `/destinations/:id`
- **Method**: DELETE
- **Description**: Delete a destination
- **Headers**: `Authorization: Bearer <token>`
- **Response**: HTTP 204 No Content

```json
{
  "success": "boolean"
}
```

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
    "source": {
      "name": "string",
      "project_id": "integer",
      "type": "string",
      "config": "string"
    },
    "destination": {
      "name": "string",
      "project_id": "integer",
      "type": "string",
      "config": "string"
    },
    "frequency": "string",
    "schema": "object",
    "project_id": "integer",
    "last_sync_state": {
      "last_run_time": "timestamp",
      "last_run_state": "string"
    }
  }
  ```

- **Response**:
  ```json
  {
    "success": "boolean",
    "data": {
      "name": "string",
      "source": {
        "name": "string",
        "project_id": "integer",
        "type": "string",
        "config": "string"
      },
      "destination": {
        "name": "string",
        "project_id": "integer",
        "type": "string",
        "config": "string"
      },
      "frequency": "string",
      "schema": "object",
      "project_id": "integer",
      "last_sync_state": {
        "last_run_time": "timestamp",
        "last_run_state": "string"
      }
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
  {
    "name": "string",
    "source": {
      "name": "string",
      "project_id": "integer",
      "type": "string",
      "config": "string"
    },
    "destination": {
      "name": "string",
      "project_id": "integer",
      "type": "string",
      "config": "string"
    },
    "frequency": "string",
    "project_id": "integer",
    "last_sync_state": {
      "last_run_time": "timestamp",
      "last_run_state": "string"
    },
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "created_by": {
      "id": "integer",
      "username": "string"
    },
    "updated_by": {
      "id": "integer",
      "username": "string"
    }
  }
  ```

### Update Job

- **Endpoint**: `/jobs/:id`
- **Method**: PUT
- **Description**: Update an existing job
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:

  ```json
  {
    "success": "boolean",
    "data": {
      "name": "string",
      "source": {
        "name": "string",
        "project_id": "integer",
        "type": "string",
        "config": "string"
      },
      "destination": {
        "name": "string",
        "project_id": "integer",
        "type": "string",
        "config": "string"
      },
      "frequency": "string",
      "schema": "object",
      "project_id": "integer",
      "last_sync_state": {
        "last_run_time": "timestamp",
        "last_run_state": "string"
      }
    }
  }
  ```

- **Response**:
  ```json
  {
    "success": "boolean",
    "data": {
      "name": "string",
      "source": {
        "name": "string",
        "project_id": "integer",
        "type": "string",
        "config": "string"
      },
      "destination": {
        "name": "string",
        "project_id": "integer",
        "type": "string",
        "config": "string"
      },
      "frequency": "string",
      "schema": "object",
      "project_id": "integer",
      "last_sync_state": {
        "last_run_time": "timestamp",
        "last_run_state": "string"
      }
    }
  }
  ```

### Delete Job

- **Endpoint**: `/jobs/:id`
- **Method**: DELETE
- **Description**: Delete a job
- **Headers**: `Authorization: Bearer <token>`
- **Response**: HTTP 204 No Content

```json
{
  "success": "boolean"
}
```

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
        "name": "string",
        "source": {
          "name": "string",
          "project_id": "integer",
          "type": "string",
          "config": "string"
        },
        "destination": {
          "name": "string",
          "project_id": "integer",
          "type": "string",
          "config": "string"
        },
        "frequency": "string",
        "project_id": "integer",
        "last_sync_state": {
          "last_run_time": "timestamp",
          "last_run_state": "string"
        },
        "created_at": "timestamp",
        "updated_at": "timestamp",
        "created_by": {
          "id": "integer",
          "username": "string"
        },
        "updated_by": {
          "id": "integer",
          "username": "string"
        }
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
        "name": "string",
        "source": {
          "name": "string",
          "project_id": "integer",
          "type": "string",
          "config": "string"
        },
        "destination": {
          "name": "string",
          "project_id": "integer",
          "type": "string",
          "config": "string"
        },
        "frequency": "string",
        "project_id": "integer",
        "last_sync_state": {
          "last_run_time": "timestamp",
          "last_run_state": "string"
        },
        "created_at": "timestamp",
        "updated_at": "timestamp",
        "created_by": {
          "id": "integer",
          "username": "string"
        },
        "updated_by": {
          "id": "integer",
          "username": "string"
        }
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

```

```
