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

### Logout
```graphql
mutation {
  logout
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

### Add Manager to Team
```bash
curl -X POST http://localhost:8080/api/teams/TEAM_ID/managers \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"managerId": "USER_ID"}'
```

### Remove Manager from Team
```bash
curl -X DELETE http://localhost:8080/api/teams/TEAM_ID/managers/MANAGER_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## REST API - Asset Management

### Create Folder
```bash
curl -X POST http://localhost:8080/api/folders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Project Documents"}'
```

### Update Folder
```bash
curl -X PUT http://localhost:8080/api/folders/FOLDER_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Folder Name"}'
```

### Delete Folder
```bash
curl -X DELETE http://localhost:8080/api/folders/FOLDER_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Create Note
```bash
curl -X POST http://localhost:8080/api/folders/FOLDER_ID/notes \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Meeting Notes", "body": "Discussion about project..."}'
```

### Update Note
```bash
curl -X PUT http://localhost:8080/api/notes/NOTE_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Updated Title", "body": "Updated content..."}'
```

### Delete Note
```bash
curl -X DELETE http://localhost:8080/api/notes/NOTE_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Share Folder
```bash
curl -X POST http://localhost:8080/api/folders/FOLDER_ID/share \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"userId": "USER_ID", "access": "write"}'
```

### Share Note
```bash
curl -X POST http://localhost:8080/api/notes/NOTE_ID/share \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"userId": "USER_ID", "access": "read"}'
```

### Revoke Folder Share
```bash
curl -X DELETE http://localhost:8080/api/folders/FOLDER_ID/share/USER_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Revoke Note Share
```bash
curl -X DELETE http://localhost:8080/api/notes/NOTE_ID/share/USER_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Manager-only APIs

### Get User Assets
```bash
curl -X GET http://localhost:8080/api/users/USER_ID/assets \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Get Team Assets
```bash
curl -X GET http://localhost:8080/api/teams/TEAM_ID/assets \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## CSV Import Feature

### Import Users from CSV
```bash
curl -X POST http://localhost:8080/api/import-users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@users.csv"
```

CSV format:
```csv
username,email,password,role
john_doe,john@example.com,password123,manager
jane_smith,jane@example.com,password123,member
```