# Grippy

Grippy is a Go backend service for user authentication, to-do management, and user profile updates with avatar storage support.

## Features

- User registration and login with JWT access tokens and refresh token cookies
- Personal to-do CRUD operations behind authentication
- User profile update and avatar upload
- PostgreSQL database for persistent storage
- Redis for session and refresh token storage
- AWS S3-compatible avatar storage support

## Requirements

- Go 1.26.1
- PostgreSQL
- Redis
- AWS S3 or compatible storage for avatar uploads (optional)
- Docker and Docker Compose (optional, recommended for local development)

## Environment Variables

Create a `.env` file in the project root with the following values:

```env
ADDR=localhost
PORT=8080
SECRET_KEY=your_jwt_secret_key

POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=Grippy

REDIS_HOST=localhost
REDIS_PORT=6379

AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_REGION=your_region
AWS_BUCKET_NAME=your_bucket_name
AWS_ENDPOINT_URL=https://s3.your-provider.com
```

> Note: `AWS_ENDPOINT_URL` can be left empty when using AWS S3. For S3-compatible services, set it accordingly.

## Running Locally

### Using Docker Compose

```bash
docker-compose up --build
```

This starts PostgreSQL, Redis, and the Grippy backend.

### Running with Go

```bash
go run ./cmd/server
```

Or build and run the binary:

```bash
go build -o grippy ./cmd/server
./grippy
```

## Database Migrations

Use the provided `Taskfile.yml` tasks:

```bash
task migrate_up
task migrate_down
```

To create a new migration:

```bash
task create_migration
```

## API Overview

### Authentication

- `POST /auth/sign_up`
  - Body: `{ "email": "...", "username": "...", "password": "..." }`
  - Registers a new user.

- `POST /auth/sign_in`
  - Body: `{ "email": "...", "password": "..." }`
  - Returns `access_token` and sets `refresh_token` cookie.

- `POST /auth/refresh_token`
  - Uses `refresh_token` cookie.
  - Returns a new `access_token` and refreshes the cookie.

- `POST /auth/logout`
  - Clears the refresh token cookie.

### To-Do Management

Authentication is required for all `/todos` routes.
Include the access token in the `Authorization` header:

```http
Authorization: Bearer <access_token>
```

- `POST /todos/create`
  - Body: `{ "title": "...", "description": "..." }`

- `PUT /todos/update`
  - Body: `{ "id": "uuid", "user_id": "uuid", "title": "...", "description": "...", "is_completed": true }`

- `DELETE /todos/delete`
  - Query: `?id=<uuid>` or JSON body: `{ "id": "uuid" }`

- `GET /todos/get`
  - Query: `?id=<uuid>`

- `GET /todos/get_all`
  - Returns all todos for the authenticated user.

### User Profile

- `PUT /user/update_info`
  - Body: `{ "email": "...", "username": "..." }`

- `PUT /user/upload-avatar`
  - Multipart form upload with key `avatar`
  - Accepts JPEG or PNG files, up to 10 MB.

## Notes

- The app loads `.env` automatically when present.
- The server defaults to `localhost:8080` if `ADDR` or `PORT` are not set.
- `SECRET_KEY` is required for JWT generation.

## [License](LICENSE)
