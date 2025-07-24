# API Usage Examples

## Mở http://localhost:8080/graphql để chạy lệnh graphql
## GraphQL - User Management

### Create User
```graphql
mutation {
  createUser(username: "john_doe", email: "john@example.com", password: "password123", role: "manager") {
    userId
    username
    email
    role
  }
}
```

### Fetch Users
```graphql
query {
  fetchUsers {
    userId
    username
    email
    role
  }
}
```

### Login
```graphql
mutation {
  login(email: "john@example.com", password: "password123") {
    token
    user {
      userId
      username
      role
    }
  }
}
```
# eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI4YTJhMzkwY2NhNGY1ODNmZjllZThhYzA3YzQwNDczNiIsInJvbGUiOiJtYW5hZ2VyIiwiZXhwIjoxNzUzNDEzNjQwLCJpYXQiOjE3NTMzMjcyNDB9.Rlr_tRVdP5N5sXCdCTKNOqPQX1szuHGQK7w-YJVd7XA

## REST API - Team Management

### Create Team
```bash
curl -X POST http://localhost:8080/api/teams \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "teamName": "Development Team",
    "managers": [{"managerId": "user123", "managerName": "John"}],
    "members": [{"memberId": "user456", "memberName": "Jane"}]
  }'
```

### Create Folder
```bash
curl -X POST http://localhost:8080/api/folders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Project Documents"}'
```

### Share Folder
```bash
curl -X POST http://localhost:8080/api/folders/FOLDER_ID/share \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"userId": "USER_ID", "access": "write"}'
```