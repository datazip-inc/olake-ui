# Olake Server API Contract

### For now use olake as project id, later on it can be used to make multitenant system

## Base URL

```
http://localhost:8000
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
    "success": "boolean",
    "message": "string",
    "data": {
      "username": "string"
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
    "email": "string",
    "username": "string",
    "password": "string"
  }
  ```
- **Response**:

  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "email": "string",
      "username": "string"
    }
  }
  ```

### Check Authentication

- **Endpoint**: `/auth`
- **Method**: GET
- **Description**: Verify if user is authenticated
- **Headers**: `Authorization: Bearer <token>` // we are using cookie currently so frontend take care accordingly
- **Response**:
  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "username": "string" // can be user id let us discuss, else let us store username in session
    }
  }
  ```

## Sources

### Get Spec Of Source

- **Endpoint**: `/api/v1/project/:projectid/sources/spec`
- **Method**: GET
- **Description**: Give spec based on source type
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
  ```json
  {
    "type": "string",
    "version": "string"
  }
  ```
- **Response**:

  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "type": "string",
      "version": "string",
      "spec": {
        "jsonschema": "object",
        "uischema": "string"
      }
    }
  }
  ```

### Source Test Connection

- **Endpoint**: `/api/v1/project/:projectid/sources/test`
- **Method**: POST
- **Description**: Test configured source configuration
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:

  ```json
  {
    "type": "string",
    "version": "string",
    "config": "json"
  }
  ```

- **Response**:

  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "type": "string",
      "version": "string",
      "config": "json"
    }
  }
  ```

### Create Source

- **Endpoint**: `/api/v1/project/:projectid/sources`
- **Method**: POST
- **Description**: Create a new source
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
  ```json
  {
    "name": "string", // we have to make sure in database that it must also unique according to project id (for doubt let us discuss)
    "type": "string",
    "version": "string", // this field need to be shown on frontend as well, we discussed at time of design as well
    "config": "json"
  }
  ```
- **Response**:
  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      // whatever received send back
      "name": "string",
      "type": "string",
      "version": "string",
      "config": "json"
    }
  }
  ```

### Get All Sources

- **Endpoint**: `/api/v1/project/:projectid/sources`
- **Method**: GET
- **Description**: Retrieve all sources
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": [
      {
        "id": "integer",
        "name": "string",
        "type": "string",
        "version": "string",
        "config": "json",
        "created_at": "timestamp",
        "updated_at": "timestamp",
        "created_by": "string", // only username of user
        "updated_by": "string", // only username of user,
        "jobs": [
          {
            "name": "string",
            "id": "integer",
            "activate": "boolean",
            "last_run_time": "timestamp",
            "last_run_state": "string",
            "dest_name": "string"
          }
        ]
      }
    ]
  }
  ```

### Update Source

- **Endpoint**: `/api/v1/project/:projectid/sources/:id`
- **Method**: PUT
- **Description**: Update an existing source
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
  ```json
  {
    "name": "string",
    "type": "string",
    "version": "string",
    "config": "json"
  }
  ```
- **Response**:
  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      // send same back
      "name": "string",
      "type": "string",
      "version": "string",
      "config": "json"
    }
  }
  ```

### Delete Source

- **Endpoint**: `/api/v1/project/:projectid/sources/:id`
- **Method**: DELETE
- **Description**: Delete a source
- **Headers**: `Authorization: Bearer <token>`
- **Response**:

```json
{
  // Note: it is soft delete not hard delete
  "success": "boolean",
  "message": "string",
  "data": {
    "name": "string" // name of source deleted
  }
}
```

## Destinations

### Destination Spec

- **Endpoint**: `/api/v1/project/:projectid/destinations/spec`
- **Method**: GET
- **Description**: Give spec based on destination type
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
  ```json
  {
    "type": "string",
    "version": "string"
  }
  ```
- **Response**:
  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "type": "string",
      "version": "string",
      "spec": {
        "jsonschema": "object",
        "uischema": "string"
      }
    }
  }
  ```

// currently this is not avaialable in olake will build this

### Test Destination

- **Endpoint**: `/api/v1/project/:projectid/destinations/test`
- **Method**: POST
- **Description**: Test configured destination configuration
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:

  ```json
  {
    "type": "string",
    "version": "string",
    "config": "json"
  }
  ```

- **Response**:
  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "type": "string",
      "version": "string",
      "config": "json"
    }
  }
  ```

### Create Destination

- **Endpoint**: `/api/v1/project/:projectid/destinations`
- **Method**: POST
- **Description**: Create a new destination
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
  ```json
  {
    "name": "string",
    "type": "string",
    "config": "json",
    "version": "string"
  }
  ```
- **Response**:

  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "name": "string",
      "type": "string",
      "config": "json",
      "version": "string" // to create a job same version of destination and same version of source required
    }
  }
  ```

