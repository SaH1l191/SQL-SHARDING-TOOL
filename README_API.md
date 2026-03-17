# SQL Sharding Tool API Documentation

## Overview

This API follows clean architecture principles with the following layers:
- **Handler Layer**: HTTP request/response handling
- **Service Layer**: Business logic implementation
- **Repository Layer**: Database operations
- **Router Layer**: Route configuration and middleware

## Base URL
```
http://localhost:8080/api/v1
```

## Project Endpoints

### Create Project
**POST** `/projects`

Request Body:
```json
{
  "name": "My E-commerce Project",
  "description": "A sample e-commerce application"
}
```

Response:
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "My E-commerce Project",
  "status": "inactive",
  "shard_count": 0,
  "created_at": "2026-03-17 23:00:00",
  "updated_at": ""
}
```

### Get All Projects
**GET** `/projects`

Response:
```json
[
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "My E-commerce Project",
    "status": "active",
    "shard_count": 4,
    "created_at": "2026-03-17 23:00:00",
    "updated_at": "2026-03-17 23:30:00"
  }
]
```

### Delete Project
**DELETE** `/projects/{id}`

Response: `204 No Content`

### Activate Project
**PUT** `/projects/{id}/activate`

Response:
```json
{
  "message": "Project activated successfully"
}
```

### Deactivate Project
**PUT** `/projects/{id}/deactivate`

Response:
```json
{
  "message": "Project deactivated successfully"
}
```

### Get Project Status
**GET** `/projects/{id}/status`

Response:
```json
{
  "project_id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "active"
}
```

## Health Check
**GET** `/health`

Response:
```json
{
  "status": "ok",
  "message": "SQL Sharding Tool API is running"
}
```

## Error Responses

All endpoints return appropriate HTTP status codes and error messages:

```json
{
  "error": "Error description"
}
```

Common status codes:
- `200`: Success
- `201`: Created
- `204`: No Content
- `400`: Bad Request
- `404`: Not Found
- `500`: Internal Server Error

## Architecture Layers

### 1. Repository Layer (`internal/repository`)
- Handles database operations
- Contains data models and SQL queries
- Interacts directly with the database

### 2. Service Layer (`internal/service`)
- Contains business logic
- Validates inputs and enforces business rules
- Coordinates between repository and handler layers

### 3. Handler Layer (`internal/handler`)
- Handles HTTP requests and responses
- Validates request data
- Calls appropriate service methods
- Formats responses

### 4. Router Layer (`internal/router`)
- Configures routes and middleware
- Sets up CORS, logging, and recovery middleware
- Maps HTTP methods to handler functions

## Usage Examples

### Create a new project
```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Project",
    "description": "A test project for demonstration"
  }'
```

### Get all projects
```bash
curl http://localhost:8080/api/v1/projects
```

### Activate a project
```bash
curl -X PUT http://localhost:8080/api/v1/projects/{project-id}/activate
```

## Dependencies

To run this API, you need to add Gin to your project:

```bash
go get github.com/gin-gonic/gin
```

## Environment Variables

Make sure your `.env` file contains the required database configuration:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_username
DB_PASSWORD=your_password
DB_NAME=sql_sharding_tool
```
