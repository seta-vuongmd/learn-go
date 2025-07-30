package handlers

import (
	"net/http"
	"user-team-asset-management/internal/models"
	"user-team-asset-management/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AssetHandler struct {
	DB *gorm.DB
}

func (h *AssetHandler) CreateFolder(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	folder := models.Folder{
		ID:      utils.GenerateID(),
		Name:    req.Name,
		OwnerID: userID,
	}

	if err := h.DB.Create(&folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		return
	}

	c.JSON(http.StatusCreated, folder)
}

func (h *AssetHandler) CreateNote(c *gin.Context) {
	folderID := c.Param("folderId")
	var req struct {
		Title string `json:"title" binding:"required"`
		Body  string `json:"body"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")

	if !h.canWriteToFolder(userID, folderID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No write access to this folder"})
		return
	}

	note := models.Note{
		ID:       utils.GenerateID(),
		Title:    req.Title,
		Body:     req.Body,
		FolderID: folderID,
		OwnerID:  userID,
	}

	if err := h.DB.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note"})
		return
	}

	c.JSON(http.StatusCreated, note)
}

func (h *AssetHandler) ShareFolder(c *gin.Context) {
	folderID := c.Param("folderId")
	var req struct {
		UserID string `json:"userId" binding:"required"`
		Access string `json:"access" binding:"required,oneof=read write"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	if !h.ownsFolder(userID, folderID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only folder owner can share"})
		return
	}

	share := models.FolderShare{
		FolderID: folderID,
		UserID:   req.UserID,
		Access:   req.Access,
	}

	if err := h.DB.Create(&share).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share folder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder shared successfully"})
}

func (h *AssetHandler) GetTeamAssets(c *gin.Context) {
	teamID := c.Param("teamId")
	userID := c.GetString("userID")

	if !h.isTeamManager(userID, teamID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Manager access required"})
		return
	}

	// Get all team members
	var memberIDs []string
	h.DB.Model(&models.TeamMember{}).Where("team_id = ?", teamID).Pluck("user_id", &memberIDs)

	// Get folders owned by team members
	var folders []models.Folder
	h.DB.Preload("Notes").Where("owner_id IN ?", memberIDs).Find(&folders)

	// Get shared folders accessible by team members
	var sharedFolders []models.Folder
	h.DB.Preload("Notes").
		Joins("JOIN folder_shares ON folders.id = folder_shares.folder_id").
		Where("folder_shares.user_id IN ?", memberIDs).
		Find(&sharedFolders)

	c.JSON(http.StatusOK, gin.H{
		"ownedFolders":  folders,
		"sharedFolders": sharedFolders,
	})
}

func (h *AssetHandler) ownsFolder(userID, folderID string) bool {
	var count int64
	h.DB.Model(&models.Folder{}).Where("id = ? AND owner_id = ?", folderID, userID).Count(&count)
	return count > 0
}

func (h *AssetHandler) canWriteToFolder(userID, folderID string) bool {
	if h.ownsFolder(userID, folderID) {
		return true
	}

	var count int64
	h.DB.Model(&models.FolderShare{}).
		Where("folder_id = ? AND user_id = ? AND access = 'write'", folderID, userID).
		Count(&count)
	return count > 0
}

func (h *AssetHandler) isTeamManager(userID, teamID string) bool {
	var count int64
	h.DB.Model(&models.TeamManager{}).Where("team_id = ? AND user_id = ?", teamID, userID).Count(&count)
	return count > 0
}

func (h *AssetHandler) GetUserFolders(c *gin.Context) {
	userID := c.GetString("userID")

	// Get owned folders
	var ownedFolders []models.Folder
	h.DB.Preload("Notes").Where("owner_id = ?", userID).Find(&ownedFolders)

	// Get shared folders
	var sharedFolders []models.Folder
	h.DB.Preload("Notes").
		Joins("JOIN folder_shares ON folders.id = folder_shares.folder_id").
		Where("folder_shares.user_id = ?", userID).
		Find(&sharedFolders)

	c.JSON(http.StatusOK, gin.H{
		"ownedFolders":  ownedFolders,
		"sharedFolders": sharedFolders,
	})
}

func (h *AssetHandler) GetFolder(c *gin.Context) {
	folderID := c.Param("folderId")
	userID := c.GetString("userID")

	if !h.canReadFolder(userID, folderID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No access to this folder"})
		return
	}

	var folder models.Folder
	if err := h.DB.Preload("Notes").Where("id = ?", folderID).First(&folder).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	c.JSON(http.StatusOK, folder)
}

func (h *AssetHandler) UpdateFolder(c *gin.Context) {
	folderID := c.Param("folderId")
	userID := c.GetString("userID")

	if !h.ownsFolder(userID, folderID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only folder owner can update"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Model(&models.Folder{}).Where("id = ?", folderID).Update("name", req.Name).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update folder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder updated successfully"})
}

func (h *AssetHandler) DeleteFolder(c *gin.Context) {
	folderID := c.Param("folderId")
	userID := c.GetString("userID")

	if !h.ownsFolder(userID, folderID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only folder owner can delete"})
		return
	}

	// Delete all notes in folder first
	h.DB.Where("folder_id = ?", folderID).Delete(&models.Note{})
	// Delete folder shares
	h.DB.Where("folder_id = ?", folderID).Delete(&models.FolderShare{})
	// Delete folder
	h.DB.Where("id = ?", folderID).Delete(&models.Folder{})

	c.JSON(http.StatusOK, gin.H{"message": "Folder deleted successfully"})
}