### Get All Destinations

- **Endpoint**: `/api/v1/project/:projectid/destinations`
- **Method**: GET
- **Description**: Retrieve all destinations
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": [
      {
        "id": "integer",
        "name": "string",
        "type": "string",
        "config": "json",
        "version": "string",
        "created_at": "timestamp",
        "updated_at": "timestamp",
        "created_by": "string", // username only
        "updated_by": "string", // username only
        "jobs": [
          {
            "name": "string",
            "id": "integer",
            "activate": "boolean",
            "last_run_time": "timestamp",
            "last_run_state": "string",
            "source_name": "string"
          }
        ]
      }
    ]
  }
  ```

### Update Destination

- **Endpoint**: `/api/v1/project/:projectid/destinations/:id`
- **Method**: PUT
- **Description**: Update an existing destination
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
  ```json
  {
    "name": "string",
    "type": "string",
    "config": "json",
    "version": "string"
  }
  ```
- **Response**:
  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "name": "string",
      "type": "string",
      "version": "string",
      "config": "json"
    }
  }
  ```

### Delete Destination

- **Endpoint**: `/api/v1/project/:projectid/destinations/:id`
- **Method**: DELETE
- **Description**: Delete a destination
- **Headers**: `Authorization: Bearer <token>`
- **Response**:

```json
{
  // NOTE: this is only soft delete not hard
  "success": "boolean",
  "message": "string",
  "data": {
    "name": "string"
  }
}
```

## Jobs

### Create Job

- **Endpoint**: `/api/v1/project/:projectid/jobs`
- **Method**: POST
- **Description**: Create a new job
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:

  ```json
  {
    "name": "string",
    "source": {
      "name": "string",
      "type": "string",
      "config": "json",
      "version": "string"
    },
    "destination": {
      "name": "string",
      "type": "string",
      "config": "string",
      "version": "string"
    },
    "frequency": "string",
    "streams_config": "json"
  }
  ```

- **Response**:
  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      // request body as it is
    }
  }
  ```

### Get All Jobs

- **Endpoint**: `/api/v1/project/:projectid/jobs` // also use endpoint for filter such as /jobs/dest_id="some_id" or /jobs/source_id="some_id"
- **Method**: GET
- **Description**: Retrieve all jobs
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": [
      {
        "id": "int",
        "name": "string",
        "source": {
          "name": "string",
          "type": "string",
          "config": "json",
          "version": "string"
        },
        "destination": {
          "name": "string",
          "type": "string",
          "config": "json",
          "version": "string"
        },
        "streams_config": "json",
        "frequency": "string",
        "last_run_time": "timestamp",
        "last_run_state": "string",
        "last_run_type": "string",
        "created_at": "timestamp",
        "updated_at": "timestamp",
        "created_by": "string", // username
        "updated_by": "string" // username
        // can also send state but if it is required
      }
    ]
  }
  ```

### Update Job

- **Endpoint**: `/api/v1/project/:projectid/jobs/:id`
- **Method**: PUT
- **Description**: Update an existing job
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:

  ```json
  {
    "name": "string",
    "source": {
      "name": "string",
      "type": "string",
      "config": "json",
      "version": "string"
    },
    "destination": {
      "name": "string",
      "type": "string",
      "config": "json",
      "version": "string"
    },
    "frequency": "string",
    "streams_config": "json",
    "difference_streams": "string",
    "activate": "boolean" // send this to activate or deactivate job
  }
  ```

- **Response**:
  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "name": "string",
      "source": {
        "name": "string",
        "type": "string",
        "config": "json",
        "version": "string"
      },
      "destination": {
        "name": "string",
        "type": "string",
        "config": "json",
        "version": "string"
      },
      "frequency": "string",
      "streams_config": "json",
      "activate": "boolean"
    }
  }
  ```

### Delete Job

- **Endpoint**: `/api/v1/project/:projectid/jobs/:id`
- **Method**: DELETE
- **Description**: Delete a job
- **Headers**: `Authorization: Bearer <token>`
- **Response**:

```json
{
  "success": "boolean",
  "message": "string",
  "data": {
    "name": "boolean"
  }
}
```

### Source Associated Streams (Discover Catalog)

- **Endpoint**: `/api/v1/project/:projectid/source/streams`
- **Method**: GET
- **Description**: Give the streams details
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:

  ```json
  {
    "name": "string",
    "type": "string",
    "version": "string",
    "config": "json",
    "job_id": "integer",
    "job_name": "string"
  }
  ```

- **Response**:

  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "streams_config": "json"
    }
  }
  ```

