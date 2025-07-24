# API Usage Examples

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