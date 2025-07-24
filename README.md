# User, Team & Asset Management System

## Prerequisites

- Go 1.21 or higher
- PostgreSQL database
- Git

## Installation & Setup

### 1. Clone and Install Dependencies

```bash
git clone <your-repo>
cd user-team-asset-management
go mod tidy
```

### 2. Database Setup

Create a PostgreSQL database:

```sql
CREATE DATABASE user_team_asset_db;
CREATE USER app_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE user_team_asset_db TO app_user;
```

### 3. Environment Configuration

Create a `.env` file in the root directory:

```env
DATABASE_URL=postgres://app_user:your_password@localhost:5432/user_team_asset_db?sslmode=disable
JWT_SECRET=your-super-secret-jwt-key-here
PORT=8080
```

Or set environment variables directly:

```bash
export DATABASE_URL="postgres://app_user:your_password@localhost:5432/user_team_asset_db?sslmode=disable"
export JWT_SECRET="your-super-secret-jwt-key-here"
export PORT="8080"
```

### 4. Run the Application

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### GraphQL Playground
- **URL**: `http://localhost:8080/graphql`
- **Purpose**: User management (create, login, fetch users)

### REST API Base
- **URL**: `http://localhost:8080/api`
- **Purpose**: Team and asset management

## Quick Start Guide

### 1. Create a User (GraphQL)

Open `http://localhost:8080/graphql` and run:

```graphql
mutation {
  createUser(
    username: "manager1"
    email: "manager@example.com"
    password: "password123"
    role: "manager"
  ) {
    userId
    username
    email
    role
  }
}
```

### 2. Login to Get Token

```graphql
mutation {
  login(email: "manager@example.com", password: "password123") {
    token
    user {
      userId
      username
      role
    }
  }
}
```

### 3. Use Token for REST API

Copy the token from login response and use it in REST API calls:

```bash
# Create a team
curl -X POST http://localhost:8080/api/teams \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "teamName": "Development Team",
    "managers": [],
    "members": []
  }'
```

## Development

### Project Structure
```
user-team-asset-management/
├── cmd/server/main.go          # Application entry point
├── internal/
│   ├── config/                 # Configuration management
│   ├── database/               # Database connection
│   ├── models/                 # Data models
│   ├── auth/                   # JWT authentication
│   ├── middleware/             # HTTP middleware
│   ├── graphql/                # GraphQL schema & resolvers
│   └── handlers/               # REST API handlers
├── examples/                   # API usage examples
├── go.mod                      # Go modules
└── README.md
```

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -o bin/server cmd/server/main.go
./bin/server
```