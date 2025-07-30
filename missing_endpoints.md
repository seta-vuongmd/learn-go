# Missing API Endpoints

## REST API - Asset Management

### Missing Folder Endpoints:
- PUT /folders/:folderId (Update folder)
- DELETE /folders/:folderId (Delete folder)

### Missing Note Endpoints:
- GET /notes/:noteId (View note)
- PUT /notes/:noteId (Update note)  
- DELETE /notes/:noteId (Delete note)

### Missing Sharing Endpoints:
- DELETE /folders/:folderId/share/:userId (Revoke folder sharing)
- POST /notes/:noteId/share (Share single note)
- DELETE /notes/:noteId/share/:userId (Revoke note sharing)

### Missing Manager-only APIs:
- GET /users/:userId/assets (View user assets)

### Missing Team Management:
- POST /teams/:teamId/managers (Add manager)
- DELETE /teams/:teamId/managers/:managerId (Remove manager)

## Advanced Features Missing:
- POST /import-users (CSV import with goroutines)
- Logging và Monitoring setup
- Error handling improvements