func (h *AssetHandler) GetNote(c *gin.Context) {
	noteID := c.Param("noteId")
	userID := c.GetString("userID")

	if !h.canReadNote(userID, noteID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No access to this note"})
		return
	}

	var note models.Note
	if err := h.DB.Where("id = ?", noteID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	c.JSON(http.StatusOK, note)
}

func (h *AssetHandler) UpdateNote(c *gin.Context) {
	noteID := c.Param("noteId")
	userID := c.GetString("userID")

	if !h.canWriteToNote(userID, noteID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No write access to this note"})
		return
	}

	var req struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Body != "" {
		updates["body"] = req.Body
	}

	if err := h.DB.Model(&models.Note{}).Where("id = ?", noteID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note updated successfully"})
}

func (h *AssetHandler) DeleteNote(c *gin.Context) {
	noteID := c.Param("noteId")
	userID := c.GetString("userID")

	if !h.ownsNote(userID, noteID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only note owner can delete"})
		return
	}

	// Delete note shares first
	h.DB.Where("note_id = ?", noteID).Delete(&models.NoteShare{})
	// Delete note
	h.DB.Where("id = ?", noteID).Delete(&models.Note{})

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}

func (h *AssetHandler) ShareNote(c *gin.Context) {
	noteID := c.Param("noteId")
	var req struct {
		UserID string `json:"userId" binding:"required"`
		Access string `json:"access" binding:"required,oneof=read write"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	if !h.ownsNote(userID, noteID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only note owner can share"})
		return
	}

	share := models.NoteShare{
		NoteID: noteID,
		UserID: req.UserID,
		Access: req.Access,
	}

	if err := h.DB.Create(&share).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note shared successfully"})
}

func (h *AssetHandler) RevokeFolderShare(c *gin.Context) {
	folderID := c.Param("folderId")
	shareUserID := c.Param("userId")
	userID := c.GetString("userID")

	if !h.ownsFolder(userID, folderID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only folder owner can revoke sharing"})
		return
	}

	if err := h.DB.Where("folder_id = ? AND user_id = ?", folderID, shareUserID).Delete(&models.FolderShare{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke folder sharing"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder sharing revoked successfully"})
}

func (h *AssetHandler) RevokeNoteShare(c *gin.Context) {
	noteID := c.Param("noteId")
	shareUserID := c.Param("userId")
	userID := c.GetString("userID")

	if !h.ownsNote(userID, noteID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only note owner can revoke sharing"})
		return
	}

	if err := h.DB.Where("note_id = ? AND user_id = ?", noteID, shareUserID).Delete(&models.NoteShare{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke note sharing"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note sharing revoked successfully"})
}

func (h *AssetHandler) GetUserAssets(c *gin.Context) {
	targetUserID := c.Param("userId")
	currentUserID := c.GetString("userID")
	currentRole := c.GetString("role")

	// Only managers can view other users' assets
	if currentRole != "manager" && currentUserID != targetUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Manager role required"})
		return
	}

	// Get owned folders
	var ownedFolders []models.Folder
	h.DB.Preload("Notes").Where("owner_id = ?", targetUserID).Find(&ownedFolders)

	// Get shared folders
	var sharedFolders []models.Folder
	h.DB.Preload("Notes").
		Joins("JOIN folder_shares ON folders.id = folder_shares.folder_id").
		Where("folder_shares.user_id = ?", targetUserID).
		Find(&sharedFolders)

	c.JSON(http.StatusOK, gin.H{
		"ownedFolders":  ownedFolders,
		"sharedFolders": sharedFolders,
	})
}

func (h *AssetHandler) canReadFolder(userID, folderID string) bool {
	// Check if user owns the folder
	var count int64
	h.DB.Model(&models.Folder{}).Where("id = ? AND owner_id = ?", folderID, userID).Count(&count)
	if count > 0 {
		return true
	}

	// Check if folder is shared with user
	h.DB.Model(&models.FolderShare{}).Where("folder_id = ? AND user_id = ?", folderID, userID).Count(&count)
	return count > 0
}

func (h *AssetHandler) canReadNote(userID, noteID string) bool {
	// Check if user owns the note
	var count int64
	h.DB.Model(&models.Note{}).Where("id = ? AND owner_id = ?", noteID, userID).Count(&count)
	if count > 0 {
		return true
	}

	// Check if note is shared with user
	h.DB.Model(&models.NoteShare{}).Where("note_id = ? AND user_id = ?", noteID, userID).Count(&count)
	return count > 0
}

func (h *AssetHandler) canWriteToNote(userID, noteID string) bool {
	// Check if user owns the note
	var count int64
	h.DB.Model(&models.Note{}).Where("id = ? AND owner_id = ?", noteID, userID).Count(&count)
	if count > 0 {
		return true
	}

	// Check if note is shared with write access
	h.DB.Model(&models.NoteShare{}).
		Where("note_id = ? AND user_id = ? AND access = 'write'", noteID, userID).
		Count(&count)
	return count > 0
}

func (h *AssetHandler) ownsNote(userID, noteID string) bool {
	var count int64
	h.DB.Model(&models.Note{}).Where("id = ? AND owner_id = ?", noteID, userID).Count(&count)
	return count > 0
}