### Job Sync

- **Endpoint**: `/api/v1/project/:projectid/jobs/:id/sync`
- **Method**: POST
- **Description**: Sync the job
- **Headers**: `Authorization: Bearer <token>`
- **Response**:

  ```json
  {
    "success": "boolean",
    "message": "string"
  }
  ```

### Activate/Inactivate Job

- **Endpoint**: `/api/v1/project/:projectid/jobs/:id/activate`

- **Method**: POST
- **Description**: Update the activate status of job
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:

  ```json
  {
   "activate":boolean
  }
  ```

- **Response**:
  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      // request body as it is
    }
  }
  ```

### Job Tasks

- **Endpoint**: `/api/v1/project/:projectid/jobs/:jobid/tasks`
- **Method**: GET
- **Description**: Give the History of jobs
- **Headers**: `Authorization: Bearer <token>`

- **Response**:

  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": [
      {
        "file_path": "string",
        "start_time": "timestamp",
        "runtime": "integer",
        "status": "string",
        "job_type": "string"
      }
    ]
  }
  ```

### cancel Job workflow

- **Endpoint**: `/api/v1/project/:projectid/jobs/:jobid/cancel`
- **Method**: GET
- **Description**: Cancel the job workflow
- **Headers**: `Authorization: Bearer <token>`

- **Response**:

  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "message": "string"
    }
  }
  ```

  ### Job Sync

- **Endpoint**: `/api/v1/project/:projectid/jobs/:id/sync`
- **Method**: POST
- **Description**: Sync the job
- **Headers**: `Authorization: Bearer <token>`
- **Response**:

  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": null
  }
  ```

  ###Activate/Inactivate Job

- **Endpoint**: `/api/v1/project/:projectid/jobs/:id/activate`
- **Method**: POST
- **Description**: Update the activate status of job
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:

  ```json
  {
   "activate":boolean
  }
  ```

  - **Response**:

  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "activate": "boolean"
    }
  }
  ```

- **Endpoint**: `/api/v1/project/:projectid/jobs/:jobid/task/:id/logs`
- **Method**: GET
- **Description**: Give the Logs of that particular Job
- **Headers**: `Authorization: Bearer <token>`

- **Response**:

  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "task_logs": "json"
    }
  }
  ```

### job name unique check
---

- **Endpoint**: `api/v1/project/:projectid/jobs/check-unique`
- **Method**: POST
- **Description**: check if job name is unique
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:

  ```json
  {
    "job_name": "string"
  }
  ```

- **Response**:

  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
     "unique": "boolean"
    }
  }
  ```


### Clear destination for a Job
---

- **Endpoint**: `/api/v1/project/:projectid/jobs/:id/clear-destination`
- **Method**: POST
- **Description**: clears the destination data
- **Headers**: `Authorization: Bearer <token>`

- **Response**: 

  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "message": "string"
    }
  }
  ```

### Clear destination status
---

- **Endpoint**: `/api/v1/project/:projectid/jobs/:id/clear-destination`
- **Method**: GET
- **Description**: gets the clear-destination status, either true (running) or false (not running)
- **Headers**: `Authorization: Bearer <token>`

- **Response**: 

  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "running": "boolean"
    }
  }
  ```

### Difference Streams
- **Endpoint**: `/api/v1/project/:projectid/jobs/:id/difference-streams`
- **Method**: POST
- **Description**: returns the stream difference bewtween the saved and the updated streams
- **Headers**: `Authorization: Bearer <token>`
- - **Request Body**:

  ```json
  {
    "updated_streams_config": "json"
  }
  ```

- **Response**: 

  ```json
  {
    "success": "boolean",
    "message": "string",
    "data": {
      "difference_streams": "json"
    }
  }
  ```

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request

```json
{
  "success": false,
  "message": "Bad request"
}
```

### 401 Unauthorized

```json
{
  "success": false,
  "message": "Authentication required"
}
```

### 403 Forbidden

```json
{
  "success": false,
  "message": "Insufficient permissions"
}
```

### 404 Not Found

```json
{
  "success": false,
  "message": "Resource not found"
}
```

### 500 Internal Server Error

```json
{
  "success": false,
  "message": "Internal server error"
}
```

## CORS Configuration

The API allows requests from:

- Origin: `http://localhost:8000`
- Methods: GET, POST, PUT, DELETE, OPTIONS
- Headers: Origin, Content-Type, Accept, Authorization
- Credentials: true
